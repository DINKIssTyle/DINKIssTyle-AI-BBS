// Created by DINKIssTyle on 2025. Copyright (C) 2025 DINKI'ssTyle. All rights reserved.

package main

import (
	"embed"

	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
	"github.com/wailsapp/wails/v2/pkg/options/windows"
)

//go:embed all:frontend/dist
var assets embed.FS

func main() {
	println("[DEBUG] Main function started")
	// 앱 인스턴스 생성
	app := NewApp()
	println("[DEBUG] App instance created")

	// Wails 앱 실행
	err := wails.Run(&options.App{
		Title:     "DINKIssTyle AI BBS",
		Width:     1330,
		Height:    768,
		MinWidth:  800,
		MinHeight: 600,
		AssetServer: &assetserver.Options{
			Assets: assets,
		},
		BackgroundColour: &options.RGBA{R: 0, G: 27, B: 51, A: 1}, // #001B33
		OnStartup:        app.startup,
		OnShutdown:       app.shutdown,
		Bind: []interface{}{
			app,
		},
		Windows: &windows.Options{
			WebviewIsTransparent: false,
			WindowIsTranslucent:  false,
			DisableWindowIcon:    false,
			WebviewUserDataPath:  "./webview_data",
		},
	})

	if err != nil {
		println("Error:", err.Error())
	}
}
