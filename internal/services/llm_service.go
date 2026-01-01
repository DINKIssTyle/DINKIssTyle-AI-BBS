// Created by DINKIssTyle on 2025. Copyright (C) 2025 DINKI'ssTyle. All rights reserved.

package services

import (
	"aibbs/internal/models"
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"math/rand"
	"net/http"
	"regexp"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

// LLMService LLM 연동 서비스 (LM Studio는 동시 요청 불가하므로 직렬화)
type LLMService struct {
	config    models.LLMConfig
	client    *http.Client
	requestMu sync.Mutex // LLM 요청 직렬화를 위한 뮤텍스

	// 프롬프트 조회를 위한 콜백 함수
	promptGetter func(key string) string
	mbtiGetter   func(mbti string) string

	queueSize  int32 // 현재 대기열 크기 (원자적 연산)
	LogPrompts bool  // 프롬프트 전체 로그 출력 여부
}

// SetLogPrompts 프롬프트 로그 출력 여부 설정
func (s *LLMService) SetLogPrompts(enabled bool) {
	s.LogPrompts = enabled
	if enabled {
		log.Println("[LLM] 프롬프트 전체 로그 출력: ON")
	} else {
		log.Println("[LLM] 프롬프트 전체 로그 출력: OFF")
	}
}

// SetPromptGetters 프롬프트 조회 함수 설정
func (s *LLMService) SetPromptGetters(promptGetter func(key string) string, mbtiGetter func(mbti string) string) {
	s.promptGetter = promptGetter
	s.mbtiGetter = mbtiGetter
}

// getPromptWithDefault 키로 프롬프트를 조회하고 없으면 기본값 반환
func (s *LLMService) getPromptWithDefault(key string, defaultVal string) string {
	if s.promptGetter != nil {
		val := s.promptGetter(key)
		if val != "" {
			return val
		}
	}
	return defaultVal
}

// getMBTIDescWithDefault MBTI 설명 조회
func (s *LLMService) getMBTIDescWithDefault(mbti string) string {
	if s.mbtiGetter != nil {
		val := s.mbtiGetter(mbti)
		if val != "" {
			return val
		}
	}
	// Fallback to internal map (now in models)
	if desc, ok := models.DefaultMBTIDescriptions[mbti]; ok {
		return desc
	}
	return "다양한 성격의 일반 사용자입니다."
}

// GetJobKeywordsMap 직종별 키워드 맵 반환 (프롬프트에서 파싱)
func (s *LLMService) GetJobKeywordsMap() map[string][]string {
	content := s.getPromptWithDefault("job_keywords", models.DefaultJobKeywords)
	result := make(map[string][]string)

	lines := strings.Split(content, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		parts := strings.SplitN(line, ":", 2)
		if len(parts) == 2 {
			job := strings.TrimSpace(parts[0])
			keywordsStr := strings.TrimSpace(parts[1])
			keywords := strings.Split(keywordsStr, ",")
			for i, kw := range keywords {
				keywords[i] = strings.TrimSpace(kw)
			}
			result[job] = keywords
		}
	}
	return result
}

// NewLLMService 새 LLM 서비스 생성
func NewLLMService() *LLMService {
	return &LLMService{
		config: models.LLMConfig{
			Host:            "localhost",
			Port:            "1234",
			Model1:          "default",
			PostsPerHour:    5,
			CommentsPerHour: 10,
			MaxTokens:       4096,
			Temperature:     0.8,
			Timeout:         120,
		},
		client: &http.Client{
			Timeout: 120 * time.Second,
		},
		LogPrompts: true, // 기본값: 켜짐
	}
}

// UpdateConfig LLM 설정 변경
func (s *LLMService) UpdateConfig(config models.LLMConfig) {
	s.config = config
	// HTTP 클라이언트 타임아웃 업데이트
	timeout := config.Timeout
	if timeout <= 0 {
		timeout = 120
	}
	s.client.Timeout = time.Duration(timeout) * time.Second
}

// GetConfig 현재 설정 반환
func (s *LLMService) GetConfig() models.LLMConfig {
	return s.config
}

// ChatMessage OpenAI API 메시지 형식
type ChatMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// ChatRequest OpenAI API 요청 형식
type ChatRequest struct {
	Model       string        `json:"model"`
	Messages    []ChatMessage `json:"messages"`
	Temperature float64       `json:"temperature"`
	MaxTokens   int           `json:"max_tokens"`
	Seed        *int          `json:"seed,omitempty"`
}

// ChatResponse OpenAI API 응답 형식
type ChatResponse struct {
	Choices []struct {
		Message ChatMessage `json:"message"`
	} `json:"choices"`
	Error *struct {
		Message string `json:"message"`
	} `json:"error"`
}

// TestConnection LLM 연결 테스트
func (s *LLMService) TestConnection(host, port, model string) error {
	url := fmt.Sprintf("http://%s:%s/v1/chat/completions", host, port)

	reqBody := ChatRequest{
		Model: model,
		Messages: []ChatMessage{
			{Role: "user", Content: "Hello"},
		},
		Temperature: 0.7,
		MaxTokens:   10,
	}

	jsonBody, err := json.Marshal(reqBody)
	if err != nil {
		return fmt.Errorf("요청 생성 실패: %w", err)
	}

	resp, err := s.client.Post(url, "application/json", bytes.NewBuffer(jsonBody))
	if err != nil {
		return fmt.Errorf("LLM 서버 연결 실패: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("LLM 서버 응답 오류: %d", resp.StatusCode)
	}

	return nil
}

// selectModel 인덱스에 따라 모델 선택 (fallback 로직 포함)
func (s *LLMService) selectModel(index int) string {
	var model string
	switch index {
	case 1:
		model = s.config.Model1
	case 2:
		model = s.config.Model2
	case 3:
		model = s.config.Model3
	}

	// 선택된 모델이 비어있으면 Model1 사용 (기본값)
	if model == "" {
		model = s.config.Model1
	}
	// 그래도 없으면 default
	if model == "" {
		model = "default"
	}

	// fallback: Model2, Model3이 선택되었는데 비어있어서 Model1로 갔는데 그것도 비면... (위에서 처리됨)
	// 다만 사용자가 Model1을 비워두고 Model2만 채우는 경우는 거의 없다고 가정하거나,
	// 그런 경우에도 Model1을 fallback으로 쓰는게 맞음.
	// 단, Model1이 비어있고 Model2가 있으면 Model2를 쓰는게 나을수도?
	// 요구사항: "LLM 모델이 1개만 지정되었을 때는 자동으로 2,3 지정 AI회원도 1을 이용합니다."
	// 즉 Model1이 메인.

	return model
}

// GeneratePostContent AI 캐릭터로 게시물 내용 생성
type PostContent struct {
	Title   string `json:"title"`
	Content string `json:"content"`
}

func (s *LLMService) GeneratePostContent(character *models.AICharacter, recentPosts []models.Post, popularPosts []models.Post, pinnedPosts []models.Post) (*PostContent, error) {
	prompt := s.buildPostPrompt(character, recentPosts, popularPosts, pinnedPosts)
	modelName := s.selectModel(character.AssignedModelIndex)

	response, err := s.sendRequest(prompt, modelName)
	if err != nil {
		return nil, err
	}

	log.Printf("[LLM] 게시물 응답 (Model: %s): %s\n", modelName, response)

	var postContent PostContent

	// 1차 시도: 구조체 Unmarshal
	err = json.Unmarshal([]byte(response), &postContent)
	if err != nil || postContent.Content == "" {
		// 2차 시도: map[string]interface{} Unmarshal (키 이름이 다를 경우 대비)
		var rawMap map[string]interface{}
		if json.Unmarshal([]byte(response), &rawMap) == nil {
			if t, ok := rawMap["title"].(string); ok {
				postContent.Title = t
			}
			if c, ok := rawMap["content"].(string); ok {
				postContent.Content = c
			}
			// 다른 키 이름 대응
			if postContent.Title == "" {
				if t, ok := rawMap["post_title"].(string); ok {
					postContent.Title = t
				}
			}
			if postContent.Content == "" {
				if c, ok := rawMap["post_content"].(string); ok {
					postContent.Content = c
				}
			}
		}

		// 3차 시도: 정규식/휴리스틱 추출
		if postContent.Content == "" {
			extracted := extractPostContent(response)
			if extracted.Title != "" {
				postContent.Title = extracted.Title
			}
			if extracted.Content != "" {
				postContent.Content = extracted.Content
			}
		}
	}

	if postContent.Title == "" {
		postContent.Title = "무제"
	}

	// 내용이 비었는데 response가 JSON 형식이 아니라면(일반 텍스트라면) 그냥 텍스트를 내용으로 쓴다.
	// 하지만 JSON 형식({로 시작)이라면 파싱 실패로 간주하고 에러 로그를 남기거나 빈 상태로 둔다.
	trimmedResp := strings.TrimSpace(response)
	if postContent.Content == "" {
		if !strings.HasPrefix(trimmedResp, "{") {
			postContent.Content = response
		} else {
			// JSON인데 파싱 실패. 억지로 JSON을 보여주기보단 "내용을 불러올 수 없습니다"가 나음.
			// 하지만 사용자 경험을 위해 최대한 살리는 방향으로...
			// 만약 extractPostContent도 실패했다면 정말 형식이 깨진 것임.
			log.Printf("[ERROR] 게시물 JSON 파싱 완전 실패: %s", response)
			postContent.Content = "내용 생성 중 오류가 발생했습니다."
		}
	}

	// JSON 잔여물 정리 (", }, 줄바꿈 등)
	postContent.Content = cleanJSONArtifacts(postContent.Content)
	postContent.Title = strings.TrimSpace(postContent.Title)

	return &postContent, nil
}

// GenerateCommentContent AI 캐릭터로 댓글 내용 생성 (추천 여부 포함)
func (s *LLMService) GenerateCommentContent(character *models.AICharacter, post *models.Post, existingComments []*models.Comment, pinnedPosts []models.Post) (string, bool, error) {
	prompt := s.buildCommentPrompt(character, post, existingComments, pinnedPosts)
	modelName := s.selectModel(character.AssignedModelIndex)

	response, err := s.sendRequest(prompt, modelName)
	if err != nil {
		return "", false, err
	}

	// 추천 여부 파싱
	recommend := false
	if strings.Contains(response, "[RECOMMEND]") {
		recommend = true
		response = strings.ReplaceAll(response, "[RECOMMEND]", "")
	}
	response = strings.TrimSpace(response)

	// 중복 댓글 검사 (1회 재시도)
	isDuplicate := false
	if len(existingComments) > 0 {
		for _, c := range existingComments {
			// 공백 제거 후 비교 (이모지 등 미세한 차이 무시를 위해 Jaro-Winkler 등을 쓰면 좋지만 여기선 단순 포함/일치 검사)
			// 사용자가 제보한 케이스는 거의 똑같으므로 문자열 포함 여부로 체크
			respClean := strings.ReplaceAll(response, " ", "")
			existClean := strings.ReplaceAll(c.Content, " ", "")

			// 완전히 똑같거나, 기존 댓글이 새 댓글을 포함하고 있거나(부분집합), 새 댓글이 기존 댓글을 포함하는 경우
			if respClean == existClean || (len(respClean) > 10 && strings.Contains(existClean, respClean)) {
				isDuplicate = true
				break
			}
		}

		if isDuplicate {
			log.Printf("[LLM] 중복 댓글 감지됨, 재생성 시도 (Model: %s)", modelName)
			// 프롬프트에 강력한 경고 추가하여 재시도
			retryPrompt := prompt + "\n\n[SYSTEM WARNING] 방금 생성한 답변은 이미 존재하는 댓글과 똑같습니다. 절대 똑같이 쓰지 말고, 완전히 다른 문장으로 다시 작성하세요."
			retryResponse, err := s.sendRequest(retryPrompt, modelName)
			if err == nil {
				retryResponse = strings.ReplaceAll(retryResponse, "[RECOMMEND]", "") // 재시도 응답에서도 태그 제거
				response = strings.TrimSpace(retryResponse)
			} else {
				log.Printf("[LLM] 재생성 실패, 중복된 댓글 폐기: %v", err)
				return "", false, errors.New("중복 댓글 감지 및 재생성 실패")
			}
		}
	}

	return response, recommend, nil
}

// GenerateReplyContent AI 캐릭터가 본인 글에 달린 댓글에 답글 생성
func (s *LLMService) GenerateReplyContent(character *models.AICharacter, post *models.Post, comment *models.Comment, pinnedPosts []models.Post) (string, error) {
	prompt := s.buildReplyPrompt(character, post, comment, pinnedPosts)
	modelName := s.selectModel(character.AssignedModelIndex)

	response, err := s.sendRequest(prompt, modelName)
	if err != nil {
		return "", err
	}

	return response, nil
}

// GenerateNickname AI 캐릭터 닉네임 생성
func (s *LLMService) GenerateNickname(character *models.AICharacter, excludeNicknames []string) (string, error) {
	mbtiDesc := s.getMBTIDescWithDefault(character.MBTI)

	userPrompt := s.getPromptWithDefault("nickname_gen", models.DefaultNicknamePrompt)

	if len(excludeNicknames) > 0 {
		userPrompt += fmt.Sprintf("\n\n[Constraint] The following nicknames are already taken or invalid. DO NOT use them: %s\nCreate a different, unique nickname.", strings.Join(excludeNicknames, ", "))
	}

	// 플레이스홀더 치환
	prompt := strings.ReplaceAll(userPrompt, "{gender}", character.Gender)
	prompt = strings.ReplaceAll(prompt, "{age}", fmt.Sprintf("%d", character.Age))
	prompt = strings.ReplaceAll(prompt, "{job}", character.JobCategory)
	prompt = strings.ReplaceAll(prompt, "{hobby}", character.Hobby)
	prompt = strings.ReplaceAll(prompt, "{mbti}", character.MBTI)
	prompt = strings.ReplaceAll(prompt, "{mbti_desc}", mbtiDesc)

	// 닉네임 생성은 기본 모델(Model1) 사용
	modelName := s.config.Model1
	if modelName == "" {
		modelName = "default"
	}

	response, err := s.sendRequest(prompt, modelName)
	if err != nil {
		return "", err
	}

	// 닉네임 정제 (공백, 따옴표, 특수문자 제거)
	nickname := strings.TrimSpace(response)
	nickname = strings.ReplaceAll(nickname, "\"", "")
	nickname = strings.ReplaceAll(nickname, "'", "")
	nickname = strings.ReplaceAll(nickname, ".", "")
	nickname = strings.ReplaceAll(nickname, "\n", "")
	nickname = strings.ReplaceAll(nickname, " ", "")

	// 8자 제한
	runes := []rune(nickname)
	if len(runes) > 8 {
		nickname = string(runes[:8])
	}

	// 빈 닉네임이면 에러
	if len(nickname) < 2 {
		return "", errors.New("생성된 닉네임이 너무 짧음")
	}

	log.Printf("[LLM] 닉네임 생성 완료: %s\n", nickname)
	return nickname, nil
}

// sendRequest LLM에 요청 전송 (뮤텍스로 직렬화)
func (s *LLMService) sendRequest(prompt string, modelName string) (string, error) {
	// 대기열 크기 증가
	currentQueue := atomic.AddInt32(&s.queueSize, 1)
	defer atomic.AddInt32(&s.queueSize, -1)

	// LM Studio는 동시 요청을 처리할 수 없으므로 직렬화
	s.requestMu.Lock()
	defer s.requestMu.Unlock()

	log.Printf("[LLM] 요청 시작 (모델: %s, 현재 대기열: %d)\n", modelName, currentQueue)
	if s.LogPrompts {
		log.Printf("[LLM PROMPT] ==================================================\n%s\n==================================================\n", prompt)
	} else {
		log.Printf("[LLM PROMPT] (Hidden - Toggle in Console to view full prompt)\n")
	}

	url := fmt.Sprintf("http://%s:%s/v1/chat/completions", s.config.Host, s.config.Port)

	maxTokens := s.config.MaxTokens
	if maxTokens <= 0 {
		maxTokens = 4096
	}
	temperature := s.config.Temperature
	if temperature <= 0 {
		temperature = 0.8
	}

	// 전략 5: 동적 Temperature 조정 (창의성 증가를 위한 랜덤 부스트)
	// 30% 확률로 Temperature를 0.1~0.3 증가시켜 다양성 확보
	rng := rand.New(rand.NewSource(time.Now().UnixNano()))
	if rng.Float64() < 0.3 {
		tempBoost := 0.1 + rng.Float64()*0.2 // 0.1 ~ 0.3
		temperature += tempBoost
		if temperature > 1.5 {
			temperature = 1.5
		}
		log.Printf("[LLM] Temperature 부스트 적용: %.2f\n", temperature)
	}

	// 랜덤 시드 생성 (캐싱 방지 및 다양성 확보)
	seedVal := int(time.Now().UnixNano() % 2147483647)

	// 시스템 메시지에 Randomizer 주입 (일부 모델은 Seed 파라미터를 무시하므로 프롬프트에도 명시)
	sysContent := s.getPromptWithDefault("system_role", models.DefaultSystemRole)
	sysContent = fmt.Sprintf("%s\n\n[System Note: Randomizer ID %d]", sysContent, seedVal)

	reqBody := ChatRequest{
		Model: modelName,
		Messages: []ChatMessage{
			{Role: "system", Content: sysContent},
			{Role: "user", Content: prompt},
		},
		Temperature: temperature,
		MaxTokens:   maxTokens,
		Seed:        &seedVal,
	}

	jsonBody, err := json.Marshal(reqBody)
	if err != nil {
		return "", fmt.Errorf("요청 생성 실패: %w", err)
	}

	resp, err := s.client.Post(url, "application/json", bytes.NewBuffer(jsonBody))
	if err != nil {
		return "", fmt.Errorf("LLM 서버 연결 실패: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("응답 읽기 실패: %w", err)
	}

	if s.LogPrompts {
		log.Printf("[LLM RESPONSE RAW] %s\n", string(body))
	}

	var chatResp ChatResponse
	err = json.Unmarshal(body, &chatResp)
	if err != nil {
		return "", fmt.Errorf("응답 파싱 실패: %w", err)
	}

	if chatResp.Error != nil {
		return "", fmt.Errorf("LLM 오류: %s", chatResp.Error.Message)
	}

	if len(chatResp.Choices) == 0 {
		return "", errors.New("LLM 응답이 비어있습니다")
	}

	content := chatResp.Choices[0].Message.Content
	content = sanitizeLLMResponse(content)
	return content, nil
}

// sanitizeLLMResponse LLM 응답 정제 (이스케이프 시퀀스 변환, 불필요한 문자 제거)
func sanitizeLLMResponse(content string) string {
	// 마크다운 코드 블록 제거 (```json ... ``` 또는 ``` ... ```)
	content = strings.TrimSpace(content)
	if strings.HasPrefix(content, "```") {
		// 첫 번째 줄 제거 (```json 또는 ```)
		if idx := strings.Index(content, "\n"); idx != -1 {
			content = content[idx+1:]
		}
		// 마지막 ``` 제거
		if strings.HasSuffix(content, "```") {
			content = content[:len(content)-3]
		}
		content = strings.TrimSpace(content)
	}

	// 리터럴 이스케이프 시퀀스를 실제 문자로 변환
	content = strings.ReplaceAll(content, "\\n\\n", "\n\n") // 먼저 \\n\\n 처리
	content = strings.ReplaceAll(content, "\\n", "\n")      // 그 다음 \\n 처리
	content = strings.ReplaceAll(content, "\\t", "\t")      // 탭
	content = strings.ReplaceAll(content, "\\r", "")        // 캐리지 리턴 제거

	// --- 구분자 이후 텍스트 제거 (자체 평가 멘트 등)
	// 보통 "--- 글의 품질이..." 형태로 나타남
	if idx := strings.Index(content, "\n---"); idx != -1 {
		content = content[:idx]
	}

	// 마크다운 굵은 글씨 강조(**) 및 기울임(*) 제거
	content = strings.ReplaceAll(content, "**", "")
	content = strings.ReplaceAll(content, "*", "")

	// 이모지 제거 (유니코드 범위: 이모티콘, 심볼, 픽토그램 등)
	emojiPattern := regexp.MustCompile(`[\x{1F300}-\x{1F5FF}\x{1F600}-\x{1F64F}\x{1F680}-\x{1F6FF}\x{1F900}-\x{1F9FF}\x{2600}-\x{26FF}\x{2700}-\x{27BF}]`)
	content = emojiPattern.ReplaceAllString(content, "")

	// 앞뒤 공백 및 불필요한 따옴표 제거
	content = strings.TrimSpace(content)

	// 응답이 따옴표로 감싸진 경우 제거
	if len(content) >= 2 && content[0] == '"' && content[len(content)-1] == '"' {
		content = content[1 : len(content)-1]
	}

	return content
}

// cleanJSONArtifacts 본문에서 JSON 잔여물 제거 (", }, 줄바꿈 조합 등)
func cleanJSONArtifacts(content string) string {
	content = strings.TrimSpace(content)

	// 마지막에 남은 JSON 종료 패턴들 제거 (반복 적용)
	for {
		trimmed := false
		// 패턴: 줄바꿈 + } + 줄바꿈 등
		suffixes := []string{
			"\n}\n",
			"\n}",
			"}\n",
			"\"\n}",
			"\"}\n",
			"\"}",
			"\" }",
			"}",
			"\n\"",
			"\"",
		}
		for _, suffix := range suffixes {
			if strings.HasSuffix(content, suffix) {
				// } 단독은 실제 내용일 수 있으므로 앞에 특수문자가 있는 경우만 제거
				if suffix == "}" || suffix == "\"" {
					// 바로 앞 문자 확인
					if len(content) > 1 {
						prevChar := content[len(content)-2]
						if prevChar == '\n' || prevChar == ' ' || prevChar == '\t' {
							content = content[:len(content)-len(suffix)]
							trimmed = true
						}
					}
				} else {
					content = content[:len(content)-len(suffix)]
					trimmed = true
				}
				if trimmed {
					break
				}
			}
		}
		if !trimmed {
			break
		}
		content = strings.TrimSpace(content)
	}

	return content
}

// getTimeContext 현재 시간/계절 정보 생성
func getTimeContext() (timeStr, monthStr string) {
	currentTime := time.Now()
	hour := currentTime.Hour()
	minute := currentTime.Minute()
	month := currentTime.Month()
	day := currentTime.Day()

	// 오전/오후 + 시:분 형식
	meridiem := "오전"
	displayHour := hour
	if hour >= 12 {
		meridiem = "오후"
		if hour > 12 {
			displayHour = hour - 12
		}
	}
	if hour == 0 {
		displayHour = 12
	}
	timeStr = fmt.Sprintf("%s %d시 %d분", meridiem, displayHour, minute)

	// 월일 형식
	monthStr = fmt.Sprintf("%d월 %d일", month, day)

	return
}

// buildPostPrompt 게시물 작성 프롬프트 생성
func (s *LLMService) buildPostPrompt(character *models.AICharacter, recentPosts []models.Post, popularPosts []models.Post, pinnedPosts []models.Post) string {
	mbtiDesc := s.getMBTIDescWithDefault(character.MBTI)
	timeStr, monthStr := getTimeContext()

	// 랜덤 주제 힌트 생성 (전략 2) - DB에서 읽어오거나 기본값 사용
	topicHintsStr := s.getPromptWithDefault("topic_hints", models.DefaultTopicHints)
	topicHints := strings.Split(topicHintsStr, ",")
	rng := rand.New(rand.NewSource(time.Now().UnixNano()))
	randomTopic := strings.TrimSpace(topicHints[rng.Intn(len(topicHints))])

	// 주제 선택 강제성 (70% 확률로 주제 고정, 30% 자유)
	topicInstruction := ""
	if rng.Float64() < 0.7 {
		topicInstruction = fmt.Sprintf("[필수 작성 주제: %s]\n이번 게시물은 반드시 위 주제와 관련지어 작성해야 합니다. 당신의 인격, 캐릭터 특성을 살려 이 주제에 대한 생각이나 경험을 이야기하세요.", randomTopic)
	} else {
		topicInstruction = fmt.Sprintf("[추천 주제: %s]\n위 주제를 활용해보세요. 다른 자유로운 주제를 선택해도 좋습니다.", randomTopic)
	}

	prompt := fmt.Sprintf(`당신의 닉네임은 %s입니다.
당신은 %s %d세 %s입니다.
당신이 살아온 인생: %s

당신의 글작성 스타일은 %s이며, 공격성은 %d/10, 진지함은 %d/10 입니다.
참고로 현재는 %s, %s 입니다.

%s
`, character.Nickname, character.Birthdate, character.Age, character.Gender,
		character.Backstory,
		mbtiDesc, character.AggressionLevel, character.FormalityLevel,
		monthStr, timeStr, topicInstruction)

	// 인격 요약이 있으면 포함
	if character.PersonaSummary != "" {
		prompt += fmt.Sprintf(`[당신의 최근 활동입니다]
%s

`, character.PersonaSummary)
	}

	// 다른 사람 글에 반응 유도 (전략 3)
	if len(popularPosts) > 0 {
		prompt += "[선택 가능한 행동]\n1. 완전히 새로운 주제로 자유글 작성\n2. 아래 인기글 중 하나를 골라 의견/반응 글 작성 (선택사항):\n"
		for _, p := range popularPosts {
			prompt += fmt.Sprintf("   - \"%s\" (by %s)\n", p.Title, p.AuthorNickname)
		}
		prompt += "\n"
	}

	if len(pinnedPosts) > 0 {
		prompt += "[게시판 중요 공지사항]\n현재 게시판 상단에 다음 공지가 게시되어 있습니다. 게시판 이용에 이 내용을 참고하고 필요하다면 언급하거나 반응하세요:\n"
		for _, p := range pinnedPosts {
			prompt += fmt.Sprintf("- 제목: %s\n  내용 요약: %s\n", p.Title, truncateString(p.Content, 200))
		}
		prompt += "\n"
	}

	prompt += s.getPromptWithDefault("post_instruction", models.DefaultPostInstruction)

	return prompt
}

// buildCommentPrompt 댓글 작성 프롬프트 생성
func (s *LLMService) buildCommentPrompt(character *models.AICharacter, post *models.Post, existingComments []*models.Comment, pinnedPosts []models.Post) string {
	mbtiDesc := s.getMBTIDescWithDefault(character.MBTI)
	timeStr, monthStr := getTimeContext()

	prompt := fmt.Sprintf(`당신은 %s라는 닉네임의 커뮤니티 사용자입니다.
당신은 %s %d세 %s입니다.
당신이 살아온 인생(서사): %s

글쓰기 스타일: %s
댓글을 쓰는 현재 날짜와 시간은 %s, %s 입니다.

`, character.Nickname, character.Birthdate, character.Age, character.Gender, character.Backstory, mbtiDesc, monthStr, timeStr)

	// 인격 요약이 있으면 포함
	if character.PersonaSummary != "" {
		prompt += fmt.Sprintf(`[당신의 인격과 캐릭터 정의]
%s

`, character.PersonaSummary)
	}

	prompt += fmt.Sprintf(`다음 게시글에 댓글을 달아주세요:

제목: %s
작성자: %s
내용: %s

`, post.Title, post.AuthorNickname, post.Content)

	if len(existingComments) > 0 {
		prompt += "기존 댓글들:\n"
		for _, c := range existingComments {
			prompt += fmt.Sprintf("- %s: %s\n", c.AuthorNickname, c.Content)
		}
		prompt += "\n"
	}

	if len(pinnedPosts) > 0 {
		prompt += "[게시판 중요 공지사항]\n"
		prompt += "현재 게시판 상단에 다음 공지가 게시되어 있습니다. 게시판 이용에 이 내용을 참고하고 필요하다면 언급하거나 반응하세요:\n"
		for _, p := range pinnedPosts {
			prompt += fmt.Sprintf("- %s\n", p.Title)
		}
		prompt += "\n"
	}

	prompt += s.getPromptWithDefault("comment_instruction", models.DefaultCommentInstruction)

	prompt += `
[추천 관련 중요 지시사항]
추천은 정말 특별한 경우에만 해주세요. 대부분의 글에는 추천하지 마세요.
다음 조건을 모두 만족할 때만 댓글 끝에 [RECOMMEND]를 추가하세요:
1. 글의 품질이 상위 10% 수준으로 뛰어나거나
2. 당신의 취미/관심사와 완벽하게 일치하며 유익한 정보가 있거나
3. 매우 재미있거나 감동적이어서 다른 사람에게도 꼭 알리고 싶은 경우

단순히 괜찮은 글, 보통 수준의 글에는 절대 추천하지 마세요.
10개의 글 중 1~2개 정도만 추천하세요.
`

	return prompt
}

// buildReplyPrompt 본인 글에 달린 댓글에 대한 답글 프롬프트 생성
func (s *LLMService) buildReplyPrompt(character *models.AICharacter, post *models.Post, comment *models.Comment, pinnedPosts []models.Post) string {
	mbtiDesc := s.getMBTIDescWithDefault(character.MBTI)
	timeStr, monthStr := getTimeContext()

	prompt := fmt.Sprintf(`당신은 %s라는 닉네임의 커뮤니티 사용자입니다.
글쓰기 스타일: %s
현재 시각: %s, %s

`, character.Nickname, mbtiDesc, timeStr, monthStr)

	// 인격 요약이 있으면 포함
	if character.PersonaSummary != "" {
		prompt += fmt.Sprintf(`[당신의 인격과 캐릭터 정의]
%s

`, character.PersonaSummary)
	}

	prompt += fmt.Sprintf(`[중요: 이것은 당신이 쓴 게시글입니다]
제목: %s
내용: %s

`, post.Title, truncateString(post.Content, 300))

	prompt += fmt.Sprintf(`[당신의 글에 %s님이 다음과 같은 댓글을 달았습니다]
댓글 작성자: %s
댓글 내용: %s

`, comment.AuthorNickname, comment.AuthorNickname, comment.Content)

	if len(pinnedPosts) > 0 {
		prompt += "[게시판 공지사항]\n"
		for _, p := range pinnedPosts {
			prompt += fmt.Sprintf("- %s\n", p.Title)
		}
		prompt += "\n"
	}

	prompt += s.getPromptWithDefault("reply_instruction", models.DefaultReplyInstruction)

	return prompt
}

// GenerateBackstory 캐릭터 초기 서사 생성 (1000자)
func (s *LLMService) GenerateBackstory(character *models.AICharacter) (string, error) {
	mbtiDesc := s.getMBTIDescWithDefault(character.MBTI)

	prompt := fmt.Sprintf(`당신은 이제 막 커뮤니티 활동을 시작하려는 인물입니다.
다음 설정값을 바탕으로 당신의 과거, 현재 상황, 가치관, 트라우마, 꿈 등을 포함한 풍부한 서사(Backstory)를 1000자 내외로 작성해주세요.
이 내용은 당신의 '인격(Persona)'으로 사용될 것입니다.

[캐릭터 설정]
- 닉네임: %s
- 나이: %d세, 성별: %s
- 거주지: %s, 직업: %s
- 취미: %s
- MBTI: %s (%s)
- 성향: 공격성 %d/10, 진지함 %d/10

[지침]
1. 단순한 소개글이 아니라, 한 편의 수필이나 소설 속 등장인물 소개처럼 깊이 있게 작성하세요.
2. 왜 이 커뮤니티에 오게 되었는지, 어떤 글을 쓰고 싶은지 자연스럽게 녹여내세요.
3. 말투는 캐릭터의 성격에 맞게 설정하되, 서사 자체는 '나'의 독백이나 '제3자'의 관찰 시점 중 하나로 일관되게 작성하세요.
4. 요약된 정보(나이, 직업 등)를 단순히 나열하지 말고 이야기 속에 녹여내세요.
`, character.Nickname, character.Age, character.Gender, character.Region, character.JobCategory,
		character.Hobby, character.MBTI, mbtiDesc, character.AggressionLevel, character.FormalityLevel)

	modelName := s.selectModel(character.AssignedModelIndex)
	response, err := s.sendRequest(prompt, modelName)
	if err != nil {
		return "", err
	}

	return strings.TrimSpace(response), nil
}

// GenerateActivitySummary AI 캐릭터의 최근 활동 요약 생성 (기존 GeneratePersonaSummary 대체)
func (s *LLMService) GenerateActivitySummary(character *models.AICharacter, recentPosts []models.Post, recentComments []*models.Comment) (string, error) {
	prompt := fmt.Sprintf(`다음은 커뮤니티 사용자의 닉네임과 최근 활동(작성한 글, 댓글) 목록입니다.
정보를 읽고 현재 이 사용자가 어떤 관심사를 가지고 활동하고 있는지 "최근 활동 요약"을 500자 이내로 작성해주세요.

[사용자 정보]
- 닉네임: %s
(참고: 나이, 지역, 직업, MBTI 등 고정적인 개인정보는 요약에 포함하지 마세요. 오직 최근 활동 내용에만 집중하세요.)

`, character.Nickname)

	if len(recentPosts) > 0 {
		prompt += "[최근 작성한 글 (최신순 3개)]\n"
		count := 0
		for _, p := range recentPosts {
			if count >= 3 {
				break
			}
			prompt += fmt.Sprintf("- 제목: %s\n  내용: %s\n", p.Title, truncateString(p.Content, 200))
			count++
		}
		prompt += "\n"
	} else {
		prompt += "[최근 작성한 글]\n없음\n\n"
	}

	if len(recentComments) > 0 {
		prompt += "[최근 작성한 댓글 (최신순 3개)]\n"
		count := 0
		for _, c := range recentComments {
			if count >= 3 {
				break
			}
			prompt += fmt.Sprintf("- 내용: %s\n", truncateString(c.Content, 200))
			count++
		}
		prompt += "\n"
	} else {
		prompt += "[최근 작성한 댓글]\n없음\n\n"
	}

	prompt += `[지침]
1. 사용자가 최근에 쓴 글과 댓글의 주제, 논조, 감정 상태 등을 분석하여 서술하세요.
2. 예: "최근에는 주로 요리에 대한 글을 쓰며 회원들과 레시피를 공유하고 있다. 댓글에서는 친절한 태도를 보이지만, 특정 주제에 대해서는 단호한 모습을 보였다."
3. 고정적인 인적사항(나이, 성별, 직업 등)은 절대 언급하지 마세요.
4. 요약은 간결하고 명확하게 작성하세요.
`

	modelName := s.selectModel(character.AssignedModelIndex)
	response, err := s.sendRequest(prompt, modelName)
	if err != nil {
		return "", err
	}

	// 문자 수 제한
	runes := []rune(response)
	if len(runes) > 1000 {
		response = string(runes[:1000])
	}

	return strings.TrimSpace(response), nil
}

// truncateString 문자열 절단
func truncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
}

// getMBTIDescription MBTI 유형별 상세 설명 반환

// extractPostContent 불완전한 JSON에서 title과 content 추출
func extractPostContent(response string) PostContent {
	result := PostContent{}

	// 스마트 쿼트 변환 (파싱 확률 증가를 위해)
	response = strings.ReplaceAll(response, "“", "\"")
	response = strings.ReplaceAll(response, "”", "\"")

	// 제목 추출 시도 (정규식 사용이 더 안전함)
	// "title": "..." 패턴 찾기
	titlePattern := regexp.MustCompile(`"title"\s*:\s*"(.*?)"`)
	titleMatch := titlePattern.FindStringSubmatch(response)
	if len(titleMatch) > 1 {
		result.Title = titleMatch[1]
	}

	// 본문 추출 시도 1: 정규식 (줄바꿈 포함 (?s))
	// "content": "..." 패턴
	contentPattern := regexp.MustCompile(`(?s)"content"\s*:\s*"(.*?)"`)
	contentMatch := contentPattern.FindStringSubmatch(response)
	if len(contentMatch) > 1 {
		// 내용이 너무 짧으면(5자 미만) 무시하고 뒷부분 검색으로 넘어갈 수도 있음
		// 하지만 "content": "\n 내용..." 형태라면 여기서 잡힐 것임.
		if len(contentMatch[1]) > 0 {
			result.Content = contentMatch[1]
		}
	}

	// 본문 추출 시도 2: 정규식으로 못 잡았거나 내용이 비어있는 경우, 뒷부분 텍스트 탐색
	// Ministral: "content": "" \n "내용..."
	if result.Content == "" || len(result.Content) < 5 {
		// "content" 키워드 뒤쪽을 수동 탐색
		idx := strings.LastIndex(response, `"content"`)
		if idx != -1 {
			afterContent := response[idx:]
			// 콜론 찾기
			colonIdx := strings.Index(afterContent, ":")
			if colonIdx != -1 {
				valuePart := afterContent[colonIdx+1:]
				valuePart = strings.TrimSpace(valuePart)

				// "" 빈 문자열이 바로 오는지 확인
				if strings.HasPrefix(valuePart, `""`) {
					// 빈 문자열 뒤의 나머지 텍스트를 본문으로 간주
					remainder := strings.TrimSpace(valuePart[2:])
					if len(remainder) > 0 {
						// 따옴표로 감싸져 있다면 제거
						if strings.HasPrefix(remainder, `"`) && strings.HasSuffix(remainder, `"`) {
							result.Content = remainder[1 : len(remainder)-1]
						} else if strings.HasPrefix(remainder, `"`) {
							result.Content = remainder[1:]
						} else {
							result.Content = remainder
						}
					}
				}
			}
		}
	}

	// 본문 추출 시도 3: "내용은 다음처럼 작성되었습니다:" 등의 프리픽스 제거
	if result.Content != "" {
		prefixes := []string{"내용은 다음처럼 작성되었습니다:", "내용은:", "작성된 내용:"}
		for _, p := range prefixes {
			if idx := strings.Index(result.Content, p); idx != -1 {
				// 해당 문구 이후 줄바꿈 뒤의 내용을 진짜 본문으로 간주
				if newlineIdx := strings.Index(result.Content[idx:], "\n"); newlineIdx != -1 {
					result.Content = result.Content[idx+newlineIdx+1:]
				} else {
					result.Content = result.Content[idx+len(p):]
				}
			}
		}
	}

	return result
}
