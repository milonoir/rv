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

	workers map[string]Worker
	order   []string
	wg      sync.WaitGroup
	cancel  context.CancelFunc
	width   int
}

// NewScanner returns a fully configured scanner.
func NewScanner(ctx context.Context, rc *redis.Client, configs map[string]*Config) Scanner {
	ctx, cancel := context.WithCancel(ctx)

	s := &scanner{
		order:   make([]string, 0, len(configs)),
		workers: make(map[string]Worker, len(configs)),
		cancel:  cancel,
	}

	for name, cfg := range configs {
		w := newWorker(ctx, rc, cfg)
		s.workers[name] = w
		s.wg.Add(1)
		go func() {
			w.Run()
			defer s.wg.Done()
		}()
		s.order = append(s.order, name)
	}
	sort.Strings(s.order)

	s.List = widgets.NewList()
	s.SelectedRowStyle = ui.NewStyle(ui.ColorWhite, ui.ColorBlue)
	s.Resize(ui.TerminalDimensions())

	return s
}

// Resize implements the common.Widget interface.
func (s *scanner) Resize(width, height int) {
	s.width = width
	s.SetRect(0, 0, width, height-3)
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
	s.Title = fmt.Sprintf("Scanners [%d]", n)
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
func (s *scanner) Close() {
	s.cancel()
	s.wg.Wait()
}

// Select implements the Scanner interface.
func (s *scanner) Select() string {
	if w := s.selectWorker(); w != nil {
		p, t := w.Pattern()
		return fmt.Sprintf("(%s) %s", t, p)
	}
	return ""
}

// Enable implements the Scanner interface.
func (s *scanner) Enable() {
	if w := s.selectWorker(); s != nil {
		w.Enable()
	}
}

// Disable implements the Scanner interface.
func (s *scanner) Disable() {
	if w := s.selectWorker(); s != nil {
		w.Disable()
	}
}

func (s *scanner) selectWorker() Worker {
	name := s.order[s.List.SelectedRow]
	if w, ok := s.workers[name]; ok {
		return w
	}
	return nil
}
