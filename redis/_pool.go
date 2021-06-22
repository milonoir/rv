package redis

import (
	"fmt"
	"net"
	"time"

	r "github.com/gomodule/redigo/redis"
)

// Pool
type Pool interface {
	Do(command string, args ...interface{}) (reply interface{}, err error)
	DoScript(script *r.Script, args ...interface{}) (interface{}, error)
}

// pool represents a Redis connection pool.
type pool struct {
	*r.Pool
}

// Do sends a command to the server and returns the received reply on a connection from the pool.
// This can be used to easily minimize the time a connection is held from the pool for an individual Conn.Do.
func (p *pool) Do(command string, args ...interface{}) (reply interface{}, err error) {
	c := p.Get()
	defer c.Close()

	return c.Do(command, args...)
}

// DoScript evaluates the script on a connection from the pool.
// This can be used to easily minimize the time a connection is held from the pool for an individual Script.Do.
func (p *pool) DoScript(script *r.Script, args ...interface{}) (interface{}, error) {
	c := p.Get()
	defer c.Close()

	return script.Do(c, args...)
}

// NewPool returns a new Redis pool setup from cfg.
func NewPool(cfg *Config) Pool {
	options := make([]r.DialOption, 0, 5)
	return newPool(cfg, options)
}

func newPool(cfg *Config, options []r.DialOption) *pool {
	if cfg.ReadTimeout.Duration != 0 {
		options = append(options, r.DialReadTimeout(cfg.ReadTimeout.Duration))
	}
	if cfg.WriteTimeout.Duration != 0 {
		options = append(options, r.DialWriteTimeout(cfg.WriteTimeout.Duration))
	}
	if cfg.ConnectTimeout.Duration != 0 {
		options = append(options, r.DialConnectTimeout(cfg.ConnectTimeout.Duration))
	}
	if cfg.Password != "" {
		options = append(options, r.DialPassword(cfg.Password))
	}
	if cfg.DB != 0 {
		options = append(options, r.DialDatabase(cfg.DB))
	}

	p := &r.Pool{
		MaxIdle:     cfg.MaxIdle,
		MaxActive:   cfg.MaxActive,
		IdleTimeout: cfg.IdleTimeout.Duration,
		Wait:        cfg.Wait,
		Dial: func() (r.Conn, error) {
			return retryingDialer(cfg, options)
		},
	}

	if cfg.TestOnBorrow {
		p.TestOnBorrow = func(c r.Conn, t time.Time) error {
			_, err := c.Do("PING")
			return err
		}
	}

	return &pool{Pool: p}
}

// retryingDialer dials the Redis connection, retrying if it receives a failure.
func retryingDialer(cfg *Config, options []r.DialOption) (c r.Conn, err error) {
	for i := 0; i < 1+cfg.DialRetries; i++ {
		c, err = r.Dial("tcp", cfg.Server, options...)
		if err == nil {
			return c, nil
		}
		if !isErrRetryable(err) {
			return nil, err
		}

		err = fmt.Errorf("dial fail after %d times: %w", i+1, err)
	}

	return c, err
}

// isErrRetryable returns whether the retry should take place.
func isErrRetryable(err error) bool {
	if err, ok := err.(net.Error); ok && err.Timeout() {
		return true
	}
	return false
}
