# MediaRails

Unified AI media generation for Go. Speech, image, and video — one interface, 10 providers.

[![Go Reference](https://pkg.go.dev/badge/github.com/promptrails/mediarails.svg)](https://pkg.go.dev/github.com/promptrails/mediarails)
[![CI](https://github.com/promptrails/mediarails/actions/workflows/ci.yml/badge.svg)](https://github.com/promptrails/mediarails/actions/workflows/ci.yml)
[![Go Report Card](https://goreportcard.com/badge/github.com/promptrails/mediarails)](https://goreportcard.com/report/github.com/promptrails/mediarails)

```go
provider := elevenlabs.New("api-key")
resp, _ := provider.Generate(ctx, &mediarails.GenerateRequest{
    Type:   mediarails.TTS,
    Model:  "eleven_multilingual_v2",
    Prompt: "Hello world!",
    Config: map[string]any{"voice_id": "21m00Tcm4TlvDq8ikWAM"},
})
// resp.AssetData = audio/mpeg bytes
```

## Install

```bash
go get github.com/promptrails/mediarails
```

## Providers

### Speech

| Provider | Package | Operations | Sync/Async |
|----------|---------|-----------|------------|
| OpenAI | `speech/openai` | TTS, STT (Whisper) | Sync |
| ElevenLabs | `speech/elevenlabs` | TTS | Sync |
| Deepgram | `speech/deepgram` | TTS, STT | Sync |

### Image

| Provider | Package | Operations | Sync/Async |
|----------|---------|-----------|------------|
| OpenAI DALL-E | `image/openai` | Image gen | Sync |
| Stability AI | `image/stability` | Image gen | Sync |
| Fal AI | `image/fal` | Image gen, Video gen | Hybrid |
| Replicate | `image/replicate` | Image gen, Video gen | Async |

### Video

| Provider | Package | Operations | Sync/Async |
|----------|---------|-----------|------------|
| Runway | `video/runway` | Video gen, Img-to-video | Async |
| Pika | `video/pika` | Video gen | Async |
| Luma | `video/luma` | Video gen, Img-to-video | Async |

## Media Types

| Type | Description |
|------|-------------|
| `TTS` | Text-to-speech |
| `STT` | Speech-to-text |
| `ImageGen` | Text-to-image |
| `VideoGen` | Text-to-video |
| `VideoFromImage` | Image-to-video |

## Documentation

| | |
|---|---|
| [Getting Started](docs/getting-started.md) | Installation and quick start |
| [Providers](docs/providers.md) | All providers with config |
| [Async Jobs](docs/async.md) | Polling and WaitForCompletion |

Full docs: [promptrails.github.io/mediarails](https://promptrails.github.io/mediarails)

## Part of the PromptRails AI Toolkit

- [LangRails](https://github.com/promptrails/langrails) — Unified LLM provider
- [GuardRails](https://github.com/promptrails/guardrails) — Content safety scanning
- [MemoryRails](https://github.com/promptrails/memoryrails) — Agent memory
- **MediaRails** — AI media generation

## License

MIT — [PromptRails](https://promptrails.com)
