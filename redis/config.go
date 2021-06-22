package redis

import (
	"milonoir/rv/common"
)

// Config is the configuration for a Redis server.
type Config struct {
	Server         string          `toml:"server"`
	Password       string          `toml:"password"`
	DB             int             `toml:"db"`
	ConnectTimeout common.Duration `toml:"connect_timeout"`
	IdleTimeout    common.Duration `toml:"idle_timeout"`
	ReadTimeout    common.Duration `toml:"read_timeout"`
	WriteTimeout   common.Duration `toml:"write_timeout"`
	MaxActive      int             `toml:"max_active"`
	MaxIdle        int             `toml:"max_idle"`
	Wait           bool            `toml:"wait"`
	TestOnBorrow   bool            `toml:"test_on_borrow"`
	DialRetries    int             `toml:"dial_retries"`
}
