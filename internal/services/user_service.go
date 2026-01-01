// Created by DINKIssTyle on 2026. Copyright (C) 2026 DINKI'ssTyle. All rights reserved.
// Finalized FindUserByNickname implementation

package services

import (
	"aibbs/internal/database"
	"aibbs/internal/models"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"errors"
	"fmt"
	"log"
	"regexp"
)

// UserService 사용자 서비스
type UserService struct {
	db          *database.Database
	currentUser *models.User
}

// NewUserService 새 사용자 서비스 생성
func NewUserService(db *database.Database) *UserService {
	return &UserService{db: db}
}

// hashPassword 비밀번호 해시
func hashPassword(password string) string {
	hash := sha256.Sum256([]byte(password))
	return hex.EncodeToString(hash[:])
}

// validateNickname 닉네임 유효성 검사 (2~8자, 특수문자/공백 금지)
func validateNickname(nickname string) error {
	// 길이 검사 (한글 등 멀티바이트 문자 고려하여 Rune count 사용)
	runeCount := len([]rune(nickname))
	if runeCount < 2 || runeCount > 13 {
		return errors.New("닉네임은 2자 이상 13자 이하여야 합니다")
	}

	// 정규식 검사 (특수문자, 공백 제외 -> 한글, 영문, 숫자만 허용)
	// ^[가-힣a-zA-Z0-9]+$
	matched, err := regexp.MatchString("^[가-힣a-zA-Z0-9]+$", nickname)
	if err != nil {
		return fmt.Errorf("닉네임 검사 오류: %w", err)
	}
	if !matched {
		return errors.New("닉네임에 특수문자나 공백을 사용할 수 없습니다")
	}

	return nil
}

// CreateUser 사용자 생성
func (s *UserService) CreateUser(username, password, nickname string) error {
	if username == "" || password == "" || nickname == "" {
		return errors.New("모든 필드를 입력해주세요")
	}

	if err := validateNickname(nickname); err != nil {
		return err
	}

	db := s.db.GetDB()
	if db == nil {
		return errors.New("데이터베이스에 연결되지 않았습니다")
	}

	// 중복 확인
	var count int
	err := db.QueryRow("SELECT COUNT(*) FROM users WHERE username = ?", username).Scan(&count)
	if err != nil {
		return fmt.Errorf("사용자 확인 실패: %w", err)
	}
	if count > 0 {
		return errors.New("이미 존재하는 아이디입니다")
	}

	// 사용자 생성
	passwordHash := hashPassword(password)
	_, err = db.Exec(
		"INSERT INTO users (username, password_hash, nickname) VALUES (?, ?, ?)",
		username, passwordHash, nickname,
	)
	if err != nil {
		return fmt.Errorf("사용자 생성 실패: %w", err)
	}

	return nil
}

// GetUserByID ID로 사용자 조회
func (s *UserService) GetUserByID(id int) (*models.User, error) {
	db := s.db.GetDB()
	if db == nil {
		return nil, errors.New("데이터베이스에 연결되지 않았습니다")
	}

	user := &models.User{}
	err := db.QueryRow(
		`SELECT id, username, nickname, is_admin, created_at, updated_at,
		        COALESCE(theme, 'dark'), COALESCE(font_style, 'default'), 
		        COALESCE(timezone, 'Asia/Seoul'), COALESCE(posts_per_page, 20)
		 FROM users WHERE id = ?`, id,
	).Scan(&user.ID, &user.Username, &user.Nickname, &user.IsAdmin, &user.CreatedAt, &user.UpdatedAt,
		&user.Theme, &user.FontStyle, &user.Timezone, &user.PostsPerPage)

	if err == sql.ErrNoRows {
		return nil, errors.New("사용자를 찾을 수 없습니다")
	}
	if err != nil {
		return nil, fmt.Errorf("사용자 조회 실패: %w", err)
	}

	return user, nil
}

// FindUserByNickname 닉네임으로 사용자 조회
func (s *UserService) FindUserByNickname(nickname string) (*models.User, error) {
	db := s.db.GetDB()
	if db == nil {
		return nil, errors.New("데이터베이스에 연결되지 않았습니다")
	}

	user := &models.User{}
	err := db.QueryRow(
		`SELECT id, username, nickname, is_admin, created_at, updated_at,
		        COALESCE(theme, 'dark'), COALESCE(font_style, 'default'), 
		        COALESCE(timezone, 'Asia/Seoul'), COALESCE(posts_per_page, 20)
		 FROM users WHERE nickname = ?`, nickname,
	).Scan(&user.ID, &user.Username, &user.Nickname, &user.IsAdmin, &user.CreatedAt, &user.UpdatedAt,
		&user.Theme, &user.FontStyle, &user.Timezone, &user.PostsPerPage)

	if err == sql.ErrNoRows {
		return nil, errors.New("사용자를 찾을 수 없습니다")
	}
	if err != nil {
		return nil, fmt.Errorf("사용자 조회 실패: %w", err)
	}

	return user, nil
}

// LoginForWeb 웹용 로그인 (세션 저장 안함)
func (s *UserService) LoginForWeb(username, password string) (*models.User, error) {
	if username == "" || password == "" {
		return nil, errors.New("아이디와 비밀번호를 입력해주세요")
	}

	db := s.db.GetDB()
	if db == nil {
		return nil, errors.New("데이터베이스에 연결되지 않았습니다")
	}

	passwordHash := hashPassword(password)
	user := &models.User{}

	err := db.QueryRow(
		`SELECT id, username, nickname, is_admin, created_at, updated_at,
		        COALESCE(theme, 'dark'), COALESCE(font_style, 'default'), 
		        COALESCE(timezone, 'Asia/Seoul'), COALESCE(posts_per_page, 20)
		 FROM users WHERE username = ? AND password_hash = ?`,
		username, passwordHash,
	).Scan(&user.ID, &user.Username, &user.Nickname, &user.IsAdmin, &user.CreatedAt, &user.UpdatedAt,
		&user.Theme, &user.FontStyle, &user.Timezone, &user.PostsPerPage)

	if err == sql.ErrNoRows {
		log.Printf("[DEBUG] LoginForWeb failed: user '%s' not found or password incorrect", username)
		return nil, errors.New("아이디 또는 비밀번호가 일치하지 않습니다")
	}
	if err != nil {
		log.Printf("[ERROR] LoginForWeb scan error for user '%s': %v", username, err)
		return nil, fmt.Errorf("로그인 실패: %w", err)
	}

	return user, nil
}

// Login 로그인
func (s *UserService) Login(username, password string) (*models.User, error) {
	if username == "" || password == "" {
		return nil, errors.New("아이디와 비밀번호를 입력해주세요")
	}

	db := s.db.GetDB()
	if db == nil {
		return nil, errors.New("데이터베이스에 연결되지 않았습니다")
	}

	passwordHash := hashPassword(password)
	user := &models.User{}

	err := db.QueryRow(
		`SELECT id, username, nickname, is_admin, created_at, updated_at,
		        COALESCE(theme, 'dark'), COALESCE(font_style, 'default'), 
		        COALESCE(timezone, 'Asia/Seoul'), COALESCE(posts_per_page, 20)
		 FROM users WHERE username = ? AND password_hash = ?`,
		username, passwordHash,
	).Scan(&user.ID, &user.Username, &user.Nickname, &user.IsAdmin, &user.CreatedAt, &user.UpdatedAt,
		&user.Theme, &user.FontStyle, &user.Timezone, &user.PostsPerPage)

	if err == sql.ErrNoRows {
		log.Printf("[DEBUG] Login failed: user '%s' not found or password incorrect", username)
		return nil, errors.New("아이디 또는 비밀번호가 일치하지 않습니다")
	}
	if err != nil {
		log.Printf("[ERROR] Login scan error for user '%s': %v", username, err)
		return nil, fmt.Errorf("로그인 실패: %w", err)
	}

	s.currentUser = user
	return user, nil
}

// Logout 로그아웃
func (s *UserService) Logout() {
	s.currentUser = nil
}

// GetCurrentUser 현재 로그인한 사용자 반환
func (s *UserService) GetCurrentUser() *models.User {
	return s.currentUser
}

// IsLoggedIn 로그인 상태 확인
func (s *UserService) IsLoggedIn() bool {
	return s.currentUser != nil
}

// ChangePassword 비밀번호 변경
func (s *UserService) ChangePassword(oldPassword, newPassword string) error {
	if s.currentUser == nil {
		return errors.New("로그인이 필요합니다")
	}

	if oldPassword == "" || newPassword == "" {
		return errors.New("비밀번호를 입력해주세요")
	}

	db := s.db.GetDB()
	if db == nil {
		return errors.New("데이터베이스에 연결되지 않았습니다")
	}

	// 현재 비밀번호 확인
	oldHash := hashPassword(oldPassword)
	var count int
	err := db.QueryRow(
		"SELECT COUNT(*) FROM users WHERE id = ? AND password_hash = ?",
		s.currentUser.ID, oldHash,
	).Scan(&count)

	if err != nil {
		return fmt.Errorf("비밀번호 확인 실패: %w", err)
	}
	if count == 0 {
		return errors.New("현재 비밀번호가 일치하지 않습니다")
	}

	// 새 비밀번호로 변경
	newHash := hashPassword(newPassword)
	_, err = db.Exec(
		"UPDATE users SET password_hash = ? WHERE id = ?",
		newHash, s.currentUser.ID,
	)
	if err != nil {
		return fmt.Errorf("비밀번호 변경 실패: %w", err)
	}

	return nil
}

// ChangeNickname 닉네임 변경
func (s *UserService) ChangeNickname(newNickname string) error {
	if s.currentUser == nil {
		return errors.New("로그인이 필요합니다")
	}

	if newNickname == "" {
		return errors.New("닉네임을 입력해주세요")
	}

	if err := validateNickname(newNickname); err != nil {
		return err
	}

	db := s.db.GetDB()
	if db == nil {
		return errors.New("데이터베이스에 연결되지 않았습니다")
	}

	_, err := db.Exec(
		"UPDATE users SET nickname = ? WHERE id = ?",
		newNickname, s.currentUser.ID,
	)
	if err != nil {
		return fmt.Errorf("닉네임 변경 실패: %w", err)
	}

	s.currentUser.Nickname = newNickname
	return nil
}

// GetAllUsers 모든 사용자 조회 (관리자용)
func (s *UserService) GetAllUsers() ([]*models.User, error) {
	db := s.db.GetDB()
	if db == nil {
		return nil, errors.New("데이터베이스에 연결되지 않았습니다")
	}

	rows, err := db.Query("SELECT id, username, nickname, is_admin, created_at, updated_at FROM users ORDER BY id DESC")
	if err != nil {
		return nil, fmt.Errorf("사용자 목록 조회 실패: %w", err)
	}
	defer rows.Close()

	var users []*models.User
	for rows.Next() {
		u := &models.User{}
		err := rows.Scan(&u.ID, &u.Username, &u.Nickname, &u.IsAdmin, &u.CreatedAt, &u.UpdatedAt)
		if err != nil {
			continue
		}
		users = append(users, u)
	}

	return users, nil
}

// SetUserAdmin 관리자 권한 설정
func (s *UserService) SetUserAdmin(userID int, isAdmin bool) error {
	db := s.db.GetDB()
	if db == nil {
		return errors.New("데이터베이스에 연결되지 않았습니다")
	}

	_, err := db.Exec("UPDATE users SET is_admin = ? WHERE id = ?", isAdmin, userID)
	if err != nil {
		return fmt.Errorf("권한 설정 실패: %w", err)
	}

	return nil
}

// UpdateSettings 사용자 설정 저장
func (s *UserService) UpdateSettings(userID int, theme, fontStyle, timezone string, postsPerPage int) error {
	db := s.db.GetDB()
	if db == nil {
		return errors.New("데이터베이스에 연결되지 않았습니다")
	}

	_, err := db.Exec(
		"UPDATE users SET theme = ?, font_style = ?, timezone = ?, posts_per_page = ? WHERE id = ?",
		theme, fontStyle, timezone, postsPerPage, userID,
	)
	if err != nil {
		return fmt.Errorf("설정 저장 실패: %w", err)
	}

	// 현재 사용자인 경우 메모리에도 반영
	if s.currentUser != nil && s.currentUser.ID == userID {
		s.currentUser.Theme = theme
		s.currentUser.FontStyle = fontStyle
		s.currentUser.Timezone = timezone
		s.currentUser.PostsPerPage = postsPerPage
	}

	return nil
}

// ChangePasswordByUserID 사용자 ID 기반 비밀번호 변경 (웹용)
func (s *UserService) ChangePasswordByUserID(userID int, oldPassword, newPassword string) error {
	if oldPassword == "" || newPassword == "" {
		return errors.New("비밀번호를 입력해주세요")
	}

	db := s.db.GetDB()
	if db == nil {
		return errors.New("데이터베이스에 연결되지 않았습니다")
	}

	// 현재 비밀번호 확인
	oldHash := hashPassword(oldPassword)
	var count int
	err := db.QueryRow(
		"SELECT COUNT(*) FROM users WHERE id = ? AND password_hash = ?",
		userID, oldHash,
	).Scan(&count)

	if err != nil {
		return fmt.Errorf("비밀번호 확인 실패: %w", err)
	}
	if count == 0 {
		return errors.New("현재 비밀번호가 일치하지 않습니다")
	}

	// 새 비밀번호로 변경
	newHash := hashPassword(newPassword)
	_, err = db.Exec("UPDATE users SET password_hash = ? WHERE id = ?", newHash, userID)
	if err != nil {
		return fmt.Errorf("비밀번호 변경 실패: %w", err)
	}

	return nil
}

// ChangeNicknameByUserID 사용자 ID 기반 닉네임 변경 (웹용)
func (s *UserService) ChangeNicknameByUserID(userID int, newNickname string) error {
	if newNickname == "" {
		return errors.New("닉네임을 입력해주세요")
	}

	if err := validateNickname(newNickname); err != nil {
		return err
	}

	db := s.db.GetDB()
	if db == nil {
		return errors.New("데이터베이스에 연결되지 않았습니다")
	}

	// 중복 확인
	var count int
	err := db.QueryRow("SELECT COUNT(*) FROM users WHERE nickname = ? AND id != ?", newNickname, userID).Scan(&count)
	if err != nil {
		return fmt.Errorf("닉네임 확인 실패: %w", err)
	}
	if count > 0 {
		return errors.New("이미 사용 중인 닉네임입니다")
	}

	_, err = db.Exec("UPDATE users SET nickname = ? WHERE id = ?", newNickname, userID)
	if err != nil {
		return fmt.Errorf("닉네임 변경 실패: %w", err)
	}

	return nil
}

// DeleteUser 회원 탈퇴
func (s *UserService) DeleteUser(userID int) error {
	db := s.db.GetDB()
	if db == nil {
		return errors.New("데이터베이스에 연결되지 않았습니다")
	}

	_, err := db.Exec("DELETE FROM users WHERE id = ?", userID)
	if err != nil {
		return fmt.Errorf("회원 탈퇴 실패: %w", err)
	}

	return nil
}
