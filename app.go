// Created by DINKIssTyle on 2025. Copyright (C) 2025 DINKI'ssTyle. All rights reserved.

package main

import (
	"aibbs/internal/ai"
	"aibbs/internal/database"
	"aibbs/internal/models"
	"aibbs/internal/services"
	"aibbs/internal/web"
	"context"
	_ "embed"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strconv"

	"github.com/wailsapp/wails/v2/pkg/runtime"
)

//go:embed internal/database/schema.sql
var schemaSQL string

// App 메인 애플리케이션 구조체
type App struct {
	ctx context.Context

	// Database
	db *database.Database

	// Services
	userService      *services.UserService
	characterService *services.CharacterService
	postService      *services.PostService
	commentService   *services.CommentService
	llmService       *services.LLMService

	// AI Activity
	activityManager *ai.ActivityManager

	// Web Server
	webServer *web.WebServer
}

// NewApp 새 앱 인스턴스 생성
func NewApp() *App {
	return &App{}
}

// startup 앱 시작 시 호출
func (a *App) startup(ctx context.Context) {
	fmt.Println("[DEBUG] Startup called")
	a.ctx = ctx

	// 실행 파일 경로 기준으로 DB 파일 경로 설정
	execPath, err := os.Executable()
	if err != nil {
		// 실행 파일 경로를 가져올 수 없으면 현재 디렉토리 사용
		execPath, _ = os.Getwd()
	}
	execDir := filepath.Dir(execPath)
	dbPath := filepath.Join(execDir, "aibbs.db")

	// Database 초기화
	a.db = database.GetInstance()
	a.db.SetDBPath(dbPath)

	// 데이터베이스 연결 및 스키마 초기화
	if err := a.db.Connect(); err == nil {
		a.db.ExecuteSchema(schemaSQL)
	}

	// Services 초기화
	a.userService = services.NewUserService(a.db)
	a.characterService = services.NewCharacterService(a.db)
	a.postService = services.NewPostService(a.db, a.userService)
	a.commentService = services.NewCommentService(a.db, a.userService)
	a.llmService = services.NewLLMService()

	// LLM 서비스에 프롬프트 Getter 주입
	a.llmService.SetPromptGetters(
		// promptGetter
		func(key string) string {
			return a.GetPrompt(key)
		},
		// mbtiGetter
		func(mbti string) string {
			// DB에서 직접 조회 (임시로 DB 연결을 매번 새로 하거나 App 메서드 사용)
			// App의 GetMBTIDescriptions은 맵 전체를 반환하므로 비효율적일 수 있음.
			// 하지만 여기서는 간단히 구현.
			descs := a.GetMBTIDescriptions()
			return descs[mbti]
		},
	)

	// AI Activity Manager 초기화
	a.activityManager = ai.NewActivityManager(
		a.characterService,
		a.postService,
		a.commentService,
		a.llmService,
	)

	// 새 글/댓글 콜백 설정
	a.activityManager.SetCallbacks(
		func() { runtime.EventsEmit(a.ctx, "newPost") },
		func() { runtime.EventsEmit(a.ctx, "newComment") },
	)

	// DB에서 LLM 설정 로드
	a.loadLLMConfigFromDB()

	// DB 마이그레이션 (컬럼 추가)
	a.migrateDatabase()

	// 웹 서버 시작 여부 확인 및 자동 시작 (옵션)
	// a.StartWebServer("8080", true)

	// 웹 서버 초기화
	fmt.Println("[DEBUG] Initializing WebServer...")
	a.webServer = web.NewWebServer(a.db, a.userService, a.postService, a.commentService)
	fmt.Println("[DEBUG] WebServer initialized")
	a.loadWebConfigFromDB() // Added loading here

	// BBS 설정 로드 (WebServer 생성 후)
	a.loadBBSConfigFromDB()

	// 관리자 존재 여부 확인 및 자동 설정
	a.ensureAdminExists()

	// 프롬프트 테이블 초기화
	a.initPromptTables()
}

// initPromptTables 프롬프트 관련 테이블 생성
func (a *App) initPromptTables() {
	db := a.db.GetDB()
	if db == nil {
		return
	}

	// prompts 테이블
	_, err := db.Exec(`CREATE TABLE IF NOT EXISTS prompts (
		key_name TEXT PRIMARY KEY,
		content TEXT NOT NULL
	)`)
	if err != nil {
		log.Printf("[ERROR] prompts 테이블 생성 실패: %v", err)
	}

	// mbti_prompts 테이블
	_, err = db.Exec(`CREATE TABLE IF NOT EXISTS mbti_prompts (
		mbti TEXT PRIMARY KEY,
		content TEXT NOT NULL
	)`)
	if err != nil {
		log.Printf("[ERROR] mbti_prompts 테이블 생성 실패: %v", err)
	}
}

// GetPrompt 프롬프트 조회 (DB -> Default)
func (a *App) GetPrompt(key string) string {
	db := a.db.GetDB()
	if db != nil {
		var content string
		err := db.QueryRow("SELECT content FROM prompts WHERE key_name = ?", key).Scan(&content)
		if err == nil {
			return content
		}
	}

	// 기본값 반환
	switch key {
	case "system_role":
		return models.DefaultSystemRole
	case "nickname_gen":
		return models.DefaultNicknamePrompt
	case "post_instruction":
		return models.DefaultPostInstruction
	case "comment_instruction":
		return models.DefaultCommentInstruction
	case "reply_instruction":
		return models.DefaultReplyInstruction
	case "summary_instruction":
		return models.DefaultSummaryInstruction
	}
	return ""
}

// SavePrompt 프롬프트 저장
func (a *App) SavePrompt(key string, content string) error {
	db := a.db.GetDB()
	if db == nil {
		return fmt.Errorf("데이터베이스 연결 안됨")
	}

	_, err := db.Exec("INSERT OR REPLACE INTO prompts (key_name, content) VALUES (?, ?)", key, content)
	return err
}

// ResetPrompt 프롬프트 초기화 (DB 삭제)
func (a *App) ResetPrompt(key string) error {
	db := a.db.GetDB()
	if db == nil {
		return fmt.Errorf("데이터베이스 연결 안됨")
	}

	_, err := db.Exec("DELETE FROM prompts WHERE key_name = ?", key)
	return err
}

// GetMBTIDescriptions MBTI 설명 전체 조회
func (a *App) GetMBTIDescriptions() map[string]string {
	// 기본값 복사
	descs := make(map[string]string)
	for k, v := range models.DefaultMBTIDescriptions {
		descs[k] = v
	}

	db := a.db.GetDB()
	if db != nil {
		rows, err := db.Query("SELECT mbti, content FROM mbti_prompts")
		if err == nil {
			defer rows.Close()
			for rows.Next() {
				var mbti, content string
				if err := rows.Scan(&mbti, &content); err == nil {
					descs[mbti] = content
				}
			}
		}
	}
	return descs
}

// SaveMBTIDescription MBTI 설명 저장
func (a *App) SaveMBTIDescription(mbti string, content string) error {
	db := a.db.GetDB()
	if db == nil {
		return fmt.Errorf("데이터베이스 연결 안됨")
	}

	_, err := db.Exec("INSERT OR REPLACE INTO mbti_prompts (mbti, content) VALUES (?, ?)", mbti, content)
	return err
}

// ResetMBTIDescription MBTI 설명 초기화
func (a *App) ResetMBTIDescription(mbti string) error {
	db := a.db.GetDB()
	if db == nil {
		return fmt.Errorf("데이터베이스 연결 안됨")
	}

	_, err := db.Exec("DELETE FROM mbti_prompts WHERE mbti = ?", mbti)
	return err
}

// ensureAdminExists 관리자가 없으면 첫 번째 유저를 관리자로 설정
func (a *App) ensureAdminExists() {
	db := a.db.GetDB()
	if db == nil {
		return
	}

	var adminCount int
	err := db.QueryRow("SELECT COUNT(*) FROM users WHERE is_admin = 1").Scan(&adminCount)
	if err == nil && adminCount == 0 {
		// 관리자가 한 명도 없음, 유저가 있는지 확인
		var userCount int
		db.QueryRow("SELECT COUNT(*) FROM users").Scan(&userCount)
		if userCount > 0 {
			// 첫 번째 유저(ID가 가장 작은)를 관리자로 승격
			_, err := db.Exec("UPDATE users SET is_admin = 1 WHERE id = (SELECT MIN(id) FROM users)")
			if err == nil {
				log.Println("[INFO] 최초 관리자 자동 설정됨")
			}
		}
	}
}

// shutdown 앱 종료 시 호출
func (a *App) shutdown(ctx context.Context) {
	// AI 활동 정지
	if a.activityManager != nil {
		a.activityManager.Stop()
	}

	// DB 연결 종료
	if a.db != nil {
		a.db.Close()
	}
}

// loadLLMConfigFromDB DB에서 LLM 설정 로드
func (a *App) loadLLMConfigFromDB() {
	db := a.db.GetDB()
	if db == nil {
		return
	}

	// LLM 설정 로드
	rows, err := db.Query("SELECT key_name, value FROM settings")
	if err != nil {
		log.Printf("Failed to query settings: %v", err)
		return
	}
	defer rows.Close()

	config := models.LLMConfig{
		Host:            "localhost",
		Port:            "1234",
		Model1:          "default",
		PostsPerHour:    5,
		CommentsPerHour: 10,
		MaxTokens:       2000,
		Temperature:     0.8,
	}

	for rows.Next() {
		var key, value string
		if err := rows.Scan(&key, &value); err != nil {
			continue
		}
		switch key {
		case "llm_host":
			config.Host = value
		case "llm_port":
			config.Port = value
		case "llm_model_1":
			config.Model1 = value
		case "llm_model_2":
			config.Model2 = value
		case "llm_model_3":
			config.Model3 = value
		case "llm_model": // 레거시 호환
			if config.Model1 == "default" || config.Model1 == "" {
				config.Model1 = value
			}
		case "posts_per_hour":
			if v, err := strconv.Atoi(value); err == nil {
				config.PostsPerHour = v
			}
		case "comments_per_hour":
			if v, err := strconv.Atoi(value); err == nil {
				config.CommentsPerHour = v
			}
		case "max_tokens":
			if v, err := strconv.Atoi(value); err == nil {
				config.MaxTokens = v
			}
		case "temperature":
			if v, err := strconv.ParseFloat(value, 64); err == nil {
				config.Temperature = v
			}
		}
	}
	a.llmService.UpdateConfig(config)
	a.activityManager.SetActivityRate(config.PostsPerHour, config.CommentsPerHour)
}

// ========================================
// Database Management APIs
// ========================================

// IsDatabaseConnected 데이터베이스 연결 상태 확인
func (a *App) IsDatabaseConnected() bool {
	return a.db.IsConnected()
}

// ReconnectDatabase 데이터베이스 재연결
func (a *App) ReconnectDatabase() error {
	if err := a.db.Connect(); err != nil {
		return err
	}
	return a.db.ExecuteSchema(schemaSQL)
}

// ResetDatabase 데이터베이스 초기화
func (a *App) ResetDatabase(confirmation string) error {
	if confirmation != "데이터삭제" {
		return fmt.Errorf("확인 문구가 일치하지 않습니다")
	}

	// 1. AI 활동 중지 (접근 방지)
	if a.activityManager != nil {
		a.activityManager.Stop()
	}

	// 2. 외래 키 제약 조건 일시 해제 (SQLite)
	db := a.db.GetDB()
	if db != nil {
		_, _ = db.Exec("PRAGMA foreign_keys = OFF")
	}

	// 3. DB 초기화 실행
	if err := a.db.ResetDatabase(); err != nil {
		return err
	}

	// 4. 스키마 재생성
	if err := a.db.ExecuteSchema(schemaSQL); err != nil {
		return err
	}

	// 5. 외래 키 제약 조건 다시 활성화
	if db != nil {
		_, _ = db.Exec("PRAGMA foreign_keys = ON")
	}

	// 6. 설정들 다시 로드
	a.loadLLMConfigFromDB()
	a.loadWebConfigFromDB()
	a.loadBBSConfigFromDB()

	// 7. 관리자 계정 복구/생성
	a.ensureAdminExists()

	return nil
}

// ========================================
// User Management APIs
// ========================================

// CreateUser 사용자 생성
func (a *App) CreateUser(username, password, nickname string) error {
	return a.userService.CreateUser(username, password, nickname)
}

// Login 로그인
func (a *App) Login(username, password string) (*models.User, error) {
	return a.userService.Login(username, password)
}

// Logout 로그아웃
func (a *App) Logout() {
	a.userService.Logout()
}

// GetCurrentUser 현재 사용자 조회
func (a *App) GetCurrentUser() *models.User {
	return a.userService.GetCurrentUser()
}

// IsLoggedIn 로그인 상태 확인
func (a *App) IsLoggedIn() bool {
	return a.userService.IsLoggedIn()
}

// ChangePassword 비밀번호 변경
func (a *App) ChangePassword(oldPassword, newPassword string) error {
	return a.userService.ChangePassword(oldPassword, newPassword)
}

// ChangeNickname 닉네임 변경
func (a *App) ChangeNickname(newNickname string) error {
	return a.userService.ChangeNickname(newNickname)
}

// GetAllUsers 모든 사용자 조회
func (a *App) GetAllUsers() ([]*models.User, error) {
	return a.userService.GetAllUsers()
}

// SetUserAdmin 관리자 권한 설정
func (a *App) SetUserAdmin(userID int, isAdmin bool) error {
	return a.userService.SetUserAdmin(userID, isAdmin)
}

// ========================================
// AI Character Management APIs
// ========================================

// GenerateCharacters AI 캐릭터 자동 생성
func (a *App) GenerateCharacters(count int) ([]models.AICharacter, error) {
	return a.characterService.GenerateCharacters(count)
}

// GetAllCharacters 모든 캐릭터 조회
func (a *App) GetAllCharacters() ([]models.AICharacter, error) {
	return a.characterService.GetAllCharacters()
}

// UpdateCharacter 캐릭터 정보 수정
func (a *App) UpdateCharacter(character models.AICharacter) error {
	return a.characterService.UpdateCharacter(character)
}

// DeleteCharacter 캐릭터 삭제
func (a *App) DeleteCharacter(id int) error {
	return a.characterService.DeleteCharacter(id)
}

// CharacterStats 캐릭터 활동 통계
type CharacterStats struct {
	CharacterID  int `json:"character_id"`
	PostCount    int `json:"post_count"`
	CommentCount int `json:"comment_count"`
}

// GetAllCharacterStats 모든 캐릭터의 활동 통계 조회
func (a *App) GetAllCharacterStats() (map[int]CharacterStats, error) {
	db := a.db.GetDB()
	if db == nil {
		return nil, fmt.Errorf("데이터베이스에 연결되지 않았습니다")
	}

	result := make(map[int]CharacterStats)

	// 글 수 조회
	rows, err := db.Query(`
		SELECT author_id, COUNT(*) 
		FROM posts 
		WHERE author_type = 'ai' 
		GROUP BY author_id
	`)
	if err == nil {
		defer rows.Close()
		for rows.Next() {
			var charID, count int
			if rows.Scan(&charID, &count) == nil {
				stat := result[charID]
				stat.CharacterID = charID
				stat.PostCount = count
				result[charID] = stat
			}
		}
	}

	// 댓글 수 조회
	rows2, err := db.Query(`
		SELECT author_id, COUNT(*) 
		FROM comments 
		WHERE author_type = 'ai' 
		GROUP BY author_id
	`)
	if err == nil {
		defer rows2.Close()
		for rows2.Next() {
			var charID, count int
			if rows2.Scan(&charID, &count) == nil {
				stat := result[charID]
				stat.CharacterID = charID
				stat.CommentCount = count
				result[charID] = stat
			}
		}
	}

	return result, nil
}

// GetJobCategories 직종 목록 조회
func (a *App) GetJobCategories() []string {
	return models.JobCategories
}

// GetHobbies 취미 목록 조회
func (a *App) GetHobbies() []string {
	return models.Hobbies
}

// GetMBTITypes MBTI 목록 조회
func (a *App) GetMBTITypes() []string {
	return models.MBTITypes
}

// ========================================
// LLM Management APIs
// ========================================

// TestLLMConnection LLM 연결 테스트
func (a *App) TestLLMConnection(host, port, model string) error {
	return a.llmService.TestConnection(host, port, model)
}

// SaveLLMConfig LLM 설정 저장
func (a *App) SaveLLMConfig(host, port, model1, model2, model3 string, postsPerHour, commentsPerHour, maxTokens int, temperature float64) error {
	config := models.LLMConfig{
		Host:            host,
		Port:            port,
		Model1:          model1,
		Model2:          model2,
		Model3:          model3,
		PostsPerHour:    postsPerHour,
		CommentsPerHour: commentsPerHour,
		MaxTokens:       maxTokens,
		Temperature:     temperature,
	}
	a.llmService.UpdateConfig(config)
	a.activityManager.SetActivityRate(postsPerHour, commentsPerHour)

	// DB 저장
	db := a.db.GetDB()
	if db != nil {
		_, err := db.Exec(`INSERT OR REPLACE INTO settings (key_name, value) VALUES 
			('llm_host', ?), ('llm_port', ?), ('llm_model_1', ?), ('llm_model_2', ?), ('llm_model_3', ?),
			('posts_per_hour', ?), ('comments_per_hour', ?), ('max_tokens', ?), ('temperature', ?)`,
			host, port, model1, model2, model3,
			fmt.Sprintf("%d", postsPerHour), fmt.Sprintf("%d", commentsPerHour),
			fmt.Sprintf("%d", maxTokens), fmt.Sprintf("%f", temperature))
		if err != nil {
			log.Printf("Failed to save LLM config to DB: %v", err)
		}
	}

	return nil
}

// GetLLMConfig LLM 설정 조회
func (a *App) GetLLMConfig() *models.LLMConfig {
	config := a.llmService.GetConfig()
	return &config
}

// ========================================
// AI Activity APIs
// ========================================

// StartAIActivity AI 활동 시작
func (a *App) StartAIActivity() error {
	return a.activityManager.Start()
}

// StopAIActivity AI 활동 정지
func (a *App) StopAIActivity() {
	a.activityManager.Stop()
}

// IsAIActivityRunning AI 활동 상태 확인
func (a *App) IsAIActivityRunning() bool {
	return a.activityManager.IsRunning()
}

// ========================================
// Post Management APIs
// ========================================

func (a *App) GetPosts(page, perPage int) (*models.PostList, error) {
	return a.postService.GetPosts(page, perPage, "", "")
}

// GetPost 게시물 상세 조회
func (a *App) GetPost(id int) (*models.Post, error) {
	return a.postService.GetPost(id)
}

// CreatePost 게시물 작성
func (a *App) CreatePost(title, content string, isPinned bool) (*models.Post, error) {
	return a.postService.CreatePost(title, content, isPinned)
}

// UpdatePost 게시물 수정
func (a *App) UpdatePost(id int, title, content string, isPinned bool) error {
	return a.postService.UpdatePost(id, title, content, isPinned)
}

// GetPinnedPosts 공지사항 조회
func (a *App) GetPinnedPosts() ([]models.Post, error) {
	return a.postService.GetPinnedPosts()
}

// DeletePost 게시물 삭제
func (a *App) DeletePost(id int) error {
	return a.postService.DeletePost(id)
}

// RecommendPost 게시물 추천
func (a *App) RecommendPost(id int) error {
	return a.postService.RecommendPost(id)
}

// ========================================
// Comment Management APIs
// ========================================

// GetComments 댓글 목록 조회
func (a *App) GetComments(postID int) ([]*models.Comment, error) {
	return a.commentService.GetComments(postID)
}

// CreateComment 댓글 작성
func (a *App) CreateComment(postID int, content string, parentID *int) (*models.Comment, error) {
	return a.commentService.CreateComment(postID, content, parentID)
}

// DeleteComment 댓글 삭제
func (a *App) DeleteComment(id int) error {
	return a.commentService.DeleteComment(id)
}

// UpdateComment 댓글 수정
func (a *App) UpdateComment(id int, content string) error {
	return a.commentService.UpdateComment(id, content)
}

// ========================================
// Web Server APIs
// ========================================

// StartWebServer 웹 서버 시작
func (a *App) StartWebServer(port string, registrationOpen bool) error {
	a.webServer.SetPort(port)
	a.webServer.SetRegistrationOpen(registrationOpen)

	// 설정 저장
	db := a.db.GetDB()
	if db != nil {
		_, err := db.Exec(`INSERT OR REPLACE INTO settings (key_name, value) VALUES 
			('web_port', ?), ('web_registration', ?)`,
			port, fmt.Sprintf("%v", registrationOpen))
		if err != nil {
			log.Printf("Failed to save WebServer config: %v", err)
		}
	}

	return a.webServer.Start()
}

// StopWebServer 웹 서버 정지
func (a *App) StopWebServer() {
	a.webServer.Stop()
}

// IsWebServerRunning 웹 서버 실행 상태
func (a *App) IsWebServerRunning() bool {
	return a.webServer.IsRunning()
}

// GetWebServerConfig 웹 서버 설정 조회
func (a *App) GetWebServerConfig() map[string]interface{} {
	return map[string]interface{}{
		"port":             a.webServer.GetPort(),
		"registrationOpen": a.webServer.IsRegistrationOpen(),
		"running":          a.webServer.IsRunning(),
	}
}

// migrateDatabase DB 스키마 마이그레이션 (컬럼 추가)
func (a *App) migrateDatabase() {
	db := a.db.GetDB()
	if db == nil {
		return
	}

	// users 테이블에 is_admin 컬럼 추가
	// 이미 존재하면 에러가 발생하므로 무시함 (SQLite 특성상)
	_, _ = db.Exec("ALTER TABLE users ADD COLUMN is_admin INTEGER DEFAULT 0")

	// posts 테이블에 is_pinned 컬럼 추가
	_, _ = db.Exec("ALTER TABLE posts ADD COLUMN is_pinned INTEGER DEFAULT 0")

	// ai_characters 테이블에 assigned_model_index 컬럼 추가
	_, _ = db.Exec("ALTER TABLE ai_characters ADD COLUMN assigned_model_index INTEGER DEFAULT 1")

	// ai_characters 테이블에 새 컬럼 추가 (AI 캐릭터 확장)
	_, _ = db.Exec("ALTER TABLE ai_characters ADD COLUMN birthdate TEXT")
	_, _ = db.Exec("ALTER TABLE ai_characters ADD COLUMN region TEXT")
	_, _ = db.Exec("ALTER TABLE ai_characters ADD COLUMN post_count INTEGER DEFAULT 0")
	_, _ = db.Exec("ALTER TABLE ai_characters ADD COLUMN comment_count INTEGER DEFAULT 0")
	_, _ = db.Exec("ALTER TABLE ai_characters ADD COLUMN persona_summary TEXT")
}

// loadBBSConfigFromDB DB에서 BBS 설정 로드
func (a *App) loadBBSConfigFromDB() {
	db := a.db.GetDB()
	if db == nil {
		return
	}

	// 기본값 (WebServer 생성 시 이미 설정됨, DB 값으로 오버라이드)
	config := models.BBSConfig{
		Title:  "DINKI'ssTyle AI BBS",
		Footer: "(C) 2025 DINKI'ssTyle",
		Theme:  "blue",
		Font:   "sans",
	}

	rows, err := db.Query("SELECT key_name, value FROM settings WHERE key_name IN ('bbs_title', 'bbs_footer', 'bbs_theme', 'bbs_font', 'bbs_posts_per_page')")
	if err == nil {
		defer rows.Close()
		for rows.Next() {
			var key, val string
			if err := rows.Scan(&key, &val); err == nil {
				switch key {
				case "bbs_title":
					config.Title = val
				case "bbs_footer":
					config.Footer = val
				case "bbs_theme":
					config.Theme = val
				case "bbs_font":
					config.Font = val
				case "bbs_posts_per_page":
					if n, err := strconv.Atoi(val); err == nil {
						config.PostsPerPage = n
					}
				}
			}
		}
	}

	a.webServer.SetBBSConfig(config)
}

// loadWebConfigFromDB DB에서 웹 서버 설정 로드
func (a *App) loadWebConfigFromDB() {
	db := a.db.GetDB()
	if db == nil {
		return
	}

	rows, err := db.Query("SELECT key_name, value FROM settings WHERE key_name IN ('web_port', 'web_registration')")
	if err == nil {
		defer rows.Close()
		for rows.Next() {
			var key, val string
			if err := rows.Scan(&key, &val); err == nil {
				switch key {
				case "web_port":
					a.webServer.SetPort(val)
				case "web_registration":
					isOpen := val == "true"
					a.webServer.SetRegistrationOpen(isOpen)
				}
			}
		}
	}
}

// SaveBBSConfig BBS 설정 저장
func (a *App) SaveBBSConfig(title, footer, theme, font string, postsPerPage int, timezone string) error {
	if postsPerPage <= 0 {
		postsPerPage = 20
	}

	config := models.BBSConfig{
		Title:        title,
		Footer:       footer,
		Theme:        theme,
		Font:         font,
		PostsPerPage: postsPerPage,
		Timezone:     timezone,
	}

	// 웹서버에 즉시 적용
	a.webServer.SetBBSConfig(config)

	// DB 저장
	db := a.db.GetDB()
	if db != nil {
		_, err := db.Exec(`INSERT OR REPLACE INTO settings (key_name, value) VALUES 
			('bbs_title', ?), ('bbs_footer', ?), ('bbs_theme', ?), ('bbs_font', ?), ('bbs_posts_per_page', ?)`,
			title, footer, theme, font, fmt.Sprintf("%d", postsPerPage))
		if err != nil {
			log.Printf("Failed to save BBS config: %v", err)
			return err
		}
	}
	return nil
}

// GetBBSConfig BBS 설정 조회
func (a *App) GetBBSConfig() models.BBSConfig {
	config := models.BBSConfig{
		Title:        "DINKI'ssTyle AI BBS",
		Footer:       "(C) 2025 DINKI'ssTyle",
		Theme:        "blue",
		Font:         "sans",
		PostsPerPage: 20,
	}

	db := a.db.GetDB()
	if db == nil {
		return config
	}

	rows, err := db.Query("SELECT key_name, value FROM settings WHERE key_name IN ('bbs_title', 'bbs_footer', 'bbs_theme', 'bbs_font', 'bbs_posts_per_page')")
	if err == nil {
		defer rows.Close()
		for rows.Next() {
			var key, val string
			if err := rows.Scan(&key, &val); err == nil {
				switch key {
				case "bbs_title":
					config.Title = val
				case "bbs_footer":
					config.Footer = val
				case "bbs_theme":
					config.Theme = val
				case "bbs_font":
					config.Font = val
				case "bbs_posts_per_page":
					if n, err := strconv.Atoi(val); err == nil && n > 0 {
						config.PostsPerPage = n
					}
				}
			}
		}
	}
	return config
}
