package stability

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"

	"github.com/promptrails/mediarails"
)

const defaultBaseURL = "https://api.stability.ai/v2beta/stable-image/generate/core"

// Provider implements mediarails.Provider for Stability AI image generation.
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

// New creates a new Stability AI provider.
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

func (p *Provider) ID() string { return "stability" }
func (p *Provider) SupportedTypes() []mediarails.MediaType {
	return []mediarails.MediaType{mediarails.ImageGen}
}

// Generate creates an image from a text prompt. Returns decoded PNG bytes synchronously.
func (p *Provider) Generate(ctx context.Context, req *mediarails.GenerateRequest) (*mediarails.GenerateResponse, error) {
	var buf bytes.Buffer
	writer := multipart.NewWriter(&buf)

	_ = writer.WriteField("prompt", req.Prompt)
	_ = writer.WriteField("output_format", "png")
	if req.Model != "" {
		_ = writer.WriteField("model", req.Model)
	}
	for k, v := range req.Config {
		_ = writer.WriteField(k, fmt.Sprintf("%v", v))
	}
	_ = writer.Close()

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, p.baseURL, &buf)
	if err != nil {
		return nil, fmt.Errorf("stability: request error: %w", err)
	}
	httpReq.Header.Set("Content-Type", writer.FormDataContentType())
	httpReq.Header.Set("Authorization", "Bearer "+p.apiKey)
	httpReq.Header.Set("Accept", "application/json")

	resp, err := p.client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("stability: request failed: %w", err)
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("stability: API error (status %d): %s", resp.StatusCode, string(respBody))
	}

	// Response contains base64-encoded image
	var result struct {
		Image        string `json:"image"`
		FinishReason string `json:"finish_reason"`
		Seed         int    `json:"seed"`
	}
	if err := json.Unmarshal(respBody, &result); err != nil {
		return nil, fmt.Errorf("stability: parse error: %w", err)
	}

	imageData, err := base64.StdEncoding.DecodeString(result.Image)
	if err != nil {
		return nil, fmt.Errorf("stability: base64 decode error: %w", err)
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
			"finish_reason": result.FinishReason,
			"seed":          result.Seed,
		},
	}, nil
}

// CheckStatus returns ErrNotAsync as Stability is synchronous.
func (p *Provider) CheckStatus(_ context.Context, _ string) (*mediarails.GenerateResponse, error) {
	return nil, mediarails.ErrNotAsync
}
