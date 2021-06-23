package scanner

import (
	"github.com/milonoir/rv/common"
)

// Config is the configuration for a Scanner.
type Config struct {
	Pattern  string          `toml:"pattern"`
	Type     string          `toml:"type"`
	Interval common.Duration `toml:"interval"`
}
