package elevenlabs

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/promptrails/mediarails"
)

func TestProvider_Generate(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("xi-api-key") != "test-key" {
			t.Error("expected xi-api-key header")
		}
		_, _ = w.Write([]byte("fake-audio-data"))
	}))
	defer server.Close()

	p := New("test-key", WithBaseURL(server.URL))
	resp, err := p.Generate(context.Background(), &mediarails.GenerateRequest{
		Type:   mediarails.TTS,
		Model:  "eleven_multilingual_v2",
		Prompt: "Hello world",
		Config: map[string]any{"voice_id": "abc123"},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.Status != mediarails.JobCompleted {
		t.Error("expected completed")
	}
	if string(resp.AssetData) != "fake-audio-data" {
		t.Error("expected audio data")
	}
	if resp.ContentType != "audio/mpeg" {
		t.Errorf("expected audio/mpeg, got %q", resp.ContentType)
	}
	if resp.Usage.Unit != "characters" {
		t.Error("expected characters usage")
	}
}

func TestProvider_Generate_MissingVoiceID(t *testing.T) {
	p := New("key")
	_, err := p.Generate(context.Background(), &mediarails.GenerateRequest{
		Type:   mediarails.TTS,
		Prompt: "test",
		Config: map[string]any{},
	})
	if err == nil {
		t.Fatal("expected error for missing voice_id")
	}
}

func TestProvider_CheckStatus(t *testing.T) {
	p := New("key")
	_, err := p.CheckStatus(context.Background(), "job-1")
	if err != mediarails.ErrNotAsync {
		t.Error("expected ErrNotAsync")
	}
}

func TestProvider_ID(t *testing.T) {
	p := New("key")
	if p.ID() != "elevenlabs" {
		t.Errorf("expected elevenlabs, got %q", p.ID())
	}
}

func TestProvider_SupportedTypes(t *testing.T) {
	p := New("key")
	types := p.SupportedTypes()
	if len(types) != 1 || types[0] != mediarails.TTS {
		t.Error("expected [TTS]")
	}
}

func TestProvider_APIError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		_, _ = w.Write([]byte("unauthorized"))
	}))
	defer server.Close()

	p := New("bad-key", WithBaseURL(server.URL))
	_, err := p.Generate(context.Background(), &mediarails.GenerateRequest{
		Type:   mediarails.TTS,
		Prompt: "test",
		Config: map[string]any{"voice_id": "abc"},
	})
	if err == nil {
		t.Fatal("expected error")
	}
}
