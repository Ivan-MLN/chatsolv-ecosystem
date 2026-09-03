package whatsapp

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestSessionLifecycleOutlivesConnectRequest(t *testing.T) {
	type contextKey string
	requestCtx, cancelRequest := context.WithCancel(context.WithValue(context.Background(), contextKey("request-id"), "req-1"))
	lifecycleCtx, cancelLifecycle := detachedSessionContext(requestCtx)

	cancelRequest()

	require.NoError(t, lifecycleCtx.Err())
	require.Equal(t, "req-1", lifecycleCtx.Value(contextKey("request-id")))
	cancelLifecycle()
	require.ErrorIs(t, lifecycleCtx.Err(), context.Canceled)
}
