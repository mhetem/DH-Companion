package main

import (
	"context"
	"database/sql"
	"embed"
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/mhetem/DH-Companion/internal/db"
	"github.com/pressly/goose/v3"
	"github.com/wailsapp/wails/v2/pkg/runtime"
	_ "modernc.org/sqlite"
)

//go:embed sql/schema/*.sql
var migrations embed.FS

const (
	RoleGM     = "gm"
	RolePlayer = "player"

	settingLastRole = "last_role"
)

type App struct {
	ctx  context.Context
	q    *db.Queries
	conn *sql.DB
}

func NewApp() *App {
	return &App{}
}

func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
	q, conn, err := Open("")
	if err != nil {
		runtime.MessageDialog(ctx, runtime.MessageDialogOptions{
			Type: runtime.ErrorDialog, Title: "Database error", Message: err.Error(),
		})
		runtime.Quit(ctx)
		return
	}
	a.q, a.conn = q, conn
}

func (a *App) shutdown(ctx context.Context) {
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

func Open(dir string) (*db.Queries, *sql.DB, error) {
	if dir == "" {
		if d := os.Getenv("DH_DATA_DIR"); d != "" {
			dir = d
		} else {
			home, err := os.UserHomeDir()
			if err != nil {
				return nil, nil, err
			}
			dir = filepath.Join(home, "DH-Companion")
		}
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
