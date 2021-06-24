package scanner

import (
	"strings"

	"github.com/milonoir/rv/common"
	r "github.com/milonoir/rv/redis"
)

// Config is the configuration for a Worker.
type Config struct {
	Pattern  string          `toml:"pattern"`
	Type     r.DataType      `toml:"type"`
	Interval common.Duration `toml:"interval"`
}

// IsSingle implements the Worker interface.
func (c Config) IsSingle() bool {
	return !strings.Contains(c.Pattern, "*")
}
