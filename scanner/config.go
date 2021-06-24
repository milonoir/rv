package scanner

import (
	"strings"

	"github.com/milonoir/rv/common"
)

// Config is the configuration for a Worker.
type Config struct {
	Pattern  string          `toml:"pattern"`
	Type     string          `toml:"type"`
	Interval common.Duration `toml:"interval"`
}

// IsSingle implements the Worker interface.
func (c Config) IsSingle() bool {
	return !strings.Contains(c.Pattern, "*")
}
