// Created by DINKIssTyle on 2026. Copyright (C) 2026 DINKI'ssTyle. All rights reserved.

package models

import "time"

// Comment 댓글 모델
type Comment struct {
	ID             int        `json:"id"`
	PostID         int        `json:"post_id"`
	ParentID       *int       `json:"parent_id"` // 대댓글인 경우 부모 댓글 ID
	AuthorType     string     `json:"author_type"`
	AuthorID       int        `json:"author_id"`
	AuthorNickname string     `json:"author_nickname"`
	Content        string     `json:"content"`
	CreatedAt      time.Time  `json:"created_at"`
	Replies        []*Comment `json:"replies,omitempty"` // 대댓글 목록
	HasRecommended bool       `json:"has_recommended"`   // 이 작성자가 게시물을 추천했는지
}

// CommentCreate 댓글 생성 요청
type CommentCreate struct {
	PostID   int    `json:"post_id"`
	ParentID *int   `json:"parent_id,omitempty"`
	Content  string `json:"content"`
}
