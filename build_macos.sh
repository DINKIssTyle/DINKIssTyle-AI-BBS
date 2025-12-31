#!/bin/bash
# Created by DINKIssTyle on 2025. Copyright (C) 2025 DINKI'ssTyle. All rights reserved.
# DINKIssTyle AI BBS - macOS 빌드 스크립트

set -e

echo "================================"
echo "DINKIssTyle AI BBS Build Script"
echo "Platform: macOS"
echo "================================"

# 색상 정의
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Homebrew 설치 확인
check_homebrew() {
    if ! command -v brew &> /dev/null; then
        echo -e "${YELLOW}[INFO] Homebrew 설치 중...${NC}"
        /bin/bash -c "$(curl -fsSL https://raw.githubusercontent.com/Homebrew/install/HEAD/install.sh)"
    fi
}

# Go 설치 확인
check_go() {
    if ! command -v go &> /dev/null; then
        echo -e "${YELLOW}[INFO] Go 설치 중...${NC}"
        check_homebrew
        brew install go
    fi
    echo -e "${GREEN}[OK] Go: $(go version)${NC}"
}

# Node.js 설치 확인
check_node() {
    if ! command -v node &> /dev/null; then
        echo -e "${YELLOW}[INFO] Node.js 설치 중...${NC}"
        check_homebrew
        brew install node
    fi
    echo -e "${GREEN}[OK] Node.js: $(node --version)${NC}"
}

# Wails 설치 확인
check_wails() {
    if ! command -v wails &> /dev/null; then
        echo -e "${YELLOW}[INFO] Wails CLI 설치 중...${NC}"
        go install github.com/wailsapp/wails/v2/cmd/wails@latest
    fi
    echo -e "${GREEN}[OK] Wails: $(wails version 2>/dev/null || echo 'installed')${NC}"
}

# 의존성 확인
echo ""
echo "의존성 확인 중..."
check_go
check_node
check_wails

# Go 모듈 다운로드
echo ""
echo "[1/4] Go 모듈 다운로드..."
go mod download

# Go 모듈 정리
echo ""
echo "[2/4] Go 모듈 정리..."
go mod tidy

# 프론트엔드 의존성 설치
echo ""
echo "[3/4] 프론트엔드 의존성 설치..."
cd frontend
npm install
cd ..

# Wails 빌드
echo ""
echo "[4/4] Wails 빌드 중..."
wails build

echo ""
echo "================================"
echo -e "${GREEN}빌드 완료!${NC}"
echo "실행 파일: build/bin/DKST_AIBBS.app"
echo "================================"
