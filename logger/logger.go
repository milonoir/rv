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

// logger collects messages from different sources and stores them in a buffer.
// It also renders a preview of the last few messages into a widgets.Paragraph.
type logger struct {
	*widgets.Paragraph

	cancel   context.CancelFunc
	messages []string
	mtx      sync.Mutex
	wg       sync.WaitGroup
}

// NewLogger returns a fully configured logger.
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

// readChan is a worker goroutine for worker that pushes incoming messages from the provided
// channel into the buffer.
func (l *logger) readChan(ctx context.Context, in <-chan string) {
	for {
		select {
		case <-ctx.Done():
			return
		case m := <-in:
			l.push(m)
		}
	}
}

// push is a thread-safe method to push a message into the buffer.
func (l *logger) push(m string) {
	l.mtx.Lock()
	defer l.mtx.Unlock()

	l.messages = append(l.messages, m)
	if len(l.messages) > bufSize {
		l.messages = l.messages[1:]
	}
}

// Update implements the common.Widget interface.
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

// Resize implements the common.Widget interface.
func (l *logger) Resize(x1, y1, x2, y2 int) {
	l.SetRect(x1, y1, x2, y2)
}

// Close implements the common.Widget interface.
func (l *logger) Close() {
	l.cancel()
	l.wg.Wait()
}

// Messages implements the Logger interface.
func (l *logger) Messages() []string {
	l.mtx.Lock()
	m := make([]string, len(l.messages))
	copy(m, l.messages)
	l.mtx.Unlock()
	return m
}
