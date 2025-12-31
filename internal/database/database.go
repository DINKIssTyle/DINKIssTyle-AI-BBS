// Created by DINKIssTyle on 2025. Copyright (C) 2025 DINKI'ssTyle. All rights reserved.

package database

import (
	"database/sql"
	"fmt"
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
	_, err = db.Exec("PRAGMA busy_timeout=5000;")
	if err != nil {
		fmt.Printf("[WARNING] Busy Timeout 설정 실패: %v\n", err)
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

	return nil
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
	db.Exec("PRAGMA busy_timeout=5000;")

	return nil
}
