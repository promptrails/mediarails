package elevenlabs

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/promptrails/mediarails"
)

const defaultBaseURL = "https://api.elevenlabs.io/v1/text-to-speech"

// Provider implements mediarails.Provider for ElevenLabs TTS.
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

// New creates a new ElevenLabs provider.
func New(apiKey string, opts ...Option) *Provider {
	p := &Provider{
		apiKey:  apiKey,
		baseURL: defaultBaseURL,
		client:  &http.Client{Timeout: 60 * 1e9},
	}
	for _, opt := range opts {
		opt(p)
	}
	return p
}

func (p *Provider) ID() string { return "elevenlabs" }
func (p *Provider) SupportedTypes() []mediarails.MediaType {
	return []mediarails.MediaType{mediarails.TTS}
}

// Generate produces speech audio from text. Returns audio/mpeg bytes synchronously.
// Config options: voice_id (required), stability (default 0.5), similarity_boost (default 0.75).
func (p *Provider) Generate(ctx context.Context, req *mediarails.GenerateRequest) (*mediarails.GenerateResponse, error) {
	voiceID, _ := req.Config["voice_id"].(string)
	if voiceID == "" {
		return nil, fmt.Errorf("elevenlabs: voice_id is required in config")
	}

	stability := 0.5
	if v, ok := req.Config["stability"].(float64); ok {
		stability = v
	}
	similarityBoost := 0.75
	if v, ok := req.Config["similarity_boost"].(float64); ok {
		similarityBoost = v
	}

	body := map[string]interface{}{
		"text":     req.Prompt,
		"model_id": req.Model,
		"voice_settings": map[string]float64{
			"stability":        stability,
			"similarity_boost": similarityBoost,
		},
	}

	jsonBody, err := json.Marshal(body)
	if err != nil {
		return nil, fmt.Errorf("elevenlabs: marshal error: %w", err)
	}

	url := fmt.Sprintf("%s/%s", p.baseURL, voiceID)
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(jsonBody))
	if err != nil {
		return nil, fmt.Errorf("elevenlabs: request error: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("xi-api-key", p.apiKey)

	resp, err := p.client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("elevenlabs: request failed: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("elevenlabs: read error: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("elevenlabs: API error (status %d): %s", resp.StatusCode, string(respBody))
	}

	return &mediarails.GenerateResponse{
		Status:      mediarails.JobCompleted,
		AssetData:   respBody,
		ContentType: "audio/mpeg",
		Usage: &mediarails.Usage{
			Unit:     "characters",
			Quantity: float64(len(req.Prompt)),
		},
		Metadata: map[string]any{
			"model":    req.Model,
			"voice_id": voiceID,
		},
	}, nil
}

// CheckStatus returns ErrNotAsync as ElevenLabs is synchronous.
func (p *Provider) CheckStatus(_ context.Context, _ string) (*mediarails.GenerateResponse, error) {
	return nil, mediarails.ErrNotAsync
}
