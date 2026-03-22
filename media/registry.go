package media

import (
	"fmt"

	"github.com/promptrails/mediarails"
	"github.com/promptrails/mediarails/image/fal"
	openaiimage "github.com/promptrails/mediarails/image/openai"
	"github.com/promptrails/mediarails/image/replicate"
	"github.com/promptrails/mediarails/image/stability"
	"github.com/promptrails/mediarails/speech/deepgram"
	"github.com/promptrails/mediarails/speech/elevenlabs"
	openaispeech "github.com/promptrails/mediarails/speech/openai"
	"github.com/promptrails/mediarails/video/luma"
	"github.com/promptrails/mediarails/video/pika"
	"github.com/promptrails/mediarails/video/runway"
)

// ProviderName identifies a supported media provider.
type ProviderName string

// Speech providers.
const (
	ElevenLabs  ProviderName = "elevenlabs"
	Deepgram    ProviderName = "deepgram"
	OpenAIAudio ProviderName = "openai_audio"
)

// Image providers.
const (
	Fal         ProviderName = "fal"
	Replicate   ProviderName = "replicate"
	Stability   ProviderName = "stability"
	OpenAIImage ProviderName = "openai_image"
)

// Video providers.
const (
	Runway ProviderName = "runway"
	Pika   ProviderName = "pika"
	Luma   ProviderName = "luma"
)

// New creates a new media provider by name.
//
//	provider, err := media.New(media.ElevenLabs, "api-key")
//	provider, err := media.New(media.Runway, "api-key")
func New(name ProviderName, apiKey string) (mediarails.Provider, error) {
	switch name {
	case ElevenLabs:
		return elevenlabs.New(apiKey), nil
	case Deepgram:
		return deepgram.New(apiKey), nil
	case OpenAIAudio:
		return openaispeech.New(apiKey), nil
	case Fal:
		return fal.New(apiKey), nil
	case Replicate:
		return replicate.New(apiKey), nil
	case Stability:
		return stability.New(apiKey), nil
	case OpenAIImage:
		return openaiimage.New(apiKey), nil
	case Runway:
		return runway.New(apiKey), nil
	case Pika:
		return pika.New(apiKey), nil
	case Luma:
		return luma.New(apiKey), nil
	default:
		return nil, fmt.Errorf("mediarails: unknown provider %q", name)
	}
}

// MustNew creates a new provider and panics on error.
func MustNew(name ProviderName, apiKey string) mediarails.Provider {
	p, err := New(name, apiKey)
	if err != nil {
		panic(err)
	}
	return p
}

// AllProviders returns all registered provider names.
func AllProviders() []ProviderName {
	return []ProviderName{
		ElevenLabs, Deepgram, OpenAIAudio,
		Fal, Replicate, Stability, OpenAIImage,
		Runway, Pika, Luma,
	}
}
