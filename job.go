package mediarails

import (
	"context"
	"fmt"
	"time"
)

// JobStatus represents the lifecycle state of an async generation job.
type JobStatus string

const (
	// JobCompleted means the asset is ready.
	JobCompleted JobStatus = "completed"

	// JobProcessing means generation is in progress.
	JobProcessing JobStatus = "processing"

	// JobQueued means the job is waiting to start.
	JobQueued JobStatus = "queued"

	// JobFailed means generation failed.
	JobFailed JobStatus = "failed"
)

// IsTerminal returns true if the job is in a final state.
func (s JobStatus) IsTerminal() bool {
	return s == JobCompleted || s == JobFailed
}

// ErrNotAsync is returned by synchronous providers when CheckStatus is called.
var ErrNotAsync = fmt.Errorf("provider does not support async job polling")

// WaitForCompletion polls CheckStatus until the job completes or fails.
// It uses exponential backoff starting at interval, doubling up to maxInterval.
func WaitForCompletion(
	ctx context.Context,
	provider Provider,
	jobID string,
	interval time.Duration,
	maxInterval time.Duration,
	timeout time.Duration,
) (*GenerateResponse, error) {
	deadline := time.After(timeout)
	current := interval

	for {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-deadline:
			return nil, fmt.Errorf("mediarails: job %s timed out after %s", jobID, timeout)
		default:
		}

		resp, err := provider.CheckStatus(ctx, jobID)
		if err != nil {
			return nil, err
		}

		if resp.Status == JobCompleted || resp.Status == JobFailed {
			return resp, nil
		}

		time.Sleep(current)
		current *= 2
		if current > maxInterval {
			current = maxInterval
		}
	}
}
