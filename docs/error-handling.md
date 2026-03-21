# Error Handling

## API Errors

All providers return descriptive errors with the provider name, HTTP status code, and response body:

```go
resp, err := provider.Generate(ctx, req)
if err != nil {
    // "elevenlabs: API error (status 401): unauthorized"
    // "stability: API error (status 400): invalid prompt"
    // "replicate: request failed: context deadline exceeded"
    fmt.Println(err)
}
```

## Sync vs Async Errors

**Synchronous providers** (ElevenLabs, Deepgram, Stability, OpenAI) return errors immediately from `Generate()`.

**Async providers** (Runway, Pika, Luma, Fal, Replicate) can fail in two places:

```go
// 1. Submission error (network, auth, bad request)
resp, err := provider.Generate(ctx, req)
if err != nil {
    // Failed to submit job
}

// 2. Job failure (generation failed after submission)
status, err := provider.CheckStatus(ctx, resp.JobID)
if err != nil {
    // Failed to check status (network error)
}
if status.Status == mediarails.JobFailed {
    // Generation failed (e.g., content policy, model error)
    fmt.Println(status.Metadata) // may contain failure_reason
}
```

## ErrNotAsync

Synchronous providers return `ErrNotAsync` from `CheckStatus`:

```go
_, err := elevenlabsProvider.CheckStatus(ctx, "job-id")
if err == mediarails.ErrNotAsync {
    // This provider doesn't support async polling
}
```

## Timeouts

All providers have default timeouts:

| Provider Type | Default Timeout |
|--------------|----------------|
| Speech (sync) | 60 seconds |
| Image (sync) | 120 seconds |
| Video (async) | 600 seconds |

Override with a custom HTTP client:

```go
provider := runway.New("key", runway.WithHTTPClient(&http.Client{
    Timeout: 15 * time.Minute,
}))
```

## Context Cancellation

All providers respect `context.Context`:

```go
ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
defer cancel()

resp, err := provider.Generate(ctx, req)
// Returns context.DeadlineExceeded if timeout is hit
```

## WaitForCompletion Errors

```go
final, err := mediarails.WaitForCompletion(ctx, provider, jobID,
    2*time.Second, 30*time.Second, 10*time.Minute,
)
if err != nil {
    // Possible errors:
    // - "mediarails: job xyz timed out after 10m0s"
    // - context.Canceled (if ctx cancelled)
    // - provider-specific API error
}
if final.Status == mediarails.JobFailed {
    // Job completed but failed
}
```

## Validation Errors

Some providers validate config before making the API call:

```go
// ElevenLabs requires voice_id
_, err := provider.Generate(ctx, &mediarails.GenerateRequest{
    Type:   mediarails.TTS,
    Prompt: "Hello",
    Config: map[string]any{}, // missing voice_id
})
// err: "elevenlabs: voice_id is required in config"

// Deepgram STT requires input
_, err = deepgramProvider.Generate(ctx, &mediarails.GenerateRequest{
    Type: mediarails.STT,
    // missing InputURL and InputData
})
// err: "deepgram: STT requires InputURL or InputData"
```
