package replicate

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
		resp := map[string]interface{}{
			"id": "pred-123", "status": "succeeded",
			"output":  "https://example.com/img.png",
			"metrics": map[string]interface{}{"predict_time": 2.5},
		}
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()
	p := New("key", WithBaseURL(server.URL))
	resp, err := p.Generate(context.Background(), &mediarails.GenerateRequest{Type: mediarails.ImageGen, Model: "flux", Prompt: "a cat"})
	if err != nil {
		t.Fatalf("error: %v", err)
	}
	if resp.Status != mediarails.JobCompleted {
		t.Error("expected completed")
	}
	if resp.AssetURL != "https://example.com/img.png" {
		t.Error("expected URL")
	}
	if resp.Usage.Unit != "gpu_seconds" {
		t.Error("expected gpu_seconds")
	}
}

func TestProvider_CheckStatus(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		resp := map[string]interface{}{"id": "p1", "status": "processing"}
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()
	p := New("key", WithBaseURL(server.URL))
	resp, _ := p.CheckStatus(context.Background(), "p1")
	if resp.Status != mediarails.JobProcessing {
		t.Error("expected processing")
	}
}

func TestProvider_ID(t *testing.T) {
	if New("k").ID() != "replicate" {
		t.Error("wrong ID")
	}
}

func TestExtractURL(t *testing.T) {
	if extractURL("https://url") != "https://url" {
		t.Error("string URL")
	}
	if extractURL([]interface{}{"https://url"}) != "https://url" {
		t.Error("array URL")
	}
	if extractURL(nil) != "" {
		t.Error("nil")
	}
}
