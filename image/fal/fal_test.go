package fal

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/promptrails/mediarails"
)

func TestProvider_Generate_Sync(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		resp := map[string]interface{}{
			"images": []map[string]interface{}{{"url": "https://example.com/img.png"}},
		}
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()
	p := New("key", WithBaseURL(server.URL, server.URL))
	resp, err := p.Generate(context.Background(), &mediarails.GenerateRequest{Type: mediarails.ImageGen, Model: "flux", Prompt: "a cat"})
	if err != nil {
		t.Fatalf("error: %v", err)
	}
	if resp.Status != mediarails.JobCompleted {
		t.Error("expected completed")
	}
	if resp.AssetURL == "" {
		t.Error("expected asset URL")
	}
}

func TestProvider_Generate_Async(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		resp := map[string]interface{}{"request_id": "job-123"}
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()
	p := New("key", WithBaseURL(server.URL, server.URL))
	resp, err := p.Generate(context.Background(), &mediarails.GenerateRequest{Type: mediarails.ImageGen, Model: "flux", Prompt: "a cat"})
	if err != nil {
		t.Fatalf("error: %v", err)
	}
	if resp.Status != mediarails.JobProcessing {
		t.Error("expected processing")
	}
	if resp.JobID != "job-123" {
		t.Error("expected job ID")
	}
}

func TestProvider_ID(t *testing.T) {
	if New("k").ID() != "fal" {
		t.Error("wrong ID")
	}
}

func TestProvider_SupportedTypes(t *testing.T) {
	if len(New("k").SupportedTypes()) != 3 {
		t.Error("expected 3 types")
	}
}
