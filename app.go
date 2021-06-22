package main

import (
	"context"
	"fmt"
	"time"

	"github.com/BurntSushi/toml"
	ui "github.com/gizak/termui/v3"
	r "github.com/go-redis/redis/v8"
	"milonoir/rv/common"
	"milonoir/rv/redis"
	"milonoir/rv/scanner"
)

// config represents the application configuration.
type config struct {
	Redis *redis.Config
	Scans map[string]*scanner.Config
}

// app represents the main application.
type app struct {
	cfg *config
	rc  *r.Client
}

// newApp creates and configures a new app.
func newApp(cfgFile string) (*app, error) {
	f, err := common.LoadFile(cfgFile)
	if err != nil {
		return nil, fmt.Errorf("load config file: %w", err)
	}

	cfg := &config{}
	if _, err = toml.Decode(string(f), cfg); err != nil {
		return nil, fmt.Errorf("parse config: %w", err)
	}

	return &app{
		cfg: cfg,
	}, nil
}

// setup configures and initializes the components of the app.
func (a *app) setup() error {
	if err := a.setupRedis(); err != nil {
		return fmt.Errorf("setup Redis: %w", err)
	}

	if err := a.initUI(); err != nil {
		return fmt.Errorf("init termui: %w", err)
	}

	return nil
}

// setupRedis configures the Redis client and tests its connection to the Redis server.
func (a *app) setupRedis() error {
	a.rc = r.NewClient(&r.Options{
		Addr:     a.cfg.Redis.Server,
		Password: a.cfg.Redis.Password,
		DB:       a.cfg.Redis.DB,
	})

	// Test connection.
	reply, err := a.rc.Do(context.Background(), "PING").Text()
	if err != nil {
		return fmt.Errorf("test Redis connection ping: %w", err)
	}
	if reply != "PONG" {
		return fmt.Errorf("unexpected response from Redis: %s != PONG", reply)
	}

	return nil
}

// initUI initializes the termui.
func (a *app) initUI() error {
	return ui.Init()
}

// run is the main loop of the application.
func (a *app) run() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sw := scanner.NewWidget(ctx, a.rc, a.cfg.Scans)

	t := time.NewTicker(time.Second)
	defer t.Stop()

	uiEvents := ui.PollEvents()
	for {
		select {
		case <-t.C:
			a.update(sw)
		case e := <-uiEvents:
			switch e.ID {
			case "q", "<C-c>":
				a.handleQuit(sw)
				return
			}
		}
	}
}

// update invokes the Update() method on each provided widget.
func (a *app) update(ws ...common.Widget) {
	for _, w := range ws {
		w.Update()
	}
}

// handleQuit invokes the Close() method on each provided widget and closes termui.
func (a *app) handleQuit(ws ...common.Widget) {
	for _, w := range ws {
		w.Close()
	}
	ui.Close()
}
