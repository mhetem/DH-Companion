package main

import (
	"embed"

	"github.com/mhetem/DH-Companion/internal/dice"
	"github.com/mhetem/DH-Companion/internal/gm"
	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
)

//go:embed all:frontend/dist
var assets embed.FS

func main() {

	// Each module is bound as its own struct, so the frontend reaches the GM
	// methods at window.go.gm.Service.* instead of piling them onto App.
	gmSvc := gm.New()
	// The roller holds no state and needs no startup wiring — it's bound purely
	// to reach the pure roll functions from JS at window.go.dice.Roller.*.
	roller := dice.NewRoller()
	app := NewApp(gmSvc)

	err := wails.Run(&options.App{
		Title:     "DH-Companion",
		Width:     1920,
		Height:    1080,
		MinWidth:  minWindowWidth,
		MinHeight: minWindowHeight,
		AssetServer: &assetserver.Options{
			Assets: assets,
		},
		BackgroundColour: &options.RGBA{R: 27, G: 38, B: 54, A: 1},
		OnStartup:        app.startup,
		OnShutdown:       app.shutdown,
		Bind: []interface{}{
			app,
			gmSvc,
			roller,
		},
	})

	if err != nil {
		println("Error:", err.Error())
	}
}
