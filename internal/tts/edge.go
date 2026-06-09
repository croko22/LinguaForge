package tts

import (
	"encoding/binary"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

const edgeWSS = "wss://speech.platform.bing.com/consumer/speech/synthesize/readaloud/edge/v1"

var defaultDialer = &websocket.Dialer{
	HandshakeTimeout: 10 * time.Second,
}

type edgeClient struct {
	conn *websocket.Conn
}

func newEdgeClient() (*edgeClient, error) {
	connID := uuid.New().String()
	u := fmt.Sprintf("%s?TrustedClientToken=6A5AA1D4EAFF4E9FB37E23D68491D6F4&ConnectionId=%s", edgeWSS, connID)

	header := http.Header{
		"Origin":     {"chrome-extension://jdiccldimpdaibmpdkjnbckngbcgghdo"},
		"User-Agent": {"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36"},
	}

	conn, _, err := defaultDialer.Dial(u, header)
	if err != nil {
		return nil, fmt.Errorf("dial: %w", err)
	}

	return &edgeClient{conn: conn}, nil
}

func (c *edgeClient) Close() error {
	return c.conn.Close()
}

func (c *edgeClient) synthesize(text, voice string) ([]byte, error) {
	if err := c.sendConfig(); err != nil {
		return nil, fmt.Errorf("send config: %w", err)
	}

	if err := c.sendSSML(text, voice); err != nil {
		return nil, fmt.Errorf("send ssml: %w", err)
	}

	audio, err := c.readAudio()
	if err != nil {
		return nil, fmt.Errorf("read audio: %w", err)
	}

	return audio, nil
}

func (c *edgeClient) sendConfig() error {
	msg := "Content-Type:application/json; charset=utf-8\r\nPath:speech.config\r\n\r\n{\"context\":{\"synthesis\":{\"audio\":{\"metadataoptions\":{\"sentenceBoundaryEnabled\":false,\"wordBoundaryEnabled\":false},\"outputformat\":\"audio-24khz-48kbitrate-mono-mp3\"}}}}}"
	return c.conn.WriteMessage(websocket.TextMessage, []byte(msg))
}

func (c *edgeClient) sendSSML(text, voice string) error {
	lang := "en-US"
	if parts := strings.SplitN(voice, "-", 2); len(parts) == 2 {
		lang = parts[0] + "-" + parts[1]
	}

	ssml := fmt.Sprintf(
		"Content-Type:application/ssml+xml\r\nPath:ssml\r\n\r\n<speak version='1.0' xmlns='http://www.w3.org/2001/10/synthesis' xml:lang='%s'><voice name='%s'><prosody pitch='+0Hz' rate='+0%%' volume='+0%%'>%s</prosody></voice></speak>",
		lang, voice, escapeXML(text),
	)
	return c.conn.WriteMessage(websocket.TextMessage, []byte(ssml))
}

func (c *edgeClient) readAudio() ([]byte, error) {
	defer c.conn.Close()

	var buf []byte
	for {
		if err := c.conn.SetReadDeadline(time.Now().Add(30 * time.Second)); err != nil {
			return nil, err
		}

		_, msg, err := c.conn.ReadMessage()
		if err != nil {
			return nil, fmt.Errorf("read message: %w", err)
		}

		if len(msg) < 4 {
			continue
		}

		headerLen := binary.BigEndian.Uint32(msg[:4])
		if int(4+headerLen) > len(msg) {
			continue
		}

		header := string(msg[4 : 4+headerLen])

		if strings.Contains(header, "Turn.End") {
			break
		}

		if strings.Contains(header, "Path:audio") {
			buf = append(buf, msg[4+headerLen:]...)
		}
	}

	return buf, nil
}

func escapeXML(s string) string {
	s = strings.ReplaceAll(s, "&", "&amp;")
	s = strings.ReplaceAll(s, "<", "&lt;")
	s = strings.ReplaceAll(s, ">", "&gt;")
	s = strings.ReplaceAll(s, "'", "&apos;")
	s = strings.ReplaceAll(s, "\"", "&quot;")
	return s
}
