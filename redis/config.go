package redis

import (
	"github.com/milonoir/rv/common"
)

// Config is the configuration for a Redis server.
type Config struct {
	Server       string          `toml:"server"`
	Password     string          `toml:"password"`
	DB           int             `toml:"db"`
	DialTimeout  common.Duration `toml:"dial_timeout"`
	IdleTimeout  common.Duration `toml:"idle_timeout"`
	ReadTimeout  common.Duration `toml:"read_timeout"`
	WriteTimeout common.Duration `toml:"write_timeout"`
	MaxRetries   int             `toml:"max_retries"`
}
