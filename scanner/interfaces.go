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

	// State returns the last response and execution time of the Redis scan command and whether the scanner is enabled.
	State() ([]string, time.Time, bool)

	// Enable enables the scanner.
	Enable()

	// Disable disables the scanner.
	Disable()
}
