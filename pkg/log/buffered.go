package log

import (
	"bufio"
	"io"
	"os"
	"sync"
	"time"

	"github.com/pkg/errors"
)

type writeFlusher interface {
	io.Writer
	Flush() error
}

type MaxLatencyWriter struct {
	dst     writeFlusher
	latency time.Duration // non-zero; negative means to flush immediately

	mu           sync.Mutex // protects t, flushPending, and dst.Flush
	t            *time.Timer
	flushPending bool
}

// GetMaxLatencyWriter returns a maxLatencyWriter that logs to logfile.
// A buffered writer is used so traffic spikes don't result in permanent writes to disk.
func GetMaxLatencyWriter(logfile string) (io.Writer, func(), error) {
	f, err := os.OpenFile(logfile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0666)
	if err != nil {
		return nil, nil, errors.WithStack(err)
	}
	w := NewMaxLatencyWriter(bufio.NewWriterSize(f, 1024), 5*time.Second)
	cleanup := func() {
		// Flush the log file, stop the latency timer.
		w.Stop()
		_ = f.Close()
	}
	return w, cleanup, nil
}

func NewMaxLatencyWriter(dst writeFlusher, latency time.Duration) *MaxLatencyWriter {
	return &MaxLatencyWriter{
		dst:     dst,
		latency: latency,
	}
}

func (m *MaxLatencyWriter) Write(p []byte) (n int, err error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	n, err = m.dst.Write(p)
	if m.latency < 0 {
		m.dst.Flush()
		return
	}
	if m.flushPending {
		return
	}
	if m.t == nil {
		m.t = time.AfterFunc(m.latency, m.delayedFlush)
	} else {
		m.t.Reset(m.latency)
	}
	m.flushPending = true
	return
}

func (m *MaxLatencyWriter) delayedFlush() {
	m.mu.Lock()
	defer m.mu.Unlock()
	if !m.flushPending { // if Stop was called but AfterFunc already started this goroutine
		return
	}
	m.dst.Flush()
	m.flushPending = false
}

func (m *MaxLatencyWriter) Stop() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.flushPending = false
	if m.t != nil {
		m.t.Stop()
	}
}
