package pika

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/promptrails/mediarails"
)

const defaultBaseURL = "https://api.pika.art/v1/generate"

// Provider implements mediarails.Provider for Pika video generation.
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

// New creates a new Pika provider.
func New(apiKey string, opts ...Option) *Provider {
	p := &Provider{
		apiKey:  apiKey,
		baseURL: defaultBaseURL,
		client:  &http.Client{Timeout: 600 * 1e9},
	}
	for _, opt := range opts {
		opt(p)
	}
	return p
}

func (p *Provider) ID() string { return "pika" }
func (p *Provider) SupportedTypes() []mediarails.MediaType {
	return []mediarails.MediaType{mediarails.VideoGen}
}

// Generate submits a video generation request. Always async.
func (p *Provider) Generate(ctx context.Context, req *mediarails.GenerateRequest) (*mediarails.GenerateResponse, error) {
	body := map[string]interface{}{
		"promptText": req.Prompt,
	}
	for k, v := range req.Config {
		body[k] = v
	}

	jsonBody, _ := json.Marshal(body)
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, p.baseURL, bytes.NewReader(jsonBody))
	if err != nil {
		return nil, fmt.Errorf("pika: request error: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+p.apiKey)

	resp, err := p.client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("pika: request failed: %w", err)
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		return nil, fmt.Errorf("pika: API error (status %d): %s", resp.StatusCode, string(respBody))
	}

	var result struct {
		ID string `json:"id"`
	}
	if err := json.Unmarshal(respBody, &result); err != nil {
		return nil, fmt.Errorf("pika: parse error: %w", err)
	}

	return &mediarails.GenerateResponse{
		JobID:  result.ID,
		Status: mediarails.JobProcessing,
	}, nil
}

// CheckStatus polls the generation status.
func (p *Provider) CheckStatus(ctx context.Context, jobID string) (*mediarails.GenerateResponse, error) {
	url := fmt.Sprintf("%s/%s", p.baseURL, jobID)

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("pika: status error: %w", err)
	}
	httpReq.Header.Set("Authorization", "Bearer "+p.apiKey)

	resp, err := p.client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("pika: status failed: %w", err)
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)
	var result struct {
		Status string `json:"status"`
		Video  struct {
			URL      string  `json:"url"`
			Duration float64 `json:"duration"`
		} `json:"video"`
	}
	if err := json.Unmarshal(respBody, &result); err != nil {
		return nil, fmt.Errorf("pika: status parse error: %w", err)
	}

	switch result.Status {
	case "completed":
		return &mediarails.GenerateResponse{
			JobID:       jobID,
			Status:      mediarails.JobCompleted,
			AssetURL:    result.Video.URL,
			ContentType: "video/mp4",
			Usage: &mediarails.Usage{
				Unit:     "seconds",
				Quantity: result.Video.Duration,
			},
		}, nil
	case "failed":
		return &mediarails.GenerateResponse{JobID: jobID, Status: mediarails.JobFailed}, nil
	default:
		return &mediarails.GenerateResponse{JobID: jobID, Status: mediarails.JobProcessing}, nil
	}
}
