package outbound

import (
	"context"
	"sync/atomic"
	"testing"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

// The ticker that drives handleOutbound must keep firing after Start
// returns. Regression test: defer ticker.Stop() used to live in Start()
// instead of the goroutine, stopping the ticker immediately and leaving
// the loop blocked forever on ticker.C.
func TestOutboundHandler_Start_TickerKeepsFiringAfterStartReturns(t *testing.T) {
	logger := zap.NewNop()

	var loads atomic.Int32
	var updates atomic.Int32

	h := NewOutboundHandler(
		logger,
		nil, // producer unused: loader returns no events
		OutboundCfg{Interval: "20ms"},
		func(ctx context.Context, eventIds []uuid.UUID) error {
			updates.Add(1)
			return nil
		},
		func(ctx context.Context) ([]OutboundEvent, error) {
			loads.Add(1)
			return nil, nil
		},
	)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	if err := h.Start(ctx); err != nil {
		t.Fatalf("Start failed: %v", err)
	}

	// Give the stopped-ticker bug ~50 ticks to manifest. With the bug,
	// loads stays 0 forever and this times out.
	deadline := time.Now().Add(2 * time.Second)
	for loads.Load() == 0 && time.Now().Before(deadline) {
		time.Sleep(10 * time.Millisecond)
	}

	if loads.Load() == 0 {
		t.Fatal("ticker never fired within 2s — handleOutbound loop is dead " +
			"(ticker stopped when Start returned?)")
	}
	if updates.Load() != 0 {
		t.Fatalf("updater called with no pending events: %d", updates.Load())
	}

	cancel()
}
