package luma

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
	resp, err := p.Generate(context.Background(), &mediarails.GenerateRequest{Type: mediarails.VideoGen, Model: "dream-machine", Prompt: "ocean waves"})
	if err != nil {
		t.Fatalf("error: %v", err)
	}
	if resp.JobID != "gen-1" {
		t.Error("expected job ID")
	}
	if resp.Usage.Quantity != 5 {
		t.Error("expected 5s default")
	}
}

func TestProvider_CheckStatus(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_ = json.NewEncoder(w).Encode(map[string]interface{}{"state": "completed", "assets": map[string]interface{}{"video": map[string]interface{}{"url": "https://v.mp4"}}})
	}))
	defer server.Close()
	p := New("key", WithBaseURL(server.URL))
	resp, _ := p.CheckStatus(context.Background(), "gen-1")
	if resp.Status != mediarails.JobCompleted {
		t.Error("expected completed")
	}
	if resp.AssetURL != "https://v.mp4" {
		t.Error("expected URL")
	}
}

func TestProvider_ID(t *testing.T) {
	if New("k").ID() != "luma" {
		t.Error("wrong")
	}
}
func TestProvider_Types(t *testing.T) {
	if len(New("k").SupportedTypes()) != 2 {
		t.Error("expected 2")
	}
}
