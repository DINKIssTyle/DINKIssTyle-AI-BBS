// Created by DINKIssTyle on 2025. Copyright (C) 2025 DINKI'ssTyle. All rights reserved.

package main

import (
	"embed"

	"flag"

	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
	"github.com/wailsapp/wails/v2/pkg/options/windows"
)

//go:embed all:frontend/dist
var frontendAssets embed.FS

func main() {
	// 플래그 파싱
	modePtr := flag.String("mode", "main", "Application mode: main or char_manager")
	flag.Parse()

	println("[DEBUG] Main function started. Mode:", *modePtr)

	// 앱 인스턴스 생성
	app := NewApp(*modePtr)
	println("[DEBUG] App instance created")

	// 윈도우 크기 및 제목 설정
	title := "DINKI'ssTyle AI BBS"
	width := 900
	height := 700
	minWidth := 900
	minHeight := 700

	if *modePtr == "char_manager" {
		title = "AI 캐릭터 관리자 - DINKI'ssTyle AI BBS"
		width = 1600
		height = 900
		minWidth = 1600
		minHeight = 800
	}

	// Wails 앱 실행
	err := wails.Run(&options.App{
		Title:     title,
		Width:     width,
		Height:    height,
		MinWidth:  minWidth,
		MinHeight: minHeight,
		AssetServer: &assetserver.Options{
			Assets: frontendAssets,
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
