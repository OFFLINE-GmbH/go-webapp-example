package daemon

import (
	"context"
	"sync"
	"time"

	"go-webapp-example/pkg/log"
)

// Example daemon
type Example struct {
}

// NewExample returns a new example daemon.
func NewExampleDaemon() *Example {
	return &Example{}
}

func (g *Example) Name() string { return "example" }

// Example daemon logs to the console every few seconds.
func (g *Example) Run(ctx context.Context, wg *sync.WaitGroup, logger log.Logger) error {
	wg.Add(1)
	defer wg.Done()

	for {
		select {
		case <-time.After(10 * time.Second):
			logger.Info("example daemon is running... Doing some work...")
		case <-ctx.Done():
			logger.Info("example daemon is shutting down...")
			return nil
		}
	}
}
