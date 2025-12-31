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
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/wailsapp/wails/v2/pkg/runtime"
)

//go:embed internal/database/schema.sql
var schemaSQL string

// App 메인 애플리케이션 구조체
type App struct {
	mode string // "main" or "char_manager"
	ctx  context.Context

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

	// Logger
	logWriter *WailsLogWriter
}

// WailsLogWriter 로그를 Wails 이벤트로 전송하는 라이터
type WailsLogWriter struct {
	ctx       context.Context
	webServer *web.WebServer
}

func (w *WailsLogWriter) Write(p []byte) (n int, err error) {
	str := string(p)
	// 터미널 출력 유지
	os.Stdout.Write(p)
	// 프론트엔드 이벤트 발생
	if w.ctx != nil {
		runtime.EventsEmit(w.ctx, "log-event", str)
	}
	// 웹 서버 브로드캐스트
	if w.webServer != nil {
		w.webServer.BroadcastLog(str)
	}
	return len(p), nil
}

// NewApp 새 앱 인스턴스 생성
func NewApp(mode string) *App {
	return &App{
		mode: mode,
	}
}

// startup 앱 시작 시 호출
func (a *App) startup(ctx context.Context) {
	fmt.Println("[DEBUG] Startup called")
	a.ctx = ctx

	a.ctx = ctx

	// 로그 설정
	a.logWriter = &WailsLogWriter{ctx: ctx}
	log.SetOutput(a.logWriter) // WailsLogWriter를 기본 로거로 설정

	// 실행 파일 경로 기준으로 DB 파일 경로 설정
	execPath, err := os.Executable()
	if err != nil {
		// 실행 파일 경로를 가져올 수 없으면 현재 디렉토리 사용
		execPath, _ = os.Getwd()
	}
	execDir := filepath.Dir(execPath)

	// 마지막 사용 DB 로드 (없으면 default.db 사용)
	lastDBFile := filepath.Join(execDir, "last_db.txt")
	dbName := "default.db"
	if data, err := os.ReadFile(lastDBFile); err == nil {
		savedName := strings.TrimSpace(string(data))
		if savedName != "" {
			// 파일이 실제로 존재하는지 확인
			if _, err := os.Stat(filepath.Join(execDir, savedName)); err == nil {
				dbName = savedName
			}
		}
	}

	dbPath := filepath.Join(execDir, dbName)

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
	a.characterService.SetContext(ctx)
	a.postService = services.NewPostService(a.db, a.userService)
	a.postService.StartViewCountFlusher(ctx) // 조회수 플러시 시작
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
	a.logWriter.webServer = a.webServer // 로거에 웹 서버 연결
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
// ResetDatabase 데이터베이스 초기화 (게시물, 댓글 등 콘텐츠만 삭제)
func (a *App) ResetDatabase(confirmation string) error {
	if confirmation != "데이터삭제" {
		return fmt.Errorf("확인 문구가 일치하지 않습니다")
	}

	// 1. AI 활동 중지 (접근 방지)
	if a.activityManager != nil {
		a.activityManager.Stop()
	}

	// 2. 콘텐츠 데이터 삭제 (유저, 캐릭터, 설정 유지)
	if err := a.db.ClearContent(); err != nil {
		return err
	}

	// 로그 남기기
	log.Println("[INFO] 데이터베이스 콘텐츠 초기화 완료 (게시물, 댓글 삭제)")

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

// OpenLogFile 로그 파일 열기
func (a *App) OpenLogFile() {
	exec.Command("explorer", "debug_log.txt").Run()
}

// SelectFile 파일 선택 다이얼로그 열기
func (a *App) SelectFile(title string, filter string) (string, error) {
	selection, err := runtime.OpenFileDialog(a.ctx, runtime.OpenDialogOptions{
		Title: title,
		Filters: []runtime.FileFilter{
			{
				DisplayName: filter,
				Pattern:     "*.*",
			},
		},
	})
	if err != nil {
		return "", err
	}
	return selection, nil
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
func (a *App) StartWebServer(port string, registrationOpen bool, sslEnabled bool, sslCert string, sslKey string) error {
	a.webServer.SetPort(port)
	a.webServer.SetRegistrationOpen(registrationOpen)

	// SSL 설정 반영 (BBSConfig 업데이트)
	config := a.webServer.GetBBSConfig()
	config.SSLEnabled = sslEnabled
	config.SSLCertPath = sslCert
	config.SSLKeyPath = sslKey
	a.webServer.SetBBSConfig(config)

	// 설정 저장
	db := a.db.GetDB()
	if db != nil {
		_, err := db.Exec(`INSERT OR REPLACE INTO settings (key_name, value) VALUES 
			('web_port', ?), ('web_registration', ?), ('web_ssl_enabled', ?), ('web_ssl_cert', ?), ('web_ssl_key', ?)`,
			port, fmt.Sprintf("%v", registrationOpen), fmt.Sprintf("%v", sslEnabled), sslCert, sslKey)
		if err != nil {
			log.Printf("Failed to save WebServer config: %v", err)
		}
	}

	// BBS 설정 재로드 (서버 시작 시 DB 설정 적용)
	a.loadBBSConfigFromDB()

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
	config := a.webServer.GetBBSConfig()
	return map[string]interface{}{
		"port":             a.webServer.GetPort(),
		"registrationOpen": a.webServer.IsRegistrationOpen(),
		"running":          a.webServer.IsRunning(),
		"sslEnabled":       config.SSLEnabled,
		"sslCertPath":      config.SSLCertPath,
		"sslKeyPath":       config.SSLKeyPath,
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
		Title:    "DINKI'ssTyle AI BBS",
		Footer:   "(C) 2025 DINKI'ssTyle",
		Theme:    "blue",
		Font:     "sans",
		Timezone: "Asia/Seoul",
	}

	rows, err := db.Query("SELECT key_name, value FROM settings WHERE key_name IN ('bbs_title', 'bbs_footer', 'bbs_theme', 'bbs_font', 'bbs_posts_per_page', 'bbs_timezone')")
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
				case "bbs_timezone":
					config.Timezone = val
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

	rows, err := db.Query("SELECT key_name, value FROM settings WHERE key_name IN ('web_port', 'web_registration', 'web_ssl_enabled', 'web_ssl_cert', 'web_ssl_key')")
	if err == nil {
		defer rows.Close()
		config := a.webServer.GetBBSConfig()
		for rows.Next() {
			var key, val string
			if err := rows.Scan(&key, &val); err == nil {
				switch key {
				case "web_port":
					a.webServer.SetPort(val)
				case "web_registration":
					isOpen := val == "true"
					a.webServer.SetRegistrationOpen(isOpen)
				case "web_ssl_enabled":
					config.SSLEnabled = val == "true"
				case "web_ssl_cert":
					config.SSLCertPath = val
				case "web_ssl_key":
					config.SSLKeyPath = val
				}
			}
		}
		a.webServer.SetBBSConfig(config)
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
			('bbs_title', ?), ('bbs_footer', ?), ('bbs_theme', ?), ('bbs_font', ?), ('bbs_posts_per_page', ?), ('bbs_timezone', ?)`,
			title, footer, theme, font, fmt.Sprintf("%d", postsPerPage), timezone)
		if err != nil {
			log.Printf("Failed to save BBS config: %v", err)
			return err
		}
	}
	return nil
}

// SaveWebServerConfig 웹 서버 설정 저장
func (a *App) SaveWebServerConfig(port string, registration bool, sslEnabled bool, sslCert, sslKey string) error {
	if a.webServer != nil {
		a.webServer.SetPort(port)
		a.webServer.SetRegistrationOpen(registration)

		// SSL 설정을 BBSConfig에 반영 (기존 설정 유지하며 업데이트)
		config := a.webServer.GetBBSConfig()
		config.SSLEnabled = sslEnabled
		config.SSLCertPath = sslCert
		config.SSLKeyPath = sslKey
		a.webServer.SetBBSConfig(config)
	}

	db := a.db.GetDB()
	if db == nil {
		return fmt.Errorf("데이터베이스에 연결되지 않았습니다")
	}

	tx, err := db.Begin()
	if err != nil {
		return err
	}

	queries := map[string]string{
		"web_port":         port,
		"web_registration": strconv.FormatBool(registration),
		"web_ssl_enabled":  strconv.FormatBool(sslEnabled),
		"web_ssl_cert":     sslCert,
		"web_ssl_key":      sslKey,
	}

	for k, v := range queries {
		_, err := tx.Exec("INSERT OR REPLACE INTO settings (key_name, value) VALUES (?, ?)", k, v)
		if err != nil {
			tx.Rollback()
			return err
		}
	}

	return tx.Commit()
}

// GetBBSConfig BBS 설정 조회
func (a *App) GetBBSConfig() models.BBSConfig {
	config := models.BBSConfig{
		Title:        "DINKI'ssTyle AI BBS",
		Footer:       "(C) 2025 DINKI'ssTyle",
		Theme:        "blue",
		Font:         "sans",
		PostsPerPage: 20,
		Timezone:     "Asia/Seoul",
	}

	db := a.db.GetDB()
	if db == nil {
		return config
	}

	rows, err := db.Query("SELECT key_name, value FROM settings WHERE key_name IN ('bbs_title', 'bbs_footer', 'bbs_theme', 'bbs_font', 'bbs_posts_per_page', 'bbs_timezone')")
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
				case "bbs_timezone":
					config.Timezone = val
				}
			}
		}
	}
	return config
}

// GetAppMode 앱 실행 모드 반환
func (a *App) GetAppMode() string {
	return a.mode
}

// OpenCharacterManagerWindow 캐릭터 관리자 창(새 프로세스) 열기
func (a *App) OpenCharacterManagerWindow() error {
	execPath, err := os.Executable()
	if err != nil {
		return err
	}

	cmd := exec.Command(execPath, "-mode", "char_manager")
	return cmd.Start()
}

// GetCharacterRefValues 캐릭터 생성 참조값 조회
func (a *App) GetCharacterRefValues() map[string]string {
	result := make(map[string]string)
	db := a.db.GetDB()
	if db == nil {
		// 기본값 반환
		result["job_categories"] = strings.Join(models.JobCategories, ", ")
		result["hobbies"] = strings.Join(models.Hobbies, ", ")
		result["regions"] = strings.Join(models.Regions, ", ")
		return result
	}

	// DB에서 값 조회, 없으면 기본값
	keys := []string{"job_categories", "hobbies", "regions"}
	defaults := map[string]string{
		"job_categories": strings.Join(models.JobCategories, ", "),
		"hobbies":        strings.Join(models.Hobbies, ", "),
		"regions":        strings.Join(models.Regions, ", "),
	}

	for _, key := range keys {
		var value string
		err := db.QueryRow("SELECT value FROM settings WHERE key_name = ?", "ref_"+key).Scan(&value)
		if err != nil || value == "" {
			result[key] = defaults[key]
		} else {
			result[key] = value
		}
	}
	return result
}

// SaveCharacterRefValue 캐릭터 생성 참조값 저장 (중복 자동 제거)
func (a *App) SaveCharacterRefValue(key, value string) error {
	db := a.db.GetDB()
	if db == nil {
		return fmt.Errorf("데이터베이스에 연결되지 않았습니다")
	}

	// 줄바꿈을 쉼표로 치환하고, 쉼표로 분리 후 중복 제거
	value = strings.ReplaceAll(value, "\r\n", ",")
	value = strings.ReplaceAll(value, "\n", ",")
	items := strings.Split(value, ",")
	seen := make(map[string]bool)
	uniqueItems := []string{}
	for _, item := range items {
		trimmed := strings.TrimSpace(item)
		trimmed = strings.Trim(trimmed, "\"'") // 따옴표 자동 제거 (Foolproof)
		if trimmed != "" && !seen[trimmed] {
			seen[trimmed] = true
			uniqueItems = append(uniqueItems, trimmed)
		}
	}
	cleanedValue := strings.Join(uniqueItems, ", ")

	_, err := db.Exec(`
		INSERT OR REPLACE INTO settings (key_name, value) VALUES (?, ?)
	`, "ref_"+key, cleanedValue)
	return err
}

// SaveGenSettings 캐릭터 생성 기본 설정 저장
func (a *App) SaveGenSettings(minAge, maxAge, maleRatio int, useAge, useGender bool) error {
	db := a.db.GetDB()
	if db == nil {
		return fmt.Errorf("데이터베이스에 연결되지 않았습니다")
	}

	tx, err := db.Begin()
	if err != nil {
		return err
	}

	queries := map[string]string{
		"ref_min_age":    strconv.Itoa(minAge),
		"ref_max_age":    strconv.Itoa(maxAge),
		"ref_male_ratio": strconv.Itoa(maleRatio),
		"ref_use_age":    strconv.FormatBool(useAge),
		"ref_use_gender": strconv.FormatBool(useGender),
	}

	for k, v := range queries {
		_, err := tx.Exec("INSERT OR REPLACE INTO settings (key_name, value) VALUES (?, ?)", k, v)
		if err != nil {
			tx.Rollback()
			return err
		}
	}

	return tx.Commit()
}

// GetGenSettings 캐릭터 생성 기본 설정 조회
func (a *App) GetGenSettings() (map[string]interface{}, error) {
	db := a.db.GetDB()
	if db == nil {
		return nil, fmt.Errorf("데이터베이스에 연결되지 않았습니다")
	}

	settings := map[string]interface{}{
		"min_age":    15, // 기본값
		"max_age":    64,
		"male_ratio": 50,
		"use_age":    true, // 기본적으로는 true로 두되, DB 없으면 true
		"use_gender": true,
	}

	// Int 값 로드
	intKeys := []string{"ref_min_age", "ref_max_age", "ref_male_ratio"}
	for _, key := range intKeys {
		var valStr string
		err := db.QueryRow("SELECT value FROM settings WHERE key_name = ?", key).Scan(&valStr)
		if err == nil {
			val, err := strconv.Atoi(valStr)
			if err == nil {
				settings[strings.TrimPrefix(key, "ref_")] = val
			}
		}
	}

	// Bool 값 로드
	boolKeys := []string{"ref_use_age", "ref_use_gender"}
	for _, key := range boolKeys {
		var valStr string
		err := db.QueryRow("SELECT value FROM settings WHERE key_name = ?", key).Scan(&valStr)
		if err == nil {
			val, err := strconv.ParseBool(valStr)
			if err == nil {
				settings[strings.TrimPrefix(key, "ref_")] = val
			} else {
				// 1/0 or true/false
				if valStr == "1" || valStr == "true" {
					settings[strings.TrimPrefix(key, "ref_")] = true
				} else {
					settings[strings.TrimPrefix(key, "ref_")] = false
				}
			}
		} else {
			// DB에 값이 없으면 기본적으로 false (사용자가 의도적으로 켜야 함)
			// 요청사항: "체크한 것만 ... 그렇지 않은것은 이전처럼"
			// 이전처럼 == 15~64, 50:50.
			// 그러니 기본값 false로 두는 게 맞을 수도 있다?
			// 아니면 UI상 기본 체크 여부.
			// 여기서는 기본값을 false로 둡니다.
			settings[strings.TrimPrefix(key, "ref_")] = false
		}
	}

	return settings, nil
}

// ResetCharacterRefValue 캐릭터 생성 참조값 기본값으로 초기화
func (a *App) ResetCharacterRefValue(key string) (string, error) {
	db := a.db.GetDB()
	if db == nil {
		return "", fmt.Errorf("데이터베이스에 연결되지 않았습니다")
	}

	defaults := map[string]string{
		"job_categories": strings.Join(models.JobCategories, ", "),
		"hobbies":        strings.Join(models.Hobbies, ", "),
		"regions":        strings.Join(models.Regions, ", "),
	}

	defaultValue, ok := defaults[key]
	if !ok {
		return "", fmt.Errorf("알 수 없는 키: %s", key)
	}

	// DB에서 삭제 (다음 조회 시 기본값 반환)
	_, _ = db.Exec("DELETE FROM settings WHERE key_name = ?", "ref_"+key)
	return defaultValue, nil
}

// BatchSetCharacterActive 여러 캐릭터 활성화/비활성화
func (a *App) BatchSetCharacterActive(ids []int, active bool) error {
	db := a.db.GetDB()
	if db == nil {
		return fmt.Errorf("데이터베이스에 연결되지 않았습니다")
	}

	if len(ids) == 0 {
		return fmt.Errorf("선택된 캐릭터가 없습니다")
	}

	// 쿼리 생성
	placeholders := make([]string, len(ids))
	args := make([]interface{}, len(ids)+1)
	activeVal := 0
	if active {
		activeVal = 1
	}
	args[0] = activeVal
	for i, id := range ids {
		placeholders[i] = "?"
		args[i+1] = id
	}

	query := fmt.Sprintf("UPDATE ai_characters SET is_active = ? WHERE id IN (%s)",
		strings.Join(placeholders, ","))
	_, err := db.Exec(query, args...)
	return err
}

// BatchDeleteCharacters 여러 캐릭터 삭제
func (a *App) BatchDeleteCharacters(ids []int) error {
	db := a.db.GetDB()
	if db == nil {
		return fmt.Errorf("데이터베이스에 연결되지 않았습니다")
	}

	if len(ids) == 0 {
		return fmt.Errorf("선택된 캐릭터가 없습니다")
	}

	// 쿼리 생성
	placeholders := make([]string, len(ids))
	args := make([]interface{}, len(ids))
	for i, id := range ids {
		placeholders[i] = "?"
		args[i] = id
	}

	query := fmt.Sprintf("DELETE FROM ai_characters WHERE id IN (%s)",
		strings.Join(placeholders, ","))
	_, err := db.Exec(query, args...)
	return err
}

// ================================
// 다중 데이터베이스 관리
// ================================

// GetDatabaseList DB 파일 목록 반환
func (a *App) GetDatabaseList() []string {
	list, err := a.db.ListDatabases()
	if err != nil {
		log.Printf("DB 목록 조회 실패: %v", err)
		return []string{}
	}
	return list
}

// GetCurrentDatabase 현재 DB 이름 반환
func (a *App) GetCurrentDatabase() string {
	return a.db.GetCurrentDBName()
}

// SwitchDatabase 다른 DB로 전환
func (a *App) SwitchDatabase(name string) error {
	// AI 활동 중지
	if a.activityManager != nil {
		a.activityManager.Stop()
	}

	// 웹 서버 중지
	if a.webServer != nil {
		a.webServer.Stop()
	}

	// DB 전환
	if err := a.db.SwitchDatabase(name); err != nil {
		return err
	}

	// 마이그레이션 실행
	a.db.Migrate()

	// 마지막 사용 DB 저장
	execPath, _ := os.Executable()
	execDir := filepath.Dir(execPath)
	lastDBFile := filepath.Join(execDir, "last_db.txt")
	os.WriteFile(lastDBFile, []byte(name), 0644)

	// 서비스 재초기화 (DB 인스턴스는 동일하므로 재생성 불필요)
	// 단, 캐시된 데이터가 있다면 초기화 필요
	log.Printf("데이터베이스 전환됨: %s", name)

	return nil
}

// CreateNewDatabase 새 DB 생성
func (a *App) CreateNewDatabase(name string) error {
	// 확장자 추가
	if !strings.HasSuffix(name, ".db") {
		name = name + ".db"
	}

	// AI 활동 중지
	if a.activityManager != nil {
		a.activityManager.Stop()
	}

	// 웹 서버 중지
	if a.webServer != nil {
		a.webServer.Stop()
	}

	// 새 DB 생성
	if err := a.db.CreateNewDatabase(name, schemaSQL); err != nil {
		return err
	}

	// 마지막 사용 DB 저장
	execPath, _ := os.Executable()
	execDir := filepath.Dir(execPath)
	lastDBFile := filepath.Join(execDir, "last_db.txt")
	os.WriteFile(lastDBFile, []byte(name), 0644)

	// 프롬프트 테이블 초기화
	a.initPromptTables()

	log.Printf("새 데이터베이스 생성됨: %s", name)
	return nil
}

// DeleteDatabase DB 파일 삭제 (default.db는 재생성)
func (a *App) DeleteDatabase(confirmation string) error {
	if confirmation != "데이터베이스 삭제" {
		return fmt.Errorf("확인 문구가 올바르지 않습니다")
	}

	currentDB := a.db.GetCurrentDBName()

	// AI 활동 중지
	if a.activityManager != nil {
		a.activityManager.Stop()
	}

	// 웹 서버 중지
	if a.webServer != nil {
		a.webServer.Stop()
	}

	execPath, _ := os.Executable()
	execDir := filepath.Dir(execPath)
	dbPath := filepath.Join(execDir, currentDB)

	// DB 연결 해제
	if db := a.db.GetDB(); db != nil {
		db.Close()
	}

	// 파일 삭제
	if err := os.Remove(dbPath); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("DB 파일 삭제 실패: %w", err)
	}

	log.Printf("데이터베이스 삭제됨: %s", currentDB)

	// default.db인 경우 재생성, 아닌 경우 default.db로 전환
	if currentDB == "default.db" {
		// 재생성
		a.db.SetDBPath(dbPath)
		if err := a.db.Connect(); err != nil {
			return fmt.Errorf("DB 재생성 실패: %w", err)
		}
		a.db.ExecuteSchema(schemaSQL)
		a.db.Migrate()
		a.initPromptTables()
		log.Printf("default.db 재생성됨")
	} else {
		// default.db로 전환
		defaultPath := filepath.Join(execDir, "default.db")
		if _, err := os.Stat(defaultPath); os.IsNotExist(err) {
			// default.db가 없으면 생성
			a.db.SetDBPath(defaultPath)
			if err := a.db.Connect(); err != nil {
				return fmt.Errorf("default.db 생성 실패: %w", err)
			}
			a.db.ExecuteSchema(schemaSQL)
			a.db.Migrate()
			a.initPromptTables()
		} else {
			// default.db가 있으면 전환
			if err := a.db.SwitchDatabase("default.db"); err != nil {
				return fmt.Errorf("default.db 전환 실패: %w", err)
			}
			a.db.Migrate()
		}

		// last_db 업데이트
		lastDBFile := filepath.Join(execDir, "last_db.txt")
		os.WriteFile(lastDBFile, []byte("default.db"), 0644)
	}

	return nil
}

// GetDatabaseInfo 현재 DB 정보 반환
func (a *App) GetDatabaseInfo() map[string]interface{} {
	result := make(map[string]interface{})

	db := a.db.GetDB()
	if db == nil {
		result["error"] = "연결 안됨"
		return result
	}

	// 현재 DB 이름
	result["name"] = a.db.GetCurrentDBName()
	log.Printf("[DEBUG] GetDatabaseInfo: DB name = %s", result["name"])

	// 파일 크기
	size, err := a.db.GetDatabaseSize()
	if err == nil {
		result["size"] = size
	}

	// 게시글 수
	var postCount int
	db.QueryRow("SELECT COUNT(*) FROM posts").Scan(&postCount)
	result["postCount"] = postCount

	// 댓글 수
	var commentCount int
	db.QueryRow("SELECT COUNT(*) FROM comments").Scan(&commentCount)
	result["commentCount"] = commentCount

	// 게시판 타이틀 (settings 테이블에서 - 키는 bbs_title)
	var title string
	err = db.QueryRow("SELECT value FROM settings WHERE key_name = 'bbs_title'").Scan(&title)
	if err != nil {
		log.Printf("[DEBUG] GetDatabaseInfo: bbs_title query error: %v", err)
	}
	if title == "" {
		title = "DINKIssTyle AI BBS"
	}
	result["title"] = title
	log.Printf("[DEBUG] GetDatabaseInfo: title = %s", title)

	// 시스템 롤 (prompts 테이블에서 - 컬럼명은 content)
	var systemRole string
	err = db.QueryRow("SELECT content FROM prompts WHERE key_name = 'system_role'").Scan(&systemRole)
	if err != nil {
		log.Printf("[DEBUG] GetDatabaseInfo: system_role query error: %v", err)
	}
	if systemRole == "" {
		systemRole = "(기본값 사용)"
	}
	result["systemRole"] = systemRole
	log.Printf("[DEBUG] GetDatabaseInfo: systemRole = %s", systemRole)

	return result
}

// ExportReferenceValues 참조값 내보내기
func (a *App) ExportReferenceValues() (string, error) {
	// 현재 값 가져오기
	data := a.GetCharacterRefValues()

	// 텍스트 포맷팅
	var builder strings.Builder
	builder.WriteString("[직종]\n")
	builder.WriteString(data["job_categories"])
	builder.WriteString("\n\n")

	builder.WriteString("[취미]\n")
	builder.WriteString(data["hobbies"])
	builder.WriteString("\n\n")

	builder.WriteString("[지역]\n")
	builder.WriteString(data["regions"])
	builder.WriteString("\n")

	// 저장 다이얼로그 (DefaultFilename에 접미사 포함)
	filename, err := runtime.SaveFileDialog(a.ctx, runtime.SaveDialogOptions{
		Title:           "참조값 내보내기",
		DefaultFilename: "data_생성참조값.txt",
		Filters:         []runtime.FileFilter{{DisplayName: "Text Files (*.txt)", Pattern: "*.txt"}},
	})

	if err != nil || filename == "" {
		return "", nil // 취소됨
	}

	err = os.WriteFile(filename, []byte(builder.String()), 0644)
	if err != nil {
		return "", err
	}
	return filename, nil
}

// ImportReferenceValues 참조값 불러오기
func (a *App) ImportReferenceValues() (string, error) {
	filename, err := runtime.OpenFileDialog(a.ctx, runtime.OpenDialogOptions{
		Title:   "참조값 불러오기 (UTF-8)",
		Filters: []runtime.FileFilter{{DisplayName: "Text Files (*.txt)", Pattern: "*.txt"}},
	})

	if err != nil || filename == "" {
		return "", nil
	}

	contentBytes, err := os.ReadFile(filename)
	if err != nil {
		return "", err
	}
	content := string(contentBytes)

	// 파싱
	lines := strings.Split(strings.ReplaceAll(content, "\r\n", "\n"), "\n")
	var currentSection string
	sections := make(map[string][]string)

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		if strings.HasPrefix(line, "[") && strings.HasSuffix(line, "]") {
			currentSection = line
			continue
		}

		if currentSection != "" {
			sections[currentSection] = append(sections[currentSection], line)
		}
	}

	// DB 저장
	updates := 0
	if val, ok := sections["[직종]"]; ok {
		// 줄바꿈된 데이터들을 쉼표로 연결하여 저장 (SaveCharacterRefValue 내부에서 파싱/중복제거)
		a.SaveCharacterRefValue("job_categories", strings.Join(val, ","))
		updates++
	}
	if val, ok := sections["[취미]"]; ok {
		a.SaveCharacterRefValue("hobbies", strings.Join(val, ","))
		updates++
	}
	if val, ok := sections["[지역]"]; ok {
		a.SaveCharacterRefValue("regions", strings.Join(val, ","))
		updates++
	}

	if updates == 0 {
		return "", fmt.Errorf("유효한 데이터 섹션([직종], [취미], [지역])을 찾지 못했습니다")
	}

	return filename, nil
}

// ExportPromptSettings 프롬프트 설정 내보내기
func (a *App) ExportPromptSettings() (string, error) {
	var builder strings.Builder

	// 기본 프롬프트들
	prompts := map[string]string{
		"nickname_gen":        "[AI 캐릭터 닉네임 생성 프롬프트]",
		"system_role":         "[시스템 롤]",
		"post_instruction":    "[게시글 작성 지시문]",
		"comment_instruction": "[댓글 작성 지시문]",
		"reply_instruction":   "[답글 작성 지시문]",
		"summary_instruction": "[AI 캐릭터 요약 지시문]",
	}

	// 순서 보장을 위해 키 슬라이스 사용
	orderedKeys := []string{"nickname_gen", "system_role", "post_instruction", "comment_instruction", "reply_instruction", "summary_instruction"}

	for _, key := range orderedKeys {
		header := prompts[key]
		content := a.GetPrompt(key)
		builder.WriteString(header + "\n")
		builder.WriteString(content + "\n\n")
	}

	// MBTI 설명
	descriptions := a.GetMBTIDescriptions()
	// MBTI 순서 (모델 정의 순)
	for _, mbti := range models.MBTITypes {
		if content, ok := descriptions[mbti]; ok {
			builder.WriteString("[" + mbti + "]\n")
			builder.WriteString(content + "\n\n")
		}
	}

	// 저장 다이얼로그
	filename, err := runtime.SaveFileDialog(a.ctx, runtime.SaveDialogOptions{
		Title:           "프롬프트 설정 내보내기",
		DefaultFilename: "data_프롬프트.txt",
		Filters:         []runtime.FileFilter{{DisplayName: "Text Files (*.txt)", Pattern: "*.txt"}},
	})

	if err != nil || filename == "" {
		return "", nil
	}

	err = os.WriteFile(filename, []byte(builder.String()), 0644)
	if err != nil {
		return "", err
	}
	return filename, nil
}

// ImportPromptSettings 프롬프트 설정 불러오기
func (a *App) ImportPromptSettings() (string, error) {
	filename, err := runtime.OpenFileDialog(a.ctx, runtime.OpenDialogOptions{
		Title:   "프롬프트 설정 불러오기 (UTF-8)",
		Filters: []runtime.FileFilter{{DisplayName: "Text Files (*.txt)", Pattern: "*.txt"}},
	})

	if err != nil || filename == "" {
		return "", nil
	}

	contentBytes, err := os.ReadFile(filename)
	if err != nil {
		return "", err
	}
	content := string(contentBytes)

	// 라인 단위 파싱
	lines := strings.Split(strings.ReplaceAll(content, "\r\n", "\n"), "\n")
	var currentSection string
	var currentContentBuilder strings.Builder
	sections := make(map[string]string)

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)

		// 섹션 헤더 감지 ([...])
		if strings.HasPrefix(trimmed, "[") && strings.HasSuffix(trimmed, "]") {
			// 이전 섹션 저장
			if currentSection != "" {
				sections[currentSection] = strings.TrimSpace(currentContentBuilder.String())
			}
			// 새 섹션 시작
			currentSection = trimmed
			currentContentBuilder.Reset()
		} else {
			// 내용 누적 (헤더가 설정된 상태여야 함)
			if currentSection != "" {
				currentContentBuilder.WriteString(line + "\n")
			}
		}
	}
	// 마지막 섹션 저장
	if currentSection != "" {
		sections[currentSection] = strings.TrimSpace(currentContentBuilder.String())
	}

	// 역매핑
	headerToKey := map[string]string{
		"[AI 캐릭터 닉네임 생성 프롬프트]": "nickname_gen",
		"[시스템 롤]":         "system_role",
		"[게시글 작성 지시문]":    "post_instruction",
		"[댓글 작성 지시문]":     "comment_instruction",
		"[답글 작성 지시문]":     "reply_instruction",
		"[AI 캐릭터 요약 지시문]": "summary_instruction",
	}

	updates := 0
	for header, val := range sections {
		// 일반 프롬프트 업데이트
		if key, ok := headerToKey[header]; ok {
			a.SavePrompt(key, val)
			updates++
			continue
		}

		// MBTI 업데이트 체크
		mbti := strings.TrimSuffix(strings.TrimPrefix(header, "["), "]")
		isMBTI := false
		for _, t := range models.MBTITypes {
			if t == mbti {
				isMBTI = true
				break
			}
		}
		if isMBTI {
			a.SaveMBTIDescription(mbti, val)
			updates++
		}
	}

	if updates == 0 {
		return "", fmt.Errorf("유효한 프롬프트 데이터를 찾지 못했습니다")
	}

	return filename, nil
}
