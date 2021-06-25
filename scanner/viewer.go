package scanner

import (
	"context"
	"fmt"
	"strings"

	ui "github.com/gizak/termui/v3"
	"github.com/gizak/termui/v3/widgets"
	"github.com/go-redis/redis/v8"
	r "github.com/milonoir/rv/redis"
)

const (
	headerTemplate = `[Type](fg:cyan): %s   [Key](fg:cyan): %s`
)

var (
	keyRenderTemplate = []string{
		" " + headerTemplate,
		"[Value](fg:cyan): %s",
	}
	listRenderTemplate = []string{
		" " + headerTemplate + "   [Length](fg:cyan): %d",
		"[Items](fg:cyan):",
	}
	setRenderTemplate = []string{
		"   " + headerTemplate + "   [Length](fg:cyan): %d",
		"[Members](fg:cyan):",
	}
	zsetRenderTemplate = []string{
		"   " + headerTemplate + "   [Length](fg:cyan): %d",
		"[Members](fg:cyan):",
	}
	hashRenderTemplate = []string{
		"  " + headerTemplate + "   [Length](fg:cyan): %d",
		"[Fields](fg:cyan):",
	}
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
	v.SelectedRowStyle = ui.NewStyle(ui.ColorWhite, ui.ColorBlue)

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
		data, ok := ret.([]redis.Z)
		if !ok {
			v.sendErr(fmt.Sprintf("executor: zset data error: %v", ret))
			return
		}
		v.renderSortedSet(key, data)
	case r.TypeHash:
		data, ok := ret.(map[string]string)
		if !ok {
			v.sendErr(fmt.Sprintf("executor: hash data error: %v", ret))
			return
		}
		v.renderHash(key, data)
	default:
		data, ok := ret.([]string)
		if !ok {
			v.sendErr(fmt.Sprintf("executor: %s data error: %v", rt, ret))
			return
		}
		v.renderStrings(key, data, rt)
	}
}

func (v *viewer) renderStrings(key string, data []string, rt r.DataType) {
	switch rt {
	case r.TypeList:
		v.renderList(key, data)
	case r.TypeSet:
		v.renderSet(key, data)
	default:
		v.renderKey(key, data[0])
	}
}

func (v *viewer) renderKey(key, data string) {
	v.Rows = []string{
		fmt.Sprintf(keyRenderTemplate[0], strings.ToUpper(string(r.TypeKey)), key),
		fmt.Sprintf(keyRenderTemplate[1], data),
	}
}

func (v *viewer) renderList(key string, data []string) {
	l := len(data)
	v.Rows = make([]string, l+2)
	v.Rows[0] = fmt.Sprintf(listRenderTemplate[0], strings.ToUpper(string(r.TypeList)), key, l)
	v.Rows[1] = listRenderTemplate[1]
	for i := range data {
		v.Rows[i+2] = fmt.Sprintf("[% 5d)](fg:cyan) %s", i, data[i])
	}
}

func (v *viewer) renderSet(key string, data []string) {
	l := len(data)
	v.Rows = make([]string, l+2)
	v.Rows[0] = fmt.Sprintf(setRenderTemplate[0], strings.ToUpper(string(r.TypeSet)), key, l)
	v.Rows[1] = setRenderTemplate[1]
	for i := range data {
		v.Rows[i+2] = fmt.Sprintf("   [-](fg:cyan) %s", data[i])
	}
}

func (v *viewer) renderSortedSet(key string, data []redis.Z) {
	l := len(data)
	v.Rows = make([]string, l+2)
	v.Rows[0] = fmt.Sprintf(zsetRenderTemplate[0], strings.ToUpper(string(r.TypeSortedSet)), key, l)
	v.Rows[1] = zsetRenderTemplate[1]
	for i, z := range data {
		v.Rows[i+2] = fmt.Sprintf("[% 20f](fg:green) - %v", z.Score, z.Member)
	}
}

func (v *viewer) renderHash(key string, data map[string]string) {
	l := len(data)
	v.Rows = make([]string, l+2)
	v.Rows[0] = fmt.Sprintf(hashRenderTemplate[0], strings.ToUpper(string(r.TypeHash)), key, l)
	v.Rows[1] = hashRenderTemplate[1]
	i := 2
	for field, value := range data {
		v.Rows[i] = fmt.Sprintf("[% 20s](fg:green): %s", field, value)
		i++
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
