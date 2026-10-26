package main

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"

	"github.com/mhetem/DH-Companion/internal/gm"
	"github.com/wailsapp/wails/v3/pkg/application"
	"github.com/wailsapp/wails/v3/pkg/events"
)

const (
	tableWindowName  = "table"
	eventTableShow   = "table:show"
	eventTableWindow = "table:window"

	tableWidth  = 1280
	tableHeight = 720
)

type TableDisplay struct {
	HandoutID int64  `json:"handoutId"`
	Kind      string `json:"kind"`
	Title     string `json:"title"`
	Body      string `json:"body"`
	MediaURL  string `json:"mediaUrl"`
	MediaType string `json:"mediaType"`
}

type TableWindow struct {
	Open       bool `json:"open"`
	Fullscreen bool `json:"fullscreen"`
	Placeable  bool `json:"placeable"`
}

type TableScreen struct {
	ID      string `json:"id"`
	Label   string `json:"label"`
	Width   int    `json:"width"`
	Height  int    `json:"height"`
	Primary bool   `json:"primary"`
	Here    bool   `json:"here"`
}

func init() {
	application.RegisterEvent[TableDisplay](eventTableShow)
	application.RegisterEvent[TableWindow](eventTableWindow)
}

func canPlaceWindows() bool {
	if runtime.GOOS != "linux" || os.Getenv("GDK_BACKEND") == "x11" {
		return true
	}
	return os.Getenv("WAYLAND_DISPLAY") == ""
}

func (a *App) TableStatus() TableWindow {
	a.tableMu.Lock()
	w := a.table
	a.tableMu.Unlock()
	status := TableWindow{Placeable: canPlaceWindows()}
	if w != nil {
		status.Open = true
		status.Fullscreen = w.IsFullscreen()
	}
	return status
}

func (a *App) TableScreens() []TableScreen {
	here := a.mainScreenID()
	screens := a.wails.Screen.GetAll()
	out := make([]TableScreen, 0, len(screens))
	for i, s := range screens {
		out = append(out, TableScreen{
			ID:      s.ID,
			Label:   fmt.Sprintf("Screen %d", i+1),
			Width:   s.Size.Width,
			Height:  s.Size.Height,
			Primary: s.IsPrimary,
			Here:    s.ID == here,
		})
	}
	return out
}

func (a *App) OpenTableWindow(screenID string) TableWindow {
	a.tableMu.Lock()
	if w := a.table; w != nil {
		a.tableMu.Unlock()
		w.Show()
		w.Focus()
		return a.TableStatus()
	}

	screen := a.tableScreen(screenID)
	away := screen != nil && screen.ID != a.mainScreenID()
	width, height := tableWidth, tableHeight
	if screen != nil {
		width, height = min(width, screen.WorkArea.Width), min(height, screen.WorkArea.Height)
	}
	w := a.wails.Window.NewWithOptions(application.WebviewWindowOptions{
		Name:             tableWindowName,
		Title:            "Hilt — Player window",
		Width:            width,
		Height:           height,
		Screen:           screen,
		InitialPosition:  application.WindowCentered,
		BackgroundColour: application.NewRGB(0, 0, 0),
		URL:              "/?table",
	})
	a.table = w
	a.tableMu.Unlock()

	placeable := canPlaceWindows()
	w.OnWindowEvent(events.Common.WindowClosing, func(*application.WindowEvent) {
		a.tableMu.Lock()
		if a.table == w {
			a.table = nil
		}
		a.tableMu.Unlock()
		a.wails.Event.Emit(eventTableWindow, TableWindow{Placeable: placeable})
	})
	w.OnWindowEvent(events.Common.WindowFullscreen, func(*application.WindowEvent) {
		a.wails.Event.Emit(eventTableWindow, TableWindow{Open: true, Fullscreen: true, Placeable: placeable})
	})
	w.OnWindowEvent(events.Common.WindowUnFullscreen, func(*application.WindowEvent) {
		a.wails.Event.Emit(eventTableWindow, TableWindow{Open: true, Placeable: placeable})
	})
	if away {
		var once sync.Once
		w.OnWindowEvent(events.Common.WindowRuntimeReady, func(*application.WindowEvent) {
			once.Do(func() { w.Fullscreen() })
		})
	}

	status := TableWindow{Open: true, Placeable: placeable}
	a.wails.Event.Emit(eventTableWindow, status)
	return status
}

func (a *App) CloseTableWindow() {
	a.tableMu.Lock()
	w := a.table
	a.tableMu.Unlock()
	if w != nil {
		w.Close()
	}
}

func (a *App) ToggleTableFullscreen() {
	a.tableMu.Lock()
	w := a.table
	a.tableMu.Unlock()
	if w != nil {
		w.ToggleFullscreen()
	}
}

func (a *App) TableNow() TableDisplay {
	a.tableMu.Lock()
	defer a.tableMu.Unlock()
	return a.showing
}

func (a *App) ShowHandout(id int64) (TableDisplay, error) {
	h, err := a.gm.GetHandout(id)
	if err != nil {
		return TableDisplay{}, err
	}
	if h.MediaMissing {
		return TableDisplay{}, fmt.Errorf("the file behind %q is missing from the media folder — choose it again to show it", h.Title)
	}
	display := TableDisplay{
		HandoutID: h.ID,
		Kind:      h.Kind,
		Title:     h.Title,
		MediaURL:  h.MediaURL,
		MediaType: h.MediaType,
	}
	if h.Kind == "text" {
		display.Body = h.Body
	}
	a.show(display)
	return display, nil
}

func (a *App) ClearTable() {
	a.show(TableDisplay{})
}

func (a *App) show(display TableDisplay) {
	a.tableMu.Lock()
	defer a.tableMu.Unlock()
	a.showing = display
	a.wails.Event.Emit(eventTableShow, display)
}

func (a *App) ChooseHandoutMedia() (string, error) {
	path, err := picked(a.wails.Dialog.OpenFileWithOptions(&application.OpenFileDialogOptions{
		Title:   "Choose an image or a video",
		Filters: []application.FileFilter{{DisplayName: "Images and video", Pattern: gm.MediaFilter()}},
		Window:  a.window,
	}).PromptForSingleSelection())
	if err != nil {
		return "", fmt.Errorf("choosing a file: %w", err)
	}
	return path, nil
}

func (a *App) tableScreen(id string) *application.Screen {
	if !canPlaceWindows() {
		return nil
	}
	screens := a.wails.Screen.GetAll()
	for _, s := range screens {
		if id != "" && s.ID == id {
			return s
		}
	}
	here := a.mainScreenID()
	for _, s := range screens {
		if s.ID != here {
			return s
		}
	}
	return nil
}

func (a *App) mainScreenID() string {
	if a.window == nil {
		return ""
	}
	if s, err := a.window.GetScreen(); err == nil && s != nil {
		return s.ID
	}
	return ""
}

func (a *App) serveMedia(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		file, ok := strings.CutPrefix(r.URL.Path, gm.MediaRoute)
		if !ok {
			next.ServeHTTP(w, r)
			return
		}
		typ, ok := gm.MediaContentType(file)
		if !ok || a.media == "" {
			http.NotFound(w, r)
			return
		}
		f, err := os.Open(filepath.Join(a.media, file))
		if err != nil {
			http.NotFound(w, r)
			return
		}
		defer f.Close()
		info, err := f.Stat()
		if err != nil {
			http.NotFound(w, r)
			return
		}
		w.Header().Set("Content-Type", typ)
		w.Header().Set("Cache-Control", "private, max-age=31536000, immutable")
		http.ServeContent(w, r, file, info.ModTime(), f)
	})
}
