package signals_test

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/your-org/simple-cli/internal/signals"
)

func TestNotifyContextCancelPropagates(t *testing.T) {
	parent, parentCancel := context.WithCancel(context.Background())
	defer parentCancel()

	ctx, stop := signals.NotifyContext(parent)
	defer stop()

	// Cancel the parent — child should cancel too.
	parentCancel()

	select {
	case <-ctx.Done():
		// expected
	case <-time.After(500 * time.Millisecond):
		t.Fatal("context was not cancelled within timeout")
	}
}

func TestNotifyContextStopCancels(t *testing.T) {
	ctx, stop := signals.NotifyContext(context.Background())

	// Calling stop() should cancel the context.
	stop()

	select {
	case <-ctx.Done():
		assert.Error(t, ctx.Err())
	case <-time.After(500 * time.Millisecond):
		t.Fatal("context not cancelled after stop()")
	}
}
