package runway

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
		_ = json.NewEncoder(w).Encode(map[string]string{"id": "task-1"})
	}))
	defer server.Close()
	p := New("key", WithBaseURL(server.URL))
	resp, err := p.Generate(context.Background(), &mediarails.GenerateRequest{Type: mediarails.VideoGen, Model: "gen3", Prompt: "a sunset"})
	if err != nil {
		t.Fatalf("error: %v", err)
	}
	if resp.JobID != "task-1" {
		t.Error("expected job ID")
	}
	if resp.Status != mediarails.JobProcessing {
		t.Error("expected processing")
	}
}

func TestProvider_CheckStatus_Completed(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_ = json.NewEncoder(w).Encode(map[string]interface{}{"status": "SUCCEEDED", "output": []string{"https://video.mp4"}})
	}))
	defer server.Close()
	p := New("key", WithBaseURL(server.URL))
	resp, _ := p.CheckStatus(context.Background(), "task-1")
	if resp.Status != mediarails.JobCompleted {
		t.Error("expected completed")
	}
	if resp.AssetURL != "https://video.mp4" {
		t.Error("expected URL")
	}
}

func TestProvider_ID(t *testing.T) {
	if New("k").ID() != "runway" {
		t.Error("wrong")
	}
}
func TestProvider_Types(t *testing.T) {
	if len(New("k").SupportedTypes()) != 2 {
		t.Error("expected 2")
	}
}
