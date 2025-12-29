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
	return err
}

// ResetDatabase 데이터베이스를 초기화합니다.
func (d *Database) ResetDatabase() error {
	d.mu.Lock()
	defer d.mu.Unlock()

	if d.db == nil {
		return fmt.Errorf("데이터베이스에 연결되지 않았습니다")
	}

	tables := []string{"recommendations", "comments", "posts", "ai_characters", "users", "settings"}
	for _, table := range tables {
		_, err := d.db.Exec("DROP TABLE IF EXISTS " + table)
		if err != nil {
			return fmt.Errorf("테이블 %s 삭제 실패: %w", table, err)
		}
	}

	return nil
}
