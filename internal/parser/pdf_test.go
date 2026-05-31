package parser

import (
	"bytes"
	"testing"
)

func TestPdfParser_CanParse(t *testing.T) {
	p := NewPdfParser()
	tests := []struct {
		filename string
		expect   bool
	}{
		{"test.pdf", true},
		{"document.PDF", true},
		{"file.PdF", true},
		{"book.txt", false},
		{"book.epub", false},
		{"", false},
	}
	for _, tc := range tests {
		if got := p.CanParse(tc.filename); got != tc.expect {
			t.Errorf("CanParse(%q) = %v, want %v", tc.filename, got, tc.expect)
		}
	}
}

func TestPdfParser_Parse_Corrupt(t *testing.T) {
	p := NewPdfParser()
	data := []byte("not a pdf at all")
	_, err := p.Parse(bytes.NewReader(data), int64(len(data)))
	if err == nil {
		t.Error("expected error for corrupt PDF")
	}
}

func TestPdfParser_Parse_Empty(t *testing.T) {
	p := NewPdfParser()
	_, err := p.Parse(bytes.NewReader([]byte{}), 0)
	if err == nil {
		t.Error("expected error for empty data")
	}
}
