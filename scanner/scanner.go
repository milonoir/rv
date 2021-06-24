package scanner

import (
	"context"
	"fmt"
	"sort"
	"strconv"
	"sync"
	"time"

	ui "github.com/gizak/termui/v3"
	"github.com/gizak/termui/v3/widgets"
	"github.com/go-redis/redis/v8"
)

const (
	countWidth = 7
	ageWidth   = 10
)

var (
	ageNew    = 30 * time.Second
	ageMedium = 1 * time.Minute
)

// scanner manages a set of workers and renders their output into a widgets.Table.
type scanner struct {
	*widgets.List

	workers  map[string]Worker
	order    []string
	wg       sync.WaitGroup
	cancel   context.CancelFunc
	width    int
	messages chan string
}

// NewScanner returns a fully configured scanner.
func NewScanner(ctx context.Context, rc *redis.Client, configs map[string]*Config) Scanner {
	ctx, cancel := context.WithCancel(ctx)

	cn := len(configs)
	s := &scanner{
		order:    make([]string, 0, cn),
		workers:  make(map[string]Worker, cn),
		cancel:   cancel,
		messages: make(chan string, cn),
	}

	for name, cfg := range configs {
		w := newWorker(ctx, rc, name, cfg)
		s.workers[name] = w
		s.wg.Add(2)
		// Main worker goroutine.
		go func() {
			defer s.wg.Done()
			w.Run()
		}()
		// Worker messages fan-in goroutine.
		go func() {
			defer s.wg.Done()
			for {
				select {
				case <-ctx.Done():
					return
				case m := <-w.ErrCh():
					select {
					case s.messages <- m:
					default:
					}
				}
			}
		}()
		s.order = append(s.order, name)
	}
	sort.Strings(s.order)

	s.List = widgets.NewList()
	s.SelectedRowStyle = ui.NewStyle(ui.ColorWhite, ui.ColorBlue)

	return s
}

// Resize implements the common.Widget interface.
func (s *scanner) Resize(x1, y1, x2, y2 int) {
	s.width = x2 - x1
	s.SetRect(x1, y1, x2, y2)
}

// Update implements the common.Widget interface.
func (s *scanner) Update() {
	n := len(s.order)
	rows := make([]string, 0, n)
	now := time.Now()
	cws := s.columnWidths()
	for _, name := range s.order {
		if w, ok := s.workers[name]; ok {
			rows = append(rows, s.renderRow(name, w, now, cws))
		}
	}
	s.Title = fmt.Sprintf(" Scanners [%d] ", n)
	s.Rows = rows
	ui.Render(s)
}

func (s *scanner) columnWidths() (v [2]int) {
	width := s.width - countWidth - ageWidth - 5 // 5 = borders and separators
	v[0] = width / 3                             // 3 = 1/3 of the remaining space
	v[1] = width - v[0]
	return
}

func (s *scanner) renderRow(name string, w Worker, now time.Time, width [2]int) string {
	reply, ut, enabled := w.State()

	return fmt.Sprintf(
		"%s %s %*s %s",
		s.renderName(name, enabled, width[0]),
		s.renderPattern(w, width[1]),
		countWidth, strconv.Itoa(len(reply)),
		s.renderUpdated(ut, now),
	)
}

func (s *scanner) renderName(name string, enabled bool, length int) string {
	if len(name) > length {
		name = name[:length]
	}
	if enabled {
		return fmt.Sprintf("[%*s](fg:green)", -length, name)
	}
	return fmt.Sprintf("[%*s](fg:red)", -length, name)
}

func (s *scanner) renderPattern(w Worker, length int) string {
	pattern, _ := w.Pattern()
	p := pattern
	if len(pattern) > length {
		p = pattern[:length]
	}
	if w.IsSingle() {
		return fmt.Sprintf("[%*s](fg:cyan)", -length, p)
	}
	return fmt.Sprintf("%*s", -length, p)
}

func (s *scanner) renderUpdated(updated, now time.Time) string {
	age := now.Sub(updated).Round(time.Second)
	ageStr, color := age.String(), "red"
	switch {
	case updated.IsZero():
		ageStr, color = "n/a", "white"
	case age <= ageNew:
		color = "green"
	case age <= ageMedium:
		color = "yellow"
	default:
		color = "red"
	}
	return fmt.Sprintf("[%*s](fg:%s)", ageWidth, ageStr, color)
}

// Close implements the common.Widget interface.
// Aborts all workers and waits for them to return.
func (s *scanner) Close() {
	s.cancel()
	s.wg.Wait()
}

// Select implements the Scanner interface.
func (s *scanner) Select() string {
	if _, w := s.selectWorker(); w != nil {
		p, t := w.Pattern()
		return fmt.Sprintf("(%s) %s", t, p)
	}
	return ""
}

// Enable implements the Scanner interface.
func (s *scanner) Enable() {
	if name, w := s.selectWorker(); w != nil {
		w.Enable()
		s.messages <- fmt.Sprintf("[enabled](fg:green) worker %q", name)
	}
}

// Disable implements the Scanner interface.
func (s *scanner) Disable() {
	if name, w := s.selectWorker(); w != nil {
		w.Disable()
		s.messages <- fmt.Sprintf("[disabled](fg:red) worker %q", name)
	}
}

func (s *scanner) selectWorker() (string, Worker) {
	name := s.order[s.List.SelectedRow]
	if w, ok := s.workers[name]; ok {
		return name, w
	}
	return "", nil
}

// Messages implements the Scanner interface.
func (s *scanner) Messages() <-chan string {
	return s.messages
}
