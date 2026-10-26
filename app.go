package main

import (
	"context"
	"database/sql"
	"embed"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strconv"

	"github.com/mhetem/DH-Companion/internal/db"
	"github.com/mhetem/DH-Companion/internal/gm"
	"github.com/mhetem/DH-Companion/internal/player"
	"github.com/mhetem/DH-Companion/internal/srd"
	"github.com/pressly/goose/v3"
	"github.com/wailsapp/wails/v3/pkg/application"
	"github.com/wailsapp/wails/v3/pkg/events"
	_ "modernc.org/sqlite"
)

//go:embed sql/schema/*.sql
var migrations embed.FS

const (
	settingWindowSize = "window_size"
	settingUIScale    = "ui_scale"

	defaultWindowWidth  = 1920
	defaultWindowHeight = 1080
	minWindowWidth      = 1100
	minWindowHeight     = 700

	defaultUIScale = 100
	minUIScale     = 80
	maxUIScale     = 200
)

var uiScales = []int{80, 90, 100, 110, 125, 150, 175, 200}

type WindowSize struct {
	Label  string `json:"label"`
	Width  int    `json:"width"`
	Height int    `json:"height"`
	Fits   bool   `json:"fits"`
}

var windowPresets = []WindowSize{
	{Label: "HD", Width: 1280, Height: 720},
	{Label: "HD+", Width: 1600, Height: 900},
	{Label: "Full HD", Width: 1920, Height: 1080},
	{Label: "QHD", Width: 2560, Height: 1440},
	{Label: "4K", Width: 3840, Height: 2160},
}

type App struct {
	ctx        context.Context
	wails      *application.App
	window     *application.WebviewWindow
	q          *db.Queries
	conn       *sql.DB
	gm         *gm.Service
	player     *player.Service
	catalog    *srd.Catalog
	updaterErr error
}

func NewApp(gmSvc *gm.Service, playerSvc *player.Service) *App {
	return &App{gm: gmSvc, player: playerSvc}
}

func (a *App) ServiceStartup(ctx context.Context, _ application.ServiceOptions) error {
	a.ctx = ctx
	if title, err := a.open(); err != nil {
		a.fatal(title, err)
		return nil
	}

	width, height := a.savedWindowSize()
	a.window = a.wails.Window.NewWithOptions(application.WebviewWindowOptions{
		Name:             "main",
		Title:            "Hilt",
		Width:            width,
		Height:           height,
		MinWidth:         minWindowWidth,
		MinHeight:        minWindowHeight,
		InitialPosition:  application.WindowCentered,
		BackgroundColour: application.NewRGB(16, 13, 23),
		URL:              "/",
	})
	a.window.RegisterHook(events.Common.WindowClosing, func(*application.WindowEvent) {
		a.rememberWindowSize()
	})
	a.updaterErr = a.initUpdater()
	return nil
}

func (a *App) ServiceShutdown() error {
	if a.conn == nil {
		return nil
	}
	a.rememberWindowSize()
	return a.conn.Close()
}

func (a *App) open() (string, error) {
	q, conn, err := Open("")
	if err != nil {
		return "Database error", err
	}
	a.q, a.conn = q, conn

	catalog, err := srd.Default()
	if err != nil {
		return "Reference data error", err
	}
	a.catalog = catalog
	gm.Attach(a.gm, a.ctx, q, conn, catalog)
	player.Attach(a.player, a.ctx, q, conn, catalog)

	if err := a.gm.ReindexCards(); err != nil {
		return "Search index error", err
	}
	return "", nil
}

func (a *App) fatal(title string, err error) {
	a.wails.Event.OnApplicationEvent(events.Common.ApplicationStarted, func(*application.ApplicationEvent) {
		dialog := a.wails.Dialog.Error().SetTitle(title).SetMessage(err.Error())
		quit := dialog.AddButton("Quit").OnClick(a.wails.Quit)
		dialog.SetDefaultButton(quit).SetCancelButton(quit).Show()
	})
}

func (a *App) WindowPresets() []WindowSize {
	screenW, screenH := a.screenSize()
	out := make([]WindowSize, 0, len(windowPresets))
	for _, p := range windowPresets {
		p.Fits = (screenW == 0 || p.Width <= screenW) && (screenH == 0 || p.Height <= screenH)
		out = append(out, p)
	}
	return out
}

func (a *App) CurrentWindowSize() WindowSize {
	w, h := a.window.Size()
	return WindowSize{Label: sizeLabel(w, h), Width: w, Height: h, Fits: true}
}

func (a *App) SetWindowSize(width, height int) (WindowSize, error) {
	if width < minWindowWidth || height < minWindowHeight {
		return WindowSize{}, fmt.Errorf("the window can't be smaller than %d×%d", minWindowWidth, minWindowHeight)
	}
	width, height = a.clampToScreen(width, height)
	a.window.SetSize(width, height)
	a.window.Center()
	if err := a.saveWindowSize(width, height); err != nil {
		return WindowSize{}, err
	}
	return a.CurrentWindowSize(), nil
}

func (a *App) UIScales() []int {
	return slices.Clone(uiScales)
}

// GetUIScale is a percentage the frontend applies to the root font size. Every
// length in the app is in rem, so this scales the layout with the text rather than
// leaving controls the same size around bigger glyphs.
func (a *App) GetUIScale() int {
	raw, err := a.q.GetSetting(a.ctx, settingUIScale)
	if err != nil {
		return defaultUIScale
	}
	scale, err := strconv.Atoi(raw)
	if err != nil || scale < minUIScale || scale > maxUIScale {
		return defaultUIScale
	}
	return scale
}

func (a *App) SetUIScale(scale int) error {
	if scale < minUIScale || scale > maxUIScale {
		return fmt.Errorf("scale must be between %d%% and %d%%, got %d%%", minUIScale, maxUIScale, scale)
	}
	return a.q.SetSetting(a.ctx, db.SetSettingParams{
		Key:   settingUIScale,
		Value: strconv.Itoa(scale),
	})
}

func sizeLabel(width, height int) string {
	for _, p := range windowPresets {
		if p.Width == width && p.Height == height {
			return p.Label
		}
	}
	return "Custom"
}

func (a *App) screenSize() (int, int) {
	var screen *application.Screen
	if a.window != nil {
		screen, _ = a.window.GetScreen()
	}
	if screen == nil {
		screen = a.wails.Screen.GetPrimary()
	}
	if screen == nil {
		return 0, 0
	}
	return screen.Size.Width, screen.Size.Height
}

func (a *App) clampToScreen(width, height int) (int, int) {
	screenW, screenH := a.screenSize()
	if screenW > 0 && width > screenW {
		width = screenW
	}
	if screenH > 0 && height > screenH {
		height = screenH
	}
	return width, height
}

func (a *App) saveWindowSize(width, height int) error {
	return a.q.SetSetting(context.WithoutCancel(a.ctx), db.SetSettingParams{
		Key:   settingWindowSize,
		Value: fmt.Sprintf("%dx%d", width, height),
	})
}

func (a *App) savedWindowSize() (int, int) {
	width, height := defaultWindowWidth, defaultWindowHeight
	if raw, err := a.q.GetSetting(a.ctx, settingWindowSize); err == nil {
		var w, h int
		if _, err := fmt.Sscanf(raw, "%dx%d", &w, &h); err == nil && w >= minWindowWidth && h >= minWindowHeight {
			width, height = w, h
		}
	}
	return a.clampToScreen(width, height)
}

func (a *App) rememberWindowSize() {
	if a.q == nil || a.window == nil {
		return
	}
	if w, h := a.window.Size(); w >= minWindowWidth && h >= minWindowHeight {
		_ = a.saveWindowSize(w, h)
	}
}

func dataDir(dir string) (string, error) {
	if dir != "" {
		return dir, nil
	}
	if d := os.Getenv("DH_DATA_DIR"); d != "" {
		return d, nil
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, "DH-Companion"), nil
}

func Open(dir string) (*db.Queries, *sql.DB, error) {
	dir, err := dataDir(dir)
	if err != nil {
		return nil, nil, err
	}
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return nil, nil, err
	}

	dsn := filepath.Join(dir, "data.db") + "?_pragma=journal_mode(WAL)&_pragma=foreign_keys(1)&_pragma=busy_timeout(5000)"
	conn, err := sql.Open("sqlite", dsn)
	if err != nil {
		return nil, nil, err
	}
	if err := conn.Ping(); err != nil {
		return nil, nil, err
	}
	if err := migrate(conn); err != nil {
		return nil, nil, err
	}
	return db.New(conn), conn, nil
}

func migrate(conn *sql.DB) error {
	goose.SetBaseFS(migrations)
	if err := goose.SetDialect("sqlite3"); err != nil {
		return err
	}
	return goose.Up(conn, "sql/schema")
}
