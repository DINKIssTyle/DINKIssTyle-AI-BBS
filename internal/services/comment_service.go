// Created by DINKIssTyle on 2025. Copyright (C) 2025 DINKI'ssTyle. All rights reserved.

package services

import (
	"aibbs/internal/database"
	"aibbs/internal/models"
	"database/sql"
	"errors"
	"fmt"
)

// CommentService 댓글 서비스
type CommentService struct {
	db          *database.Database
	userService *UserService
}

// NewCommentService 새 댓글 서비스 생성
func NewCommentService(db *database.Database, userService *UserService) *CommentService {
	return &CommentService{db: db, userService: userService}
}

// DeleteCommentForWeb 댓글 삭제 (웹 전용 - 권한 체크 생략)
func (s *CommentService) DeleteCommentForWeb(id int) error {
	db := s.db.GetDB()
	if db == nil {
		return errors.New("데이터베이스에 연결되지 않았습니다")
	}

	_, err := db.Exec("DELETE FROM comments WHERE id = ?", id)
	if err != nil {
		return fmt.Errorf("댓글 삭제 실패: %w", err)
	}

	return nil
}

// UpdateCommentForWeb 댓글 수정 (웹 전용 - 권한 체크 생략)
func (s *CommentService) UpdateCommentForWeb(id int, content string) error {
	db := s.db.GetDB()
	if db == nil {
		return errors.New("데이터베이스에 연결되지 않았습니다")
	}

	_, err := db.Exec("UPDATE comments SET content = ? WHERE id = ?", content, id)
	if err != nil {
		return fmt.Errorf("댓글 수정 실패: %w", err)
	}
	return nil
}

// GetComments 게시물의 댓글 목록 조회
func (s *CommentService) GetComments(postID int) ([]*models.Comment, error) {
	db := s.db.GetDB()
	if db == nil {
		return nil, errors.New("데이터베이스에 연결되지 않았습니다")
	}

	// 모든 댓글 조회
	rows, err := db.Query(`
		SELECT 
			c.id, c.post_id, c.parent_id, c.author_type, c.author_id, c.content, c.created_at,
			CASE 
				WHEN c.author_type = 'user' THEN (SELECT nickname FROM users WHERE id = c.author_id)
				ELSE (SELECT nickname FROM ai_characters WHERE id = c.author_id)
			END as author_nickname
		FROM comments c
		WHERE c.post_id = ?
		ORDER BY c.created_at ASC
	`, postID)
	if err != nil {
		return nil, fmt.Errorf("댓글 조회 실패: %w", err)
	}
	defer rows.Close()

	commentMap := make(map[int]*models.Comment)
	var allComments []*models.Comment  // 순서 보장을 위한 슬라이스
	var rootComments []*models.Comment // 결과 반환용 슬라이스

	for rows.Next() {
		c := &models.Comment{}
		var parentID sql.NullInt64
		var authorNickname sql.NullString

		err := rows.Scan(&c.ID, &c.PostID, &parentID, &c.AuthorType, &c.AuthorID,
			&c.Content, &c.CreatedAt, &authorNickname)
		if err != nil {
			continue
		}

		if parentID.Valid {
			pid := int(parentID.Int64)
			c.ParentID = &pid
		}
		if authorNickname.Valid {
			c.AuthorNickname = authorNickname.String
		}

		commentMap[c.ID] = c
		allComments = append(allComments, c)
	}

	// 댓글 트리 구성 (allComments 순회로 순서 보장)
	for _, c := range allComments {
		if c.ParentID == nil {
			rootComments = append(rootComments, c)
		} else {
			parent, exists := commentMap[*c.ParentID]
			if exists {
				parent.Replies = append(parent.Replies, c)
			}
		}
	}

	return rootComments, nil
}

// CreateComment 댓글 작성 (사용자)
func (s *CommentService) CreateComment(postID int, content string, parentID *int) (*models.Comment, error) {
	user := s.userService.GetCurrentUser()
	if user == nil {
		return nil, errors.New("로그인이 필요합니다")
	}

	return s.CreateCommentByAuthor("user", user.ID, postID, content, parentID)
}

// CreateCommentByAuthor 댓글 작성 (작성자 지정)
func (s *CommentService) CreateCommentByAuthor(authorType string, authorID, postID int, content string, parentID *int) (*models.Comment, error) {
	if content == "" {
		return nil, errors.New("댓글 내용을 입력해주세요")
	}

	db := s.db.GetDB()
	if db == nil {
		return nil, errors.New("데이터베이스에 연결되지 않았습니다")
	}

	// 게시물 존재 확인
	var count int
	err := db.QueryRow("SELECT COUNT(*) FROM posts WHERE id = ?", postID).Scan(&count)
	if err != nil || count == 0 {
		return nil, errors.New("게시물을 찾을 수 없습니다")
	}

	// 부모 댓글 존재 확인 (대댓글인 경우)
	if parentID != nil {
		err := db.QueryRow("SELECT COUNT(*) FROM comments WHERE id = ? AND post_id = ?", *parentID, postID).Scan(&count)
		if err != nil || count == 0 {
			return nil, errors.New("부모 댓글을 찾을 수 없습니다")
		}
	}

	var result sql.Result
	if parentID != nil {
		result, err = db.Exec(`
			INSERT INTO comments (post_id, parent_id, author_type, author_id, content)
			VALUES (?, ?, ?, ?, ?)
		`, postID, *parentID, authorType, authorID, content)
	} else {
		result, err = db.Exec(`
			INSERT INTO comments (post_id, author_type, author_id, content)
			VALUES (?, ?, ?, ?)
		`, postID, authorType, authorID, content)
	}

	if err != nil {
		return nil, fmt.Errorf("댓글 작성 실패: %w", err)
	}

	id, _ := result.LastInsertId()

	// 작성한 댓글 조회
	c := &models.Comment{}
	var pID sql.NullInt64
	var authorNickname sql.NullString

	err = db.QueryRow(`
		SELECT 
			c.id, c.post_id, c.parent_id, c.author_type, c.author_id, c.content, c.created_at,
			CASE 
				WHEN c.author_type = 'user' THEN (SELECT nickname FROM users WHERE id = c.author_id)
				ELSE (SELECT nickname FROM ai_characters WHERE id = c.author_id)
			END as author_nickname
		FROM comments c
		WHERE c.id = ?
	`, id).Scan(&c.ID, &c.PostID, &pID, &c.AuthorType, &c.AuthorID,
		&c.Content, &c.CreatedAt, &authorNickname)

	if err != nil {
		return nil, fmt.Errorf("댓글 조회 실패: %w", err)
	}

	if pID.Valid {
		pid := int(pID.Int64)
		c.ParentID = &pid
	}
	if authorNickname.Valid {
		c.AuthorNickname = authorNickname.String
	}

	return c, nil
}

// DeleteComment 댓글 삭제
func (s *CommentService) DeleteComment(id int) error {
	user := s.userService.GetCurrentUser()
	if user == nil {
		return errors.New("로그인이 필요합니다")
	}

	db := s.db.GetDB()
	if db == nil {
		return errors.New("데이터베이스에 연결되지 않았습니다")
	}

	// 권한 확인
	var authorType string
	var authorID int
	err := db.QueryRow("SELECT author_type, author_id FROM comments WHERE id = ?", id).Scan(&authorType, &authorID)
	if err != nil {
		return errors.New("댓글을 찾을 수 없습니다")
	}
	if authorType != "user" || authorID != user.ID {
		return errors.New("삭제 권한이 없습니다")
	}

	_, err = db.Exec("DELETE FROM comments WHERE id = ?", id)
	if err != nil {
		return fmt.Errorf("댓글 삭제 실패: %w", err)
	}

	return nil
}

// UpdateComment 댓글 수정
func (s *CommentService) UpdateComment(id int, content string) error {
	user := s.userService.GetCurrentUser()
	if user == nil {
		return errors.New("로그인이 필요합니다")
	}

	if content == "" {
		return errors.New("댓글 내용을 입력해주세요")
	}

	db := s.db.GetDB()
	if db == nil {
		return errors.New("데이터베이스에 연결되지 않았습니다")
	}

	// 권한 확인
	var authorType string
	var authorID int
	err := db.QueryRow("SELECT author_type, author_id FROM comments WHERE id = ?", id).Scan(&authorType, &authorID)
	if err != nil {
		return errors.New("댓글을 찾을 수 없습니다")
	}
	if authorType != "user" || authorID != user.ID {
		return errors.New("수정 권한이 없습니다")
	}

	_, err = db.Exec("UPDATE comments SET content = ? WHERE id = ?", content, id)
	if err != nil {
		return fmt.Errorf("댓글 수정 실패: %w", err)
	}

	return nil
}

// GetRecentCommentsByPost 게시물의 최근 댓글 조회
func (s *CommentService) GetRecentCommentsByPost(postID, limit int) ([]*models.Comment, error) {
	db := s.db.GetDB()
	if db == nil {
		return nil, errors.New("데이터베이스에 연결되지 않았습니다")
	}

	rows, err := db.Query(`
		SELECT 
			c.id, c.post_id, c.parent_id, c.author_type, c.author_id, c.content, c.created_at,
			CASE 
				WHEN c.author_type = 'user' THEN (SELECT nickname FROM users WHERE id = c.author_id)
				ELSE (SELECT nickname FROM ai_characters WHERE id = c.author_id)
			END as author_nickname
		FROM comments c
		WHERE c.post_id = ?
		ORDER BY c.created_at DESC
		LIMIT ?
	`, postID, limit)
	if err != nil {
		return nil, fmt.Errorf("댓글 조회 실패: %w", err)
	}
	defer rows.Close()

	var comments []*models.Comment
	for rows.Next() {
		c := &models.Comment{}
		var parentID sql.NullInt64
		var authorNickname sql.NullString

		err := rows.Scan(&c.ID, &c.PostID, &parentID, &c.AuthorType, &c.AuthorID,
			&c.Content, &c.CreatedAt, &authorNickname)
		if err != nil {
			continue
		}

		if parentID.Valid {
			pid := int(parentID.Int64)
			c.ParentID = &pid
		}
		if authorNickname.Valid {
			c.AuthorNickname = authorNickname.String
		}
		comments = append(comments, c)
	}

	return comments, nil
}

// GetRecentCommentsByCharacter AI 캐릭터의 최근 댓글 조회
func (s *CommentService) GetRecentCommentsByCharacter(characterID, limit int) ([]*models.Comment, error) {
	db := s.db.GetDB()
	if db == nil {
		return nil, errors.New("데이터베이스에 연결되지 않았습니다")
	}

	rows, err := db.Query(`
		SELECT 
			c.id, c.post_id, c.parent_id, c.author_type, c.author_id, c.content, c.created_at,
			(SELECT nickname FROM ai_characters WHERE id = c.author_id) as author_nickname
		FROM comments c
		WHERE c.author_type = 'ai' AND c.author_id = ?
		ORDER BY c.created_at DESC
		LIMIT ?
	`, characterID, limit)
	if err != nil {
		return nil, fmt.Errorf("댓글 조회 실패: %w", err)
	}
	defer rows.Close()

	var comments []*models.Comment
	for rows.Next() {
		c := &models.Comment{}
		var parentID sql.NullInt64
		var authorNickname sql.NullString

		err := rows.Scan(&c.ID, &c.PostID, &parentID, &c.AuthorType, &c.AuthorID,
			&c.Content, &c.CreatedAt, &authorNickname)
		if err != nil {
			continue
		}

		if parentID.Valid {
			pid := int(parentID.Int64)
			c.ParentID = &pid
		}
		if authorNickname.Valid {
			c.AuthorNickname = authorNickname.String
		}
		comments = append(comments, c)
	}

	return comments, nil
}

// GetCommentsOnCharacterPosts AI 캐릭터가 쓴 게시글에 달린 다른 사람의 댓글 조회
// 본인이 쓴 댓글은 제외하고, 본인의 답글이 없는 댓글만 반환
func (s *CommentService) GetCommentsOnCharacterPosts(characterID, limit int) ([]*models.Comment, error) {
	db := s.db.GetDB()
	if db == nil {
		return nil, errors.New("데이터베이스에 연결되지 않았습니다")
	}

	// AI 캐릭터가 쓴 게시글에 달린 댓글 중, 본인이 아닌 다른 사람이 쓴 댓글 조회
	// 그리고 해당 댓글에 본인의 답글이 없는 경우만 조회
	rows, err := db.Query(`
		SELECT 
			c.id, c.post_id, c.parent_id, c.author_type, c.author_id, c.content, c.created_at,
			CASE 
				WHEN c.author_type = 'user' THEN (SELECT nickname FROM users WHERE id = c.author_id)
				ELSE (SELECT nickname FROM ai_characters WHERE id = c.author_id)
			END as author_nickname,
			p.title as post_title
		FROM comments c
		INNER JOIN posts p ON c.post_id = p.id
		WHERE p.author_type = 'ai' AND p.author_id = ?  -- AI 캐릭터가 쓴 게시글
		  AND NOT (c.author_type = 'ai' AND c.author_id = ?)  -- 본인이 쓴 댓글 제외
		  AND NOT EXISTS (  -- 본인의 답글이 없는 댓글만
			SELECT 1 FROM comments r 
			WHERE r.parent_id = c.id AND r.author_type = 'ai' AND r.author_id = ?
		  )
		ORDER BY c.created_at DESC
		LIMIT ?
	`, characterID, characterID, characterID, limit)
	if err != nil {
		return nil, fmt.Errorf("댓글 조회 실패: %w", err)
	}
	defer rows.Close()

	var comments []*models.Comment
	for rows.Next() {
		c := &models.Comment{}
		var parentID sql.NullInt64
		var authorNickname sql.NullString
		var postTitle string

		err := rows.Scan(&c.ID, &c.PostID, &parentID, &c.AuthorType, &c.AuthorID,
			&c.Content, &c.CreatedAt, &authorNickname, &postTitle)
		if err != nil {
			continue
		}

		if parentID.Valid {
			pid := int(parentID.Int64)
			c.ParentID = &pid
		}
		if authorNickname.Valid {
			c.AuthorNickname = authorNickname.String
		}
		comments = append(comments, c)
	}

	return comments, nil
}
