package main

import (
	"embed"
	"log"

	"github.com/mhetem/DH-Companion/internal/dice"
	"github.com/mhetem/DH-Companion/internal/gm"
	"github.com/mhetem/DH-Companion/internal/player"
	"github.com/wailsapp/wails/v3/pkg/application"
)

//go:embed all:frontend/dist
var assets embed.FS

var version = "dev"

func main() {
	gmSvc := gm.New()
	playerSvc := player.New()
	app := NewApp(gmSvc, playerSvc)

	app.wails = application.New(application.Options{
		Name:        "Hilt",
		Description: "A Daggerheart companion for both sides of the table.",
		Services: []application.Service{
			application.NewService(app),
			application.NewService(gmSvc),
			application.NewService(playerSvc),
			application.NewService(dice.NewRoller()),
		},
		Assets: application.AssetOptions{
			Handler:    application.AssetFileServerFS(assets),
			Middleware: app.serveMedia,
		},
		Mac: application.MacOptions{
			ApplicationShouldTerminateAfterLastWindowClosed: true,
		},
		Linux: application.LinuxOptions{
			ProgramName: "DH-Companion",
		},
	})

	if err := app.wails.Run(); err != nil {
		log.Fatal(err)
	}
}
