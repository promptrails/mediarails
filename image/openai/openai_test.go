package openai

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
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		resp := map[string]interface{}{
			"data": []map[string]interface{}{
				{
					"b64_json":       base64.StdEncoding.EncodeToString([]byte("fake-png")),
					"revised_prompt": "a cute cat, digital art",
				},
			},
		}
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	p := New("key", WithBaseURL(server.URL))
	resp, err := p.Generate(context.Background(), &mediarails.GenerateRequest{
		Type:   mediarails.ImageGen,
		Prompt: "a cat",
	})
	if err != nil {
		t.Fatalf("error: %v", err)
	}
	if resp.Status != mediarails.JobCompleted {
		t.Error("expected completed")
	}
	if string(resp.AssetData) != "fake-png" {
		t.Error("expected decoded image data")
	}
	if resp.Metadata["revised_prompt"] != "a cute cat, digital art" {
		t.Error("expected revised prompt in metadata")
	}
}

func TestProvider_ID(t *testing.T) {
	if New("k").ID() != "openai-dalle" {
		t.Error("wrong ID")
	}
}

func TestProvider_SupportedTypes(t *testing.T) {
	types := New("k").SupportedTypes()
	if len(types) != 1 || types[0] != mediarails.ImageGen {
		t.Error("expected [ImageGen]")
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
		_, _ = w.Write([]byte("bad"))
	}))
	defer server.Close()
	_, err := New("k", WithBaseURL(server.URL)).Generate(context.Background(), &mediarails.GenerateRequest{Type: mediarails.ImageGen, Prompt: "test"})
	if err == nil {
		t.Fatal("expected error")
	}
}
