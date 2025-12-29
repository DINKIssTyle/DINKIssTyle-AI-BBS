#!/bin/bash
# Created by DINKIssTyle on 2025. Copyright (C) 2025 DINKI'ssTyle. All rights reserved.
# DINKIssTyle AI BBS - Ubuntu/Linux 빌드 스크립트

set -e

echo "================================"
echo "DINKIssTyle AI BBS Build Script"
echo "Platform: Ubuntu/Linux"
echo "================================"

# 색상 정의
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# 시스템 패키지 매니저 확인
PACKAGE_MANAGER=""
if command -v apt-get &> /dev/null; then
    PACKAGE_MANAGER="apt"
elif command -v dnf &> /dev/null; then
    PACKAGE_MANAGER="dnf"
elif command -v pacman &> /dev/null; then
    PACKAGE_MANAGER="pacman"
fi

# Wails 필수 라이브러리 설치 (GTK, WebKit 등)
install_wails_deps() {
    echo -e "${YELLOW}[INFO] Wails 필수 라이브러리 설치 중...${NC}"
    if [ "$PACKAGE_MANAGER" = "apt" ]; then
        sudo apt-get update
        # 최신 Ubuntu 버전에서는 4.1-dev가 필요할 수 있으므로 함께 시도하거나 대체 권장
        sudo apt-get install -y libgtk-3-dev libwebkit2gtk-4.0-dev libwebkit2gtk-4.1-dev build-essential || \
        sudo apt-get install -y libgtk-3-dev libwebkit2gtk-4.0-dev build-essential
    elif [ "$PACKAGE_MANAGER" = "dnf" ]; then
        sudo dnf install -y gtk3-devel webkit2gtk3-devel gcc-c++
    elif [ "$PACKAGE_MANAGER" = "pacman" ]; then
        sudo pacman -S --noconfirm gtk3 webkit2gtk base-devel
    else
        echo -e "${RED}[ERROR] 지원하지 않는 패키지 매니저입니다.${NC}"
        echo "수동으로 GTK3, WebKit2GTK를 설치해주세요."
        exit 1
    fi
}

# Go 설치 확인
check_go() {
    if ! command -v go &> /dev/null; then
        echo -e "${YELLOW}[INFO] Go 설치 중...${NC}"
        if [ "$PACKAGE_MANAGER" = "apt" ]; then
            sudo apt-get install -y golang-go
        elif [ "$PACKAGE_MANAGER" = "dnf" ]; then
            sudo dnf install -y golang
        elif [ "$PACKAGE_MANAGER" = "pacman" ]; then
            sudo pacman -S --noconfirm go
        fi
        
        # PATH 설정
        export PATH=$PATH:/usr/local/go/bin:$HOME/go/bin
        echo 'export PATH=$PATH:/usr/local/go/bin:$HOME/go/bin' >> ~/.bashrc
    fi
    echo -e "${GREEN}[OK] Go: $(go version)${NC}"
}

# Node.js 설치 확인
check_node() {
    if ! command -v node &> /dev/null; then
        echo -e "${YELLOW}[INFO] Node.js 설치 중...${NC}"
        if [ "$PACKAGE_MANAGER" = "apt" ]; then
            curl -fsSL https://deb.nodesource.com/setup_lts.x | sudo -E bash -
            sudo apt-get install -y nodejs
        elif [ "$PACKAGE_MANAGER" = "dnf" ]; then
            sudo dnf install -y nodejs npm
        elif [ "$PACKAGE_MANAGER" = "pacman" ]; then
            sudo pacman -S --noconfirm nodejs npm
        fi
    fi
    echo -e "${GREEN}[OK] Node.js: $(node --version)${NC}"
}

# Wails 설치 확인
check_wails() {
    if ! command -v wails &> /dev/null; then
        echo -e "${YELLOW}[INFO] Wails CLI 설치 중...${NC}"
        go install github.com/wailsapp/wails/v2/cmd/wails@latest
        export PATH=$PATH:$HOME/go/bin
    fi
    echo -e "${GREEN}[OK] Wails: $(wails version 2>/dev/null || echo 'installed')${NC}"
}

# Wails 필수 라이브러리 확인
echo ""
echo "Wails 필수 라이브러리 확인..."
if ! pkg-config --exists gtk+-3.0 webkit2gtk-4.0 2>/dev/null; then
    install_wails_deps
fi

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
echo "실행 파일: build/bin/DINKIssTyle-AI-BBS"
echo "================================"
