package translator

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestDeepLTranslate_Translate(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v2/translate" {
			t.Fatalf("expected /v2/translate, got %s", r.URL.Path)
		}
		if r.Header.Get("Authorization") != "DeepL-Auth-Key test-key" {
			t.Fatalf("expected DeepL-Auth-Key test-key, got %s", r.Header.Get("Authorization"))
		}
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"translations":[{"text":"hola"}]}`))
	}))
	defer server.Close()

	client := NewDeepLTranslate(server.URL, "test-key")
	result, err := client.Translate(context.Background(), TranslateRequest{
		Word:       "hello",
		SourceLang: "EN",
		TargetLang: "ES",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Translation != "hola" {
		t.Fatalf("expected 'hola', got '%s'", result.Translation)
	}
}

func TestDeepLTranslate_EmptyWord(t *testing.T) {
	client := NewDeepLTranslate("https://api-free.deepl.com", "key")
	_, err := client.Translate(context.Background(), TranslateRequest{
		Word:       "",
		SourceLang: "EN",
		TargetLang: "ES",
	})
	if err == nil {
		t.Fatal("expected error for empty word")
	}
}

func TestDeepLTranslate_NonOKStatus(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusForbidden)
		w.Write([]byte(`{"message":"auth failed"}`))
	}))
	defer server.Close()

	client := NewDeepLTranslate(server.URL, "bad-key")
	_, err := client.Translate(context.Background(), TranslateRequest{
		Word:       "hello",
		SourceLang: "EN",
		TargetLang: "ES",
	})
	if err == nil {
		t.Fatal("expected error for non-200 status")
	}
}

func TestDeepLTranslate_EmptyResponse(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"translations":[]}`))
	}))
	defer server.Close()

	client := NewDeepLTranslate(server.URL, "key")
	_, err := client.Translate(context.Background(), TranslateRequest{
		Word:       "hello",
		SourceLang: "EN",
		TargetLang: "ES",
	})
	if err == nil {
		t.Fatal("expected error for empty translations")
	}
}
