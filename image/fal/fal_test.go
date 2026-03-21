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

func TestProvider_Generate_VideoSync(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		resp := map[string]interface{}{
			"video": map[string]interface{}{"url": "https://example.com/video.mp4"},
		}
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()
	p := New("key", WithBaseURL(server.URL, server.URL))
	resp, err := p.Generate(context.Background(), &mediarails.GenerateRequest{
		Type:   mediarails.VideoGen,
		Model:  "fal-ai/minimax-video",
		Prompt: "a sunset over the ocean",
	})
	if err != nil {
		t.Fatalf("error: %v", err)
	}
	if resp.Status != mediarails.JobCompleted {
		t.Error("expected completed")
	}
	if resp.AssetURL != "https://example.com/video.mp4" {
		t.Errorf("expected video URL, got %q", resp.AssetURL)
	}
}

func TestProvider_Generate_VideoFromImage(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var body map[string]interface{}
		_ = json.NewDecoder(r.Body).Decode(&body)
		if body["image_url"] != "https://example.com/input.jpg" {
			t.Error("expected image_url in request")
		}
		resp := map[string]interface{}{"request_id": "vid-456"}
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()
	p := New("key", WithBaseURL(server.URL, server.URL))
	resp, err := p.Generate(context.Background(), &mediarails.GenerateRequest{
		Type:     mediarails.VideoFromImage,
		Model:    "fal-ai/minimax-video",
		Prompt:   "animate this image",
		InputURL: "https://example.com/input.jpg",
	})
	if err != nil {
		t.Fatalf("error: %v", err)
	}
	if resp.JobID != "vid-456" {
		t.Error("expected job ID")
	}
}

func TestProvider_CheckStatus_Completed(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		resp := map[string]interface{}{
			"status": "COMPLETED",
			"url":    "https://example.com/result.mp4",
		}
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()
	p := New("key", WithBaseURL(server.URL, server.URL))
	resp, err := p.CheckStatus(context.Background(), "job-123")
	if err != nil {
		t.Fatalf("error: %v", err)
	}
	if resp.Status != mediarails.JobCompleted {
		t.Error("expected completed")
	}
	if resp.AssetURL != "https://example.com/result.mp4" {
		t.Errorf("expected URL, got %q", resp.AssetURL)
	}
}

func TestProvider_CheckStatus_Failed(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		resp := map[string]interface{}{"status": "FAILED"}
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()
	p := New("key", WithBaseURL(server.URL, server.URL))
	resp, _ := p.CheckStatus(context.Background(), "job-123")
	if resp.Status != mediarails.JobFailed {
		t.Error("expected failed")
	}
}

func TestProvider_SupportedTypes(t *testing.T) {
	types := New("k").SupportedTypes()
	if len(types) != 3 {
		t.Error("expected 3 types")
	}
	// Should support image, video, and video from image
	hasImage, hasVideo, hasVFI := false, false, false
	for _, typ := range types {
		switch typ {
		case mediarails.ImageGen:
			hasImage = true
		case mediarails.VideoGen:
			hasVideo = true
		case mediarails.VideoFromImage:
			hasVFI = true
		}
	}
	if !hasImage || !hasVideo || !hasVFI {
		t.Error("expected ImageGen, VideoGen, and VideoFromImage")
	}
}
