// Created by DINKIssTyle on 2025. Copyright (C) 2025 DINKI'ssTyle. All rights reserved.

package ai

import (
	"aibbs/internal/models"
	"aibbs/internal/services"
	"fmt"
	"log"
	"math"
	"math/rand"
	"sort"
	"strings"
	"sync"
	"time"
)

// ActivityManager AI 자동 활동 관리자
type ActivityManager struct {
	characterService *services.CharacterService
	postService      *services.PostService
	commentService   *services.CommentService
	llmService       *services.LLMService

	running  bool
	stopChan chan struct{}
	mu       sync.RWMutex

	postsPerHour    int
	commentsPerHour int

	onNewPost    func()
	onNewComment func()
}

// NewActivityManager 새 활동 관리자 생성
func NewActivityManager(
	characterService *services.CharacterService,
	postService *services.PostService,
	commentService *services.CommentService,
	llmService *services.LLMService,
) *ActivityManager {
	return &ActivityManager{
		characterService: characterService,
		postService:      postService,
		commentService:   commentService,
		llmService:       llmService,
		postsPerHour:     5,
		commentsPerHour:  10,
		stopChan:         make(chan struct{}),
	}
}

// SetActivityRate 시간당 활동 수 설정
func (m *ActivityManager) SetActivityRate(postsPerHour, commentsPerHour int) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.postsPerHour = postsPerHour
	m.commentsPerHour = commentsPerHour
}

// SetCallbacks 콜백 설정 (새 글/댓글 알림)
func (m *ActivityManager) SetCallbacks(onNewPost, onNewComment func()) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.onNewPost = onNewPost
	m.onNewComment = onNewComment
}

// Start AI 활동 시작
func (m *ActivityManager) Start() error {
	m.mu.Lock()
	if m.running {
		m.mu.Unlock()
		return nil
	}
	m.running = true
	m.stopChan = make(chan struct{})
	m.mu.Unlock()

	// 게시물 작성 고루틴
	go m.postActivityLoop()

	// 댓글 작성 고루틴
	go m.commentActivityLoop()

	// 게시물 브라우징 (조회수 증가) 고루틴
	go m.browsingActivityLoop()

	// 본인 글에 달린 댓글에 답글 달기 고루틴
	go m.replyToCommentsLoop()

	log.Println("AI 활동 시작됨")
	return nil
}

// Stop AI 활동 정지
func (m *ActivityManager) Stop() {
	m.mu.Lock()
	defer m.mu.Unlock()

	if !m.running {
		return
	}

	close(m.stopChan)
	m.running = false
	log.Println("AI 활동 정지됨")
}

// IsRunning 활동 중인지 확인
func (m *ActivityManager) IsRunning() bool {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.running
}

// postActivityLoop 게시물 작성 루프
func (m *ActivityManager) postActivityLoop() {
	rng := rand.New(rand.NewSource(time.Now().UnixNano()))

	// 시작 직후 첫 게시물 (1~30초 랜덤 대기)
	initialDelay := time.Duration(1+rng.Intn(30)) * time.Second
	select {
	case <-m.stopChan:
		return
	case <-time.After(initialDelay):
		m.createRandomPost()
	}

	for {
		m.mu.RLock()
		postsPerHour := m.postsPerHour
		m.mu.RUnlock()

		if postsPerHour <= 0 {
			postsPerHour = 1
		}

		// 시간당 게시물 수에 따른 간격 계산 (약간의 랜덤 추가)
		baseInterval := time.Hour / time.Duration(postsPerHour)
		jitter := time.Duration(rng.Int63n(int64(baseInterval / 2)))
		interval := baseInterval + jitter - baseInterval/4

		select {
		case <-m.stopChan:
			return
		case <-time.After(interval):
			m.createRandomPost()
		}
	}
}

// commentActivityLoop 댓글 작성 루프
func (m *ActivityManager) commentActivityLoop() {
	rng := rand.New(rand.NewSource(time.Now().UnixNano()))

	// 시작 직후 첫 댓글 (5~60초 랜덤 대기)
	initialDelay := time.Duration(5+rng.Intn(55)) * time.Second
	select {
	case <-m.stopChan:
		return
	case <-time.After(initialDelay):
		m.createRandomComment()
	}

	for {
		m.mu.RLock()
		commentsPerHour := m.commentsPerHour
		m.mu.RUnlock()

		if commentsPerHour <= 0 {
			commentsPerHour = 1
		}

		// 시간당 댓글 수에 따른 간격 계산
		baseInterval := time.Hour / time.Duration(commentsPerHour)
		jitter := time.Duration(rng.Int63n(int64(baseInterval / 2)))
		interval := baseInterval + jitter - baseInterval/4

		select {
		case <-m.stopChan:
			return
		case <-time.After(interval):
			m.createRandomComment()
		}
	}
}

// browsingActivityLoop 게시물 랜덤 브라우징 (조회수 증가) 루프
func (m *ActivityManager) browsingActivityLoop() {
	rng := rand.New(rand.NewSource(time.Now().UnixNano()))

	// 시작 직후 첫 브라우징 (5~60초 랜덤 대기)
	initialDelay := time.Duration(5+rng.Intn(55)) * time.Second
	select {
	case <-m.stopChan:
		return
	case <-time.After(initialDelay):
		m.browseRandomPost()
	}

	for {
		// 시간당 약 30~60회 브라우징 (1~2분 간격)
		baseInterval := time.Duration(60+rng.Intn(60)) * time.Second

		select {
		case <-m.stopChan:
			return
		case <-time.After(baseInterval):
			m.browseRandomPost()
		}
	}
}

// browseRandomPost 랜덤 게시물 조회 (조회수 증가) + 활동전 닉네임 변경
func (m *ActivityManager) browseRandomPost() {
	// 최근 게시물 조회
	postList, err := m.postService.GetPosts(1, 50, "", "")
	if err != nil || len(postList.Posts) == 0 {
		return
	}

	// 랜덤 게시물 선택
	rng := rand.New(rand.NewSource(time.Now().UnixNano()))
	selectedPost := postList.Posts[rng.Intn(len(postList.Posts))]

	// 조회수 증가
	err = m.postService.IncrementViewCount(selectedPost.ID)
	if err != nil {
		log.Printf("조회수 증가 실패: %v\n", err)
	}

	// 운영봇 작업: "활동전AI" 닉네임 변경
	m.updatePendingNicknames()

	// 운영봇 작업: 인격 갱신이 필요한 캐릭터 처리
	m.updatePendingPersonas()
}

// updatePendingNicknames "활동전AI" 닉네임을 가진 캐릭터를 찾아 닉네임 변경
func (m *ActivityManager) updatePendingNicknames() {
	// "활동전AI"로 시작하는 닉네임을 가진 캐릭터 1명 조회
	character, err := m.characterService.GetCharacterWithPendingNickname()
	if err != nil || character == nil {
		return // 변경 대상 없음
	}

	// 닉네임 업데이트
	m.updateNicknameIfNeeded(character)
}

// updatePendingPersonas 인격 갱신이 필요한 캐릭터를 찾아 인격 생성/갱신
func (m *ActivityManager) updatePendingPersonas() {
	// 인격 갱신이 필요한 캐릭터 1명 조회
	character, err := m.characterService.GetCharacterNeedingPersonaUpdate()
	if err != nil || character == nil {
		return // 갱신 대상 없음
	}

	// 인격 생성/갱신
	m.checkAndGeneratePersona(character)
}

// createRandomPost 랜덤 캐릭터로 게시물 작성
func (m *ActivityManager) createRandomPost() {
	// 랜덤 캐릭터 선택
	character, err := m.characterService.GetRandomActiveCharacter()
	if err != nil {
		log.Printf("랜덤 캐릭터 조회 실패: %v\n", err)
		return
	}

	// 닉네임 업데이트 체크
	m.updateNicknameIfNeeded(character)

	// 캐릭터의 최근 게시물 조회
	recentPosts, _ := m.postService.GetRecentPostsByCharacter(character.ID, 5)

	// 인기 게시물 조회
	popularPosts, _ := m.postService.GetPopularPosts(5)

	// 공지사항(고정글) 조회
	pinnedPosts, _ := m.postService.GetPinnedPosts()

	// LLM으로 게시물 내용 생성
	postContent, err := m.llmService.GeneratePostContent(character, recentPosts, popularPosts, pinnedPosts)
	if err != nil {
		log.Printf("게시물 생성 실패: %v\n", err)
		return
	}

	// 게시물 저장
	_, err = m.postService.CreatePostByAuthor("ai", character.ID, postContent.Title, postContent.Content, false)
	if err != nil {
		fmt.Printf("게시물 생성 실패: %v\n", err)
		return
	}

	log.Printf("AI 게시물 작성: [%s] %s\n", character.Nickname, postContent.Title)

	// 활동 횟수 증가
	_ = m.characterService.IncrementPostCount(character.ID)

	// 인격 생성 체크 (3회 이상 활동 시)
	m.checkAndGeneratePersona(character)

	// 콜백 호출
	m.mu.RLock()
	callback := m.onNewPost
	m.mu.RUnlock()
	if callback != nil {
		callback()
	}
}

// createRandomComment 댓글 작성 (최근 글 가중치 + 어울리는 캐릭터)
func (m *ActivityManager) createRandomComment() {
	// 최근 게시물 50개 조회 (더 넓은 풀에서 선택)
	postList, err := m.postService.GetPosts(1, 50, "", "")
	if err != nil || len(postList.Posts) == 0 {
		log.Println("게시물이 없습니다")
		return
	}

	// 시간 가중치 기반 게시물 선택 (최근 글일수록 확률 높음)
	selectedPost := m.selectPostByRecency(postList.Posts)

	// 게시물 상세 조회
	post, err := m.postService.GetPost(selectedPost.ID)
	if err != nil {
		log.Printf("게시물 조회 실패: %v\n", err)
		return
	}

	// AI가 글을 읽었으므로 조회수 증가
	_ = m.postService.IncrementViewCount(post.ID)

	// 글 제목에 가장 어울리는 캐릭터 선택
	character, err := m.selectBestCharacterForPost(post)
	if err != nil {
		log.Printf("캐릭터 선택 실패: %v\n", err)
		return
	}

	// 닉네임 업데이트 체크
	m.updateNicknameIfNeeded(character)

	// 기존 댓글 조회
	existingComments, _ := m.commentService.GetRecentCommentsByPost(post.ID, 10)

	// 공지사항(고정글) 조회
	pinnedPosts, _ := m.postService.GetPinnedPosts()

	// LLM으로 댓글 내용 생성
	commentContent, recommend, err := m.llmService.GenerateCommentContent(character, post, existingComments, pinnedPosts)
	if err != nil {
		log.Printf("댓글 생성 실패: %v\n", err)
		return
	}

	// 댓글 저장
	_, err = m.commentService.CreateCommentByAuthor("ai", character.ID, post.ID, commentContent, nil)
	if err != nil {
		log.Printf("댓글 저장 실패: %v\n", err)
		return
	}

	log.Printf("AI 댓글 작성: [%s] on [%s]\n", character.Nickname, post.Title)

	// 추천 처리
	if recommend {
		_ = m.postService.RecommendPost(post.ID)
		log.Printf("AI 추천: [%s] -> [%s]\n", character.Nickname, post.Title)
	}

	// 활동 횟수 증가
	_ = m.characterService.IncrementCommentCount(character.ID)

	// 인격 생성 체크 (3회 이상 활동 시)
	m.checkAndGeneratePersona(character)

	// 콜백 호출
	m.mu.RLock()
	callback := m.onNewComment
	m.mu.RUnlock()
	if callback != nil {
		callback()
	}
}

// selectPostByRecency 최근 글일수록 높은 확률로 선택
func (m *ActivityManager) selectPostByRecency(posts []models.Post) models.Post {
	rng := rand.New(rand.NewSource(time.Now().UnixNano()))

	if len(posts) == 0 {
		return models.Post{}
	}
	if len(posts) == 1 {
		return posts[0]
	}

	// 가중치 계산: 가장 최근 글 = 가중치 높음
	// 인덱스가 작을수록(최근) 가중치 높음
	weights := make([]float64, len(posts))
	totalWeight := 0.0

	for i := range posts {
		// 지수적 감소: 최근 글이 훨씬 높은 가중치
		// 첫 번째 글: 1.0, 두 번째: 0.8, 세 번째: 0.64 ...
		weight := math.Pow(0.9, float64(i))
		weights[i] = weight
		totalWeight += weight
	}

	// 가중치 기반 랜덤 선택
	r := rng.Float64() * totalWeight
	cumulative := 0.0

	for i, weight := range weights {
		cumulative += weight
		if r <= cumulative {
			return posts[i]
		}
	}

	return posts[0]
}

// selectBestCharacterForPost 글 제목/내용에 어울리는 캐릭터 선택
func (m *ActivityManager) selectBestCharacterForPost(post *models.Post) (*models.AICharacter, error) {
	// 활성화된 모든 캐릭터 조회
	characters, err := m.characterService.GetActiveCharacters()
	if err != nil || len(characters) == 0 {
		// 폴백: 랜덤 캐릭터
		return m.characterService.GetRandomActiveCharacter()
	}

	rng := rand.New(rand.NewSource(time.Now().UnixNano()))

	// 캐릭터 수가 적으면 랜덤 사용 (다양성 유지)
	if len(characters) <= 3 {
		return &characters[rng.Intn(len(characters))], nil
	}

	// 키워드 기반 매칭
	title := post.Title
	content := post.Content

	// 각 캐릭터에 대해 관련성 점수 계산
	type scoredChar struct {
		char  models.AICharacter
		score float64
	}
	scoredChars := make([]scoredChar, len(characters))

	for i, char := range characters {
		score := 1.0 // 기본 점수

		// 직종 관련 키워드 매칭
		if containsAny(title+content, getJobKeywords(char.JobCategory)) {
			score += 5.0
		}

		// 취미 관련 키워드 매칭
		if containsAny(title+content, getHobbyKeywords(char.Hobby)) {
			score += 3.0
		}

		// MBTI 특성 매칭
		score += getMBTIMatchScore(char.MBTI, title+content)

		// 약간의 랜덤성 추가 (완전 결정론적이지 않게)
		score += rng.Float64() * 2.0

		scoredChars[i] = scoredChar{char: char, score: score}
	}

	// 점수 기준 상위 5명 중 랜덤 선택 (다양성+관련성)
	sort.Slice(scoredChars, func(i, j int) bool {
		return scoredChars[i].score > scoredChars[j].score
	})

	topN := 5
	if len(scoredChars) < topN {
		topN = len(scoredChars)
	}

	selected := scoredChars[rng.Intn(topN)]
	return &selected.char, nil
}

// containsAny 문자열에 키워드가 포함되어 있는지 확인
func containsAny(text string, keywords []string) bool {
	textLower := strings.ToLower(text)
	for _, kw := range keywords {
		if strings.Contains(textLower, strings.ToLower(kw)) {
			return true
		}
	}
	return false
}

// getJobKeywords 직종 관련 키워드
func getJobKeywords(job string) []string {
	jobKeywords := map[string][]string{
		"IT/소프트웨어": {"프로그램", "개발", "코딩", "컴퓨터", "서버", "앱", "소프트웨어", "버그"},
		"디자인":      {"디자인", "그림", "UI", "UX", "폰트", "색", "로고"},
		"마케팅":      {"마케팅", "광고", "홍보", "브랜드", "판매"},
		"금융":       {"주식", "투자", "은행", "대출", "금리", "돈", "경제"},
		"의료":       {"병원", "의사", "건강", "약", "치료", "수술"},
		"교육":       {"학교", "공부", "수업", "선생", "학생", "시험"},
		"요리/식음료":   {"음식", "요리", "맛집", "레시피", "식당", "카페"},
		"게임":       {"게임", "플레이", "캐릭터", "랭크", "eスポーツ"},
	}
	if kws, ok := jobKeywords[job]; ok {
		return kws
	}
	return []string{job}
}

// getHobbyKeywords 취미 관련 키워드
func getHobbyKeywords(hobby string) []string {
	return []string{hobby}
}

// getMBTIMatchScore MBTI 특성에 따른 매칭 점수
func getMBTIMatchScore(mbti, text string) float64 {
	textLower := strings.ToLower(text)
	score := 0.0

	// E vs I: 사교적 주제
	if strings.ContainsAny(mbti, "E") && containsAny(textLower, []string{"파티", "모임", "같이", "함께"}) {
		score += 1.0
	}

	// T vs F: 논리 vs 감정
	if strings.ContainsAny(mbti, "T") && containsAny(textLower, []string{"분석", "논리", "이유", "왜"}) {
		score += 1.0
	}
	if strings.ContainsAny(mbti, "F") && containsAny(textLower, []string{"느낌", "감정", "마음", "사랑"}) {
		score += 1.0
	}

	return score
}

// checkAndGeneratePersona 인격 생성/갱신 체크 (활동량 기반)
func (m *ActivityManager) checkAndGeneratePersona(character *models.AICharacter) {
	// 활동 통계 확인
	postCount, commentCount, err := m.characterService.GetActivityStats(character.ID)
	if err != nil {
		return
	}

	totalActivity := postCount + commentCount

	// 인격 존재 여부 확인
	hasPersona, err := m.characterService.HasPersonaSummary(character.ID)
	if err != nil {
		return
	}

	// 인격 생성/갱신 조건:
	// 1. 인격이 없고 글+댓글 합이 3개 이상이면 최초 생성
	// 2. 인격이 있고 총 활동량이 6, 12, 24... (3의 배수+3)을 초과할 때마다 갱신
	shouldGenerate := false

	if !hasPersona && totalActivity >= 3 {
		// 최초 생성
		shouldGenerate = true
		log.Printf("인격 최초 생성 시작: [%s] (글 %d, 댓글 %d)\n", character.Nickname, postCount, commentCount)
	} else if hasPersona {
		// 갱신 조건: 마지막 갱신 이후 활동량이 3개 이상 증가했을 때
		// 간단히 6, 12, 24... 매 3회마다 갱신 (확률적으로)
		// 6, 9, 12, 15... 등 3의 배수마다 갱신 기회
		if totalActivity >= 6 && totalActivity%3 == 0 {
			shouldGenerate = true
			log.Printf("인격 갱신 시작: [%s] (글 %d, 댓글 %d, 총 %d)\n", character.Nickname, postCount, commentCount, totalActivity)
		}
	}

	if !shouldGenerate {
		return
	}

	// 최근 글/댓글 조회
	recentPosts, _ := m.postService.GetRecentPostsByCharacter(character.ID, 5)
	recentComments, _ := m.commentService.GetRecentCommentsByCharacter(character.ID, 5)

	// 인격 생성/갱신
	summary, err := m.llmService.GeneratePersonaSummary(character, recentPosts, recentComments)
	if err != nil {
		log.Printf("인격 생성 실패: %v\n", err)
		return
	}

	// 인격 저장
	err = m.characterService.UpdatePersonaSummary(character.ID, summary)
	if err != nil {
		log.Printf("인격 저장 실패: %v\n", err)
		return
	}

	log.Printf("인격 생성/갱신 완료: [%s]\n", character.Nickname)
}

// updateNicknameIfNeeded "활동전AI" 닉네임 변경
func (m *ActivityManager) updateNicknameIfNeeded(character *models.AICharacter) {
	if !strings.HasPrefix(character.Nickname, "활동전AI") {
		return
	}

	oldNickname := character.Nickname
	success := false

	// 최대 3회 재시도
	for i := 0; i < 3; i++ {
		// LLM으로 새 닉네임 생성
		newNickname, err := m.llmService.GenerateNickname(character)
		if err != nil {
			log.Printf("닉네임 생성 실패 (시도 %d/3): %v\n", i+1, err)
			continue
		}

		character.Nickname = newNickname
		err = m.characterService.UpdateCharacter(*character)
		if err == nil {
			log.Printf("닉네임 변경 완료: %s -> %s\n", oldNickname, newNickname)
			success = true
			break
		}

		// 실패 시 (중복 등) 로그 남기고 재시도
		log.Printf("닉네임 변경 실패 (중복 등, 시도 %d/3): %s -> %s (%v)\n", i+1, oldNickname, newNickname, err)

		// 닉네임 원복 후 재시도
		character.Nickname = oldNickname

		// 약간의 대기 (연속 요청 방지)
		time.Sleep(500 * time.Millisecond)
	}

	if !success {
		log.Printf("닉네임 변경 최종 실패: %s\n", oldNickname)
	}
}

// replyToCommentsLoop 본인 글에 달린 댓글에 답글 달기 루프
func (m *ActivityManager) replyToCommentsLoop() {
	rng := rand.New(rand.NewSource(time.Now().UnixNano()))

	// 시작 직후 첫 반응 (30~120초 랜덤 대기)
	initialDelay := time.Duration(30+rng.Intn(90)) * time.Second
	select {
	case <-m.stopChan:
		return
	case <-time.After(initialDelay):
		m.respondToComments()
	}

	for {
		// 댓글 반응 간격: 10~30분 (랜덤)
		interval := time.Duration(10+rng.Intn(20)) * time.Minute

		select {
		case <-m.stopChan:
			return
		case <-time.After(interval):
			m.respondToComments()
		}
	}
}

// respondToComments 모든 AI 캐릭터의 게시글에 달린 댓글에 답글 달기
func (m *ActivityManager) respondToComments() {
	characters, err := m.characterService.GetAllCharacters()
	if err != nil || len(characters) == 0 {
		return
	}

	// 랜덤하게 캐릭터 순서 섞기
	rng := rand.New(rand.NewSource(time.Now().UnixNano()))
	rng.Shuffle(len(characters), func(i, j int) {
		characters[i], characters[j] = characters[j], characters[i]
	})

	// 최대 2명의 캐릭터만 반응
	limit := 2
	if len(characters) < limit {
		limit = len(characters)
	}

	for i := 0; i < limit; i++ {
		character := &characters[i]
		m.respondToCommentForCharacter(character)
	}
}

// respondToCommentForCharacter 특정 AI 캐릭터의 게시글에 달린 댓글에 답글 달기
func (m *ActivityManager) respondToCommentForCharacter(character *models.AICharacter) {
	// 본인 글에 달린 다른 사람의 댓글 중 아직 답글이 없는 것 조회
	comments, err := m.commentService.GetCommentsOnCharacterPosts(character.ID, 5)
	if err != nil || len(comments) == 0 {
		return
	}

	// 랜덤하게 하나의 댓글 선택
	rng := rand.New(rand.NewSource(time.Now().UnixNano()))
	comment := comments[rng.Intn(len(comments))]

	// 게시글 정보 조회
	post, err := m.postService.GetPost(comment.PostID)
	if err != nil {
		log.Printf("게시글 조회 실패: %v\n", err)
		return
	}

	// 공지사항 조회
	pinnedPosts, _ := m.postService.GetPinnedPosts()

	// LLM으로 답글 내용 생성
	replyContent, err := m.llmService.GenerateReplyContent(character, post, comment, pinnedPosts)
	if err != nil {
		log.Printf("답글 생성 실패: %v\n", err)
		return
	}

	// 답글 작성 (대댓글로)
	_, err = m.commentService.CreateCommentByAuthor("ai", character.ID, post.ID, replyContent, &comment.ID)
	if err != nil {
		log.Printf("답글 저장 실패: %v\n", err)
		return
	}

	log.Printf("[답글] %s -> %s의 댓글에 답글: \"%s\"...\n", character.Nickname, comment.AuthorNickname, truncateString(replyContent, 30))

	// 콜백 호출
	if m.onNewComment != nil {
		m.onNewComment()
	}
}

func truncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
}
