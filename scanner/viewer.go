package scanner

import (
	"context"
	"fmt"

	ui "github.com/gizak/termui/v3"
	"github.com/gizak/termui/v3/widgets"
	"github.com/go-redis/redis/v8"
	r "github.com/milonoir/rv/redis"
)

type viewer struct {
	*widgets.List

	executor Executor
	err      chan string
}

func NewViewer(rc *redis.Client) *viewer {
	v := &viewer{
		List:     widgets.NewList(),
		executor: newExecutor(rc),
		err:      make(chan string, 1),
	}
	v.Title = " Details "

	return v
}

// Update implements the common.Widget interface.
func (v *viewer) Update() {
	ui.Render(v)
}

// Resize implements the common.Widget interface.
func (v *viewer) Resize(x1, y1, x2, y2 int) {
	v.SetRect(x1, y1, x2, y2)
}

// Close implements the common.Widget interface.
func (v *viewer) Close() {}

// View implements the Viewer interface.
func (v *viewer) View(ctx context.Context, key string, rt r.DataType) {
	ret, err := v.executor.Execute(ctx, key, rt)
	if err != nil {
		v.sendErr(err.Error())
		return
	}
	switch rt {
	case r.TypeSortedSet:
		data, ok := ret.(map[string]float64)
		if !ok {
			v.sendErr(fmt.Sprintf("executor: zset data error: %v", ret))
			return
		}
		v.renderSortedSet(data)
	case r.TypeHash:
		data, ok := ret.(map[string]string)
		if !ok {
			v.sendErr(fmt.Sprintf("executor: hash data error: %v", ret))
			return
		}
		v.renderHash(data)
	default:
		data, ok := ret.([]string)
		if !ok {
			v.sendErr(fmt.Sprintf("executor: %s data error: %v", rt, ret))
			return
		}
		v.renderStrings(data, rt)
	}
}

func (v *viewer) renderStrings(data []string, rt r.DataType) {
	v.Rows = data
}

func (v *viewer) renderSortedSet(data map[string]float64) {
	v.Rows = make([]string, 0, len(data))
	for member, score := range data {
		v.Rows = append(v.Rows, fmt.Sprintf("member: %s, score: %f", member, score))
	}
}

func (v *viewer) renderHash(data map[string]string) {
	v.Rows = make([]string, 0, len(data))
	for field, value := range data {
		v.Rows = append(v.Rows, fmt.Sprintf("key: %s, value: %s", field, value))
	}
}

func (v *viewer) sendErr(err string) {
	select {
	case v.err <- err:
	default:
	}
}

// Messages implements the common.Messenger interface.
func (v *viewer) Messages() <-chan string {
	return v.err
}
