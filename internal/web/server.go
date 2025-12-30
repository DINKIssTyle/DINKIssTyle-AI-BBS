// Created by DINKIssTyle on 2025. Copyright (C) 2025 DINKI'ssTyle. All rights reserved.

package web

import (
	"aibbs/internal/database"
	"aibbs/internal/models"
	"aibbs/internal/services"
	"bytes"
	"context"
	"database/sql"
	"fmt"
	"html/template"
	"io"
	"log"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"sync"
	"time"

	"golang.org/x/text/encoding/korean"
	"golang.org/x/text/transform"
)

// WebServer 웹 서버 구조체
type WebServer struct {
	db             *database.Database
	userService    *services.UserService
	postService    *services.PostService
	commentService *services.CommentService

	server           *http.Server
	templates        *template.Template
	port             string
	registrationOpen bool
	running          bool

	bbsConfig    models.BBSConfig // 게시판 설정
	themeManager *ThemeManager    // 테마 및 기기 관리
	mu           sync.RWMutex

	// Log Broadcasting (SSE)
	logClients   map[chan string]bool
	logClientsMu sync.Mutex
}

// NewWebServer 새 웹 서버 생성
func NewWebServer(db *database.Database, userService *services.UserService, postService *services.PostService, commentService *services.CommentService) *WebServer {
	fmt.Println("[DEBUG] NewWebServer called")
	ws := &WebServer{
		db:               db,
		userService:      userService,
		postService:      postService,
		commentService:   commentService,
		port:             "8080",
		registrationOpen: true,
		// 기본 설정
		bbsConfig: models.BBSConfig{
			Title:  "DINKI'ssTyle AI BBS",
			Footer: "(C) 2025 DINKI'ssTyle",
			Theme:  "blue",
			Font:   "sans",
		},
		logClients: make(map[chan string]bool),
	}

	// 테마 매니저 초기화
	ws.themeManager = NewThemeManager(ws.bbsConfig)

	// 템플릿 로드
	ws.loadTemplates()

	return ws
}

// loadTemplates 템플릿 로드 (문자열에서)
func (ws *WebServer) loadTemplates() {
	fmt.Println("[DEBUG] loadTemplates called")
	funcMap := template.FuncMap{
		"add": func(a, b int) int { return a + b },
		"sub": func(a, b int) int { return a - b },
		"lt":  func(a, b int) bool { return a < b },
		"gt":  func(a, b int) bool { return a > b },
		"iterate": func(start, end int) []int {
			var result []int
			for i := start; i <= end; i++ {
				result = append(result, i)
			}
			return result
		},
		"nl2br": func(text string) template.HTML {
			return template.HTML(strings.ReplaceAll(text, "\n", "<br>"))
		},
		"multiply": func(a interface{}, b float64) float64 {
			switch v := a.(type) {
			case int:
				return float64(v) * b
			case int64:
				return float64(v) * b
			case float64:
				return v * b
			default:
				return 0
			}
		},
		"formatDate": func(t time.Time) string {
			loc, _ := time.LoadLocation(ws.bbsConfig.Timezone)
			if loc == nil {
				loc = time.Local
			}
			return t.In(loc).Format("2006-01-02 15:04")
		},
		"till": func(from, to int) []int {
			res := make([]int, 0, to-from+1)
			for i := from; i <= to; i++ {
				res = append(res, i)
			}
			return res
		},
		"formatDateList": func(t time.Time) template.HTML {
			loc, _ := time.LoadLocation(ws.bbsConfig.Timezone)
			if loc == nil {
				loc = time.Local
			}
			now := time.Now().In(loc)
			target := t.In(loc)

			// 오늘인 경우 시간만 표시
			if now.Year() == target.Year() && now.Month() == target.Month() && now.Day() == target.Day() {
				return template.HTML(target.Format("15:04"))
			}
			// 오늘이 아닌 경우 날짜와 시간을 분리하여 표시 (CSS로 줄바꿈 제어 가능)
			return template.HTML(target.Format("2006-01-02") + "<span class=\"date-br\"> </span>" + target.Format("15:04"))
		},
	}

	tmpl := template.New("").Funcs(funcMap)
	// ... (rest of template parsing) ...
	// Classic Templates
	tmpl = template.Must(tmpl.New("classic/board.html").Parse(boardTemplateClassic))
	tmpl = template.Must(tmpl.New("classic/post.html").Parse(postTemplateClassic))
	tmpl = template.Must(tmpl.New("classic/write.html").Parse(writeTemplateClassic))
	tmpl = template.Must(tmpl.New("classic/login.html").Parse(loginTemplateClassic))
	tmpl = template.Must(tmpl.New("classic/register.html").Parse(registerTemplateClassic))
	tmpl = template.Must(tmpl.New("classic/edit.html").Parse(editTemplateClassic))
	tmpl = template.Must(tmpl.New("classic/comment_edit.html").Parse(commentEditTemplateClassic))

	// Unified Responsive Templates (데스크톱/모바일 통합)
	tmpl = template.Must(tmpl.New("unified/board.html").Parse(boardTemplateUnified))
	tmpl = template.Must(tmpl.New("unified/post.html").Parse(postTemplateUnified))
	tmpl = template.Must(tmpl.New("unified/write.html").Parse(writeTemplateUnified))
	tmpl = template.Must(tmpl.New("unified/login.html").Parse(loginTemplateUnified))
	tmpl = template.Must(tmpl.New("unified/register.html").Parse(registerTemplateUnified))
	tmpl = template.Must(tmpl.New("unified/edit.html").Parse(writeTemplateUnified))
	tmpl = template.Must(tmpl.New("unified/profile.html").Parse(profileTemplateUnified))
	tmpl = template.Must(tmpl.New("unified/user_comments.html").Parse(userCommentsTemplateUnified))

	// Fallback for classic (can use unified content for now if classic not specifically needed)
	tmpl = template.Must(tmpl.New("classic/profile.html").Parse(profileTemplateUnified))
	tmpl = template.Must(tmpl.New("classic/user_comments.html").Parse(userCommentsTemplateUnified))
	fmt.Println("[DEBUG] Templates parsed successfully")

	ws.templates = tmpl
}

// renderTemplate 템플릿 렌더링 + 기기별 인코딩 및 템플릿 선택
func (ws *WebServer) renderTemplate(w http.ResponseWriter, tmplName string, data interface{}) {
	// 데이터에서 디바이스 정보 추출
	m, ok := data.(map[string]interface{})
	if !ok {
		log.Printf("[ERROR] Invalid data type for template")
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	device, _ := m["Device"].(DeviceType)
	if device == "" {
		device = DeviceOld
	}

	// 템플릿 이름 결정 (device/tmplName)
	fullTmplName := string(device) + "/" + tmplName
	if ws.templates.Lookup(fullTmplName) == nil {
		fullTmplName = "classic/" + tmplName // 폴백
	}

	// 1. Render to UTF-8 Buffer
	var buf bytes.Buffer
	err := ws.templates.ExecuteTemplate(&buf, fullTmplName, data)
	if err != nil {
		log.Printf("[ERROR] Template execution failed (%s): %v", fullTmplName, err)
		http.Error(w, fmt.Sprintf("Template Error: %s - %v", fullTmplName, err), http.StatusInternalServerError)
		return
	}

	// 2. Device가 Old인 경우에만 EUC-KR 변환
	if device == DeviceOld {
		eucBuf, err := ws.convertToEUCKR(buf.Bytes())
		if err != nil {
			log.Printf("[WARN] EUC-KR conversion failed, fallback to UTF-8: %v", err)
			w.Header().Set("Content-Type", "text/html; charset=utf-8")
			w.Write(buf.Bytes())
			return
		}
		w.Header().Set("Content-Type", "text/html; charset=euc-kr")
		w.Header().Set("Content-Length", strconv.Itoa(len(eucBuf)))
		w.Write(eucBuf)
	} else {
		// Modern/Mobile은 UTF-8 사용
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.Header().Set("Content-Length", strconv.Itoa(buf.Len()))
		w.Header().Set("Content-Language", "ko")
		w.Write(buf.Bytes())
	}
}

// SetPort 포트 설정
func (ws *WebServer) SetPort(port string) {
	ws.mu.Lock()
	defer ws.mu.Unlock()
	ws.port = port
}

// convertToEUCKR UTF-8 바이트배열을 EUC-KR로 변환
func (ws *WebServer) convertToEUCKR(utf8Bytes []byte) ([]byte, error) {
	reader := transform.NewReader(bytes.NewReader(utf8Bytes), korean.EUCKR.NewEncoder())
	return io.ReadAll(reader)
}

// SetBBSConfig 게시판 설정 업데이트
func (ws *WebServer) SetBBSConfig(config models.BBSConfig) {
	ws.mu.Lock()
	ws.bbsConfig = config
	ws.themeManager.SetConfig(config)
	ws.mu.Unlock()

	// 템플릿 재로드 (날짜 포맷 등 설정 반영)
	ws.loadTemplates()
}

// GetBBSConfig 게시판 설정 조회
func (ws *WebServer) GetBBSConfig() models.BBSConfig {
	ws.mu.RLock()
	defer ws.mu.RUnlock()
	return ws.bbsConfig
}

// SetRegistrationOpen 회원가입 허용 설정
func (ws *WebServer) SetRegistrationOpen(open bool) {
	ws.mu.Lock()
	defer ws.mu.Unlock()
	ws.registrationOpen = open
}

// GetPort 포트 조회
func (ws *WebServer) GetPort() string {
	ws.mu.RLock()
	defer ws.mu.RUnlock()
	return ws.port
}

// IsRegistrationOpen 회원가입 허용 여부
func (ws *WebServer) IsRegistrationOpen() bool {
	ws.mu.RLock()
	defer ws.mu.RUnlock()
	return ws.registrationOpen
}

// IsRunning 실행 중인지 확인
func (ws *WebServer) IsRunning() bool {
	ws.mu.RLock()
	defer ws.mu.RUnlock()
	return ws.running
}

// GetWebServerConfig 현재 설정 조회 (API용)
func (ws *WebServer) GetWebServerConfig() (string, bool, bool) {
	ws.mu.RLock()
	defer ws.mu.RUnlock()
	return ws.port, ws.registrationOpen, ws.running
}

// Start 웹 서버 시작
func (ws *WebServer) Start() error {
	ws.mu.Lock()
	if ws.running {
		ws.mu.Unlock()
		return nil
	}

	mux := http.NewServeMux()

	// 라우트 등록
	mux.HandleFunc("/", ws.handleBoard)
	mux.HandleFunc("/post/", ws.handlePost)
	mux.HandleFunc("/post/edit/", ws.handlePostEdit)
	mux.HandleFunc("/post/delete/", ws.handlePostDelete)
	mux.HandleFunc("/comment/edit/", ws.handleCommentEdit) // Added route
	mux.HandleFunc("/comment/delete/", ws.handleCommentDelete)
	mux.HandleFunc("/write", ws.handleWrite)
	mux.HandleFunc("/post/recommend/", ws.handleRecommend)
	mux.HandleFunc("/login", ws.handleLogin)
	mux.HandleFunc("/logout", ws.handleLogout)
	mux.HandleFunc("/register", ws.handleRegister)
	mux.HandleFunc("/profile/", ws.handleUserProfile)
	mux.HandleFunc("/comments/user/", ws.handleUserComments)
	mux.HandleFunc("/events/logs", ws.handleLogStream) // Log Stream Route

	ws.server = &http.Server{
		Addr:    ":" + ws.port,
		Handler: mux,
	}

	ws.running = true
	ws.mu.Unlock()

	go func() {
		protocol := "http"
		if ws.bbsConfig.SSLEnabled {
			protocol = "https"
		}
		log.Printf("웹 서버 시작: %s://localhost:%s\n", protocol, ws.port)

		var err error
		if ws.bbsConfig.SSLEnabled && ws.bbsConfig.SSLCertPath != "" && ws.bbsConfig.SSLKeyPath != "" {
			err = ws.server.ListenAndServeTLS(ws.bbsConfig.SSLCertPath, ws.bbsConfig.SSLKeyPath)
		} else {
			err = ws.server.ListenAndServe()
		}

		if err != http.ErrServerClosed {
			log.Printf("웹 서버 오류: %v\n", err)
		}
	}()

	return nil
}

// Stop 웹 서버 정지
func (ws *WebServer) Stop() {
	ws.mu.Lock()
	defer ws.mu.Unlock()

	if !ws.running || ws.server == nil {
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	ws.server.Shutdown(ctx)
	ws.running = false
	log.Println("웹 서버 정지됨")
}

// 세션 관리
func (ws *WebServer) getSessionUser(r *http.Request) *models.User {
	cookie, err := r.Cookie("user_id")
	if err != nil {
		return nil
	}

	userID, err := strconv.Atoi(cookie.Value)
	if err != nil {
		return nil
	}

	user, _ := ws.userService.GetUserByID(userID)
	return user
}

func (ws *WebServer) setSessionUser(w http.ResponseWriter, userID int) {
	http.SetCookie(w, &http.Cookie{
		Name:     "user_id",
		Value:    strconv.Itoa(userID),
		Path:     "/",
		MaxAge:   86400 * 7, // 7일
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	})
}

func (ws *WebServer) clearSession(w http.ResponseWriter) {
	http.SetCookie(w, &http.Cookie{
		Name:   "user_id",
		Value:  "",
		Path:   "/",
		MaxAge: -1,
	})
}

// getThemeColors -> theme_manager.go로 이동됨 (삭제됨)

// 공통 데이터 생성 헬퍼
func (ws *WebServer) getCommonData(r *http.Request) map[string]interface{} {
	ws.mu.RLock()
	config := ws.bbsConfig
	regOpen := ws.registrationOpen
	ws.mu.RUnlock()

	user := ws.getSessionUser(r)
	device := ws.themeManager.DetectDevice(r)
	colors := ws.themeManager.GetColors(config.Theme)
	fontFace := ws.themeManager.GetFontFace()

	return map[string]interface{}{
		"Config":           config,
		"Device":           device,
		"ResultColors":     colors,
		"FontFace":         fontFace,
		"User":             user,
		"RegistrationOpen": regOpen,
		"PageTitle":        config.Title,
	}
}

// handleBoard 게시판 목록
func (ws *WebServer) handleBoard(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}

	data := ws.getCommonData(r)

	page := 1
	if p := r.URL.Query().Get("page"); p != "" {
		if n, err := strconv.Atoi(p); err == nil && n > 0 {
			page = n
		}
	}

	// 검색 파라미터 처리
	searchType := r.URL.Query().Get("type")
	keyword := r.URL.Query().Get("q")

	perPage := ws.bbsConfig.PostsPerPage
	if perPage <= 0 {
		perPage = 20
	}
	posts, err := ws.postService.GetPosts(page, perPage, searchType, keyword)
	if err != nil {
		http.Error(w, "게시물 로딩 실패", http.StatusInternalServerError)
		return
	}

	data["Posts"] = posts.Posts
	data["Page"] = page
	data["TotalPages"] = posts.TotalPages
	data["SearchType"] = searchType
	data["Keyword"] = keyword

	ws.renderTemplate(w, "board.html", data)
}

// handlePost 게시물 상세
func (ws *WebServer) handlePost(w http.ResponseWriter, r *http.Request) {
	idStr := r.URL.Path[len("/post/"):]
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.NotFound(w, r)
		return
	}

	post, err := ws.postService.GetPost(id)
	if err != nil {
		http.NotFound(w, r)
		return
	}

	// 조회수 증가 (쿠키 기반 중복 방지)
	viewedPosts, err := r.Cookie("viewed_posts")
	isViewed := false
	cookieVal := ""
	if err == nil {
		cookieVal = viewedPosts.Value
		ids := strings.Split(cookieVal, ",")
		for _, vId := range ids {
			if vId == strconv.Itoa(id) {
				isViewed = true
				break
			}
		}
	}

	if !isViewed {
		ws.postService.IncrementViewCount(id)
		newCookieVal := strconv.Itoa(id)
		if cookieVal != "" {
			newCookieVal = cookieVal + "," + newCookieVal
		}
		http.SetCookie(w, &http.Cookie{
			Name:     "viewed_posts",
			Value:    newCookieVal,
			Path:     "/",
			Expires:  time.Now().Add(24 * time.Hour),
			HttpOnly: true,
		})
	}

	data := ws.getCommonData(r)
	data["Post"] = post
	// 상세 페이지 타이틀: "글제목 - 게시판이름"
	config := data["Config"].(models.BBSConfig)
	data["PageTitle"] = fmt.Sprintf("%s - %s", post.Title, config.Title)

	comments, _ := ws.commentService.GetComments(id)
	data["Comments"] = comments

	// 작성자 확인 및 관리자 여부
	user := data["User"].(*models.User)
	isAuthor := false
	isAdmin := false
	if user != nil {
		isAdmin = user.IsAdmin
		if post.AuthorType == "user" && post.AuthorID == user.ID {
			isAuthor = true
		}
	}
	data["IsAuthor"] = isAuthor
	data["IsAdmin"] = isAdmin
	data["UserID"] = 0
	if user != nil {
		data["UserID"] = user.ID
	}

	// 댓글 작성 처리
	if r.Method == "POST" && user != nil {
		content := r.FormValue("content")
		if content != "" {
			ws.commentService.CreateCommentByAuthor("user", user.ID, id, content, nil)
			http.Redirect(w, r, fmt.Sprintf("/post/%d", id), http.StatusSeeOther)
			return
		}
	}

	ws.renderTemplate(w, "post.html", data)
}

// handleWrite 글쓰기
func (ws *WebServer) handleWrite(w http.ResponseWriter, r *http.Request) {
	data := ws.getCommonData(r)
	user := data["User"].(*models.User)

	if user == nil {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	data["PageTitle"] = "글쓰기"

	var errMsg string

	if r.Method == "POST" {
		title := r.FormValue("title")
		content := r.FormValue("content")
		isPinned := r.FormValue("is_pinned") == "1"

		// 관리자가 아니면 공지 설정 무시
		if isPinned && !user.IsAdmin {
			isPinned = false
		}

		if title != "" && content != "" {
			post, err := ws.postService.CreatePostByAuthor("user", user.ID, title, content, isPinned)
			if err == nil {
				http.Redirect(w, r, fmt.Sprintf("/post/%d", post.ID), http.StatusFound) // 302
				return
			}
			errMsg = "게시물 작성 실패: " + err.Error()
		} else {
			errMsg = "제목과 내용을 모두 입력해주세요."
		}
	}

	data["Error"] = errMsg
	ws.renderTemplate(w, "write.html", data)
}

// handlePostEdit 게시물 수정
func (ws *WebServer) handlePostEdit(w http.ResponseWriter, r *http.Request) {
	idStr := r.URL.Path[len("/post/edit/"):]
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.NotFound(w, r)
		return
	}

	data := ws.getCommonData(r)
	user := data["User"].(*models.User)

	if user == nil {
		http.Redirect(w, r, "/login", http.StatusFound) // 302
		return
	}

	post, err := ws.postService.GetPost(id)
	if err != nil {
		http.NotFound(w, r)
		return
	}

	// 작성자 본인 확인
	if post.AuthorType != "user" || post.AuthorID != user.ID {
		http.Error(w, "권한이 없습니다.", http.StatusForbidden)
		return
	}

	if r.Method == "POST" {
		title := r.FormValue("title")
		content := r.FormValue("content")
		isPinned := r.FormValue("is_pinned") == "1"

		if title != "" && content != "" {
			// 권한 확인 (본인 또는 관리자)
			if post.AuthorType == "user" && post.AuthorID == user.ID {
				// 관리자가 아니면 공지 설정 권한 없음
				if isPinned && !user.IsAdmin {
					isPinned = post.IsPinned // 기존 값 유지
				}
				err = ws.postService.UpdatePostForWeb(id, title, content, isPinned)
			} else if user.IsAdmin {
				err = ws.postService.UpdatePostForWeb(id, title, content, isPinned)
			} else {
				http.Error(w, "권한이 없습니다.", http.StatusForbidden)
				return
			}

			if err == nil {
				http.Redirect(w, r, fmt.Sprintf("/post/%d", id), http.StatusFound) // 302
				return
			}
			http.Error(w, "게시물 수정 실패: "+err.Error(), http.StatusInternalServerError)
			return
		}
	}

	data["Post"] = post
	data["PageTitle"] = "게시물 수정"
	log.Printf("[DEBUG] handlePostEdit: Rendering edit.html for Post ID %d\n", id)
	ws.renderTemplate(w, "edit.html", data)
}

// handlePostDelete 게시물 삭제
func (ws *WebServer) handlePostDelete(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "잘못된 요청입니다.", http.StatusMethodNotAllowed)
		return
	}

	idStr := r.URL.Path[len("/post/delete/"):]
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "잘못된 게시물 ID", http.StatusBadRequest)
		return
	}

	user := ws.getSessionUser(r)
	if user == nil {
		log.Println("[DEBUG] handlePostDelete: User session missing or invalid")
		http.Redirect(w, r, "/login", http.StatusFound) // 302
		return
	}

	post, err := ws.postService.GetPost(id)
	if err != nil {
		log.Printf("[DEBUG] handlePostDelete: Post not found ID=%d\n", id)
		http.Error(w, "게시물을 찾을 수 없습니다", http.StatusNotFound)
		return
	}

	// 작성자 본인 확인
	log.Printf("[DEBUG] handlePostDelete: UserID=%d, PostAuthorID=%d, AuthorType=%s\n", user.ID, post.AuthorID, post.AuthorType)
	if post.AuthorType != "user" || post.AuthorID != user.ID {
		log.Println("[DEBUG] handlePostDelete: Permission denied")
		http.Error(w, "권한이 없습니다.", http.StatusForbidden)
		return
	}

	err = ws.postService.DeletePostForWeb(id)
	if err != nil {
		log.Printf("[DEBUG] handlePostDelete: DeletePost error: %v\n", err)
		http.Error(w, "삭제 실패: "+err.Error(), http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/", http.StatusFound) // 302
}

// handleCommentEdit 댓글 수정
func (ws *WebServer) handleCommentEdit(w http.ResponseWriter, r *http.Request) {
	idStr := r.URL.Path[len("/comment/edit/"):]
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.NotFound(w, r)
		return
	}

	data := ws.getCommonData(r)
	user := data["User"].(*models.User)

	if user == nil {
		http.Redirect(w, r, "/login", http.StatusFound)
		return
	}

	// 댓글 정보 조회 (권한 확인용, DB 직접 조회)
	comment := new(models.Comment)
	db := ws.db.GetDB()
	err = db.QueryRow("SELECT id, author_type, author_id, content, post_id FROM comments WHERE id = ?", id).Scan(&comment.ID, &comment.AuthorType, &comment.AuthorID, &comment.Content, &comment.PostID)
	if err != nil {
		http.NotFound(w, r)
		return
	}

	if comment.AuthorType != "user" || comment.AuthorID != user.ID {
		http.Error(w, "권한이 없습니다.", http.StatusForbidden)
		return
	}

	if r.Method == "POST" {
		content := r.FormValue("content")
		if content != "" {
			err := ws.commentService.UpdateCommentForWeb(id, content)
			if err == nil {
				http.Redirect(w, r, fmt.Sprintf("/post/%d", comment.PostID), http.StatusFound)
				return
			}
			http.Error(w, "댓글 수정 실패: "+err.Error(), http.StatusInternalServerError)
			return
		}
	}

	data["Comment"] = comment
	data["PageTitle"] = "댓글 수정"
	ws.renderTemplate(w, "comment_edit.html", data)
}

// handleCommentDelete 댓글 삭제
func (ws *WebServer) handleCommentDelete(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "잘못된 요청입니다.", http.StatusMethodNotAllowed)
		return
	}

	idStr := r.URL.Path[len("/comment/delete/"):]
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "잘못된 댓글 ID", http.StatusBadRequest)
		return
	}

	user := ws.getSessionUser(r)
	if user == nil {
		http.Redirect(w, r, "/login", http.StatusFound) // 302
		return
	}

	// 댓글 조회 로직이 필요함 (Service에 GetCommentById 없으면 추가하거나, 직접 DB 조회)
	// 여기선 편의상 commentService를 통해 가져온다고 가정.
	// 하지만 현재 commentService에는 GetCommentByID가 없음.
	// 임시로 PostID 리다이렉트를 위해 form value 받음.
	postIDStr := r.FormValue("post_id")

	// 본인 확인 로직 필요: Service에 DeleteCommentV2 같은걸 만들어서 user_id 체크를 넣거나
	// 아니면 여기서 DB 조회를 해야함.
	// 간단히 구현하기 위해: CommentService에 AuthorID 확인 후 삭제하는 로직이 없으니,
	// 여기서 DB 조회를 추가하거나, Service에 'DeleteCommentByOwner' 메서드를 추가하는게 정석.
	// 우선은 간단히 CommentService에 의존하되, 보안이 약간 취약할 수 있음 (ID만 알면 삭제 시도 가능) -> 절대 안됨.
	// 해결책: Database 객체 접근하여 직접 확인.

	comment := new(models.Comment)
	// ws.db 접근 필요. 하지만 ws.db는 *database.Database 타입.
	// 직접 쿼리:
	db := ws.db.GetDB()
	err = db.QueryRow("SELECT id, author_type, author_id, post_id FROM comments WHERE id = ?", id).Scan(&comment.ID, &comment.AuthorType, &comment.AuthorID, &comment.PostID)
	if err != nil {
		http.Error(w, "댓글을 찾을 수 없습니다", http.StatusNotFound)
		return
	}

	if comment.AuthorType != "user" || comment.AuthorID != user.ID {
		http.Error(w, "권한이 없습니다.", http.StatusForbidden)
		return
	}

	err = ws.commentService.DeleteCommentForWeb(id)
	if err != nil {
		http.Error(w, "삭제 실패: "+err.Error(), http.StatusInternalServerError)
		return
	}

	redirectURL := fmt.Sprintf("/post/%d", comment.PostID)
	// 폼에서 받은 post_id가 있으면 우선 사용 (DB조회 했으니 comment.PostID가 정확함)
	if postIDStr != "" {
		// 이미 DB에서 가져왔으니 무시해도 됨
	}

	http.Redirect(w, r, redirectURL, http.StatusFound) // 302
}

// handleLogin 로그인
func (ws *WebServer) handleLogin(w http.ResponseWriter, r *http.Request) {
	data := ws.getCommonData(r)
	data["PageTitle"] = "로그인"

	var errMsg string

	if r.Method == "POST" {
		username := r.FormValue("username")
		password := r.FormValue("password")

		user, err := ws.userService.LoginForWeb(username, password)
		if err == nil {
			ws.setSessionUser(w, user.ID)
			http.Redirect(w, r, "/", http.StatusFound) // 302
			return
		}
		errMsg = "로그인 실패: 아이디 또는 비밀번호가 올바르지 않습니다."
	}

	data["Error"] = errMsg
	ws.renderTemplate(w, "login.html", data)
}

// handleLogout 로그아웃
func (ws *WebServer) handleLogout(w http.ResponseWriter, r *http.Request) {
	ws.clearSession(w)
	http.Redirect(w, r, "/", http.StatusFound) // 302
}

// handleRegister 회원가입
func (ws *WebServer) handleRegister(w http.ResponseWriter, r *http.Request) {
	data := ws.getCommonData(r)
	data["PageTitle"] = "회원가입"

	regOpen := data["RegistrationOpen"].(bool)
	if !regOpen {
		http.Error(w, "회원가입이 비활성화되어 있습니다.", http.StatusForbidden)
		return
	}

	var errMsg string

	if r.Method == "POST" {
		username := r.FormValue("username")
		password := r.FormValue("password")
		nickname := r.FormValue("nickname")

		if username != "" && password != "" && nickname != "" {
			err := ws.userService.CreateUser(username, password, nickname)
			if err == nil {
				http.Redirect(w, r, "/login", http.StatusSeeOther)
				return
			}
			errMsg = err.Error()
		}
	}

	data["Error"] = errMsg
	ws.renderTemplate(w, "register.html", data)
}

// handleRecommend 게시물 추천
func (ws *WebServer) handleRecommend(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}

	// URL에서 게시물 ID 추출
	path := r.URL.Path
	idStr := strings.TrimPrefix(path, "/post/recommend/")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "잘못된 게시물 ID", http.StatusBadRequest)
		return
	}

	// 추천 처리
	data := ws.getCommonData(r)
	user, _ := data["User"].(*models.User)
	if user == nil {
		log.Printf("[ERROR] 추천 실패: 로그인이 필요합니다")
		http.Redirect(w, r, "/login", http.StatusFound)
		return
	}

	err = ws.postService.RecommendPostByUser("user", user.ID, id)
	if err != nil {
		log.Printf("[ERROR] 추천 실패: %v", err)
	}

	// 원래 게시물로 리다이렉트
	http.Redirect(w, r, fmt.Sprintf("/post/%d", id), http.StatusFound)
}

// handleUserProfile 사용자 프로필 및 캐릭터 정보
func (ws *WebServer) handleUserProfile(w http.ResponseWriter, r *http.Request) {
	rawNickname := r.URL.Path[len("/profile/"):]
	nickname, _ := url.PathUnescape(rawNickname)
	log.Printf("[DEBUG] handleUserProfile: raw=%s, unescaped=%s", rawNickname, nickname)
	if nickname == "" {
		http.Redirect(w, r, "/", http.StatusFound)
		return
	}

	// 캐릭터 정보 조회
	db := ws.db.GetDB()
	c := &models.AICharacter{}
	var personaUpdatedAt sql.NullTime
	err := db.QueryRow(`
		SELECT id, nickname, gender, age, COALESCE(birthdate, ''), COALESCE(region, ''),
		       hobby, job_category, mbti, aggression_level, formality_level, roleplay_level,
		       COALESCE(persona_summary, ''), persona_updated_at, created_at
		FROM ai_characters WHERE nickname = ?
	`, nickname).Scan(&c.ID, &c.Nickname, &c.Gender, &c.Age, &c.Birthdate, &c.Region,
		&c.Hobby, &c.JobCategory, &c.MBTI, &c.AggressionLevel, &c.FormalityLevel, &c.RoleplayLevel,
		&c.PersonaSummary, &personaUpdatedAt, &c.CreatedAt)

	if err != nil {
		log.Printf("[DEBUG] handleUserProfile: AI character query failed for '%s': %v", nickname, err)
		// 캐릭터가 없으면 사용자인지 확인 (사용자 프로필은 아직 간단히 처리)
		user, userErr := ws.userService.FindUserByNickname(nickname)
		if userErr != nil {
			log.Printf("[DEBUG] handleUserProfile: User query also failed for '%s': %v", nickname, userErr)
		}
		if user == nil {
			http.NotFound(w, r)
			return
		}
		// 사용자용 캐릭터 객체 생성 (일부 필드 비움)
		c = &models.AICharacter{
			ID:        user.ID,
			Nickname:  user.Nickname,
			CreatedAt: user.CreatedAt,
		}
	} else {
		// NULL-safe 처리: NullTime에서 실제 Time으로 변환
		if personaUpdatedAt.Valid {
			c.PersonaUpdatedAt = personaUpdatedAt.Time
		}
	}

	data := ws.getCommonData(r)
	data["Character"] = c
	data["PageTitle"] = fmt.Sprintf("%s 님의 프로필", c.Nickname)

	ws.renderTemplate(w, "profile.html", data)
}

// UserCommentInfo 댓글 목록용 확장 구조체
type UserCommentInfo struct {
	ID        int
	PostID    int
	PostTitle string
	Content   string
	CreatedAt time.Time
}

// handleUserComments 사용자가 작성한 댓글 목록
func (ws *WebServer) handleUserComments(w http.ResponseWriter, r *http.Request) {
	rawNickname := r.URL.Path[len("/comments/user/"):]
	nickname, _ := url.PathUnescape(rawNickname)
	log.Printf("[DEBUG] handleUserComments: raw=%s, unescaped=%s", rawNickname, nickname)
	if nickname == "" {
		http.Redirect(w, r, "/", http.StatusFound)
		return
	}

	page := 1
	if p := r.URL.Query().Get("page"); p != "" {
		if n, err := strconv.Atoi(p); err == nil && n > 0 {
			page = n
		}
	}
	perPage := 20

	db := ws.db.GetDB()

	// 전체 댓글 수 확인
	var totalCount int
	err := db.QueryRow(`
		SELECT COUNT(*) FROM comments c
		LEFT JOIN ai_characters a ON c.author_type = 'ai' AND c.author_id = a.id
		LEFT JOIN users u ON c.author_type = 'user' AND c.author_id = u.id
		WHERE (c.author_type = 'ai' AND a.nickname = ?) OR (c.author_type = 'user' AND u.nickname = ?)
	`, nickname, nickname).Scan(&totalCount)

	if err != nil {
		http.Error(w, "댓글 조회 중 오류가 발생했습니다.", http.StatusInternalServerError)
		return
	}

	totalPages := (totalCount + perPage - 1) / perPage
	offset := (page - 1) * perPage

	// 댓글 목록 조회 (게시물 제목 포함)
	rows, err := db.Query(`
		SELECT c.id, c.post_id, p.title, c.content, c.created_at
		FROM comments c
		JOIN posts p ON c.post_id = p.id
		LEFT JOIN ai_characters a ON c.author_type = 'ai' AND c.author_id = a.id
		LEFT JOIN users u ON c.author_type = 'user' AND c.author_id = u.id
		WHERE (c.author_type = 'ai' AND a.nickname = ?) OR (c.author_type = 'user' AND u.nickname = ?)
		ORDER BY c.created_at DESC
		LIMIT ? OFFSET ?
	`, nickname, nickname, perPage, offset)

	if err != nil {
		http.Error(w, "댓글 로딩 실패", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var comments []UserCommentInfo
	for rows.Next() {
		var ci UserCommentInfo
		if err := rows.Scan(&ci.ID, &ci.PostID, &ci.PostTitle, &ci.Content, &ci.CreatedAt); err == nil {
			comments = append(comments, ci)
		}
	}

	data := ws.getCommonData(r)
	data["Nickname"] = nickname
	data["Comments"] = comments
	data["CurrentPage"] = page
	data["TotalPages"] = totalPages
	data["TotalCount"] = totalCount
	data["PageTitle"] = fmt.Sprintf("%s 님의 작성 댓글", nickname)

	ws.renderTemplate(w, "user_comments.html", data)
}

// ========================================
// Log Broadcasting (SSE)
// ========================================

// BroadcastLog 연결된 모든 클라이언트에게 로그 브로드캐스트
func (ws *WebServer) BroadcastLog(message string) {
	ws.logClientsMu.Lock()
	defer ws.logClientsMu.Unlock()

	for client := range ws.logClients {
		select {
		case client <- message:
		default:
			// 채널이 꽉 찼거나 클라이언트가 느린 경우 스킵 (블로킹 방지)
		}
	}
}

// handleLogStream SSE를 이용한 로그 스트리밍 핸들러
func (ws *WebServer) handleLogStream(w http.ResponseWriter, r *http.Request) {
	// SSE 헤더 설정
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	// 클라이언트 채널 생성
	clientChan := make(chan string, 100)

	// 클라이언트 등록
	ws.logClientsMu.Lock()
	ws.logClients[clientChan] = true
	ws.logClientsMu.Unlock()

	// 클라이언트 연결 종료 시 제거
	defer func() {
		ws.logClientsMu.Lock()
		delete(ws.logClients, clientChan)
		ws.logClientsMu.Unlock()
		close(clientChan)
	}()

	// 연결 유지 (Flusher)
	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "Streaming unsupported", http.StatusInternalServerError)
		return
	}

	// 초기 메시지 전송
	fmt.Fprintf(w, "data: Connected to log stream\n\n")
	flusher.Flush()

	// 로그 전송 루프
	for {
		select {
		case msg, open := <-clientChan:
			if !open {
				return
			}
			// SSE 포맷으로 전송 (data: message\n\n)
			// 여러 줄일 경우 각 줄마다 data: prefix를 붙이거나 JSON으로 보낼 수 있음.
			// 여기서는 단순히 줄바꿈을 제거하고 한 줄로 전송하거나 그대로 전송.
			// JS EventSource는 \n\n을 메시지 구분자로 사용하므로, msg 내부의 \n은 처리 필요.
			lines := strings.Split(msg, "\n")
			for _, line := range lines {
				if strings.TrimSpace(line) == "" {
					continue
				}
				fmt.Fprintf(w, "data: %s\n\n", line)
			}
			flusher.Flush()
		case <-r.Context().Done():
			return
		}
	}
}
