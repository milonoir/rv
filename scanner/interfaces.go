package scanner

import (
	"context"
	"time"

	"github.com/milonoir/rv/common"
	r "github.com/milonoir/rv/redis"
)

// Worker provides an interface to interact with workers.
// A worker periodically executes a Redis SCAN command with a configured pattern and interval.
type Worker interface {
	common.Messenger

	// Run executes the configured Redis scan command.
	Run()

	// IsSingle returns true if the worker scans a single Redis key.
	IsSingle() bool

	// Pattern returns the configured pattern of the Redis scan command and the type of the matching keys.
	Pattern() (string, r.DataType)

	// State returns the last response and execution time of the Redis scan command and whether the worker is enabled.
	State() ([]string, time.Time, bool)

	// Enable enables the worker.
	Enable()

	// Disable disables the worker.
	Disable()

	// ErrCh is a channel where worker sends its error messages.
	ErrCh() <-chan string
}

// Executor provides an interface with the Redis command executor.
type Executor interface {
	// Execute executes a Redis read-only command based on the data type.
	Execute(context.Context, string, r.DataType) (interface{}, error)
}

// Scanner provides an interface to interact with the scanner widget.
type Scanner interface {
	common.Widget
	common.Messenger
	common.Scrollable

	// Select returns data from the selected worker.
	Select() ([]string, r.DataType)

	// Enable enables the selected worker.
	Enable()

	// Disable disables the selected worker.
	Disable()
}

// Selector provides an interface to interact with the selector widget.
type Selector interface {
	common.Widget
	common.Scrollable

	// Select returns the selected Redis key and data type from the list.
	Select() (string, r.DataType)

	// SetItems sets the list rows and Redis data type.
	SetItems([]string, r.DataType)
}

// Viewer provides an interface to interact with the viewer widget.
type Viewer interface {
	common.Widget
	common.Messenger
	common.Scrollable

	// View shows the details of the provided Redis key based on its data type.
	View(context.Context, string, r.DataType)
}
