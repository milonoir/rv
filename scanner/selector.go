package scanner

import (
	"fmt"
	"sort"
	"sync"

	ui "github.com/gizak/termui/v3"
	"github.com/gizak/termui/v3/widgets"
	r "github.com/milonoir/rv/redis"
)

type selector struct {
	*widgets.List

	width int
	rtype r.DataType
	mtx   sync.Mutex
}

func newSelector() *selector {
	s := &selector{
		List: widgets.NewList(),
	}
	s.SelectedRowStyle = ui.NewStyle(ui.ColorWhite, ui.ColorBlue)

	return s
}

func (s *selector) Update() {
	ui.Render(s)
}

func (s *selector) Resize(x1, y1, x2, y2 int) {
	s.width = x2 - x1
	s.SetRect(x1, y1, x2, y2)
}

func (s *selector) Close() {}

func (s *selector) SetItems(items []string, rtype r.DataType) {
	s.Title = fmt.Sprintf(" Select an item to inspect [%s] ", rtype)
	sort.Strings(items)
	s.mtx.Lock()
	s.rtype = rtype
	s.Rows = items
	s.mtx.Unlock()
}

func (s *selector) Select() (item string, rtype r.DataType) {
	s.mtx.Lock()
	defer s.mtx.Unlock()
	return s.Rows[s.SelectedRow], s.rtype
}
