package scanner

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/go-redis/redis/v8"
	r "github.com/milonoir/rv/redis"
)

// worker implements the Worker interface.
type worker struct {
	*Config

	rc      *redis.Client
	ctx     context.Context
	name    string
	enabled bool
	reply   []string
	updated time.Time
	mtx     sync.Mutex
	err     chan string
}

// newWorker returns a configured worker.
func newWorker(ctx context.Context, rc *redis.Client, name string, cfg *Config) Worker {
	return &worker{
		Config:  cfg,
		rc:      rc,
		ctx:     ctx,
		name:    name,
		enabled: true,
		err:     make(chan string, 1),
	}
}

// Run implements the Worker interface.
func (w *worker) Run() {
	t := time.NewTicker(w.Interval.Duration)
	defer t.Stop()
	defer close(w.err)

	w.run()
	for {
		select {
		case <-w.ctx.Done():
			// Worker has been aborted.
			return
		case <-t.C:
			if w.enabled {
				w.run()
			}
		}
	}
}

// run executes the configured Redis scan command and saves its response and time of execution.
func (w *worker) run() {
	reply := make([]string, 0, 100)

	iter := w.rc.Scan(w.ctx, 0, w.Config.Pattern, 0).Iterator()
	for iter.Next(w.ctx) {
		reply = append(reply, iter.Val())
	}
	if err := iter.Err(); err != nil {
		// Try sending the error.
		select {
		case w.err <- fmt.Sprintf("(worker: %s): %s", w.name, err.Error()):
		default:
		}
	}

	w.mtx.Lock()
	w.reply = reply
	w.updated = time.Now().Local()
	w.mtx.Unlock()
}

// Pattern implements the Worker interface.
func (w *worker) Pattern() (string, r.DataType) {
	return w.Config.Pattern, w.Type
}

// LastReply implements the Worker interface.
func (w *worker) State() ([]string, time.Time, bool) {
	w.mtx.Lock()
	defer w.mtx.Unlock()

	cpy := make([]string, len(w.reply))
	copy(cpy, w.reply)
	return cpy, w.updated, w.enabled
}

// Enable implements the Worker interface.
func (w *worker) Enable() {
	w.mtx.Lock()
	w.enabled = true
	w.mtx.Unlock()
}

// Disable implements the Worker interface.
func (w *worker) Disable() {
	w.mtx.Lock()
	w.enabled = false
	w.mtx.Unlock()
}

// ErrCh implements the Worker interface.
func (w *worker) ErrCh() <-chan string {
	return w.err
}
