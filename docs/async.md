# Async Jobs

Some providers (Runway, Pika, Luma, Fal, Replicate) return a job ID instead of an immediate result. Use `CheckStatus` or `WaitForCompletion` to poll for the result.

## Manual Polling

```go
resp, _ := provider.Generate(ctx, req)

if resp.Status != mediarails.JobCompleted {
    // Poll until done
    for {
        status, _ := provider.CheckStatus(ctx, resp.JobID)
        if status.Status == mediarails.JobCompleted {
            fmt.Println("Asset:", status.AssetURL)
            break
        }
        if status.Status == mediarails.JobFailed {
            fmt.Println("Failed!")
            break
        }
        time.Sleep(5 * time.Second)
    }
}
```

## WaitForCompletion Helper

```go
resp, _ := provider.Generate(ctx, req)

if resp.Status != mediarails.JobCompleted {
    final, err := mediarails.WaitForCompletion(ctx, provider, resp.JobID,
        2*time.Second,   // initial interval
        30*time.Second,  // max interval (exponential backoff)
        10*time.Minute,  // timeout
    )
    if err != nil {
        log.Fatal(err)
    }
    if final.Status == mediarails.JobCompleted {
        fmt.Println("Asset:", final.AssetURL)
    }
}
```

## Job Status Lifecycle

```
queued → processing → completed
                    → failed
```

- `JobCompleted` — asset is ready (`AssetURL` or `AssetData`)
- `JobProcessing` — still generating
- `JobQueued` — waiting to start
- `JobFailed` — generation failed

## Provider Behavior

| Provider | Generate() Returns | Needs Polling |
|----------|-------------------|---------------|
| ElevenLabs | Completed immediately | No |
| Deepgram | Completed immediately | No |
| Stability | Completed immediately | No |
| Fal | Completed or JobID | Sometimes |
| Replicate | Completed or JobID | Sometimes (Prefer: wait) |
| Runway | Always JobID | Yes |
| Pika | Always JobID | Yes |
| Luma | Always JobID | Yes |

Synchronous providers return `ErrNotAsync` from `CheckStatus`.
