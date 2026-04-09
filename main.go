package main

import (
	"embed"
	"fmt"

	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/menu"
	"github.com/wailsapp/wails/v2/pkg/menu/keys"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
	"github.com/wailsapp/wails/v2/pkg/options/mac"
)

// Version is set via -ldflags at build time
var version = "dev"

//go:embed all:frontend/dist
var assets embed.FS

func main() {
	app := NewApp()
	appMenu := createMenu(app)

	err := wails.Run(&options.App{
		Title:            "Hey Nanobot 🐈",
		Width:            1200,
		Height:           800,
		MinWidth:         800,
		MinHeight:        600,
		BackgroundColour: &options.RGBA{R: 30, G: 30, B: 30, A: 255},
		AssetServer: &assetserver.Options{
			Assets: assets,
		},
		OnStartup:     app.startup,
		OnDomReady:    app.domReady,
		OnShutdown:    app.shutdown,
		OnBeforeClose: app.beforeClose,
		Menu:          appMenu,
		Bind: []interface{}{
			app,
		},
		Mac: &mac.Options{
			TitleBar: &mac.TitleBar{
				TitlebarAppearsTransparent: true,
				HideTitle:                  true,
				HideTitleBar:               false,
				FullSizeContent:            true,
				UseToolbar:                 false,
				HideToolbarSeparator:       true,
			},
			About: &mac.AboutInfo{
				Title:   "Hey Nanobot",
				Message: fmt.Sprintf("Personal AI Assistant 🐈\n\nMulti-Bot Desktop Client\nv%s", version),
			},
			WebviewIsTransparent: false,
			WindowIsTranslucent:  false,
		},
		SingleInstanceLock: &options.SingleInstanceLock{
			UniqueId: "hey-nanobot-e3b0c44298fc1c14",
		},
	})

	if err != nil {
		println("Error:", err.Error())
	}
}

func createMenu(app *App) *menu.Menu {
	appMenu := menu.NewMenu()

	// App menu
	fileMenu := appMenu.AddSubmenu("Hey Nanobot")
	fileMenu.AddText("About Hey Nanobot", keys.CmdOrCtrl(""), func(_ *menu.CallbackData) {})
	fileMenu.AddSeparator()
	fileMenu.AddText("Preferences...", keys.CmdOrCtrl(","), func(_ *menu.CallbackData) {
		app.NavigateTo("config")
	})
	fileMenu.AddSeparator()
	fileMenu.AddText("Hide nanobot", keys.CmdOrCtrl("h"), func(_ *menu.CallbackData) {
		app.HideWindow()
	})
	fileMenu.AddSeparator()
	fileMenu.AddText("Quit nanobot", keys.CmdOrCtrl("q"), func(_ *menu.CallbackData) {
		app.Quit()
	})

	// View menu
	viewMenu := appMenu.AddSubmenu("View")
	viewMenu.AddText("Chat", keys.CmdOrCtrl("1"), func(_ *menu.CallbackData) {
		app.NavigateTo("chat")
	})
	viewMenu.AddText("Channels", keys.CmdOrCtrl("2"), func(_ *menu.CallbackData) {
		app.NavigateTo("channels")
	})
	viewMenu.AddText("Sessions", keys.CmdOrCtrl("3"), func(_ *menu.CallbackData) {
		app.NavigateTo("sessions")
	})
	viewMenu.AddText("Providers", keys.CmdOrCtrl("4"), func(_ *menu.CallbackData) {
		app.NavigateTo("providers")
	})
	viewMenu.AddText("Config", keys.CmdOrCtrl("5"), func(_ *menu.CallbackData) {
		app.NavigateTo("config")
	})
	viewMenu.AddText("Gateway", keys.CmdOrCtrl("6"), func(_ *menu.CallbackData) {
		app.NavigateTo("gateway")
	})
	viewMenu.AddSeparator()
	viewMenu.AddText("Toggle Full Screen", keys.Key("F11"), func(_ *menu.CallbackData) {
		app.ToggleFullscreen()
	})

	// Window menu
	winMenu := appMenu.AddSubmenu("Window")
	winMenu.AddText("Minimize", keys.CmdOrCtrl("m"), func(_ *menu.CallbackData) {
		app.HideWindow()
	})

	// Help menu
	helpMenu := appMenu.AddSubmenu("Help")
	helpMenu.AddText("nanobot Documentation", nil, func(_ *menu.CallbackData) {
		app.OpenURL("https://github.com/HKUDS/nanobot")
	})

	return appMenu
}
