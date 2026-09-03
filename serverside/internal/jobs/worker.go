package jobs

import (
	"context"
	"errors"
	"math"
	"math/rand/v2"
	"time"
)

var ErrNoJob = errors.New("no job available")

type Event struct {
	ID, WorkspaceID, Type, AggregateID string
	Attempts                           int
}
type Handler func(context.Context, string) error
type Queue interface {
	Claim(context.Context) (Event, error)
	Complete(context.Context, string) error
	Retry(context.Context, string, time.Time, string) error
	Dead(context.Context, Event, string) error
}
type Worker struct {
	queue       Queue
	handlers    map[string]Handler
	maxAttempts int
	now         func() time.Time
}

func NewWorker(queue Queue, handlers map[string]Handler, maxAttempts int, now func() time.Time) *Worker {
	return &Worker{queue, handlers, maxAttempts, now}
}
func (w *Worker) RunOnce(ctx context.Context) error {
	event, err := w.queue.Claim(ctx)
	if errors.Is(err, ErrNoJob) {
		return nil
	}
	if err != nil {
		return err
	}
	handler, ok := w.handlers[event.Type]
	if !ok {
		return w.queue.Dead(ctx, event, "UNKNOWN_JOB_TYPE")
	}
	if err = handler(ctx, event.AggregateID); err == nil {
		return w.queue.Complete(ctx, event.ID)
	}
	println("JOB HANDLER ERROR:", event.Type, err.Error())
	if event.Attempts >= w.maxAttempts {
		return w.queue.Dead(ctx, event, "JOB_FAILED")
	}
	seconds := math.Pow(2, float64(event.Attempts))
	jitter := time.Duration(rand.IntN(1000)) * time.Millisecond
	return w.queue.Retry(ctx, event.ID, w.now().Add(time.Duration(seconds)*time.Second+jitter), "TRANSIENT_FAILURE")
}
func (w *Worker) Run(ctx context.Context, poll time.Duration) error {
	ticker := time.NewTicker(poll)
	defer ticker.Stop()
	for {
		if err := w.RunOnce(ctx); err != nil {
			return err
		}
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
		}
	}
}
