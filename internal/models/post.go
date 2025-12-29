// Created by DINKIssTyle on 2025. Copyright (C) 2025 DINKI'ssTyle. All rights reserved.

package models

import "time"

// Post 게시물 모델
type Post struct {
	ID             int       `json:"id"`
	AuthorType     string    `json:"author_type"` // "user" 또는 "ai"
	AuthorID       int       `json:"author_id"`
	AuthorNickname string    `json:"author_nickname"` // JOIN으로 가져오는 필드
	Title          string    `json:"title"`
	Content        string    `json:"content"`
	ViewCount      int       `json:"view_count"`
	RecommendCount int       `json:"recommend_count"`
	CommentCount   int       `json:"comment_count"` // JOIN으로 가져오는 필드
	IsPinned       bool      `json:"is_pinned"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}

// PostCreate 게시물 생성 요청
type PostCreate struct {
	Title    string `json:"title"`
	Content  string `json:"content"`
	IsPinned bool   `json:"is_pinned"`
}

// PostUpdate 게시물 수정 요청
type PostUpdate struct {
	Title    string `json:"title"`
	Content  string `json:"content"`
	IsPinned bool   `json:"is_pinned"`
}

// PostList 게시물 목록 응답
type PostList struct {
	Posts      []Post `json:"posts"`
	TotalCount int    `json:"total_count"`
	Page       int    `json:"page"`
	PerPage    int    `json:"per_page"`
	TotalPages int    `json:"total_pages"`
}
