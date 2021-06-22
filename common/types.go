package common

import (
	"time"
)

// Duration is a wrapper around time.Duration that implements encoding.TextUnmarshaler.
type Duration struct {
	time.Duration
}

// UnmarshalText implements the encoding.TextUnmarshaler interface.
func (d *Duration) UnmarshalText(text []byte) error {
	var err error
	d.Duration, err = time.ParseDuration(string(text))
	return err
}
