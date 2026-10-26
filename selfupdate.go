package main

import (
	_ "embed"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/wailsapp/wails/v3/pkg/updater"
	"github.com/wailsapp/wails/v3/pkg/updater/providers/endpoint"
)

const manifestURL = "https://github.com/mhetem/DH-Companion/releases/latest/download/manifest.json"

//go:embed build/updater.pub
var updaterKey []byte

type SelfUpdate struct {
	Supported bool   `json:"supported"`
	Reason    string `json:"reason"`
}

func (a *App) initUpdater() error {
	if version == "dev" {
		return errors.New("this is an unreleased build, so there's nothing to update it to")
	}
	provider, err := endpoint.New(endpoint.Config{URL: manifestURL})
	if err != nil {
		return err
	}
	return a.wails.Updater.Init(updater.Config{
		CurrentVersion: strings.TrimPrefix(version, "v"),
		Providers:      []updater.Provider{provider},
		PublicKey:      updaterKey,
		Window:         updater.WindowNone,
	})
}

func (a *App) SelfUpdateSupport() SelfUpdate {
	if err := a.selfUpdateBlocker(); err != nil {
		return SelfUpdate{Reason: err.Error()}
	}
	return SelfUpdate{Supported: true}
}

func (a *App) InstallUpdate() error {
	if err := a.selfUpdateBlocker(); err != nil {
		return err
	}
	rel, err := a.wails.Updater.Check(a.ctx)
	if err != nil {
		return fmt.Errorf("checking for the update: %w", err)
	}
	if rel == nil {
		return errors.New("the latest release can't be installed from inside the app — use the download page")
	}
	if err := a.wails.Updater.DownloadAndInstall(a.ctx); err != nil {
		return fmt.Errorf("installing %s: %w", rel.Version, err)
	}
	return a.wails.Updater.Restart(a.ctx)
}

func (a *App) selfUpdateBlocker() error {
	if a.updaterErr != nil {
		return a.updaterErr
	}
	if runtime.GOOS == "linux" && os.Getenv("APPIMAGE") == "" {
		return errors.New("this copy came from the .deb or the tarball, so update it the same way")
	}
	target, err := updateTarget()
	if err != nil {
		return fmt.Errorf("couldn't find where Hilt is installed: %w", err)
	}
	probe, err := os.MkdirTemp(filepath.Dir(target), ".hilt-update-check-*")
	if err != nil {
		return errors.New("the folder Hilt is installed in isn't writable, so it can't replace itself")
	}
	_ = os.Remove(probe)
	return nil
}

func updateTarget() (string, error) {
	exe, err := os.Executable()
	if err != nil {
		return "", err
	}
	img, mount := os.Getenv("APPIMAGE"), os.Getenv("APPDIR")
	if img != "" && mount != "" && strings.HasPrefix(exe, filepath.Clean(mount)+string(filepath.Separator)) {
		return img, nil
	}
	if i := strings.Index(exe, ".app"+string(filepath.Separator)); i >= 0 {
		return exe[:i+len(".app")], nil
	}
	return exe, nil
}
