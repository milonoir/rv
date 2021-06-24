package logger

import (
	"github.com/milonoir/rv/common"
)

// Logger provides an interface to interact with the logger widget.
type Logger interface {
	common.Widget

	// Messages returns all the messages from the logger's buffer.
	Messages() []string
}
