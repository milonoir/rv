package scanner

import (
	"context"
	"log"
	"sync"
	"time"

	"github.com/go-redis/redis/v8"
)

// scanner implements the Scanner interface.
type scanner struct {
	rc      *redis.Client
	cfg     *Config
	ctx     context.Context
	enabled bool
	reply   []string
	updated time.Time
	mtx     sync.Mutex
}

// NewScanner returns a configured scanner.
func NewScanner(ctx context.Context, rc *redis.Client, cfg *Config) Scanner {
	return &scanner{
		rc:      rc,
		cfg:     cfg,
		ctx:     ctx,
		enabled: true,
	}
}

// Run implements the Scanner interface.
func (s *scanner) Run() {
	t := time.NewTicker(s.cfg.Interval.Duration)
	defer t.Stop()

	s.run()
	for {
		select {
		case <-s.ctx.Done():
			// Scanner has been aborted.
			return
		case <-t.C:
			if s.enabled {
				s.run()
			}
		}
	}
}

// run executes the configured Redis scan command and saves its response and time of execution.
func (s *scanner) run() {
	reply := make([]string, 0, 100)

	iter := s.rc.Scan(s.ctx, 0, s.cfg.Pattern, 0).Iterator()
	for iter.Next(s.ctx) {
		reply = append(reply, iter.Val())
	}
	if err := iter.Err(); err != nil {
		// TODO: logging...
		log.Println(err)
	}

	s.mtx.Lock()
	s.reply = reply
	s.updated = time.Now().Local()
	s.mtx.Unlock()
}

// Pattern implements the Scanner interface.
func (s *scanner) Pattern() string {
	return s.cfg.Pattern
}

// LastReply implements the Scanner interface.
func (s *scanner) LastReply() ([]string, time.Time) {
	s.mtx.Lock()
	defer s.mtx.Unlock()

	cpy := make([]string, len(s.reply))
	copy(cpy, s.reply)
	return cpy, s.updated
}

// Enable implements the Scanner interface.
func (s *scanner) Enable() {
	s.enabled = true
}

// Disable implements the Scanner interface.
func (s *scanner) Disable() {
	s.enabled = false
}
