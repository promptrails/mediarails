package runway

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/promptrails/mediarails"
)

const defaultBaseURL = "https://api.dev.runwayml.com/v1"

// Provider implements mediarails.Provider for Runway ML video generation.
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

// New creates a new Runway provider.
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

func (p *Provider) ID() string { return "runway" }
func (p *Provider) SupportedTypes() []mediarails.MediaType {
	return []mediarails.MediaType{mediarails.VideoGen, mediarails.VideoFromImage}
}

// Generate submits a video generation task. Always async.
// Config options: duration (default 5, supports 5 or 10).
func (p *Provider) Generate(ctx context.Context, req *mediarails.GenerateRequest) (*mediarails.GenerateResponse, error) {
	body := map[string]interface{}{
		"promptText": req.Prompt,
		"model":      req.Model,
	}
	if req.InputURL != "" {
		body["promptImage"] = req.InputURL
	}

	duration := 5
	if d, ok := req.Config["duration"].(int); ok {
		duration = d
	}
	body["duration"] = duration

	for k, v := range req.Config {
		if k != "duration" {
			body[k] = v
		}
	}

	jsonBody, _ := json.Marshal(body)
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, p.baseURL+"/image_to_video", bytes.NewReader(jsonBody))
	if err != nil {
		return nil, fmt.Errorf("runway: request error: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+p.apiKey)
	httpReq.Header.Set("X-Runway-Version", "2024-11-06")

	resp, err := p.client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("runway: request failed: %w", err)
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		return nil, fmt.Errorf("runway: API error (status %d): %s", resp.StatusCode, string(respBody))
	}

	var result struct {
		ID string `json:"id"`
	}
	if err := json.Unmarshal(respBody, &result); err != nil {
		return nil, fmt.Errorf("runway: parse error: %w", err)
	}

	return &mediarails.GenerateResponse{
		JobID:  result.ID,
		Status: mediarails.JobProcessing,
		Usage: &mediarails.Usage{
			Unit:     "seconds",
			Quantity: float64(duration),
		},
	}, nil
}

// CheckStatus polls the task status.
func (p *Provider) CheckStatus(ctx context.Context, jobID string) (*mediarails.GenerateResponse, error) {
	url := fmt.Sprintf("%s/tasks/%s", p.baseURL, jobID)

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("runway: status error: %w", err)
	}
	httpReq.Header.Set("Authorization", "Bearer "+p.apiKey)
	httpReq.Header.Set("X-Runway-Version", "2024-11-06")

	resp, err := p.client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("runway: status failed: %w", err)
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)
	var result struct {
		ID     string   `json:"id"`
		Status string   `json:"status"`
		Output []string `json:"output"`
	}
	if err := json.Unmarshal(respBody, &result); err != nil {
		return nil, fmt.Errorf("runway: status parse error: %w", err)
	}

	switch result.Status {
	case "SUCCEEDED":
		assetURL := ""
		if len(result.Output) > 0 {
			assetURL = result.Output[0]
		}
		return &mediarails.GenerateResponse{
			JobID:       jobID,
			Status:      mediarails.JobCompleted,
			AssetURL:    assetURL,
			ContentType: "video/mp4",
		}, nil
	case "FAILED":
		return &mediarails.GenerateResponse{JobID: jobID, Status: mediarails.JobFailed}, nil
	default:
		return &mediarails.GenerateResponse{JobID: jobID, Status: mediarails.JobProcessing}, nil
	}
}
