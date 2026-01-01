// Created by DINKIssTyle on 2026. Copyright (C) 2026 DINKI'ssTyle. All rights reserved.

package models

import "time"

// User 사용자 모델
type User struct {
	ID           int       `json:"id"`
	Username     string    `json:"username"`
	PasswordHash string    `json:"-"`
	Nickname     string    `json:"nickname"`
	IsAdmin      bool      `json:"is_admin"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
	// 사용자 설정
	Theme        string `json:"theme"`          // 'dark', 'light'
	FontStyle    string `json:"font_style"`     // 'default', 'serif', 'monospace'
	Timezone     string `json:"timezone"`       // 'Asia/Seoul' 등
	PostsPerPage int    `json:"posts_per_page"` // 기본 20
}

// UserCreate 사용자 생성 요청
type UserCreate struct {
	Username string `json:"username"`
	Password string `json:"password"`
	Nickname string `json:"nickname"`
}

// UserLogin 로그인 요청
type UserLogin struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

// PasswordChange 비밀번호 변경 요청
type PasswordChange struct {
	OldPassword string `json:"old_password"`
	NewPassword string `json:"new_password"`
}
