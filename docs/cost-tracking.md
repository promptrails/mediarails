# Cost Tracking

Every provider returns `Usage` data with each response, enabling cost calculation and metering.

## Usage Structure

```go
type Usage struct {
    Unit     string  // "characters", "seconds", "images", "gpu_seconds"
    Quantity float64 // Amount consumed
}
```

## Calculating Cost

```go
resp, _ := provider.Generate(ctx, req)

if resp.Usage != nil {
    unitPrice := 0.015 // your price per unit
    cost := resp.Usage.Cost(unitPrice)
    fmt.Printf("Used: %.1f %s → $%.4f\n", resp.Usage.Quantity, resp.Usage.Unit, cost)
}
```

## Metering by Provider

| Provider | Unit | What's Measured |
|----------|------|-----------------|
| **ElevenLabs** | characters | Length of input text |
| **Deepgram TTS** | characters | Length of input text |
| **Deepgram STT** | seconds | Audio duration |
| **OpenAI TTS** | characters | Length of input text |
| **OpenAI DALL-E** | images | 1 per request |
| **Stability** | images | 1 per request |
| **Fal** | images/videos | 1 per request |
| **Replicate** | gpu_seconds | Actual GPU time |
| **Runway** | seconds | Video duration (5 or 10) |
| **Pika** | seconds | Video duration |
| **Luma** | seconds | Default 5 seconds |

## Tracking Over Time

```go
type CostTracker struct {
    mu    sync.Mutex
    total float64
    items []CostItem
}

type CostItem struct {
    Provider string
    Type     string
    Usage    mediarails.Usage
    Cost     float64
    Time     time.Time
}

func (t *CostTracker) Track(provider string, resp *mediarails.GenerateResponse, unitPrice float64) {
    if resp.Usage == nil {
        return
    }
    t.mu.Lock()
    defer t.mu.Unlock()

    cost := resp.Usage.Cost(unitPrice)
    t.total += cost
    t.items = append(t.items, CostItem{
        Provider: provider,
        Type:     resp.Usage.Unit,
        Usage:    *resp.Usage,
        Cost:     cost,
        Time:     time.Now(),
    })
}

func (t *CostTracker) Total() float64 {
    t.mu.Lock()
    defer t.mu.Unlock()
    return t.total
}
```

## Pricing Reference (approximate)

These are rough estimates — check provider pricing pages for current rates:

| Provider | Model | Approx. Price |
|----------|-------|---------------|
| ElevenLabs | Multilingual v2 | $0.30 / 1K chars |
| Deepgram | Nova-2 STT | $0.0043 / min |
| OpenAI | TTS-1 | $0.015 / 1K chars |
| OpenAI | DALL-E 3 1024x1024 | $0.040 / image |
| Stability | SD Core | $0.03 / image |
| Replicate | Flux Schnell | ~$0.003 / image |
| Runway | Gen-3 Alpha | $0.05 / second |
| Luma | Dream Machine | $0.032 / second |
