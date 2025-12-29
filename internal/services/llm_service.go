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
	"net/http"
	"strings"
	"sync"
	"time"
)

// LLMService LLM 연동 서비스 (LM Studio는 동시 요청 불가하므로 직렬화)
type LLMService struct {
	config    models.LLMConfig
	client    *http.Client
	requestMu sync.Mutex // LLM 요청 직렬화를 위한 뮤텍스
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
			MaxTokens:       2000,
			Temperature:     0.8,
		},
		client: &http.Client{
			Timeout: 120 * time.Second,
		},
	}
}

// UpdateConfig LLM 설정 변경
func (s *LLMService) UpdateConfig(config models.LLMConfig) {
	s.config = config
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

	return &postContent, nil
}

// GenerateCommentContent AI 캐릭터로 댓글 내용 생성
func (s *LLMService) GenerateCommentContent(character *models.AICharacter, post *models.Post, existingComments []*models.Comment, pinnedPosts []models.Post) (string, error) {
	prompt := s.buildCommentPrompt(character, post, existingComments, pinnedPosts)
	modelName := s.selectModel(character.AssignedModelIndex)

	response, err := s.sendRequest(prompt, modelName)
	if err != nil {
		return "", err
	}

	return response, nil
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
func (s *LLMService) GenerateNickname(character *models.AICharacter) (string, error) {
	mbtiDesc := getMBTIDescription(character.MBTI)
	prompt := fmt.Sprintf(`다음 페르소나를 가진 인물의 닉네임을 하나만 지어주세요.
특성: 성별 %s, 나이 %d세, 직업 %s, MBTI %s (%s).
조건:
1. 2~8글자의 한글 (특수문자, 공백 제외)
2. 설명 없이 오직 닉네임 단어 하나만 응답할 것
3. 창의적이고 개성있는 닉네임`,
		character.Gender, character.Age, character.JobCategory, character.MBTI, mbtiDesc)

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
	// LM Studio는 동시 요청을 처리할 수 없으므로 직렬화
	s.requestMu.Lock()
	defer s.requestMu.Unlock()

	log.Printf("[LLM] 요청 시작 (모델: %s)\n", modelName)

	url := fmt.Sprintf("http://%s:%s/v1/chat/completions", s.config.Host, s.config.Port)

	maxTokens := s.config.MaxTokens
	if maxTokens <= 0 {
		maxTokens = 2000
	}
	temperature := s.config.Temperature
	if temperature <= 0 {
		temperature = 0.8
	}

	reqBody := ChatRequest{
		Model: modelName,
		Messages: []ChatMessage{
			{Role: "system", Content: "당신은 한국어로 대화하는 BBS 게시판 사용자입니다. 자연스럽고 인간적인 글을 작성합니다."},
			{Role: "user", Content: prompt},
		},
		Temperature: temperature,
		MaxTokens:   maxTokens,
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

	return chatResp.Choices[0].Message.Content, nil
}

// getTimeContext 현재 시간/계절 정보 생성
func getTimeContext() (timeStr, monthStr string) {
	currentTime := time.Now()
	hour := currentTime.Hour()
	minute := currentTime.Minute()
	month := currentTime.Month()

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

	// 월 형식
	monthStr = fmt.Sprintf("%d월", month)

	return
}

// buildPostPrompt 게시물 작성 프롬프트 생성
func (s *LLMService) buildPostPrompt(character *models.AICharacter, recentPosts []models.Post, popularPosts []models.Post, pinnedPosts []models.Post) string {
	mbtiDesc := getMBTIDescription(character.MBTI)
	timeStr, monthStr := getTimeContext()

	prompt := fmt.Sprintf(`당신은 %s라는 닉네임의 BBS 게시판 사용자입니다.
당신의 특성(참고용):
- 성별: %s, 나이: %d세, 거주지: %s, 취미: %s, 직종: %s
- MBTI: %s (공격성 %d/10, 진지함 %d/10)

[현재 시간]
- 시각: %s
- 월: %s

[글쓰기 스타일]
%s

`, character.Nickname, character.Gender, character.Age, character.Region, character.Hobby,
		character.JobCategory, character.MBTI, character.AggressionLevel,
		character.FormalityLevel, timeStr, monthStr, mbtiDesc)

	// 인격 요약이 있으면 포함
	if character.PersonaSummary != "" {
		prompt += fmt.Sprintf(`[당신의 인격 정의]
%s

`, character.PersonaSummary)
	}

	if len(recentPosts) > 0 {
		prompt += "당신이 최근 작성한 글:\n"
		for _, p := range recentPosts {
			prompt += fmt.Sprintf("- %s\n", p.Title)
		}
		prompt += "\n"
	}

	if len(popularPosts) > 0 {
		prompt += "현재 게시판의 인기 글:\n"
		for _, p := range popularPosts {
			prompt += fmt.Sprintf("- %s (by %s)\n", p.Title, p.AuthorNickname)
		}
		prompt += "\n"
	}

	if len(pinnedPosts) > 0 {
		prompt += "[게시판 중요 공지사항]\n"
		prompt += "현재 게시판 상단에 다음 공지가 게시되어 있습니다. 글 작성 시 이 내용을 참고하고 필요하다면 언급하거나 반응하세요:\n"
		for _, p := range pinnedPosts {
			prompt += fmt.Sprintf("- 제목: %s\n  내용 요약: %s\n", p.Title, truncateString(p.Content, 200))
		}
		prompt += "\n"
	}

	prompt += `일상적인 주제로 자연스러운 게시글을 작성해주세요.

[중요 규칙]
- 자기소개 금지 (나이, 직업, MBTI, 취미 등을 언급하지 마세요)
- "안녕하세요, 저는 ~입니다" 형태의 인사 금지
- 일상 이야기, 질문, 정보 공유, 잡담 등 자연스러운 글 작성
- 글쓰기 스타일만 성격에 맞게 반영
- 글 작성할 때 시각, 현재가 몇 월인지 참조할 수도 있습니다.

JSON 형식으로 응답: {"title": "제목", "content": "본문"}`

	return prompt
}

// buildCommentPrompt 댓글 작성 프롬프트 생성
func (s *LLMService) buildCommentPrompt(character *models.AICharacter, post *models.Post, existingComments []*models.Comment, pinnedPosts []models.Post) string {
	mbtiDesc := getMBTIDescription(character.MBTI)
	timeStr, monthStr := getTimeContext()

	prompt := fmt.Sprintf(`당신은 %s라는 닉네임의 BBS 사용자입니다.
글쓰기 스타일: %s
현재 시각: %s, %s

`, character.Nickname, mbtiDesc, timeStr, monthStr)

	// 인격 요약이 있으면 포함
	if character.PersonaSummary != "" {
		prompt += fmt.Sprintf(`[당신의 인격 정의]
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
		prompt += "현재 게시판 상단에 다음 공지가 고정되어 있습니다. 댓글 작성 시 이 내용을 인지하고 필요 시 참고하세요:\n"
		for _, p := range pinnedPosts {
			prompt += fmt.Sprintf("- %s\n", p.Title)
		}
		prompt += "\n"
	}

	prompt += `게시글에 대한 자연스러운 댓글을 작성해주세요.

[중요 규칙]
- 자기소개 금지 (나이, 직업, MBTI 등 언급 금지)
- 게시글 내용에 대한 반응, 의견, 질문만 작성
- 짧고 자연스럽게 (1~3문장)
- 시간대와 계절에 맞는 말투
- 댓글 내용만 작성 (JSON 형식 아님)`

	return prompt
}

// buildReplyPrompt 본인 글에 달린 댓글에 대한 답글 프롬프트 생성
func (s *LLMService) buildReplyPrompt(character *models.AICharacter, post *models.Post, comment *models.Comment, pinnedPosts []models.Post) string {
	mbtiDesc := getMBTIDescription(character.MBTI)
	timeStr, monthStr := getTimeContext()

	prompt := fmt.Sprintf(`당신은 %s라는 닉네임의 BBS 사용자입니다.
글쓰기 스타일: %s
현재 시각: %s, %s

`, character.Nickname, mbtiDesc, timeStr, monthStr)

	// 인격 요약이 있으면 포함
	if character.PersonaSummary != "" {
		prompt += fmt.Sprintf(`[당신의 인격 정의]
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

	prompt += `이 댓글에 대한 답글을 작성해주세요.

[중요 규칙]
- 본인의 글에 달린 댓글에 대한 답변이므로, 글 작성자로서 자연스럽게 응대하세요
- 댓글 작성자의 의견이나 질문에 대해 친절하게 반응하세요
- 자기소개 금지 (나이, 직업, MBTI 등 언급 금지)
- 짧고 자연스럽게 (1~3문장)
- 시간대와 계절에 맞는 말투
- 답글 내용만 작성 (JSON 형식 아님)`

	return prompt
}

// GeneratePersonaSummary AI 캐릭터의 인격 요약 생성
func (s *LLMService) GeneratePersonaSummary(character *models.AICharacter, recentPosts []models.Post, recentComments []*models.Comment) (string, error) {
	prompt := fmt.Sprintf(`다음은 BBS 사용자의 정보와 최근 활동입니다:

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

	prompt += `위 정보를 바탕으로 이 사용자의 가상 인격을 300자 이내로 정의해주세요.
- 말투, 성격, 관심사, 특징적인 표현 방식 등을 포함
- 이후 이 사용자가 글을 쓸 때 이 인격을 반영합니다
- 인격 설명만 작성 (다른 내용 불필요)`

	modelName := s.selectModel(character.AssignedModelIndex)
	response, err := s.sendRequest(prompt, modelName)
	if err != nil {
		return "", err
	}

	// 300자 제한
	if len(response) > 300 {
		response = response[:300]
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
func getMBTIDescription(mbti string) string {
	descriptions := map[string]string{
		"INTJ": `전략가. 논리적이고 체계적. 단정적 어조.`,
		"INTP": `논리술사. 호기심 많음. 질문을 많이 던짐.`,
		"ENTJ": `지도자. 결단력 있음. 직설적 어투.`,
		"ENTP": `변론가. 창의적이고 논쟁적. 반어법 사용.`,
		"INFJ": `옹호자. 통찰력 있고 부드러운 어조.`,
		"INFP": `중재자. 감성적이고 시적인 표현.`,
		"ENFJ": `선도자. 격려하고 긍정적인 표현.`,
		"ENFP": `활동가. 열정적이고 활발. 느낌표 사용.`,
		"ISTJ": `논리주의자. 사실적이고 격식있는 어조.`,
		"ISFJ": `수호자. 친절하고 배려하는 말투.`,
		"ESTJ": `경영자. 직설적이고 규칙 강조.`,
		"ESFJ": `집정관. 사교적이고 친근한 말투.`,
		"ISTP": `장인. 간결하고 실용적인 말투.`,
		"ISFP": `모험가. 감각적이고 부드러운 느낌.`,
		"ESTP": `사업가. 에너지 넘치고 경험담 위주.`,
		"ESFP": `연예인. 재미있고 유머러스. 이모티콘.`,
	}

	if desc, ok := descriptions[mbti]; ok {
		return desc
	}
	return "다양한 성격의 일반 사용자입니다."
}

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
