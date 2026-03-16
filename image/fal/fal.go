package fal

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/promptrails/mediarails"
)

const (
	defaultQueueURL  = "https://queue.fal.run"
	defaultStatusURL = "https://queue.fal.run/requests"
)

// Provider implements mediarails.Provider for Fal AI (image + video).
type Provider struct {
	apiKey    string
	queueURL  string
	statusURL string
	client    *http.Client
}

// Option configures the provider.
type Option func(*Provider)

// WithBaseURL sets custom queue and status URLs.
func WithBaseURL(queueURL, statusURL string) Option {
	return func(p *Provider) { p.queueURL = queueURL; p.statusURL = statusURL }
}

// WithHTTPClient sets a custom HTTP client.
func WithHTTPClient(c *http.Client) Option { return func(p *Provider) { p.client = c } }

// New creates a new Fal AI provider.
func New(apiKey string, opts ...Option) *Provider {
	p := &Provider{
		apiKey:    apiKey,
		queueURL:  defaultQueueURL,
		statusURL: defaultStatusURL,
		client:    &http.Client{Timeout: 300 * 1e9},
	}
	for _, opt := range opts {
		opt(p)
	}
	return p
}

func (p *Provider) ID() string { return "fal" }
func (p *Provider) SupportedTypes() []mediarails.MediaType {
	return []mediarails.MediaType{mediarails.ImageGen, mediarails.VideoGen, mediarails.VideoFromImage}
}

// Generate submits a generation request. Returns completed result or job ID for polling.
func (p *Provider) Generate(ctx context.Context, req *mediarails.GenerateRequest) (*mediarails.GenerateResponse, error) {
	body := map[string]interface{}{
		"prompt": req.Prompt,
	}
	if req.InputURL != "" {
		body["image_url"] = req.InputURL
	}
	for k, v := range req.Config {
		body[k] = v
	}

	jsonBody, _ := json.Marshal(body)
	url := fmt.Sprintf("%s/%s", p.queueURL, req.Model)

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(jsonBody))
	if err != nil {
		return nil, fmt.Errorf("fal: request error: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Key "+p.apiKey)

	resp, err := p.client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("fal: request failed: %w", err)
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("fal: API error (status %d): %s", resp.StatusCode, string(respBody))
	}

	var result map[string]interface{}
	if err := json.Unmarshal(respBody, &result); err != nil {
		return nil, fmt.Errorf("fal: parse error: %w", err)
	}

	// Check if completed immediately
	if images, ok := result["images"].([]interface{}); ok && len(images) > 0 {
		if img, ok := images[0].(map[string]interface{}); ok {
			return &mediarails.GenerateResponse{
				Status:   mediarails.JobCompleted,
				AssetURL: fmt.Sprintf("%v", img["url"]),
				Usage:    &mediarails.Usage{Unit: "images", Quantity: 1},
			}, nil
		}
	}
	if video, ok := result["video"].(map[string]interface{}); ok {
		return &mediarails.GenerateResponse{
			Status:   mediarails.JobCompleted,
			AssetURL: fmt.Sprintf("%v", video["url"]),
			Usage:    &mediarails.Usage{Unit: "videos", Quantity: 1},
		}, nil
	}

	// Async — return job ID
	requestID, _ := result["request_id"].(string)
	return &mediarails.GenerateResponse{
		JobID:  requestID,
		Status: mediarails.JobProcessing,
	}, nil
}

// CheckStatus polls the status of an async generation job.
func (p *Provider) CheckStatus(ctx context.Context, jobID string) (*mediarails.GenerateResponse, error) {
	url := fmt.Sprintf("%s/%s/status", p.statusURL, jobID)

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("fal: status request error: %w", err)
	}
	httpReq.Header.Set("Authorization", "Key "+p.apiKey)

	resp, err := p.client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("fal: status request failed: %w", err)
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)
	var result map[string]interface{}
	if err := json.Unmarshal(respBody, &result); err != nil {
		return nil, fmt.Errorf("fal: status parse error: %w", err)
	}

	status, _ := result["status"].(string)
	switch status {
	case "COMPLETED":
		assetURL, _ := result["url"].(string)
		return &mediarails.GenerateResponse{
			JobID:    jobID,
			Status:   mediarails.JobCompleted,
			AssetURL: assetURL,
		}, nil
	case "FAILED":
		return &mediarails.GenerateResponse{
			JobID:  jobID,
			Status: mediarails.JobFailed,
		}, nil
	default:
		return &mediarails.GenerateResponse{
			JobID:  jobID,
			Status: mediarails.JobProcessing,
		}, nil
	}
}
