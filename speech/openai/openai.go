package openai

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"

	"github.com/promptrails/mediarails"
)

const defaultBaseURL = "https://api.openai.com/v1"

// Provider implements mediarails.Provider for OpenAI TTS and Whisper STT.
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

// New creates a new OpenAI speech provider.
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

func (p *Provider) ID() string { return "openai" }
func (p *Provider) SupportedTypes() []mediarails.MediaType {
	return []mediarails.MediaType{mediarails.TTS, mediarails.STT}
}

// Generate handles TTS and STT based on req.Type.
func (p *Provider) Generate(ctx context.Context, req *mediarails.GenerateRequest) (*mediarails.GenerateResponse, error) {
	switch req.Type {
	case mediarails.TTS:
		return p.generateTTS(ctx, req)
	case mediarails.STT:
		return p.generateSTT(ctx, req)
	default:
		return nil, fmt.Errorf("openai speech: unsupported type: %s", req.Type)
	}
}

func (p *Provider) generateTTS(ctx context.Context, req *mediarails.GenerateRequest) (*mediarails.GenerateResponse, error) {
	model := req.Model
	if model == "" {
		model = "tts-1"
	}
	voice, _ := req.Config["voice"].(string)
	if voice == "" {
		voice = "alloy"
	}

	body, _ := json.Marshal(map[string]string{
		"model": model,
		"input": req.Prompt,
		"voice": voice,
	})

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, p.baseURL+"/audio/speech", bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("openai speech: request error: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+p.apiKey)

	resp, err := p.client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("openai speech: request failed: %w", err)
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("openai speech: API error (status %d): %s", resp.StatusCode, string(respBody))
	}

	return &mediarails.GenerateResponse{
		Status:      mediarails.JobCompleted,
		AssetData:   respBody,
		ContentType: "audio/mpeg",
		Usage: &mediarails.Usage{
			Unit:     "characters",
			Quantity: float64(len(req.Prompt)),
		},
		Metadata: map[string]any{"model": model, "voice": voice},
	}, nil
}

func (p *Provider) generateSTT(ctx context.Context, req *mediarails.GenerateRequest) (*mediarails.GenerateResponse, error) {
	if len(req.InputData) == 0 {
		return nil, fmt.Errorf("openai speech: STT requires InputData (audio bytes)")
	}

	model := req.Model
	if model == "" {
		model = "whisper-1"
	}

	var buf bytes.Buffer
	writer := multipart.NewWriter(&buf)
	_ = writer.WriteField("model", model)

	part, err := writer.CreateFormFile("file", "audio.mp3")
	if err != nil {
		return nil, fmt.Errorf("openai speech: multipart error: %w", err)
	}
	_, _ = part.Write(req.InputData)
	_ = writer.Close()

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, p.baseURL+"/audio/transcriptions", &buf)
	if err != nil {
		return nil, fmt.Errorf("openai speech: request error: %w", err)
	}
	httpReq.Header.Set("Content-Type", writer.FormDataContentType())
	httpReq.Header.Set("Authorization", "Bearer "+p.apiKey)

	resp, err := p.client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("openai speech: request failed: %w", err)
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("openai speech: API error (status %d): %s", resp.StatusCode, string(respBody))
	}

	var result struct {
		Text string `json:"text"`
	}
	if err := json.Unmarshal(respBody, &result); err != nil {
		return nil, fmt.Errorf("openai speech: parse error: %w", err)
	}

	return &mediarails.GenerateResponse{
		Status:     mediarails.JobCompleted,
		TextOutput: result.Text,
		Metadata:   map[string]any{"model": model},
	}, nil
}

// CheckStatus returns ErrNotAsync as OpenAI speech is synchronous.
func (p *Provider) CheckStatus(_ context.Context, _ string) (*mediarails.GenerateResponse, error) {
	return nil, mediarails.ErrNotAsync
}
