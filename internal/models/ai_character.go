// Created by DINKIssTyle on 2025. Copyright (C) 2025 DINKI'ssTyle. All rights reserved.

package models

import "time"

// AICharacter AI 캐릭터 모델
type AICharacter struct {
	ID                 int       `json:"id"`
	Nickname           string    `json:"nickname"`
	Gender             string    `json:"gender"`
	Age                int       `json:"age"`
	Birthdate          string    `json:"birthdate"` // YYYY-MM-DD 형식
	Region             string    `json:"region"`    // 거주 지역
	Hobby              string    `json:"hobby"`
	JobCategory        string    `json:"job_category"`
	MBTI               string    `json:"mbti"`
	AggressionLevel    int       `json:"aggression_level"`
	FormalityLevel     int       `json:"formality_level"`
	RoleplayLevel      int       `json:"roleplay_level"`
	AssignedModelIndex int       `json:"assigned_model_index"`
	PostCount          int       `json:"post_count"`      // 작성한 글 수
	CommentCount       int       `json:"comment_count"`   // 작성한 댓글 수
	PersonaSummary     string    `json:"persona_summary"` // 인격 요약 (1000자 이내)
	PersonaUpdatedAt   time.Time `json:"persona_updated_at"`
	IsActive           bool      `json:"is_active"`
	CreatedAt          time.Time `json:"created_at"`
}

// 기본값은 defaults.go에 정의되어 있습니다:
// - JobCategories (직종 목록)
// - Hobbies (취미 목록)
// - MBTITypes (MBTI 목록)
// - Regions (지역 목록)
