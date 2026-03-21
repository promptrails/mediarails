package openai

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/promptrails/mediarails"
)

const defaultBaseURL = "https://api.openai.com/v1/images/generations"

// Provider implements mediarails.Provider for OpenAI DALL-E image generation.
type Provider struct {
	apiKey  string
	baseURL string
	client  *http.Client
}

// Option configures the provider.
type Option func(*Provider)

// WithBaseURL sets a custom API URL.
func WithBaseURL(url string) Option { return func(p *Provider) { p.baseURL = url } }

// WithHTTPClient sets a custom HTTP client.
func WithHTTPClient(c *http.Client) Option { return func(p *Provider) { p.client = c } }

// New creates a new OpenAI DALL-E provider.
func New(apiKey string, opts ...Option) *Provider {
	p := &Provider{
		apiKey:  apiKey,
		baseURL: defaultBaseURL,
		client:  &http.Client{Timeout: 120 * 1e9},
	}
	for _, opt := range opts {
		opt(p)
	}
	return p
}

func (p *Provider) ID() string { return "openai-dalle" }
func (p *Provider) SupportedTypes() []mediarails.MediaType {
	return []mediarails.MediaType{mediarails.ImageGen}
}

// Generate creates an image from a text prompt using DALL-E.
// Config options: size (default "1024x1024"), quality (default "standard"), style.
func (p *Provider) Generate(ctx context.Context, req *mediarails.GenerateRequest) (*mediarails.GenerateResponse, error) {
	model := req.Model
	if model == "" {
		model = "dall-e-3"
	}

	size := "1024x1024"
	if v, ok := req.Config["size"].(string); ok {
		size = v
	}
	quality := "standard"
	if v, ok := req.Config["quality"].(string); ok {
		quality = v
	}

	body := map[string]interface{}{
		"model":           model,
		"prompt":          req.Prompt,
		"n":               1,
		"size":            size,
		"quality":         quality,
		"response_format": "b64_json",
	}

	if v, ok := req.Config["style"].(string); ok {
		body["style"] = v
	}

	jsonBody, _ := json.Marshal(body)
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, p.baseURL, bytes.NewReader(jsonBody))
	if err != nil {
		return nil, fmt.Errorf("openai dalle: request error: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+p.apiKey)

	resp, err := p.client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("openai dalle: request failed: %w", err)
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("openai dalle: API error (status %d): %s", resp.StatusCode, string(respBody))
	}

	var result struct {
		Data []struct {
			B64JSON       string `json:"b64_json"`
			RevisedPrompt string `json:"revised_prompt"`
		} `json:"data"`
	}
	if err := json.Unmarshal(respBody, &result); err != nil {
		return nil, fmt.Errorf("openai dalle: parse error: %w", err)
	}

	if len(result.Data) == 0 {
		return nil, fmt.Errorf("openai dalle: no images returned")
	}

	imageData, err := base64.StdEncoding.DecodeString(result.Data[0].B64JSON)
	if err != nil {
		return nil, fmt.Errorf("openai dalle: base64 decode error: %w", err)
	}

	return &mediarails.GenerateResponse{
		Status:      mediarails.JobCompleted,
		AssetData:   imageData,
		ContentType: "image/png",
		Usage: &mediarails.Usage{
			Unit:     "images",
			Quantity: 1,
		},
		Metadata: map[string]any{
			"model":          model,
			"revised_prompt": result.Data[0].RevisedPrompt,
			"size":           size,
			"quality":        quality,
		},
	}, nil
}

// CheckStatus returns ErrNotAsync as DALL-E is synchronous.
func (p *Provider) CheckStatus(_ context.Context, _ string) (*mediarails.GenerateResponse, error) {
	return nil, mediarails.ErrNotAsync
}
