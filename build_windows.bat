@echo off
REM Created by DINKIssTyle on 2025. Copyright (C) 2025 DINKI'ssTyle. All rights reserved.
REM DINKIssTyle AI BBS - Windows 빌드 스크립트

echo ================================
echo DINKIssTyle AI BBS Build Script
echo Platform: Windows
echo ================================

REM Go 설치 확인
where go >nul 2>&1
if %ERRORLEVEL% neq 0 (
    echo [ERROR] Go가 설치되어 있지 않습니다.
    echo Go를 다운로드하세요: https://go.dev/dl/
    pause
    exit /b 1
)

REM Wails 설치 확인 및 자동 설치
where wails >nul 2>&1
if %ERRORLEVEL% neq 0 (
    echo [INFO] Wails CLI 설치 중...
    go install github.com/wailsapp/wails/v2/cmd/wails@latest
    if %ERRORLEVEL% neq 0 (
        echo [ERROR] Wails 설치 실패
        pause
        exit /b 1
    )
)

REM Node.js 설치 확인
where node >nul 2>&1
if %ERRORLEVEL% neq 0 (
    echo [ERROR] Node.js가 설치되어 있지 않습니다.
    echo Node.js를 다운로드하세요: https://nodejs.org/
    pause
    exit /b 1
)

REM npm 설치 확인
where npm >nul 2>&1
if %ERRORLEVEL% neq 0 (
    echo [ERROR] npm이 설치되어 있지 않습니다.
    pause
    exit /b 1
)

echo.
echo [0/4] 아이콘 리소스 업데이트...
copy /Y "icon\app_icon_256.ico" "build\windows\icon.ico" >nul
copy /Y "icon\app_icon_512.png" "build\appicon.png" >nul
if exist "build\darwin" (
    copy /Y "icon\app_icon.icns" "build\darwin\icon.icns" >nul
)

echo.
echo [1/4] Go 모듈 다운로드...
go mod download
if %ERRORLEVEL% neq 0 (
    echo [ERROR] Go 모듈 다운로드 실패
    pause
    exit /b 1
)

echo.
echo [2/4] Go 모듈 정리...
go mod tidy

echo.
echo [3/4] 프론트엔드 의존성 설치...
cd frontend
call npm install
if %ERRORLEVEL% neq 0 (
    echo [ERROR] npm install 실패
    cd ..
    pause
    exit /b 1
)
cd ..

echo.
echo [4/4] Wails 빌드 중...
wails build
if %ERRORLEVEL% neq 0 (
    echo [ERROR] Wails 빌드 실패
    pause
    exit /b 1
)

echo.
echo ================================
echo 빌드 완료!
echo 실행 파일: build\bin\DINKIssTyle-AI-BBS.exe
echo ================================
pause
