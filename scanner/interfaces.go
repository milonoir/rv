package scanner

import (
	"time"
)

// Scanner provides an interface to interact with scanners.
// A scanner periodically executes a Redis SCAN command with a configured pattern and interval.
type Scanner interface {
	// Run executes the configured Redis scan command.
	Run()

	// Pattern returns the configured pattern of the Redis scan command.
	Pattern() string

	// LastReply returns the last response and execution time of the Redis scan command.
	LastReply() ([]string, time.Time)

	// Enable enables the scanner.
	Enable()

	// Disable disables the scanner.
	Disable()
}
