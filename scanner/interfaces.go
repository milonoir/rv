package scanner

import (
	"time"

	"github.com/milonoir/rv/common"
)

// Worker provides an interface to interact with workers.
// A worker periodically executes a Redis SCAN command with a configured pattern and interval.
type Worker interface {
	// Run executes the configured Redis scan command.
	Run()

	// IsSingle returns true if the worker scans a single Redis key.
	IsSingle() bool

	// Pattern returns the configured pattern of the Redis scan command and the type of the matching keys.
	Pattern() (string, string)

	// State returns the last response and execution time of the Redis scan command and whether the worker is enabled.
	State() ([]string, time.Time, bool)

	// Enable enables the worker.
	Enable()

	// Disable disables the worker.
	Disable()

	// ErrCh is a channel where worker sends its error messages.
	ErrCh() <-chan string
}

// Scanner provides an interface to interact with the scanner scanner.
type Scanner interface {
	common.Widget
	common.Scrollable

	// Select returns data from the selected worker.
	Select() string

	// Enable enables the selected worker.
	Enable()

	// Disable disables the selected worker.
	Disable()
}
