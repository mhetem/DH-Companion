package main

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/mhetem/DH-Companion/internal/gm"
	"github.com/mhetem/DH-Companion/internal/player"
	"github.com/mhetem/DH-Companion/internal/srd"
	"github.com/mhetem/DH-Companion/internal/update"
	"github.com/wailsapp/wails/v2/pkg/runtime"
)

func stamp() string { return time.Now().Format("2006-01-02-1504") }

func (a *App) Version() string { return version }

func (a *App) CheckForUpdate() (update.Release, error) {
	return update.Check(a.ctx, version)
}

func (a *App) OpenReleasesPage(url string) {
	if url == "" {
		url = "https://github.com/mhetem/DH-Companion/releases"
	}
	runtime.BrowserOpenURL(a.ctx, url)
}

func (a *App) DataDirectory() (string, error) { return dataDir("") }

func (a *App) ExportDatabase() (string, error) {
	target, err := runtime.SaveFileDialog(a.ctx, runtime.SaveDialogOptions{
		Title:           "Save a database backup",
		DefaultFilename: fmt.Sprintf("hilt-backup-%s.db", stamp()),
		Filters:         []runtime.FileFilter{{DisplayName: "SQLite database (*.db)", Pattern: "*.db"}},
	})
	if err != nil {
		return "", fmt.Errorf("choosing a location: %w", err)
	}
	if target == "" {
		return "", nil
	}
	if err := a.snapshot(target); err != nil {
		return "", err
	}
	return target, nil
}

func (a *App) snapshot(target string) error {
	if a.conn == nil {
		return fmt.Errorf("the database isn't open")
	}
	if err := os.Remove(target); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("clearing %s: %w", target, err)
	}
	if _, err := a.conn.ExecContext(a.ctx, "VACUUM INTO ?", target); err != nil {
		return fmt.Errorf("writing the backup: %w", err)
	}
	return nil
}

func (a *App) ImportDatabase() (string, error) {
	source, err := runtime.OpenFileDialog(a.ctx, runtime.OpenDialogOptions{
		Title:   "Open a Hilt database",
		Filters: []runtime.FileFilter{{DisplayName: "SQLite database (*.db)", Pattern: "*.db"}},
	})
	if err != nil {
		return "", fmt.Errorf("choosing a file: %w", err)
	}
	if source == "" {
		return "", nil
	}
	if err := verifyDatabase(source); err != nil {
		return "", err
	}

	dir, err := dataDir("")
	if err != nil {
		return "", err
	}
	live := filepath.Join(dir, "data.db")
	backup := filepath.Join(dir, fmt.Sprintf("data-replaced-%s.db", stamp()))
	if err := a.snapshot(backup); err != nil {
		return "", fmt.Errorf("the current database could not be backed up, so nothing was replaced: %w", err)
	}

	a.conn.Close()
	a.q, a.conn = nil, nil

	if err := copyFile(source, live); err != nil {
		return "", fmt.Errorf("replacing the database: %w (your data is still at %s)", err, backup)
	}
	for _, sidecar := range []string{live + "-wal", live + "-shm"} {
		if err := os.Remove(sidecar); err != nil && !os.IsNotExist(err) {
			return "", fmt.Errorf("clearing %s: %w", sidecar, err)
		}
	}

	if err := a.reopen(); err != nil {
		return "", err
	}
	return backup, nil
}

func (a *App) reopen() error {
	q, conn, err := Open("")
	if err != nil {
		return fmt.Errorf("reopening the database: %w", err)
	}
	a.q, a.conn = q, conn

	catalog := a.catalog
	if catalog == nil {
		catalog, err = srd.Default()
		if err != nil {
			return fmt.Errorf("reference data error: %w", err)
		}
		a.catalog = catalog
	}
	gm.Attach(a.gm, a.ctx, q, conn, catalog)
	player.Attach(a.player, a.ctx, q, conn, catalog)
	if err := a.gm.ReindexCards(); err != nil {
		return fmt.Errorf("rebuilding the search index: %w", err)
	}
	return nil
}

func verifyDatabase(path string) error {
	conn, err := sql.Open("sqlite", path+"?_pragma=busy_timeout(2000)&mode=ro")
	if err != nil {
		return fmt.Errorf("that file could not be opened as a database")
	}
	defer conn.Close()
	if err := conn.Ping(); err != nil {
		return fmt.Errorf("that file is not a SQLite database")
	}

	var tables int
	err = conn.QueryRow(
		"SELECT count(*) FROM sqlite_master WHERE type = 'table' AND name IN ('settings', 'goose_db_version')",
	).Scan(&tables)
	if err != nil {
		return fmt.Errorf("that file could not be read as a database")
	}
	if tables < 2 {
		return fmt.Errorf("that database wasn't made by Hilt")
	}
	return nil
}

func copyFile(src, dst string) error {
	data, err := os.ReadFile(src)
	if err != nil {
		return err
	}
	return os.WriteFile(dst, data, 0o644)
}

func (a *App) ExportLibrary() (string, error) {
	payload, err := a.gm.ExportLibraryJSON()
	if err != nil {
		return "", err
	}
	target, err := runtime.SaveFileDialog(a.ctx, runtime.SaveDialogOptions{
		Title:           "Export your library",
		DefaultFilename: fmt.Sprintf("hilt-library-%s.json", stamp()),
		Filters:         []runtime.FileFilter{{DisplayName: "JSON (*.json)", Pattern: "*.json"}},
	})
	if err != nil {
		return "", fmt.Errorf("choosing a location: %w", err)
	}
	if target == "" {
		return "", nil
	}
	if err := os.WriteFile(target, []byte(payload), 0o644); err != nil {
		return "", fmt.Errorf("writing the export: %w", err)
	}
	return target, nil
}

func (a *App) ImportLibrary() (*gm.ImportReport, error) {
	source, err := runtime.OpenFileDialog(a.ctx, runtime.OpenDialogOptions{
		Title:   "Import a Hilt library",
		Filters: []runtime.FileFilter{{DisplayName: "JSON (*.json)", Pattern: "*.json"}},
	})
	if err != nil {
		return nil, fmt.Errorf("choosing a file: %w", err)
	}
	if source == "" {
		return nil, nil
	}
	raw, err := os.ReadFile(source)
	if err != nil {
		return nil, fmt.Errorf("reading the export: %w", err)
	}
	report, err := a.gm.ImportLibraryJSON(string(raw))
	if err != nil {
		return nil, err
	}
	return &report, nil
}
