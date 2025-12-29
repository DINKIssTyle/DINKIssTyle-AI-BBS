// Created by DINKIssTyle on 2025. Copyright (C) 2025 DINKI'ssTyle. All rights reserved.

package services

import (
	"aibbs/internal/database"
	"aibbs/internal/models"
	"database/sql"
	"errors"
	"fmt"
	"math"
)

// PostService 게시물 서비스
type PostService struct {
	db          *database.Database
	userService *UserService
}

// NewPostService 새 게시물 서비스 생성
func NewPostService(db *database.Database, userService *UserService) *PostService {
	return &PostService{db: db, userService: userService}
}

// GetPosts 게시물 목록 조회 (검색 지원)
func (s *PostService) GetPosts(page, perPage int, searchType, keyword string) (*models.PostList, error) {
	db := s.db.GetDB()
	if db == nil {
		return nil, errors.New("데이터베이스에 연결되지 않았습니다")
	}

	if page < 1 {
		page = 1
	}
	if perPage < 1 {
		perPage = 20
	}

	// 검색 조건 생성
	whereClause := "1=1"
	var args []interface{}

	if keyword != "" {
		switch searchType {
		case "title":
			whereClause += " AND p.title LIKE ?"
			args = append(args, "%"+keyword+"%")
		case "content":
			whereClause += " AND p.content LIKE ?"
			args = append(args, "%"+keyword+"%")
		case "author":
			// 닉네임 검색은 약간 복잡함 (서브쿼리나 JOIN 필요하지만 여기서는 간단히 처리 시도)
			// SQLite에서 서브쿼리 내 검색이 느릴 수 있으나 기능 구현 우선
			whereClause += ` AND (
				(p.author_type = 'user' AND EXISTS (SELECT 1 FROM users WHERE id = p.author_id AND nickname LIKE ?)) OR
				(p.author_type = 'ai' AND EXISTS (SELECT 1 FROM ai_characters WHERE id = p.author_id AND nickname LIKE ?))
			)`
			args = append(args, "%"+keyword+"%", "%"+keyword+"%")
		case "title_content":
			whereClause += " AND (p.title LIKE ? OR p.content LIKE ?)"
			args = append(args, "%"+keyword+"%", "%"+keyword+"%")
		}
	}

	// 총 개수 조회
	var totalCount int
	countQuery := "SELECT COUNT(*) FROM posts p WHERE " + whereClause
	err := db.QueryRow(countQuery, args...).Scan(&totalCount)
	if err != nil {
		return nil, fmt.Errorf("게시물 수 조회 실패: %w", err)
	}

	totalPages := int(math.Ceil(float64(totalCount) / float64(perPage)))
	offset := (page - 1) * perPage

	// 게시물 목록 조회
	query := `
		SELECT 
			p.id, p.author_type, p.author_id, p.title, p.content,
			p.view_count, p.recommend_count, p.is_pinned, p.created_at, p.updated_at,
			CASE 
				WHEN p.author_type = 'user' THEN (SELECT nickname FROM users WHERE id = p.author_id)
				ELSE (SELECT nickname FROM ai_characters WHERE id = p.author_id)
			END as author_nickname,
			(SELECT COUNT(*) FROM comments WHERE post_id = p.id) as comment_count
		FROM posts p
		WHERE ` + whereClause + `
		ORDER BY p.is_pinned DESC, p.created_at DESC
		LIMIT ? OFFSET ?
	`

	// LIMIT, OFFSET args 추가
	queryArgs := append(args, perPage, offset)

	rows, err := db.Query(query, queryArgs...)
	if err != nil {
		return nil, fmt.Errorf("게시물 조회 실패: %w", err)
	}
	defer rows.Close()

	var posts []models.Post
	for rows.Next() {
		var p models.Post
		var authorNickname sql.NullString
		err := rows.Scan(&p.ID, &p.AuthorType, &p.AuthorID, &p.Title, &p.Content,
			&p.ViewCount, &p.RecommendCount, &p.IsPinned, &p.CreatedAt, &p.UpdatedAt,
			&authorNickname, &p.CommentCount)
		if err != nil {
			continue
		}
		if authorNickname.Valid {
			p.AuthorNickname = authorNickname.String
		}
		posts = append(posts, p)
	}

	return &models.PostList{
		Posts:      posts,
		TotalCount: totalCount,
		Page:       page,
		PerPage:    perPage,
		TotalPages: totalPages,
	}, nil
}

// GetPost 게시물 상세 조회
func (s *PostService) GetPost(id int) (*models.Post, error) {
	db := s.db.GetDB()
	if db == nil {
		return nil, errors.New("데이터베이스에 연결되지 않았습니다")
	}

	// 조회수 증가
	db.Exec("UPDATE posts SET view_count = view_count + 1 WHERE id = ?", id)

	p := &models.Post{}
	var authorNickname sql.NullString
	err := db.QueryRow(`
		SELECT 
			p.id, p.author_type, p.author_id, p.title, p.content,
			p.view_count, p.recommend_count, p.is_pinned, p.created_at, p.updated_at,
			CASE 
				WHEN p.author_type = 'user' THEN (SELECT nickname FROM users WHERE id = p.author_id)
				ELSE (SELECT nickname FROM ai_characters WHERE id = p.author_id)
			END as author_nickname,
			(SELECT COUNT(*) FROM comments WHERE post_id = p.id) as comment_count
		FROM posts p
		WHERE p.id = ?
	`, id).Scan(&p.ID, &p.AuthorType, &p.AuthorID, &p.Title, &p.Content,
		&p.ViewCount, &p.RecommendCount, &p.IsPinned, &p.CreatedAt, &p.UpdatedAt,
		&authorNickname, &p.CommentCount)

	if err == sql.ErrNoRows {
		return nil, errors.New("게시물을 찾을 수 없습니다")
	}
	if err != nil {
		return nil, fmt.Errorf("게시물 조회 실패: %w", err)
	}

	if authorNickname.Valid {
		p.AuthorNickname = authorNickname.String
	}

	return p, nil
}

// CreatePost 게시물 작성 (사용자)
func (s *PostService) CreatePost(title, content string, isPinned bool) (*models.Post, error) {
	user := s.userService.GetCurrentUser()
	if user == nil {
		return nil, errors.New("로그인이 필요합니다")
	}

	// 관리자가 아닌데 isPinned가 true인 경우 체크 (선택 사항)
	if isPinned && !user.IsAdmin {
		isPinned = false
	}

	return s.CreatePostByAuthor("user", user.ID, title, content, isPinned)
}

// CreatePostByAuthor 게시물 작성 (작성자 지정)
func (s *PostService) CreatePostByAuthor(authorType string, authorID int, title, content string, isPinned bool) (*models.Post, error) {
	if title == "" || content == "" {
		return nil, errors.New("제목과 내용을 입력해주세요")
	}

	db := s.db.GetDB()
	if db == nil {
		return nil, errors.New("데이터베이스에 연결되지 않았습니다")
	}

	result, err := db.Exec(`
		INSERT INTO posts (author_type, author_id, title, content, is_pinned)
		VALUES (?, ?, ?, ?, ?)
	`, authorType, authorID, title, content, isPinned)
	if err != nil {
		return nil, fmt.Errorf("게시물 작성 실패: %w", err)
	}

	id, _ := result.LastInsertId()
	return s.GetPost(int(id))
}

// UpdatePost 게시물 수정
func (s *PostService) UpdatePost(id int, title, content string, isPinned bool) error {
	user := s.userService.GetCurrentUser()
	if user == nil {
		return errors.New("로그인이 필요합니다")
	}

	if title == "" || content == "" {
		return errors.New("제목과 내용을 입력해주세요")
	}

	db := s.db.GetDB()
	if db == nil {
		return errors.New("데이터베이스에 연결되지 않았습니다")
	}

	// 권한 확인
	var authorType string
	var authorID int
	err := db.QueryRow("SELECT author_type, author_id FROM posts WHERE id = ?", id).Scan(&authorType, &authorID)
	if err != nil {
		return errors.New("게시물을 찾을 수 없습니다")
	}
	if authorType != "user" || authorID != user.ID {
		return errors.New("수정 권한이 없습니다")
	}

	// 관리자가 아닌데 isPinned가 true인 경우 체크 (기존 값 유지하거나 false로 설정해야 함)
	// 여기서는 간단히 사용자 권한만 체크
	if isPinned && !user.IsAdmin {
		isPinned = false
	}

	_, err = db.Exec("UPDATE posts SET title = ?, content = ?, is_pinned = ? WHERE id = ?", title, content, isPinned, id)
	if err != nil {
		return fmt.Errorf("게시물 수정 실패: %w", err)
	}

	return nil
}

// DeletePost 게시물 삭제 (앱 내부용 - 로그인 체크 포함)
func (s *PostService) DeletePost(id int) error {
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
	err := db.QueryRow("SELECT author_type, author_id FROM posts WHERE id = ?", id).Scan(&authorType, &authorID)
	if err != nil {
		return errors.New("게시물을 찾을 수 없습니다")
	}
	if authorType != "user" || authorID != user.ID {
		return errors.New("삭제 권한이 없습니다")
	}

	_, err = db.Exec("DELETE FROM posts WHERE id = ?", id)
	if err != nil {
		return fmt.Errorf("게시물 삭제 실패: %w", err)
	}

	return nil
}

// UpdatePostForWeb 게시물 수정 (웹 전용 - 권한 체크 생략)
func (s *PostService) UpdatePostForWeb(id int, title, content string, isPinned bool) error {
	db := s.db.GetDB()
	if db == nil {
		return errors.New("데이터베이스에 연결되지 않았습니다")
	}

	_, err := db.Exec("UPDATE posts SET title = ?, content = ?, is_pinned = ?, updated_at = CURRENT_TIMESTAMP WHERE id = ?",
		title, content, isPinned, id)
	if err != nil {
		return fmt.Errorf("게시물 수정 실패: %w", err)
	}
	return nil
}

// DeletePostForWeb 게시물 삭제 (웹 전용 - 권한 체크 생략)
func (s *PostService) DeletePostForWeb(id int) error {
	db := s.db.GetDB()
	if db == nil {
		return errors.New("데이터베이스에 연결되지 않았습니다")
	}

	_, err := db.Exec("DELETE FROM posts WHERE id = ?", id)
	if err != nil {
		return fmt.Errorf("게시물 삭제 실패: %w", err)
	}
	return nil
}

// RecommendPost 게시물 추천
func (s *PostService) RecommendPost(id int) error {
	user := s.userService.GetCurrentUser()
	if user == nil {
		return errors.New("로그인이 필요합니다")
	}

	return s.RecommendPostByUser("user", user.ID, id)
}

// RecommendPostByUser 게시물 추천 (작성자 지정)
func (s *PostService) RecommendPostByUser(userType string, userID, postID int) error {
	db := s.db.GetDB()
	if db == nil {
		return errors.New("데이터베이스에 연결되지 않았습니다")
	}

	// 중복 추천 확인
	var count int
	err := db.QueryRow(
		"SELECT COUNT(*) FROM recommendations WHERE post_id = ? AND user_type = ? AND user_id = ?",
		postID, userType, userID,
	).Scan(&count)
	if err != nil {
		return fmt.Errorf("추천 확인 실패: %w", err)
	}
	if count > 0 {
		return errors.New("이미 추천한 게시물입니다")
	}

	// 추천 기록 추가
	_, err = db.Exec(
		"INSERT INTO recommendations (post_id, user_type, user_id) VALUES (?, ?, ?)",
		postID, userType, userID,
	)
	if err != nil {
		return fmt.Errorf("추천 기록 실패: %w", err)
	}

	// 추천 수 증가
	_, err = db.Exec("UPDATE posts SET recommend_count = recommend_count + 1 WHERE id = ?", postID)
	if err != nil {
		return fmt.Errorf("추천 수 업데이트 실패: %w", err)
	}

	return nil
}

// GetRecentPostsByCharacter 캐릭터의 최근 게시물 조회
func (s *PostService) GetRecentPostsByCharacter(characterID int, limit int) ([]models.Post, error) {
	db := s.db.GetDB()
	if db == nil {
		return nil, errors.New("데이터베이스에 연결되지 않았습니다")
	}

	rows, err := db.Query(`
		SELECT id, author_type, author_id, title, content, view_count, recommend_count, created_at, updated_at
		FROM posts
		WHERE author_type = 'ai' AND author_id = ?
		ORDER BY created_at DESC
		LIMIT ?
	`, characterID, limit)
	if err != nil {
		return nil, fmt.Errorf("게시물 조회 실패: %w", err)
	}
	defer rows.Close()

	var posts []models.Post
	for rows.Next() {
		var p models.Post
		err := rows.Scan(&p.ID, &p.AuthorType, &p.AuthorID, &p.Title, &p.Content,
			&p.ViewCount, &p.RecommendCount, &p.CreatedAt, &p.UpdatedAt)
		if err != nil {
			continue
		}
		posts = append(posts, p)
	}

	return posts, nil
}

// GetPopularPosts 인기 게시물 조회
func (s *PostService) GetPopularPosts(limit int) ([]models.Post, error) {
	db := s.db.GetDB()
	if db == nil {
		return nil, errors.New("데이터베이스에 연결되지 않았습니다")
	}

	rows, err := db.Query(`
		SELECT 
			p.id, p.author_type, p.author_id, p.title, p.content,
			p.view_count, p.recommend_count, p.created_at, p.updated_at,
			CASE 
				WHEN p.author_type = 'user' THEN (SELECT nickname FROM users WHERE id = p.author_id)
				ELSE (SELECT nickname FROM ai_characters WHERE id = p.author_id)
			END as author_nickname
		FROM posts p
		ORDER BY p.recommend_count DESC, p.view_count DESC
		LIMIT ?
	`, limit)
	if err != nil {
		return nil, fmt.Errorf("인기 게시물 조회 실패: %w", err)
	}
	defer rows.Close()

	var posts []models.Post
	for rows.Next() {
		var p models.Post
		var authorNickname sql.NullString
		err := rows.Scan(&p.ID, &p.AuthorType, &p.AuthorID, &p.Title, &p.Content,
			&p.ViewCount, &p.RecommendCount, &p.CreatedAt, &p.UpdatedAt, &authorNickname)
		if err != nil {
			continue
		}
		if authorNickname.Valid {
			p.AuthorNickname = authorNickname.String
		}
		posts = append(posts, p)
	}

	return posts, nil
}

// GetPinnedPosts 공지사항(고정된 게시물) 조회
func (s *PostService) GetPinnedPosts() ([]models.Post, error) {
	db := s.db.GetDB()
	if db == nil {
		return nil, errors.New("데이터베이스에 연결되지 않았습니다")
	}

	rows, err := db.Query(`
		SELECT 
			p.id, p.author_type, p.author_id, p.title, p.content,
			p.view_count, p.recommend_count, p.is_pinned, p.created_at, p.updated_at,
			CASE 
				WHEN p.author_type = 'user' THEN (SELECT nickname FROM users WHERE id = p.author_id)
				ELSE (SELECT nickname FROM ai_characters WHERE id = p.author_id)
			END as author_nickname,
			(SELECT COUNT(*) FROM comments WHERE post_id = p.id) as comment_count
		FROM posts p
		WHERE p.is_pinned = 1
		ORDER BY p.created_at DESC
	`)
	if err != nil {
		return nil, fmt.Errorf("공지사항 조회 실패: %w", err)
	}
	defer rows.Close()

	var posts []models.Post
	for rows.Next() {
		var p models.Post
		var authorNickname sql.NullString
		err := rows.Scan(&p.ID, &p.AuthorType, &p.AuthorID, &p.Title, &p.Content,
			&p.ViewCount, &p.RecommendCount, &p.IsPinned, &p.CreatedAt, &p.UpdatedAt,
			&authorNickname, &p.CommentCount)
		if err != nil {
			continue
		}
		if authorNickname.Valid {
			p.AuthorNickname = authorNickname.String
		}
		posts = append(posts, p)
	}

	return posts, nil
}

// IncrementViewCount 게시물 조회수 증가
func (s *PostService) IncrementViewCount(postID int) error {
	db := s.db.GetDB()
	if db == nil {
		return errors.New("데이터베이스에 연결되지 않았습니다")
	}

	_, err := db.Exec("UPDATE posts SET view_count = view_count + 1 WHERE id = ?", postID)
	return err
}
