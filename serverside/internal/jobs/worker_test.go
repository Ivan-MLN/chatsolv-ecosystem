package jobs

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

type fakeQueue struct {
	event     Event
	claimed   bool
	completed bool
	retried   bool
	dead      bool
}

func (f *fakeQueue) Claim(context.Context) (Event, error) {
	if f.claimed {
		return Event{}, ErrNoJob
	}
	f.claimed = true
	return f.event, nil
}
func (f *fakeQueue) Complete(context.Context, string) error { f.completed = true; return nil }
func (f *fakeQueue) Retry(context.Context, string, time.Time, string) error {
	f.retried = true
	return nil
}
func (f *fakeQueue) Dead(context.Context, Event, string) error { f.dead = true; return nil }
func TestWorkerCompletesProvisionEvent(t *testing.T) {
	q := &fakeQueue{event: Event{ID: "e", Type: "workspace.provision", AggregateID: "w", Attempts: 1}}
	called := false
	w := NewWorker(q, map[string]Handler{"workspace.provision": func(context.Context, string) error { called = true; return nil }}, 3, time.Now)
	require.NoError(t, w.RunOnce(context.Background()))
	require.True(t, called)
	require.True(t, q.completed)
}
func TestWorkerRetriesTransientFailureThenDeadLetters(t *testing.T) {
	q := &fakeQueue{event: Event{ID: "e", Type: "workspace.provision", AggregateID: "w", Attempts: 1}}
	w := NewWorker(q, map[string]Handler{"workspace.provision": func(context.Context, string) error { return errors.New("temporary") }}, 3, time.Now)
	require.NoError(t, w.RunOnce(context.Background()))
	require.True(t, q.retried)
	q = &fakeQueue{event: Event{ID: "e", Type: "workspace.provision", AggregateID: "w", Attempts: 3}}
	w = NewWorker(q, map[string]Handler{"workspace.provision": func(context.Context, string) error { return errors.New("temporary") }}, 3, time.Now)
	require.NoError(t, w.RunOnce(context.Background()))
	require.True(t, q.dead)
}
