package mediarails

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"
)

// scriptedProvider returns a predefined sequence of statuses from CheckStatus,
// repeating the last one once exhausted. A non-nil err is returned instead.
type scriptedProvider struct {
	statuses []JobStatus
	calls    int
	err      error
}

func (p *scriptedProvider) ID() string                  { return "mock" }
func (p *scriptedProvider) SupportedTypes() []MediaType { return []MediaType{VideoGen} }

func (p *scriptedProvider) Generate(context.Context, *GenerateRequest) (*GenerateResponse, error) {
	return nil, nil
}

func (p *scriptedProvider) CheckStatus(_ context.Context, jobID string) (*GenerateResponse, error) {
	p.calls++
	if p.err != nil {
		return nil, p.err
	}
	idx := p.calls - 1
	if idx >= len(p.statuses) {
		idx = len(p.statuses) - 1
	}
	return &GenerateResponse{JobID: jobID, Status: p.statuses[idx]}, nil
}

func TestWaitForCompletion_Completes(t *testing.T) {
	p := &scriptedProvider{statuses: []JobStatus{JobQueued, JobProcessing, JobCompleted}}

	resp, err := WaitForCompletion(context.Background(), p, "job-1",
		time.Millisecond, 2*time.Millisecond, time.Second)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.Status != JobCompleted {
		t.Errorf("Status = %q, want completed", resp.Status)
	}
	if p.calls != 3 {
		t.Errorf("expected 3 polls, got %d", p.calls)
	}
}

// A failed job is a normal (non-error) terminal return — the caller inspects Status.
func TestWaitForCompletion_Fails(t *testing.T) {
	p := &scriptedProvider{statuses: []JobStatus{JobProcessing, JobFailed}}

	resp, err := WaitForCompletion(context.Background(), p, "job-1",
		time.Millisecond, 2*time.Millisecond, time.Second)
	if err != nil {
		t.Fatalf("a failed job should not produce an error, got: %v", err)
	}
	if resp.Status != JobFailed {
		t.Errorf("Status = %q, want failed", resp.Status)
	}
}

func TestWaitForCompletion_CheckStatusError(t *testing.T) {
	wantErr := errors.New("api exploded")
	p := &scriptedProvider{err: wantErr}

	_, err := WaitForCompletion(context.Background(), p, "job-1",
		time.Millisecond, 2*time.Millisecond, time.Second)
	if !errors.Is(err, wantErr) {
		t.Errorf("err = %v, want %v", err, wantErr)
	}
}

func TestWaitForCompletion_ContextCanceled(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // already canceled

	p := &scriptedProvider{statuses: []JobStatus{JobProcessing}}
	_, err := WaitForCompletion(ctx, p, "job-1",
		time.Millisecond, 2*time.Millisecond, time.Second)
	if !errors.Is(err, context.Canceled) {
		t.Errorf("err = %v, want context.Canceled", err)
	}
}

func TestWaitForCompletion_Timeout(t *testing.T) {
	// Never terminal + interval >= timeout → the deadline fires on a later poll.
	p := &scriptedProvider{statuses: []JobStatus{JobProcessing}}
	_, err := WaitForCompletion(context.Background(), p, "job-1",
		5*time.Millisecond, 5*time.Millisecond, time.Millisecond)
	if err == nil || !strings.Contains(err.Error(), "timed out") {
		t.Errorf("err = %v, want a timeout error", err)
	}
}
