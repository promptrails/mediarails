# Getting Started

## Installation

```bash
go get github.com/promptrails/mediarails
```

## Quick Start (Registry)

The easiest way to create any media provider:

```go
import (
    "github.com/promptrails/mediarails"
    "github.com/promptrails/mediarails/media"
)

// Create a provider using the registry
provider, err := media.New(media.ElevenLabs, "your-api-key")
if err != nil {
    log.Fatal(err)
}

// Generate speech
resp, err := provider.Generate(ctx, &mediarails.GenerateRequest{
    Type:   mediarails.TTS,
    Model:  "eleven_multilingual_v2",
    Prompt: "Hello, welcome to MediaRails!",
    Config: map[string]any{"voice_id": "21m00Tcm4TlvDq8ikWAM"},
})
```

Available constants:

| Category | Constants |
|----------|-----------|
| Speech | `media.ElevenLabs`, `media.Deepgram`, `media.OpenAIAudio` |
| Image | `media.Fal`, `media.Replicate`, `media.Stability`, `media.OpenAIImage` |
| Video | `media.Runway`, `media.Pika`, `media.Luma` |

## Direct Provider Import

For provider-specific options, import the provider directly:

## Text-to-Speech (ElevenLabs)

```go
import "github.com/promptrails/mediarails/speech/elevenlabs"

provider := elevenlabs.New("your-api-key")
resp, err := provider.Generate(ctx, &mediarails.GenerateRequest{
    Type:   mediarails.TTS,
    Model:  "eleven_multilingual_v2",
    Prompt: "Hello, welcome to MediaRails!",
    Config: map[string]any{"voice_id": "21m00Tcm4TlvDq8ikWAM"},
})
// resp.AssetData = audio/mpeg bytes
// resp.ContentType = "audio/mpeg"
os.WriteFile("output.mp3", resp.AssetData, 0644)
```

## Image Generation (Stability AI)

```go
import "github.com/promptrails/mediarails/image/stability"

provider := stability.New("your-api-key")
resp, err := provider.Generate(ctx, &mediarails.GenerateRequest{
    Type:   mediarails.ImageGen,
    Prompt: "A cat wearing sunglasses, digital art",
})
// resp.AssetData = PNG bytes
os.WriteFile("output.png", resp.AssetData, 0644)
```

## Video Generation (Runway — Async)

```go
import (
    "github.com/promptrails/mediarails/video/runway"
    "github.com/promptrails/mediarails"
)

provider := runway.New("your-api-key")

// Submit job
resp, err := provider.Generate(ctx, &mediarails.GenerateRequest{
    Type:   mediarails.VideoGen,
    Model:  "gen3a_turbo",
    Prompt: "A sunset over the ocean, cinematic",
})
// resp.JobID = "task-abc123"
// resp.Status = "processing"

// Wait for completion
final, err := mediarails.WaitForCompletion(ctx, provider, resp.JobID,
    2*time.Second,   // initial poll interval
    30*time.Second,  // max interval
    10*time.Minute,  // timeout
)
// final.AssetURL = "https://..."
```

## Cost Tracking

```go
resp, _ := provider.Generate(ctx, req)
if resp.Usage != nil {
    fmt.Printf("Used: %.1f %s\n", resp.Usage.Quantity, resp.Usage.Unit)
    fmt.Printf("Cost: $%.4f\n", resp.Usage.Cost(0.01)) // unit price
}
```
