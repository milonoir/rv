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

const (
	countWidth = 7
	ageWidth   = 10
)

var (
	ageNew    = 30 * time.Second
	ageMedium = 1 * time.Minute
)

// widget manages a set of Scanners and renders their output to a widgets.Table.
type widget struct {
	*widgets.List

	scanners map[string]Scanner
	order    []string
	wg       sync.WaitGroup
	cancel   context.CancelFunc
	width    int
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

	w.List = widgets.NewList()
	w.SelectedRowStyle = ui.NewStyle(ui.ColorWhite, ui.ColorBlue)
	w.Resize(ui.TerminalDimensions())

	return w
}

// Resize implements the common.Widget interface.
func (w *widget) Resize(width, height int) {
	w.width = width
	w.SetRect(0, 0, width, height)
}

// Update implements the common.Widget interface.
func (w *widget) Update() {
	n := len(w.order)
	rows := make([]string, 0, n)
	now := time.Now()
	cws := w.columnWidths()
	for _, name := range w.order {
		if scr, ok := w.scanners[name]; ok {
			rows = append(rows, w.renderRow(name, scr, now, cws))
		}
	}
	w.Title = fmt.Sprintf("Scanners [%d]", n)
	w.Rows = rows
	ui.Render(w)
}

func (w *widget) columnWidths() (v [2]int) {
	width := w.width - countWidth - ageWidth - 5 // 5 = borders and separators
	v[0] = width / 3                             // 3 = 1/3 of the remaining space
	v[1] = width - v[0]
	return
}

func (w *widget) renderRow(name string, scr Scanner, now time.Time, width [2]int) string {
	reply, ut, enabled := scr.State()

	return fmt.Sprintf(
		"%s %s %*s %s",
		w.renderName(name, enabled, width[0]),
		w.renderPattern(scr.Pattern(), width[1]),
		countWidth, strconv.Itoa(len(reply)),
		w.renderUpdated(ut, now),
	)
}

func (w *widget) renderName(name string, enabled bool, length int) string {
	if len(name) > length {
		name = name[:length]
	}
	if enabled {
		return fmt.Sprintf("[%*s](fg:green)", -length, name)
	}
	return fmt.Sprintf("[%*s](fg:red)", -length, name)
}

func (w *widget) renderPattern(pattern string, length int) string {
	p := pattern
	if len(pattern) > length {
		p = pattern[:length]
	}
	if !strings.Contains(pattern, "*") {
		return fmt.Sprintf("[%*s](fg:cyan)", -length, p)
	}
	return fmt.Sprintf("%*s", -length, p)
}

func (w *widget) renderUpdated(updated, now time.Time) string {
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
func (w *widget) Close() {
	w.cancel()
	w.wg.Wait()
}
