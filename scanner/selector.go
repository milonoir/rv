package scanner

import (
	"fmt"
	"sort"
	"strings"
	"sync"

	ui "github.com/gizak/termui/v3"
	"github.com/gizak/termui/v3/widgets"
	r "github.com/milonoir/rv/redis"
)

const (
	typeWidth = 6
)

type selector struct {
	*widgets.List

	items     []string
	itemWidth int
	rtype     r.DataType
	mtx       sync.Mutex
}

func NewSelector() *selector {
	s := &selector{
		List: widgets.NewList(),
	}
	s.Title = " Select an item to inspect "
	s.SelectedRowStyle = ui.NewStyle(ui.ColorWhite, ui.ColorBlue)

	return s
}

func (s *selector) Update() {
	ui.Render(s)
}

func (s *selector) Resize(x1, y1, x2, y2 int) {
	s.itemWidth = x2 - x1 - typeWidth - 3 // 3 = borders + separator
	s.SetRect(x1, y1, x2, y2)
}

func (s *selector) Close() {}

func (s *selector) SetItems(items []string, rtype r.DataType) {
	s.mtx.Lock()
	defer s.mtx.Unlock()

	s.SelectedRow = 0
	sort.Strings(items)
	s.items = items
	s.rtype = rtype
	rt := s.renderType()
	s.Rows = make([]string, len(items))
	for i, item := range items {
		s.Rows[i] = s.renderRow(item, rt)
	}
}

func (s *selector) renderRow(item, rt string) string {
	return fmt.Sprintf("%s %*s", rt, -s.itemWidth, item)
}

func (s *selector) renderType() string {
	return fmt.Sprintf("[%*s](fg:green)", -typeWidth, strings.ToUpper(string(s.rtype)))
}

func (s *selector) Select() (item string, rtype r.DataType) {
	s.mtx.Lock()
	defer s.mtx.Unlock()
	return s.items[s.SelectedRow], s.rtype
}
