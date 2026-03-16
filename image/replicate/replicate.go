package replicate

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/promptrails/mediarails"
)

const defaultBaseURL = "https://api.replicate.com/v1/predictions"

// Provider implements mediarails.Provider for Replicate (image + video).
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

// New creates a new Replicate provider.
func New(apiKey string, opts ...Option) *Provider {
	p := &Provider{
		apiKey:  apiKey,
		baseURL: defaultBaseURL,
		client:  &http.Client{Timeout: 300 * 1e9},
	}
	for _, opt := range opts {
		opt(p)
	}
	return p
}

func (p *Provider) ID() string { return "replicate" }
func (p *Provider) SupportedTypes() []mediarails.MediaType {
	return []mediarails.MediaType{mediarails.ImageGen, mediarails.VideoGen}
}

// Generate submits a prediction. Uses Prefer: wait for optimistic sync return.
func (p *Provider) Generate(ctx context.Context, req *mediarails.GenerateRequest) (*mediarails.GenerateResponse, error) {
	input := map[string]interface{}{
		"prompt": req.Prompt,
	}
	for k, v := range req.Config {
		input[k] = v
	}

	body := map[string]interface{}{
		"model": req.Model,
		"input": input,
	}

	jsonBody, _ := json.Marshal(body)
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, p.baseURL, bytes.NewReader(jsonBody))
	if err != nil {
		return nil, fmt.Errorf("replicate: request error: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+p.apiKey)
	httpReq.Header.Set("Prefer", "wait")

	resp, err := p.client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("replicate: request failed: %w", err)
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		return nil, fmt.Errorf("replicate: API error (status %d): %s", resp.StatusCode, string(respBody))
	}

	return p.parseResponse(respBody)
}

// CheckStatus polls the status of a prediction.
func (p *Provider) CheckStatus(ctx context.Context, jobID string) (*mediarails.GenerateResponse, error) {
	url := fmt.Sprintf("%s/%s", p.baseURL, jobID)

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("replicate: status request error: %w", err)
	}
	httpReq.Header.Set("Authorization", "Bearer "+p.apiKey)

	resp, err := p.client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("replicate: status request failed: %w", err)
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)
	return p.parseResponse(respBody)
}

func (p *Provider) parseResponse(data []byte) (*mediarails.GenerateResponse, error) {
	var result struct {
		ID      string      `json:"id"`
		Status  string      `json:"status"`
		Output  interface{} `json:"output"`
		Metrics struct {
			PredictTime float64 `json:"predict_time"`
		} `json:"metrics"`
	}

	if err := json.Unmarshal(data, &result); err != nil {
		return nil, fmt.Errorf("replicate: parse error: %w", err)
	}

	resp := &mediarails.GenerateResponse{
		JobID: result.ID,
	}

	switch result.Status {
	case "succeeded":
		resp.Status = mediarails.JobCompleted
		resp.AssetURL = extractURL(result.Output)
		resp.Usage = &mediarails.Usage{
			Unit:     "gpu_seconds",
			Quantity: result.Metrics.PredictTime,
		}
	case "failed", "canceled":
		resp.Status = mediarails.JobFailed
	default:
		resp.Status = mediarails.JobProcessing
	}

	return resp, nil
}

// extractURL handles Replicate's flexible output format (string or array).
func extractURL(output interface{}) string {
	switch v := output.(type) {
	case string:
		return v
	case []interface{}:
		if len(v) > 0 {
			if s, ok := v[0].(string); ok {
				return s
			}
		}
	}
	return ""
}
