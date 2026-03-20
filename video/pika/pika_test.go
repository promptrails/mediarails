package pika

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/promptrails/mediarails"
)

func TestProvider_Generate(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_ = json.NewEncoder(w).Encode(map[string]string{"id": "gen-1"})
	}))
	defer server.Close()
	p := New("key", WithBaseURL(server.URL))
	resp, err := p.Generate(context.Background(), &mediarails.GenerateRequest{Type: mediarails.VideoGen, Prompt: "a sunset"})
	if err != nil {
		t.Fatalf("error: %v", err)
	}
	if resp.JobID != "gen-1" {
		t.Error("expected job ID")
	}
}

func TestProvider_CheckStatus(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_ = json.NewEncoder(w).Encode(map[string]interface{}{"status": "completed", "video": map[string]interface{}{"url": "https://v.mp4", "duration": 4.0}})
	}))
	defer server.Close()
	p := New("key", WithBaseURL(server.URL))
	resp, _ := p.CheckStatus(context.Background(), "gen-1")
	if resp.Status != mediarails.JobCompleted {
		t.Error("expected completed")
	}
	if resp.Usage.Quantity != 4.0 {
		t.Error("expected 4s duration")
	}
}

func TestProvider_ID(t *testing.T) {
	if New("k").ID() != "pika" {
		t.Error("wrong")
	}
}
