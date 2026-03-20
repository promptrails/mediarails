package deepgram

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
		if r.Header.Get("Authorization") != "Token test-key" {
			t.Error("expected Token auth")
		}
		_, _ = w.Write([]byte("fake-audio"))
	}))
	defer server.Close()

	p := New("test-key", WithBaseURL(server.URL))
	resp, err := p.Generate(context.Background(), &mediarails.GenerateRequest{
		Type:   mediarails.TTS,
		Model:  "aura-asteria-en",
		Prompt: "Hello",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.Status != mediarails.JobCompleted {
		t.Error("expected completed")
	}
	if resp.ContentType != "audio/mp3" {
		t.Errorf("expected audio/mp3, got %q", resp.ContentType)
	}
}

func TestProvider_STT(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		resp := map[string]interface{}{
			"results": map[string]interface{}{
				"channels": []map[string]interface{}{
					{"alternatives": []map[string]interface{}{
						{"transcript": "hello world", "confidence": 0.95},
					}},
				},
			},
			"metadata": map[string]interface{}{
				"duration": 2.5,
			},
		}
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	p := New("key", WithBaseURL(server.URL))
	resp, err := p.Generate(context.Background(), &mediarails.GenerateRequest{
		Type:      mediarails.STT,
		Model:     "nova-2",
		InputData: []byte("fake-audio-data"),
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.TextOutput != "hello world" {
		t.Errorf("expected 'hello world', got %q", resp.TextOutput)
	}
	if resp.Usage.Unit != "seconds" {
		t.Error("expected seconds usage")
	}
}

func TestProvider_STT_NoInput(t *testing.T) {
	p := New("key")
	_, err := p.Generate(context.Background(), &mediarails.GenerateRequest{
		Type:  mediarails.STT,
		Model: "nova-2",
	})
	if err == nil {
		t.Fatal("expected error for missing input")
	}
}

func TestProvider_UnsupportedType(t *testing.T) {
	p := New("key")
	_, err := p.Generate(context.Background(), &mediarails.GenerateRequest{
		Type: mediarails.ImageGen,
	})
	if err == nil {
		t.Fatal("expected error for unsupported type")
	}
}

func TestProvider_ID(t *testing.T) {
	if New("k").ID() != "deepgram" {
		t.Error("wrong ID")
	}
}

func TestProvider_SupportedTypes(t *testing.T) {
	types := New("k").SupportedTypes()
	if len(types) != 2 {
		t.Errorf("expected 2 types, got %d", len(types))
	}
}

func TestProvider_CheckStatus(t *testing.T) {
	_, err := New("k").CheckStatus(context.Background(), "job")
	if err != mediarails.ErrNotAsync {
		t.Error("expected ErrNotAsync")
	}
}
