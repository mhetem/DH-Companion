package main

import (
	"context"
	"database/sql"
	"embed"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strconv"

	"github.com/mhetem/DH-Companion/internal/db"
	"github.com/mhetem/DH-Companion/internal/gm"
	"github.com/mhetem/DH-Companion/internal/srd"
	"github.com/pressly/goose/v3"
	"github.com/wailsapp/wails/v2/pkg/runtime"
	_ "modernc.org/sqlite"
)

//go:embed sql/schema/*.sql
var migrations embed.FS

const (
	RoleGM     = "gm"
	RolePlayer = "player"

	settingLastRole   = "last_role"
	settingWindowSize = "window_size"
	settingUIScale    = "ui_scale"

	minWindowWidth  = 1100
	minWindowHeight = 700

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
	ctx     context.Context
	q       *db.Queries
	conn    *sql.DB
	gm      *gm.Service
	catalog *srd.Catalog
}

func NewApp(gmSvc *gm.Service) *App {
	return &App{gm: gmSvc}
}

func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
	q, conn, err := Open("")
	if err != nil {
		a.fatal(ctx, "Database error", err)
		return
	}
	a.q, a.conn = q, conn

	catalog, err := srd.Default()
	if err != nil {
		a.fatal(ctx, "Reference data error", err)
		return
	}
	a.catalog = catalog
	gm.Attach(a.gm, ctx, q, conn, catalog)

	if err := a.gm.ReindexCards(); err != nil {
		a.fatal(ctx, "Search index error", err)
		return
	}

	a.restoreWindowSize()
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
	w, h := runtime.WindowGetSize(a.ctx)
	return WindowSize{Label: sizeLabel(w, h), Width: w, Height: h, Fits: true}
}

func (a *App) SetWindowSize(width, height int) (WindowSize, error) {
	if width < minWindowWidth || height < minWindowHeight {
		return WindowSize{}, fmt.Errorf("the window can't be smaller than %d×%d", minWindowWidth, minWindowHeight)
	}
	width, height = a.clampToScreen(width, height)
	runtime.WindowSetSize(a.ctx, width, height)
	runtime.WindowCenter(a.ctx)
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

// screenSize reports the logical size of the screen the window is on. Screen.Size
// is the logical-pixel space WindowSetSize works in; Width/Height are physical and
// deprecated. Zero means the platform wouldn't say, and callers skip clamping.
func (a *App) screenSize() (int, int) {
	screens, err := runtime.ScreenGetAll(a.ctx)
	if err != nil {
		return 0, 0
	}
	for _, s := range screens {
		if s.IsCurrent {
			return s.Size.Width, s.Size.Height
		}
	}
	for _, s := range screens {
		if s.IsPrimary {
			return s.Size.Width, s.Size.Height
		}
	}
	return 0, 0
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
	return a.q.SetSetting(a.ctx, db.SetSettingParams{
		Key:   settingWindowSize,
		Value: fmt.Sprintf("%dx%d", width, height),
	})
}

func (a *App) restoreWindowSize() {
	raw, err := a.q.GetSetting(a.ctx, settingWindowSize)
	if err != nil {
		return
	}
	var width, height int
	if _, err := fmt.Sscanf(raw, "%dx%d", &width, &height); err != nil {
		return
	}
	if width < minWindowWidth || height < minWindowHeight {
		return
	}
	width, height = a.clampToScreen(width, height)
	runtime.WindowSetSize(a.ctx, width, height)
	runtime.WindowCenter(a.ctx)
}

func (a *App) fatal(ctx context.Context, title string, err error) {
	runtime.MessageDialog(ctx, runtime.MessageDialogOptions{
		Type: runtime.ErrorDialog, Title: title, Message: err.Error(),
	})
	runtime.Quit(ctx)
}

// shutdown records whatever size the window was left at, so dragging the frame is
// remembered the same way picking a preset is.
func (a *App) shutdown(ctx context.Context) {
	if a.q != nil {
		if w, h := runtime.WindowGetSize(ctx); w >= minWindowWidth && h >= minWindowHeight {
			_ = a.saveWindowSize(w, h)
		}
	}
	if a.conn != nil {
		a.conn.Close()
	}
}

func (a *App) GetRole() (string, error) {
	role, err := a.q.GetSetting(a.ctx, settingLastRole)
	if errors.Is(err, sql.ErrNoRows) {
		return "", nil
	}
	if err != nil {
		return "", err
	}
	return role, nil
}

func (a *App) SetRole(role string) error {
	if role != RoleGM && role != RolePlayer {
		return fmt.Errorf("unknown role %q", role)
	}
	return a.q.SetSetting(a.ctx, db.SetSettingParams{Key: settingLastRole, Value: role})
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
