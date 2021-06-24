package main

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/BurntSushi/toml"
	ui "github.com/gizak/termui/v3"
	r "github.com/go-redis/redis/v8"
	"github.com/milonoir/rv/common"
	"github.com/milonoir/rv/logger"
	"github.com/milonoir/rv/redis"
	"github.com/milonoir/rv/scanner"
	"github.com/milonoir/rv/textbox"
)

const (
	scannerUsage = `  [<Up>](fg:yellow)/[<Down>](fg:yellow)   move selection up/down   [<Enter>](fg:yellow) select            [<m>](fg:yellow) view messages 
[<PgUp>](fg:yellow)/[<PgDown>](fg:yellow) scroll up/down           [<e>](fg:yellow)     enable scanner
[<Home>](fg:yellow)/[<End>](fg:yellow)    move to top/bottom       [<d>](fg:yellow)     disable scanner   [<q>](fg:yellow) quit`
	messagesUsage = `[<Esc>](fg:yellow) back
  [<q>](fg:yellow) quit`
)

var (
	updateInterval = 100 * time.Millisecond
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

	scanner  scanner.Scanner
	helper   textbox.TextBox
	messages textbox.TextBox
	logger   logger.Logger

	messagesActive bool

	ch chan string
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

// initWidgets initializes the widgets.
func (a *app) initWidgets(ctx context.Context) {
	a.ch = make(chan string, 1)

	// Scanner widget
	a.scanner = scanner.NewScanner(ctx, a.rc, a.cfg.Scans)

	// Helper widget
	a.helper = textbox.NewTextBox(" Help ")
	a.helper.SetText(scannerUsage)

	// Logger widget
	a.logger = logger.NewLogger(ctx, a.ch)

	// Messages widget
	a.messages = textbox.NewTextBox(" Messages ")

	a.resize(ui.TerminalDimensions())
}

// run is the main event loop of the application.
func (a *app) run() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	a.initWidgets(ctx)

	t := time.NewTicker(updateInterval)
	defer t.Stop()

	uiEvents := ui.PollEvents()
	for {
		select {
		case <-t.C:
			a.update()
		case e := <-uiEvents:
			switch e.ID {
			case "<Resize>":
				payload := e.Payload.(ui.Resize)
				a.resize(payload.Width, payload.Height)
			case "q", "<C-c>":
				a.handleQuit()
				return
			}

			switch {
			case a.messagesActive:
				switch e.ID {
				case "<Escape>":
					a.messagesActive = false
					a.helper.SetText(scannerUsage)
				}
			default:
				switch e.ID {
				case "<Up>":
					a.scanner.ScrollUp()
				case "<Down>":
					a.scanner.ScrollDown()
				case "<PageUp>":
					a.scanner.ScrollPageUp()
				case "<PageDown>":
					a.scanner.ScrollPageDown()
				case "<Home>":
					a.scanner.ScrollTop()
				case "<End>":
					a.scanner.ScrollBottom()
				case "<Enter>":
					// TODO: send selection to viewer
					a.ch <- a.scanner.Select()
				case "e":
					a.scanner.Enable()
				case "d":
					a.scanner.Disable()
				case "m":
					a.messages.SetText(strings.Join(a.logger.Messages(), "\n"))
					a.helper.SetText(messagesUsage)
					a.messagesActive = true
				}
			}
		}
	}
}

// update invokes the Update() method on each widget.
func (a *app) update() {
	a.helper.Update()
	a.logger.Update()
	if a.messagesActive {
		a.messages.Update()
	} else {
		a.scanner.Update()
	}
}

// resize resizes all widgets.
func (a *app) resize(w, h int) {
	a.scanner.Resize(0, 0, w, h-5)
	a.helper.Resize(0, h-5, w/2, h)
	a.logger.Resize(w/2, h-5, w, h)
	a.messages.Resize(0, 0, w, h-5)

	ui.Clear()
}

// handleQuit invokes the Close() method on each widget and closes termui.
func (a *app) handleQuit() {
	close(a.ch)

	a.messages.Close()
	a.scanner.Close()
	a.helper.Close()
	a.logger.Close()

	ui.Clear()
	ui.Close()
}
