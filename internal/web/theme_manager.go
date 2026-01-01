// Created by DINKIssTyle on 2026. Copyright (C) 2026 DINKI'ssTyle. All rights reserved.

package web

import (
	"aibbs/internal/models"
	"net/http"
	"strings"
)

// DeviceType 브라우저 기기 유형
type DeviceType string

const (
	DeviceOld     DeviceType = "old"
	DeviceUnified DeviceType = "unified" // 반응형 템플릿 (데스크톱/모바일 통합)
)

// ThemeColors 테마 색상 정의
type ThemeColors struct {
	BgColor       string
	TextColor     string
	LinkColor     string
	VLinkColor    string
	ALinkColor    string
	TableBgColor  string
	HeaderBgColor string
	BorderColor   string
	PointColor    string // 강조색
}

// ThemeManager 테마 및 디바이스 감지 관리자
type ThemeManager struct {
	config models.BBSConfig
}

func NewThemeManager(config models.BBSConfig) *ThemeManager {
	return &ThemeManager{config: config}
}

func (tm *ThemeManager) SetConfig(config models.BBSConfig) {
	tm.config = config
}

// DetectDevice User-Agent 기반 기기 감지 (Old 브라우저만 분리)
func (tm *ThemeManager) DetectDevice(r *http.Request) DeviceType {
	ua := r.Header.Get("User-Agent")
	if ua == "" {
		return DeviceOld
	}

	uaLower := strings.ToLower(ua)

	// Old 브라우저 감지 (Netscape, MSIE 9 미만 등)
	if strings.Contains(uaLower, "mozilla/2.0") || strings.Contains(uaLower, "mozilla/3.0") ||
		strings.Contains(uaLower, "mozilla/4.0") || strings.Contains(uaLower, "msie") {
		return DeviceOld
	}

	// 나머지는 모두 통합 반응형 템플릿 사용
	return DeviceUnified
}

// GetColors 현재 설정된 테마의 색상 반환
func (tm *ThemeManager) GetColors(theme string) ThemeColors {
	if theme == "" {
		theme = tm.config.Theme
	}

	switch theme {
	case "red":
		return ThemeColors{
			BgColor: "#330000", TextColor: "#FFFFFF", LinkColor: "#FF6666", VLinkColor: "#FF9999", ALinkColor: "#FFCC00",
			TableBgColor: "#441111", HeaderBgColor: "#660000", BorderColor: "#880000", PointColor: "#FF4444",
		}
	case "green":
		return ThemeColors{
			BgColor: "#002200", TextColor: "#FFFFFF", LinkColor: "#66FF66", VLinkColor: "#99FF99", ALinkColor: "#FFFF00",
			TableBgColor: "#113311", HeaderBgColor: "#004400", BorderColor: "#006600", PointColor: "#44FF44",
		}
	case "purple":
		return ThemeColors{
			BgColor: "#220022", TextColor: "#FFFFFF", LinkColor: "#FF66FF", VLinkColor: "#FF99FF", ALinkColor: "#FFFF00",
			TableBgColor: "#331133", HeaderBgColor: "#440044", BorderColor: "#660066", PointColor: "#FF44FF",
		}
	case "gray":
		return ThemeColors{
			BgColor: "#222222", TextColor: "#CCCCCC", LinkColor: "#AAAAAA", VLinkColor: "#888888", ALinkColor: "#FFFFFF",
			TableBgColor: "#333333", HeaderBgColor: "#444444", BorderColor: "#555555", PointColor: "#DDDDDD",
		}
	default: // blue (default)
		return ThemeColors{
			BgColor: "#001B33", TextColor: "#FFFFFF", LinkColor: "#00BFFF", VLinkColor: "#87CEEB", ALinkColor: "#FFD700",
			TableBgColor: "#002244", HeaderBgColor: "#003366", BorderColor: "#004488", PointColor: "#00BFFF",
		}
	}
}

// GetFontFace 폰트 페이스 반환
func (tm *ThemeManager) GetFontFace() string {
	if tm.config.Font == "serif" {
		return "Batang"
	}
	return "Gulim"
}
