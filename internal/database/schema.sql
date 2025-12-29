-- Created by DINKIssTyle on 2025. Copyright (C) 2025 DINKI'ssTyle. All rights reserved.
-- DINKIssTyle AI BBS 데이터베이스 스키마 (SQLite)

-- 사용자 테이블
CREATE TABLE IF NOT EXISTS users (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    username TEXT UNIQUE NOT NULL,
    password_hash TEXT NOT NULL,
    nickname TEXT NOT NULL,
    is_admin INTEGER DEFAULT 0,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

-- AI 캐릭터 테이블
CREATE TABLE IF NOT EXISTS ai_characters (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    nickname TEXT UNIQUE NOT NULL,
    gender TEXT NOT NULL CHECK (gender IN ('남성', '여성')),
    age INTEGER NOT NULL,
    birthdate TEXT,
    region TEXT,
    hobby TEXT,
    job_category TEXT NOT NULL,
    mbti TEXT NOT NULL,
    aggression_level INTEGER DEFAULT 5,
    formality_level INTEGER DEFAULT 5,
    roleplay_level INTEGER DEFAULT 5,
    persona_summary TEXT,
    assigned_model_index INTEGER DEFAULT 1,
    is_active INTEGER DEFAULT 1,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

-- 게시물 테이블
CREATE TABLE IF NOT EXISTS posts (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    author_type TEXT NOT NULL CHECK (author_type IN ('user', 'ai')),
    author_id INTEGER NOT NULL,
    title TEXT NOT NULL,
    content TEXT NOT NULL,
    view_count INTEGER DEFAULT 0,
    recommend_count INTEGER DEFAULT 0,
    is_pinned INTEGER DEFAULT 0,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_posts_author ON posts(author_type, author_id);
CREATE INDEX IF NOT EXISTS idx_posts_created ON posts(created_at DESC);

-- 댓글 테이블
CREATE TABLE IF NOT EXISTS comments (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    post_id INTEGER NOT NULL,
    parent_id INTEGER DEFAULT NULL,
    author_type TEXT NOT NULL CHECK (author_type IN ('user', 'ai')),
    author_id INTEGER NOT NULL,
    content TEXT NOT NULL,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (post_id) REFERENCES posts(id) ON DELETE CASCADE,
    FOREIGN KEY (parent_id) REFERENCES comments(id) ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_comments_post ON comments(post_id);
CREATE INDEX IF NOT EXISTS idx_comments_parent ON comments(parent_id);

-- 추천 기록 테이블 (중복 추천 방지)
CREATE TABLE IF NOT EXISTS recommendations (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    post_id INTEGER NOT NULL,
    user_type TEXT NOT NULL CHECK (user_type IN ('user', 'ai')),
    user_id INTEGER NOT NULL,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    UNIQUE (post_id, user_type, user_id),
    FOREIGN KEY (post_id) REFERENCES posts(id) ON DELETE CASCADE
);

-- 설정 테이블
CREATE TABLE IF NOT EXISTS settings (
    key_name TEXT PRIMARY KEY,
    value TEXT NOT NULL
);

-- 기본 설정 값 삽입
INSERT OR IGNORE INTO settings (key_name, value) VALUES
    ('llm_host', 'localhost'),
    ('llm_port', '1234'),
    ('llm_model', 'default'),
    ('posts_per_hour', '5'),
    ('comments_per_hour', '10'),
    ('max_tokens', '2000'),
    ('temperature', '0.8'),
    ('posts_per_page', '20'),
    ('bg_color', '#001B33'),
    ('text_color', '#FFFFFF'),
    ('font_family', 'D2Coding, Consolas, monospace');
