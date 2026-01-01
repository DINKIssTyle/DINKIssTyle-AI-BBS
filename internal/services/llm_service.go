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

	queueSize int32 // 현재 대기열 크기 (원자적 연산)
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
	err = json.Unmarshal([]byte(response), &postContent)
	if err != nil {
		postContent = extractPostContent(response)
	}

	if postContent.Title == "" {
		postContent.Title = "무제"
	}
	if postContent.Content == "" {
		postContent.Content = response
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

	return strings.TrimSpace(response), recommend, nil
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
	log.Printf("[LLM PROMPT] ==================================================\n%s\n==================================================\n", prompt)

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

	// 마크다운 굵은 글씨 강조(**) 제거
	content = strings.ReplaceAll(content, "**", "")

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
		topicInstruction = fmt.Sprintf("[필수 작성 주제: %s]\n이번 게시물은 반드시 위 주제와 관련지어 작성해야 합니다. 당신의 캐릭터 특성을 살려 이 주제에 대한 생각이나 경험을 이야기하세요.", randomTopic)
	} else {
		topicInstruction = fmt.Sprintf("[추천 주제: %s]\n특별히 쓸 내용이 없다면 위 주제를 활용해보세요. 물론 다른 자유로운 주제를 선택해도 좋습니다.", randomTopic)
	}

	prompt := fmt.Sprintf(`당신은 %s라는 닉네임의 글 작성자입니다.
당신은 %d세 %s으로 %s에 거주하며, 취미는 %s, 직종은 %s입니다.
당신의 MBTI는 %s이며, 공격성은 %d/10, 진지함은 %d/10 입니다.
글을 작성하고 있는 오늘은 %s, %s 입니다.
글작성 스타일은 %s입니다.

%s
`, character.Nickname, character.Age, character.Gender, character.Region, character.Hobby,
		character.JobCategory, character.MBTI, character.AggressionLevel,
		character.FormalityLevel, monthStr, timeStr, mbtiDesc, topicInstruction)

	// 인격 요약이 있으면 포함
	if character.PersonaSummary != "" {
		prompt += fmt.Sprintf(`[최근 보정 된 당신의 요약 정보]
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
글쓰기 스타일: %s
댓글을 쓰는 현재 날짜와 시간은 %s, %s 입니다.

`, character.Nickname, mbtiDesc, monthStr, timeStr)

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

// GeneratePersonaSummary AI 캐릭터의 인격 요약 생성
func (s *LLMService) GeneratePersonaSummary(character *models.AICharacter, recentPosts []models.Post, recentComments []*models.Comment) (string, error) {
	prompt := fmt.Sprintf(`다음은 커뮤니티 사용자의 정보와 최근 활동입니다:

[사용자 정보]
- 닉네임: %s
- 성별: %s, 나이: %d세
- 거주지: %s
- 취미: %s
- 직종: %s
- MBTI: %s
- 공격성: %d/10, 진지함: %d/10

`, character.Nickname, character.Gender, character.Age, character.Region,
		character.Hobby, character.JobCategory, character.MBTI,
		character.AggressionLevel, character.FormalityLevel)

	if len(recentPosts) > 0 {
		prompt += "[최근 작성한 글]\n"
		for i, p := range recentPosts {
			if i >= 5 {
				break
			}
			prompt += fmt.Sprintf("- %s: %s\n", p.Title, truncateString(p.Content, 100))
		}
		prompt += "\n"
	}

	if len(recentComments) > 0 {
		prompt += "[최근 작성한 댓글]\n"
		for i, c := range recentComments {
			if i >= 5 {
				break
			}
			prompt += fmt.Sprintf("- %s\n", truncateString(c.Content, 100))
		}
		prompt += "\n"
	}

	prompt += s.getPromptWithDefault("summary_instruction", models.DefaultSummaryInstruction)

	modelName := s.selectModel(character.AssignedModelIndex)
	response, err := s.sendRequest(prompt, modelName)
	if err != nil {
		return "", err
	}

	// 2000자 제한 (바이트가 아닌 실제 문자 수 기준)
	runes := []rune(response)
	if len(runes) > 2000 {
		response = string(runes[:2000])
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
	titleStart := strings.Index(response, `"title"`)
	if titleStart != -1 {
		colonIdx := strings.Index(response[titleStart:], ":")
		if colonIdx != -1 {
			valueStart := titleStart + colonIdx + 1
			firstQuote := strings.Index(response[valueStart:], `"`)
			if firstQuote != -1 {
				valueStart += firstQuote + 1
				closing := strings.Index(response[valueStart:], `"`)
				if closing != -1 {
					result.Title = response[valueStart : valueStart+closing]
				}
			}
		}
	}
	contentStart := strings.Index(response, `"content"`)
	if contentStart != -1 {
		colonIdx := strings.Index(response[contentStart:], ":")
		if colonIdx != -1 {
			valueStart := contentStart + colonIdx + 1
			firstQuote := strings.Index(response[valueStart:], `"`)
			if firstQuote != -1 {
				valueStart += firstQuote + 1
				closing := strings.LastIndex(response[valueStart:], `"`)
				if closing != -1 {
					result.Content = response[valueStart : valueStart+closing]
				} else {
					result.Content = response[valueStart:]
				}
			}
		}
	}
	return result
}
