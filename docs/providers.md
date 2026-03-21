# Providers

All providers implement the `Provider` interface:

```go
type Provider interface {
    ID() string
    SupportedTypes() []MediaType
    Generate(ctx context.Context, req *GenerateRequest) (*GenerateResponse, error)
    CheckStatus(ctx context.Context, jobID string) (*GenerateResponse, error)
}
```

## Speech

### OpenAI (TTS + Whisper STT)

```go
import "github.com/promptrails/mediarails/speech/openai"

p := openai.New("api-key")
```

**Operations**: TTS + STT (sync)
**TTS Config**: `voice` (default "alloy", options: alloy, echo, fable, onyx, nova, shimmer)
**TTS Models**: `tts-1`, `tts-1-hd`
**STT Models**: `whisper-1`
**STT Input**: `InputData` (raw audio bytes)

### ElevenLabs

```go
import "github.com/promptrails/mediarails/speech/elevenlabs"

p := elevenlabs.New("api-key")
```

**Operations**: TTS (sync)
**Config**: `voice_id` (required), `stability` (default 0.5), `similarity_boost` (default 0.75)
**Output**: audio/mpeg bytes
**Metering**: characters

### Deepgram

```go
import "github.com/promptrails/mediarails/speech/deepgram"

p := deepgram.New("api-key")
```

**Operations**: TTS + STT (sync)
**TTS Output**: audio/mp3 bytes
**STT Input**: `InputURL` (JSON) or `InputData` (raw audio bytes)
**STT Output**: `TextOutput` (transcript)
**Metering**: characters (TTS), seconds (STT)

## Image

### OpenAI DALL-E

```go
import "github.com/promptrails/mediarails/image/openai"

p := openai.New("api-key")
```

**Operations**: Image gen (sync)
**Models**: `dall-e-3` (default), `dall-e-2`
**Config**: `size` (default "1024x1024"), `quality` ("standard" or "hd"), `style` ("vivid" or "natural")
**Output**: decoded PNG bytes
**Metering**: 1 image per request
**Metadata**: includes `revised_prompt` (DALL-E 3's rewritten prompt)

### Stability AI

```go
import "github.com/promptrails/mediarails/image/stability"

p := stability.New("api-key")
```

**Operations**: Image gen (sync)
**Output**: decoded PNG bytes
**Metering**: 1 image per request

### Fal AI

Fal supports both image and video generation. Despite being in the `image/` package, it fully handles video and image-to-video.

```go
import "github.com/promptrails/mediarails/image/fal"

p := fal.New("api-key")

// Image generation
resp, _ := p.Generate(ctx, &mediarails.GenerateRequest{
    Type:   mediarails.ImageGen,
    Model:  "fal-ai/flux/schnell",
    Prompt: "a cat in space",
})

// Video generation
resp, _ = p.Generate(ctx, &mediarails.GenerateRequest{
    Type:   mediarails.VideoGen,
    Model:  "fal-ai/minimax-video",
    Prompt: "ocean waves at sunset",
})

// Image-to-video
resp, _ = p.Generate(ctx, &mediarails.GenerateRequest{
    Type:     mediarails.VideoFromImage,
    Model:    "fal-ai/minimax-video",
    Prompt:   "animate this scene",
    InputURL: "https://example.com/photo.jpg",
})
```

**Operations**: ImageGen, VideoGen, VideoFromImage (hybrid async)
**Model**: passed as URL path (e.g., `fal-ai/flux/schnell`, `fal-ai/minimax-video`)
**Output**: `AssetURL` when completed, `JobID` when async (poll via `CheckStatus`)
**Metering**: 1 per request

### Replicate

```go
import "github.com/promptrails/mediarails/image/replicate"

p := replicate.New("api-key")
```

**Operations**: Image gen, Video gen (async with `Prefer: wait`)
**Output**: `AssetURL` (string or first element of array)
**Metering**: gpu_seconds

## Video

### Runway

```go
import "github.com/promptrails/mediarails/video/runway"

p := runway.New("api-key")
```

**Operations**: Video gen, Image-to-video (always async)
**Config**: `duration` (5 or 10 seconds, default 5)
**Image-to-video**: set `InputURL` to source image
**Metering**: seconds

### Pika

```go
import "github.com/promptrails/mediarails/video/pika"

p := pika.New("api-key")
```

**Operations**: Video gen (always async)
**Metering**: seconds (from video duration)

### Luma Dream Machine

```go
import "github.com/promptrails/mediarails/video/luma"

p := luma.New("api-key")
```

**Operations**: Video gen, Image-to-video (always async)
**Image-to-video**: set `InputURL` (uses keyframes internally)
**Metering**: 5 seconds (default)

## Common Options

All providers support:

```go
// Custom base URL
provider := elevenlabs.New("key", elevenlabs.WithBaseURL("https://custom"))

// Custom HTTP client
provider := elevenlabs.New("key", elevenlabs.WithHTTPClient(&http.Client{Timeout: 2 * time.Minute}))
```
