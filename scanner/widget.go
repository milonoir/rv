package scanner

import (
	"context"
	"sort"
	"strconv"
	"sync"

	ui "github.com/gizak/termui/v3"
	"github.com/gizak/termui/v3/widgets"
	"github.com/go-redis/redis/v8"
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
	table.Title = " Scanners "
	table.RowSeparator = false
	w.Table = table

	return w
}

// Update implements the common.Widget interface.
func (w *widget) Update() {
	rows := [][]string{
		{"name", "pattern", "count", "updated"},
	}

	for _, name := range w.order {
		if scr, ok := w.scanners[name]; ok {
			reply, ut := scr.LastReply()
			updated := "-"
			if !ut.IsZero() {
				updated = ut.String()
			}
			rows = append(rows, []string{name, scr.Pattern(), strconv.Itoa(len(reply)), updated})
		}
	}

	w.Rows = rows
	ui.Render(w)
}

// Close implements the common.Widget interface.
func (w *widget) Close() {
	w.cancel()
	w.wg.Wait()
}
