package deepgram

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/promptrails/mediarails"
)

const defaultBaseURL = "https://api.deepgram.com/v1"

// Provider implements mediarails.Provider for Deepgram TTS and STT.
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

// New creates a new Deepgram provider.
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

func (p *Provider) ID() string { return "deepgram" }
func (p *Provider) SupportedTypes() []mediarails.MediaType {
	return []mediarails.MediaType{mediarails.TTS, mediarails.STT}
}

// Generate handles both TTS and STT based on req.Type.
func (p *Provider) Generate(ctx context.Context, req *mediarails.GenerateRequest) (*mediarails.GenerateResponse, error) {
	switch req.Type {
	case mediarails.TTS:
		return p.generateTTS(ctx, req)
	case mediarails.STT:
		return p.generateSTT(ctx, req)
	default:
		return nil, fmt.Errorf("deepgram: unsupported media type: %s", req.Type)
	}
}

func (p *Provider) generateTTS(ctx context.Context, req *mediarails.GenerateRequest) (*mediarails.GenerateResponse, error) {
	url := fmt.Sprintf("%s/speak?model=%s", p.baseURL, req.Model)

	body, _ := json.Marshal(map[string]string{"text": req.Prompt})
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("deepgram: request error: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Token "+p.apiKey)

	resp, err := p.client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("deepgram: request failed: %w", err)
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("deepgram: API error (status %d): %s", resp.StatusCode, string(respBody))
	}

	return &mediarails.GenerateResponse{
		Status:      mediarails.JobCompleted,
		AssetData:   respBody,
		ContentType: "audio/mp3",
		Usage: &mediarails.Usage{
			Unit:     "characters",
			Quantity: float64(len(req.Prompt)),
		},
	}, nil
}

func (p *Provider) generateSTT(ctx context.Context, req *mediarails.GenerateRequest) (*mediarails.GenerateResponse, error) {
	url := fmt.Sprintf("%s/listen?model=%s&smart_format=true", p.baseURL, req.Model)

	var httpReq *http.Request
	var err error

	if req.InputURL != "" {
		body, _ := json.Marshal(map[string]string{"url": req.InputURL})
		httpReq, err = http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
		if err != nil {
			return nil, fmt.Errorf("deepgram: request error: %w", err)
		}
		httpReq.Header.Set("Content-Type", "application/json")
	} else if len(req.InputData) > 0 {
		httpReq, err = http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(req.InputData))
		if err != nil {
			return nil, fmt.Errorf("deepgram: request error: %w", err)
		}
		httpReq.Header.Set("Content-Type", "audio/wav")
	} else {
		return nil, fmt.Errorf("deepgram: STT requires InputURL or InputData")
	}

	httpReq.Header.Set("Authorization", "Token "+p.apiKey)

	resp, err := p.client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("deepgram: request failed: %w", err)
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("deepgram: API error (status %d): %s", resp.StatusCode, string(respBody))
	}

	var result struct {
		Results struct {
			Channels []struct {
				Alternatives []struct {
					Transcript string  `json:"transcript"`
					Confidence float64 `json:"confidence"`
				} `json:"alternatives"`
			} `json:"channels"`
		} `json:"results"`
		Metadata struct {
			Duration float64 `json:"duration"`
		} `json:"metadata"`
	}

	if err := json.Unmarshal(respBody, &result); err != nil {
		return nil, fmt.Errorf("deepgram: parse error: %w", err)
	}

	transcript := ""
	if len(result.Results.Channels) > 0 && len(result.Results.Channels[0].Alternatives) > 0 {
		transcript = result.Results.Channels[0].Alternatives[0].Transcript
	}

	return &mediarails.GenerateResponse{
		Status:     mediarails.JobCompleted,
		TextOutput: transcript,
		Usage: &mediarails.Usage{
			Unit:     "seconds",
			Quantity: result.Metadata.Duration,
		},
		Metadata: map[string]any{
			"duration": result.Metadata.Duration,
		},
	}, nil
}

// CheckStatus returns ErrNotAsync as Deepgram is synchronous.
func (p *Provider) CheckStatus(_ context.Context, _ string) (*mediarails.GenerateResponse, error) {
	return nil, mediarails.ErrNotAsync
}
