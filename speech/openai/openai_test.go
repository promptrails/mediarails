package openai

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/promptrails/mediarails"
)

func TestProvider_TTS(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != "Bearer test-key" {
			t.Error("expected Bearer auth")
		}
		_, _ = w.Write([]byte("fake-audio"))
	}))
	defer server.Close()

	p := New("test-key", WithBaseURL(server.URL))
	resp, err := p.Generate(context.Background(), &mediarails.GenerateRequest{
		Type:   mediarails.TTS,
		Prompt: "Hello",
		Config: map[string]any{"voice": "nova"},
	})
	if err != nil {
		t.Fatalf("error: %v", err)
	}
	if resp.Status != mediarails.JobCompleted {
		t.Error("expected completed")
	}
	if resp.ContentType != "audio/mpeg" {
		t.Errorf("expected audio/mpeg, got %q", resp.ContentType)
	}
}

func TestProvider_STT(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_ = json.NewEncoder(w).Encode(map[string]string{"text": "hello world"})
	}))
	defer server.Close()

	p := New("key", WithBaseURL(server.URL))
	resp, err := p.Generate(context.Background(), &mediarails.GenerateRequest{
		Type:      mediarails.STT,
		InputData: []byte("fake-audio"),
	})
	if err != nil {
		t.Fatalf("error: %v", err)
	}
	if resp.TextOutput != "hello world" {
		t.Errorf("expected 'hello world', got %q", resp.TextOutput)
	}
}

func TestProvider_STT_NoInput(t *testing.T) {
	p := New("key")
	_, err := p.Generate(context.Background(), &mediarails.GenerateRequest{Type: mediarails.STT})
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestProvider_UnsupportedType(t *testing.T) {
	p := New("key")
	_, err := p.Generate(context.Background(), &mediarails.GenerateRequest{Type: mediarails.ImageGen})
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestProvider_ID(t *testing.T) {
	if New("k").ID() != "openai" {
		t.Error("wrong ID")
	}
}

func TestProvider_SupportedTypes(t *testing.T) {
	if len(New("k").SupportedTypes()) != 2 {
		t.Error("expected 2 types")
	}
}

func TestProvider_CheckStatus(t *testing.T) {
	_, err := New("k").CheckStatus(context.Background(), "job")
	if err != mediarails.ErrNotAsync {
		t.Error("expected ErrNotAsync")
	}
}
