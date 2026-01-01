// Created by DINKIssTyle on 2025. Copyright (C) 2025 DINKI'ssTyle. All rights reserved.

package database

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sync"

	_ "modernc.org/sqlite"
)

// Database 구조체는 SQLite 연결을 관리합니다.
type Database struct {
	db     *sql.DB
	dbPath string
	mu     sync.RWMutex
}

var (
	instance *Database
	once     sync.Once
)

// GetInstance 싱글톤 패턴으로 Database 인스턴스를 반환합니다.
func GetInstance() *Database {
	once.Do(func() {
		instance = &Database{}
	})
	return instance
}

// SetDBPath 데이터베이스 파일 경로를 설정합니다.
func (d *Database) SetDBPath(path string) {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.dbPath = path
}

// GetDBPath 데이터베이스 파일 경로를 반환합니다.
func (d *Database) GetDBPath() string {
	d.mu.RLock()
	defer d.mu.RUnlock()
	return d.dbPath
}

// Connect 데이터베이스에 연결합니다.
func (d *Database) Connect() error {
	d.mu.Lock()
	defer d.mu.Unlock()

	if d.db != nil {
		return nil
	}

	// 데이터베이스 파일 경로 확인
	if d.dbPath == "" {
		// 기본 경로: 실행 파일과 같은 디렉토리
		execPath, err := os.Executable()
		if err != nil {
			execPath = "."
		}
		d.dbPath = filepath.Join(filepath.Dir(execPath), "aibbs.db")
	}

	// 디렉토리가 없으면 생성
	dir := filepath.Dir(d.dbPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("디렉토리 생성 실패: %w", err)
	}

	// SQLite 연결 (modernc.org/sqlite 드라이버는 "sqlite"로 등록됨)
	db, err := sql.Open("sqlite", d.dbPath+"?_foreign_keys=on")
	if err != nil {
		return fmt.Errorf("데이터베이스 연결 실패: %w", err)
	}

	// 연결 테스트
	if err := db.Ping(); err != nil {
		return fmt.Errorf("데이터베이스 핑 실패: %w", err)
	}

	d.db = db

	// SQLite 최적화 및 동시성 설정
	_, err = db.Exec("PRAGMA journal_mode=WAL;")
	if err != nil {
		fmt.Printf("[WARNING] WAL 모드 설정 실패: %v\n", err)
	}
	_, err = db.Exec("PRAGMA busy_timeout=10000;")
	if err != nil {
		fmt.Printf("[WARNING] Busy Timeout 설정 실패: %v\n", err)
	}

	// 외래키 제약조건 활성화 (중요)
	_, err = db.Exec("PRAGMA foreign_keys = ON;")
	if err != nil {
		fmt.Printf("[WARNING] 외래키 제약조건 활성화 실패: %v\n", err)
	}

	return nil
}

// Close 데이터베이스 연결을 종료합니다.
func (d *Database) Close() error {
	d.mu.Lock()
	defer d.mu.Unlock()

	if d.db != nil {
		err := d.db.Close()
		d.db = nil
		return err
	}
	return nil
}

// GetDB 데이터베이스 연결을 반환합니다.
func (d *Database) GetDB() *sql.DB {
	d.mu.RLock()
	defer d.mu.RUnlock()
	return d.db
}

// IsConnected 데이터베이스 연결 상태를 확인합니다.
func (d *Database) IsConnected() bool {
	d.mu.RLock()
	defer d.mu.RUnlock()
	return d.db != nil
}

// ExecuteSchema 스키마 SQL을 실행합니다.
func (d *Database) ExecuteSchema(schemaSQL string) error {
	d.mu.RLock()
	defer d.mu.RUnlock()

	if d.db == nil {
		return fmt.Errorf("데이터베이스에 연결되지 않았습니다")
	}

	_, err := d.db.Exec(schemaSQL)
	if err != nil {
		return err
	}

	// 마이그레이션 실행
	return d.Migrate()
}

// Migrate 데이터베이스 마이그레이션 (필요한 컬럼 추가 등)
func (d *Database) Migrate() error {
	// ai_characters 테이블에 persona_updated_at 컬럼이 없으면 추가
	rows, err := d.db.Query("PRAGMA table_info(ai_characters)")
	if err != nil {
		return err
	}
	defer rows.Close()

	hasPersonaUpdatedAt := false
	for rows.Next() {
		var cid int
		var name, dtype string
		var notnull int
		var dfltValue interface{}
		var pk int
		if err := rows.Scan(&cid, &name, &dtype, &notnull, &dfltValue, &pk); err != nil {
			return err
		}
		if name == "persona_updated_at" {
			hasPersonaUpdatedAt = true
			break
		}
	}

	if !hasPersonaUpdatedAt {
		_, err := d.db.Exec("ALTER TABLE ai_characters ADD COLUMN persona_updated_at DATETIME")
		if err != nil {
			return fmt.Errorf("ai_characters 마이그레이션 실패: %w", err)
		}
		fmt.Println("[DEBUG] ai_characters 테이블에 persona_updated_at 컬럼을 추가했습니다.")
	}

	// ai_characters 테이블에 avatar_image 컬럼이 없으면 추가
	rows, err = d.db.Query("PRAGMA table_info(ai_characters)")
	if err != nil {
		return err
	}
	defer rows.Close()

	hasAvatarImage := false
	for rows.Next() {
		var cid int
		var name, dtype string
		var notnull int
		var dfltValue interface{}
		var pk int
		if err := rows.Scan(&cid, &name, &dtype, &notnull, &dfltValue, &pk); err != nil {
			return err
		}
		if name == "avatar_image" {
			hasAvatarImage = true
			break
		}
	}

	if !hasAvatarImage {
		_, err := d.db.Exec("ALTER TABLE ai_characters ADD COLUMN avatar_image TEXT")
		if err != nil {
			// 이미 존재할 수도 있으므로 에러 무시 혹은 로깅
			fmt.Printf("[WARNING] avatar_image 컬럼 추가 실패 (이미 존재할 수 있음): %v\n", err)
		} else {
			fmt.Println("[DEBUG] ai_characters 테이블에 avatar_image 컬럼을 추가했습니다.")
		}
	}

	// posts 테이블에 is_pinned 컬럼이 없으면 추가
	rows, err = d.db.Query("PRAGMA table_info(posts)")
	if err != nil {
		return err
	}
	defer rows.Close()

	hasIsPinned := false
	for rows.Next() {
		var cid int
		var name, dtype string
		var notnull int
		var dfltValue interface{}
		var pk int
		if err := rows.Scan(&cid, &name, &dtype, &notnull, &dfltValue, &pk); err != nil {
			return err
		}
		if name == "is_pinned" {
			hasIsPinned = true
			break
		}
	}

	if !hasIsPinned {
		_, err := d.db.Exec("ALTER TABLE posts ADD COLUMN is_pinned INTEGER DEFAULT 0")
		if err != nil {
			return fmt.Errorf("posts 마이그레이션 실패 (is_pinned): %w", err)
		}
		fmt.Println("[DEBUG] posts 테이블에 is_pinned 컬럼을 추가했습니다.")
	}

	// users 테이블에 설정 컬럼들 추가
	userSettingsColumns := []struct {
		name     string
		sqlType  string
		defValue string
	}{
		{"theme", "TEXT", "'dark'"},
		{"font_style", "TEXT", "'default'"},
		{"timezone", "TEXT", "'Asia/Seoul'"},
		{"posts_per_page", "INTEGER", "20"},
	}

	for _, col := range userSettingsColumns {
		rows, err = d.db.Query("PRAGMA table_info(users)")
		if err != nil {
			return err
		}

		hasColumn := false
		for rows.Next() {
			var cid int
			var name, dtype string
			var notnull int
			var dfltValue interface{}
			var pk int
			if err := rows.Scan(&cid, &name, &dtype, &notnull, &dfltValue, &pk); err != nil {
				rows.Close()
				return err
			}
			if name == col.name {
				hasColumn = true
				break
			}
		}
		rows.Close()

		if !hasColumn {
			_, err := d.db.Exec(fmt.Sprintf("ALTER TABLE users ADD COLUMN %s %s DEFAULT %s", col.name, col.sqlType, col.defValue))
			if err != nil {
				return fmt.Errorf("users 마이그레이션 실패 (%s): %w", col.name, err)
			}
			fmt.Printf("[DEBUG] users 테이블에 %s 컬럼을 추가했습니다.\n", col.name)
		}
	}

	// ai_characters 테이블에 avatar_image 컬럼 추가 (없을 경우)
	{
		var count int
		err := d.db.QueryRow("SELECT COUNT(*) FROM pragma_table_info('ai_characters') WHERE name='avatar_image'").Scan(&count)
		if err == nil && count == 0 {
			_, err = d.db.Exec("ALTER TABLE ai_characters ADD COLUMN avatar_image TEXT")
			if err != nil {
				log.Printf("[WARNING] avatar_image 컬럼 추가 실패: %v", err)
			} else {
				log.Println("[INFO] ai_characters 테이블에 avatar_image 컬럼을 추가했습니다.")
			}
		}
	}

	// ai_characters 테이블에 backstory 컬럼 추가 (없을 경우)
	{
		var count int
		err := d.db.QueryRow("SELECT COUNT(*) FROM pragma_table_info('ai_characters') WHERE name='backstory'").Scan(&count)
		if err == nil && count == 0 {
			_, err = d.db.Exec("ALTER TABLE ai_characters ADD COLUMN backstory TEXT")
			if err != nil {
				log.Printf("[WARNING] backstory 컬럼 추가 실패: %v", err)
			} else {
				log.Println("[INFO] ai_characters 테이블에 backstory 컬럼을 추가했습니다.")
			}
		}
	}

	// 기본 타임아웃 설정 추가 (없을 경우)
	_, err = d.db.Exec("INSERT OR IGNORE INTO settings (key_name, value) VALUES ('timeout', '120')")
	if err != nil {
		fmt.Printf("[WARNING] 타임아웃 설정 추가 실패: %v\n", err)
	}

	return nil
}

// ClearContent 게시물, 댓글, 추천 기록 등 콘텐츠만 삭제합니다. (설정, 유저, 캐릭터 유지)
func (d *Database) ClearContent() error {
	d.mu.Lock()
	defer d.mu.Unlock()

	if d.db == nil {
		return fmt.Errorf("데이터베이스에 연결되지 않았습니다")
	}

	tx, err := d.db.Begin()
	if err != nil {
		return err
	}

	// 순서 중요 (FK 제약조건 등)
	queries := []string{
		"DELETE FROM recommendations",
		"DELETE FROM comments",
		"DELETE FROM posts",
		"DELETE FROM sqlite_sequence WHERE name IN ('posts', 'comments', 'recommendations')",
		"UPDATE ai_characters SET post_count = 0, comment_count = 0",
	}

	for _, query := range queries {
		if _, err := tx.Exec(query); err != nil {
			tx.Rollback()
			return fmt.Errorf("쿼리 실행 실패 (%s): %w", query, err)
		}
	}

	return tx.Commit()
}

// ResetDatabase 데이터베이스를 초기화합니다 (파일 삭제 후 재생성).
func (d *Database) ResetDatabase() error {
	d.mu.Lock()
	defer d.mu.Unlock()

	if d.db == nil {
		return fmt.Errorf("데이터베이스에 연결되지 않았습니다")
	}

	dbPath := d.dbPath

	// 1. 기존 연결 닫기
	d.db.Close()
	d.db = nil

	// 2. DB 파일 삭제
	if err := os.Remove(dbPath); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("DB 파일 삭제 실패: %w", err)
	}

	// 3. 새 DB 연결
	db, err := sql.Open("sqlite", dbPath+"?_foreign_keys=on")
	if err != nil {
		return fmt.Errorf("DB 재연결 실패: %w", err)
	}
	d.db = db

	// SQLite 최적화 및 동시성 설정 (재연결 후에도 적용)
	db.Exec("PRAGMA journal_mode=WAL;")
	db.Exec("PRAGMA busy_timeout=10000;")

	return nil
}

// ListDatabases 실행 파일 디렉토리에서 .db 파일 목록을 반환합니다.
func (d *Database) ListDatabases() ([]string, error) {
	d.mu.RLock()
	defer d.mu.RUnlock()

	if d.dbPath == "" {
		return nil, fmt.Errorf("데이터베이스 경로가 설정되지 않았습니다")
	}

	dir := filepath.Dir(d.dbPath)
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, fmt.Errorf("디렉토리 읽기 실패: %w", err)
	}

	var dbFiles []string
	for _, entry := range entries {
		if !entry.IsDir() && filepath.Ext(entry.Name()) == ".db" {
			dbFiles = append(dbFiles, entry.Name())
		}
	}

	return dbFiles, nil
}

// SwitchDatabase 다른 데이터베이스로 전환합니다.
func (d *Database) SwitchDatabase(name string) error {
	d.mu.Lock()
	defer d.mu.Unlock()

	// 기존 연결 닫기
	if d.db != nil {
		d.db.Close()
		d.db = nil
	}

	// 새 경로 설정
	dir := filepath.Dir(d.dbPath)
	newPath := filepath.Join(dir, name)
	d.dbPath = newPath

	// 파일 존재 확인
	if _, err := os.Stat(newPath); os.IsNotExist(err) {
		return fmt.Errorf("데이터베이스 파일이 존재하지 않습니다: %s", name)
	}

	// 새 DB 연결
	db, err := sql.Open("sqlite", newPath+"?_foreign_keys=on")
	if err != nil {
		return fmt.Errorf("DB 연결 실패: %w", err)
	}

	if err := db.Ping(); err != nil {
		return fmt.Errorf("DB 핑 실패: %w", err)
	}

	d.db = db

	// SQLite 최적화
	db.Exec("PRAGMA journal_mode=WAL;")
	db.Exec("PRAGMA busy_timeout=10000;")

	fmt.Printf("[DEBUG] 데이터베이스 전환됨: %s\n", name)
	return nil
}

// CreateNewDatabase 새 데이터베이스 파일을 생성하고 스키마를 적용합니다.
func (d *Database) CreateNewDatabase(name string, schemaSQL string) error {
	d.mu.Lock()
	defer d.mu.Unlock()

	dir := filepath.Dir(d.dbPath)
	newPath := filepath.Join(dir, name)

	// 이미 존재하는지 확인
	if _, err := os.Stat(newPath); err == nil {
		return fmt.Errorf("데이터베이스가 이미 존재합니다: %s", name)
	}

	// 기존 연결 닫기
	if d.db != nil {
		d.db.Close()
		d.db = nil
	}

	// 새 DB 생성 및 연결
	db, err := sql.Open("sqlite", newPath+"?_foreign_keys=on")
	if err != nil {
		return fmt.Errorf("DB 생성 실패: %w", err)
	}

	d.db = db
	d.dbPath = newPath

	// SQLite 최적화
	db.Exec("PRAGMA journal_mode=WAL;")
	db.Exec("PRAGMA busy_timeout=10000;")

	// 스키마 적용
	if schemaSQL != "" {
		if _, err := db.Exec(schemaSQL); err != nil {
			return fmt.Errorf("스키마 적용 실패: %w", err)
		}
	}

	// 마이그레이션 실행
	d.mu.Unlock()
	err = d.Migrate()
	d.mu.Lock()
	if err != nil {
		return fmt.Errorf("마이그레이션 실패: %w", err)
	}

	fmt.Printf("[DEBUG] 새 데이터베이스 생성됨: %s\n", name)
	return nil
}

// GetDatabaseSize 현재 데이터베이스 파일 크기를 바이트 단위로 반환합니다.
func (d *Database) GetDatabaseSize() (int64, error) {
	d.mu.RLock()
	defer d.mu.RUnlock()

	info, err := os.Stat(d.dbPath)
	if err != nil {
		return 0, fmt.Errorf("파일 정보 조회 실패: %w", err)
	}

	return info.Size(), nil
}

// GetCurrentDBName 현재 데이터베이스 파일 이름만 반환합니다.
func (d *Database) GetCurrentDBName() string {
	d.mu.RLock()
	defer d.mu.RUnlock()
	return filepath.Base(d.dbPath)
}
