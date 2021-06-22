package scanner

import (
	"context"
	"fmt"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	ui "github.com/gizak/termui/v3"
	"github.com/gizak/termui/v3/widgets"
	"github.com/go-redis/redis/v8"
)

var (
	ageNew    = 30 * time.Second
	ageMedium = 5 * time.Minute
)

// widget manages a set of Scanners and renders their output to a widgets.Table.
type widget struct {
	*widgets.Table

	scanners map[string]Scanner
	order    []string
	wg       sync.WaitGroup
	cancel   context.CancelFunc
}

// NewWidget returns a fully configured scanner widget.
func NewWidget(ctx context.Context, rc *redis.Client, scans map[string]*Config) *widget {
	ctx, cancel := context.WithCancel(ctx)

	w := &widget{
		order:    make([]string, 0, len(scans)),
		scanners: make(map[string]Scanner, len(scans)),
		cancel:   cancel,
	}

	for name, scan := range scans {
		s := NewScanner(ctx, rc, scan)
		w.scanners[name] = s
		w.wg.Add(1)
		go func() {
			s.Run()
			defer w.wg.Done()
		}()
		w.order = append(w.order, name)
	}
	sort.Strings(w.order)

	table := widgets.NewTable()
	width, height := ui.TerminalDimensions()
	table.SetRect(0, 0, width, height)
	table.RowSeparator = false
	w.Table = table

	return w
}

// Update implements the common.Widget interface.
func (w *widget) Update() {
	rows := [][]string{
		{"Name", "Pattern", "Count", "Last scanned"},
	}

	now := time.Now()
	for _, name := range w.order {
		if scr, ok := w.scanners[name]; ok {
			rows = append(rows, w.renderRow(name, scr, now))
		}
	}
	w.Title = fmt.Sprintf(" Scanners [%d] ", len(w.order))
	w.Rows = rows
	ui.Render(w)
}

func (w *widget) renderRow(name string, scr Scanner, now time.Time) []string {
	reply, ut, enabled := scr.State()

	return []string{
		w.renderName(name, enabled),
		w.renderPattern(scr.Pattern()),
		strconv.Itoa(len(reply)),
		w.renderUpdated(ut, now),
	}
}

func (w *widget) renderName(name string, enabled bool) string {
	if enabled {
		return fmt.Sprintf("[%s](fg:green)", name)
	}
	return fmt.Sprintf("[%s](fg:red)", name)
}

func (w *widget) renderPattern(pattern string) string {
	return strings.ReplaceAll(pattern, "*", "[*](fg:yellow)")
}

func (w *widget) renderUpdated(updated, now time.Time) string {
	age := now.Sub(updated).Round(time.Second)
	switch {
	case updated.IsZero():
		return "-"
	case age <= ageNew:
		return fmt.Sprintf("[%s](fg:green) ago", age)
	case age <= ageMedium:
		return fmt.Sprintf("[%s](fg:yellow) ago", age)
	default:
		return fmt.Sprintf("[%s](fg:red) ago", age)
	}
}

// Close implements the common.Widget interface.
func (w *widget) Close() {
	w.cancel()
	w.wg.Wait()
}
