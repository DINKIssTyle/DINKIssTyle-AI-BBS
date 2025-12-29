// Created by DINKIssTyle on 2025. Copyright (C) 2025 DINKI'ssTyle. All rights reserved.

package models

// Settings 앱 설정 모델
type Settings struct {
	// MySQL 설정
	DBHost     string `json:"db_host"`
	DBPort     string `json:"db_port"`
	DBUser     string `json:"db_user"`
	DBPassword string `json:"db_password"`

	// LLM 설정
	LLMHost         string `json:"llm_host"`
	LLMPort         string `json:"llm_port"`
	LLMModel        string `json:"llm_model"`
	PostsPerHour    int    `json:"posts_per_hour"`
	CommentsPerHour int    `json:"comments_per_hour"`

	// BBS 설정
	PostsPerPage string `json:"posts_per_page"`
	BgColor      string `json:"bg_color"`
	TextColor    string `json:"text_color"`
	FontFamily   string `json:"font_family"`
}

// LLMConfig LLM 연결 설정
type LLMConfig struct {
	Host            string  `json:"host"`
	Port            string  `json:"port"`
	Model1          string  `json:"model1"`
	Model2          string  `json:"model2"`
	Model3          string  `json:"model3"`
	PostsPerHour    int     `json:"posts_per_hour"`
	CommentsPerHour int     `json:"comments_per_hour"`
	MaxTokens       int     `json:"max_tokens"`
	Temperature     float64 `json:"temperature"`
}

// BBSConfig 게시판 설정 (타이틀, 테마, 폰트, 푸터)
type BBSConfig struct {
	Title        string `json:"title"`
	Footer       string `json:"footer"`
	Theme        string `json:"theme"` // blue, red, green, purple, gray
	Font         string `json:"font"`  // sans, serif
	PostsPerPage int    `json:"posts_per_page"`
	Timezone     string `json:"timezone"` // UTC+9, UTC+0, etc.
}
