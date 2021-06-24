package logger

import (
	"context"
	"strings"
	"sync"

	ui "github.com/gizak/termui/v3"
	"github.com/gizak/termui/v3/widgets"
)

const (
	bufSize  = 100
	prevSize = 3
)

type logger struct {
	*widgets.Paragraph

	cancel   context.CancelFunc
	messages []string
	mtx      sync.Mutex
	wg       sync.WaitGroup
}

func NewLogger(ctx context.Context, channels ...<-chan string) *logger {
	ctx, cancel := context.WithCancel(ctx)

	l := &logger{
		Paragraph: widgets.NewParagraph(),
		messages:  make([]string, 0, bufSize),
		cancel:    cancel,
	}
	l.Title = " Messages "
	l.WrapText = true

	for _, ch := range channels {
		l.wg.Add(1)
		go func(c <-chan string) {
			defer l.wg.Done()
			l.readChan(ctx, c)
		}(ch)
	}

	return l
}

func (l *logger) readChan(ctx context.Context, in <-chan string) {
	for {
		select {
		case <-ctx.Done():
			return
		case m := <-in:
			l.mtx.Lock()
			l.messages = append(l.messages, m)
			l.mtx.Unlock()
		}
	}
}

func (l *logger) push(m string) {
	l.mtx.Lock()
	defer l.mtx.Unlock()

	l.messages = append(l.messages, m)
	if len(l.messages) > bufSize {
		l.messages = l.messages[1:]
	}
}

func (l *logger) Update() {
	l.mtx.Lock()
	n := len(l.messages) - prevSize
	if n < 0 {
		n = 0
	}
	l.Text = strings.Join(l.messages[n:], "\n")
	l.mtx.Unlock()
	ui.Render(l)
}

func (l *logger) Resize(x1, y1, x2, y2 int) {
	l.SetRect(x1, y1, x2, y2)
}

func (l *logger) Close() {
	l.cancel()
	l.wg.Wait()
}

func (l *logger) Messages() []string {
	l.mtx.Lock()
	m := make([]string, len(l.messages))
	copy(m, l.messages)
	l.mtx.Unlock()
	return m
}
