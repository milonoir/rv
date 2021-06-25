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
)

const (
	scannerUsage = `  [<Up>](fg:yellow)/[<Down>](fg:yellow)   move selection up/down   [<Enter>](fg:yellow) select            [<m>](fg:yellow) view messages
[<PgUp>](fg:yellow)/[<PgDown>](fg:yellow) scroll up/down           [<e>](fg:yellow)     enable scanner
[<Home>](fg:yellow)/[<End>](fg:yellow)    move to top/bottom       [<d>](fg:yellow)     disable scanner   [<q>](fg:yellow) quit`
	selectorUsage = `  [<Up>](fg:yellow)/[<Down>](fg:yellow)   move selection up/down   [<Enter>](fg:yellow) select
[<PgUp>](fg:yellow)/[<PgDown>](fg:yellow) scroll up/down           [<Esc>](fg:yellow)   go back
[<Home>](fg:yellow)/[<End>](fg:yellow)    move to top/bottom       [<q>](fg:yellow)     quit`
	viewerUsage = `  [<Up>](fg:yellow)/[<Down>](fg:yellow)   move selection up/down
[<PgUp>](fg:yellow)/[<PgDown>](fg:yellow) scroll up/down           [<Esc>](fg:yellow)   go back
[<Home>](fg:yellow)/[<End>](fg:yellow)    move to top/bottom       [<q>](fg:yellow)     quit`
	messagesUsage = `[<Esc>](fg:yellow) go back
  [<q>](fg:yellow) quit`
)

var (
	updateInterval = 100 * time.Millisecond
	viewerTimeout  = 3 * time.Second
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
	selector scanner.Selector
	viewer   scanner.Viewer
	helper   common.TextBox
	messages common.TextBox
	logger   logger.Logger

	messagesVisible bool
	selectorVisible bool
	viewerVisible   bool

	msgCh chan string
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
	a.msgCh = make(chan string, 1)

	// Scanner widget
	a.scanner = scanner.NewScanner(ctx, a.rc, a.cfg.Scans)

	// Selector widget
	a.selector = scanner.NewSelector()

	// Viewer widget
	a.viewer = scanner.NewViewer(a.rc)

	// Helper widget
	a.helper = common.NewTextBox(" Help ")
	a.helper.SetText(scannerUsage)

	// Logger widget
	a.logger = logger.NewLogger(ctx, a.msgCh, a.scanner.Messages(), a.viewer.Messages())

	// Messages widget
	a.messages = common.NewTextBox(" Messages ")

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
			case a.viewerVisible:
				a.handleViewerEvents(e)
			case a.selectorVisible:
				a.handleSelectorEvents(ctx, e)
			case a.messagesVisible:
				a.handleMessagesEvents(e)
			default:
				a.handleScannerEvents(e)
			}
		}
	}
}

func (a *app) handleScannerEvents(e ui.Event) {
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
		items, rt := a.scanner.Select()
		switch {
		case items == nil:
			a.msgCh <- fmt.Sprintf("Error in selection")
		case len(items) == 0:
			a.msgCh <- fmt.Sprintf("No matching keys")
		default:
			a.selector.SetItems(items, rt)
			a.helper.SetText(selectorUsage)
			a.selectorVisible = true
		}
	case "e":
		a.scanner.Enable()
	case "d":
		a.scanner.Disable()
	case "m":
		a.messages.SetText(strings.Join(a.logger.Messages(), "\n"))
		a.helper.SetText(messagesUsage)
		a.messagesVisible = true
	}
}

func (a *app) handleSelectorEvents(ctx context.Context, e ui.Event) {
	switch e.ID {
	case "<Up>":
		a.selector.ScrollUp()
	case "<Down>":
		a.selector.ScrollDown()
	case "<PageUp>":
		a.selector.ScrollPageUp()
	case "<PageDown>":
		a.selector.ScrollPageDown()
	case "<Home>":
		a.selector.ScrollTop()
	case "<End>":
		a.selector.ScrollBottom()
	case "<Enter>":
		c, cancel := context.WithTimeout(ctx, viewerTimeout)
		defer cancel()
		key, rt := a.selector.Select()
		a.msgCh <- fmt.Sprintf("rendering view for key: %s, type: %s", key, rt)
		a.viewer.View(c, key, rt)
		a.helper.SetText(viewerUsage)
		a.selectorVisible = false
		a.viewerVisible = true
	case "<Escape>":
		a.selectorVisible = false
		a.helper.SetText(scannerUsage)
	}
}

func (a *app) handleViewerEvents(e ui.Event) {
	switch e.ID {
	case "<Escape>":
		a.viewerVisible = false
		a.selectorVisible = true
		a.helper.SetText(selectorUsage)
	}
}

func (a *app) handleMessagesEvents(e ui.Event) {
	switch e.ID {
	case "<Escape>":
		a.messagesVisible = false
		a.helper.SetText(scannerUsage)
	}
}

// update invokes the Update() method on each widget.
func (a *app) update() {
	a.helper.Update()
	a.logger.Update()
	switch {
	case a.viewerVisible:
		a.viewer.Update()
	case a.selectorVisible:
		a.selector.Update()
	case a.messagesVisible:
		a.messages.Update()
	default:
		a.scanner.Update()
	}
}

// resize resizes all widgets.
func (a *app) resize(w, h int) {
	a.scanner.Resize(0, 0, w, h-5)
	a.selector.Resize(0, 0, w, h-5)
	a.viewer.Resize(0, 0, w, h-5)
	a.messages.Resize(0, 0, w, h-5)
	a.helper.Resize(0, h-5, w/2, h)
	a.logger.Resize(w/2, h-5, w, h)

	ui.Clear()
}

// handleQuit invokes the Close() method on each widget and closes termui.
func (a *app) handleQuit() {
	close(a.msgCh)

	a.messages.Close()
	a.viewer.Close()
	a.selector.Close()
	a.scanner.Close()
	a.helper.Close()
	a.logger.Close()

	ui.Clear()
	ui.Close()
}
