package stability

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/promptrails/mediarails"
)

func TestProvider_Generate(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != "Bearer test-key" {
			t.Error("expected Bearer auth")
		}
		resp := map[string]interface{}{
			"image":         base64.StdEncoding.EncodeToString([]byte("fake-png")),
			"finish_reason": "SUCCESS",
			"seed":          42,
		}
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	p := New("test-key", WithBaseURL(server.URL))
	resp, err := p.Generate(context.Background(), &mediarails.GenerateRequest{
		Type:   mediarails.ImageGen,
		Prompt: "a cat",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.Status != mediarails.JobCompleted {
		t.Error("expected completed")
	}
	if string(resp.AssetData) != "fake-png" {
		t.Error("expected decoded PNG data")
	}
	if resp.Usage.Unit != "images" {
		t.Error("expected images usage")
	}
}

func TestProvider_ID(t *testing.T) {
	if New("k").ID() != "stability" {
		t.Error("wrong ID")
	}
}

func TestProvider_SupportedTypes(t *testing.T) {
	types := New("k").SupportedTypes()
	if len(types) != 1 || types[0] != mediarails.ImageGen {
		t.Error("wrong types")
	}
}

func TestProvider_CheckStatus(t *testing.T) {
	_, err := New("k").CheckStatus(context.Background(), "job")
	if err != mediarails.ErrNotAsync {
		t.Error("expected ErrNotAsync")
	}
}

func TestProvider_APIError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte("bad request"))
	}))
	defer server.Close()
	_, err := New("k", WithBaseURL(server.URL)).Generate(context.Background(), &mediarails.GenerateRequest{Type: mediarails.ImageGen, Prompt: "test"})
	if err == nil {
		t.Fatal("expected error")
	}
}
