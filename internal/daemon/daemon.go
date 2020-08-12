package daemon

import (
	"context"
	"sync"
	"time"

	"go-webapp-example/pkg/clock"
	"go-webapp-example/pkg/log"

	"github.com/cenkalti/backoff/v3"
)

// prefix is used to prefix daemon logs.
const prefix = "dmn"

// Daemon defines all required methods of a background daemon process.
type Daemon interface {
	Name() string
	Run(ctx context.Context, wg *sync.WaitGroup, logger log.Logger) error
}

// Manager takes care of starting and orchestrating daemon processes.
type Manager struct {
	log log.Logger
	// wg is used to keep track of running daemons.
	wg *sync.WaitGroup
	// ctx is used to signal cancellation to running daemons.
	ctx context.Context
}

// NewManager returns a new manager instance.
func NewManager(ctx context.Context, wg *sync.WaitGroup, l log.Logger) *Manager {
	return &Manager{
		ctx: ctx,
		wg:  wg,
		log: l,
	}
}

// Start starts a daemon.
func (m *Manager) Start(d Daemon) {
	m.wg.Add(1)
	defer m.wg.Done()

	var wg sync.WaitGroup
	var try int

	logger := m.log.WithPrefix(prefix + "." + d.Name())
	go m.recoverPanic(d, logger)

	ticker := backoff.NewTicker(newExponentialBackOff())
	for range ticker.C {
		select {
		// If the daemon should stop, wait for the wg to be done, then
		// stop the restart ticket and exit the method.
		case <-m.ctx.Done():
			wg.Wait()
			logger.Infof(`daemon "%s" shutdown complete`, d.Name())
			ticker.Stop()
			return
		default:
			try++
			logger.Tracef(`starting daemon "%s" (%d. try)...`, d.Name(), try)
			// Run the daemon. If it crashes, continue to the next ticker iteration.
			if err := d.Run(m.ctx, &wg, logger); err != nil {
				logger.Warnf(`daemon "%s" crashed: %s`, d.Name(), err)
				continue
			}
			// If the Run method returned without an error, reset the try counter
			// and restart the daemon again. All daemons are run forever, even
			// if the return without an error.
			logger.Debugf(`daemon "%s" exited without errors`, d.Name())
			try = 0
		}
	}
}

// recoverPanic recovers a crashed daemon and restarts it.
func (m *Manager) recoverPanic(d Daemon, logger log.Logger) {
	if err := recover(); err != nil {
		logger.Errorf("daemon exited with panic (restarting in 5 seconds): %s", err)
		time.Sleep(5 * time.Second)
		m.Start(d)
	}
}

// newExponentialBackOff makes sure that the timeout for
// restarting crashed daemons gets exponentially longer.
func newExponentialBackOff() backoff.BackOff {
	b := &backoff.ExponentialBackOff{
		InitialInterval:     500 * time.Millisecond,
		RandomizationFactor: 0.5,
		Multiplier:          1.25,
		MaxInterval:         30 * time.Second,
		MaxElapsedTime:      0,
		Clock:               clock.FromTime(time.Now()),
	}
	b.Reset()
	return b
}
