# 프롬프트 설정 추가 가이드

새로운 프롬프트 설정을 추가할 때 수정해야 하는 파일들과 단계를 정리한 문서입니다.

> **예시**: `topic_hints` (주제 힌트) 추가 사례 기준

---

## 수정해야 할 파일 목록

| 순서 | 파일 | 역할 |
|:----:|------|------|
| 1 | `internal/models/defaults.go` | 기본값 상수 정의 |
| 2 | `internal/services/llm_service.go` | 프롬프트 사용 코드 |
| 3 | `app.go` - GetPrompt | 프롬프트 조회 시 기본값 반환 |
| 4 | `app.go` - ExportPromptSettings | txt 내보내기 |
| 5 | `app.go` - 불러오기 함수 | txt 불러오기 |
| 6 | `internal/web/server.go` | 웹 관리자 핸들러 |
| 7 | `internal/web/templates_unified.go` | 웹 관리자 UI |
| 8 | `frontend/index.html` | 데스크탑 앱 UI |
| 9 | `frontend/src/main.js` | PROMPT_MAP 추가 |

---

## 단계별 상세 설명

### 1. defaults.go - 기본값 상수 정의

```go
// 파일: internal/models/defaults.go
const DefaultTopicHints = `음식/맛집,날씨/계절,뉴스/이슈,...`
```

- 프롬프트 키 네이밍: `Default` + `PascalCase이름`
- 백틱(`)으로 멀티라인 문자열 정의

---

### 2. llm_service.go - 프롬프트 사용

```go
// 파일: internal/services/llm_service.go
topicHintsStr := s.getPromptWithDefault("topic_hints", models.DefaultTopicHints)
```

- `getPromptWithDefault(키, 기본값)` 함수 사용
- DB에 값이 없으면 기본값 반환

---

### 3. app.go - GetPrompt 함수

```go
// 파일: app.go (GetPrompt 함수 내 switch문)
case "topic_hints":
    return models.DefaultTopicHints
```

- 프롬프트 키에 대한 case 추가
- 데스크탑 앱/웹에서 초기화 시 기본값 반환용

---

### 4. app.go - ExportPromptSettings 함수

```go
// prompts 맵에 추가
prompts := map[string]string{
    // ...
    "topic_hints": "[주제 힌트]",
}

// orderedKeys 슬라이스에 추가 (순서 보장)
orderedKeys := []string{..., "topic_hints", ...}
```

- `prompts`: 키 → 헤더 매핑
- `orderedKeys`: 내보내기 순서

---

### 5. app.go - 불러오기 함수 (headerToKey 맵)

```go
// headerToKey 맵에 추가
headerToKey := map[string]string{
    // ...
    "[주제 힌트]": "topic_hints",
}
```

- 헤더 → 키 역매핑 (4번과 반대)

---

### 6. server.go - 웹 관리자 핸들러

```go
// promptKeys 슬라이스에 추가
promptKeys := []string{..., "topic_hints", ...}

// defaults 맵에 추가
defaults := map[string]string{
    // ...
    "topic_hints": models.DefaultTopicHints,
}
```

- 웹 관리자 페이지에서 로드할 프롬프트 목록

---

### 7. templates_unified.go - 웹 관리자 UI

```html
<div class="prompt-section">
    <h4>주제 힌트 (Topic Hints)</h4>
    <form method="POST">
        <input type="hidden" name="action" value="prompt">
        <input type="hidden" name="prompt_key" value="topic_hints">
        <div class="form-group">
            <textarea name="prompt_content" rows="2">{{.Prompts.topic_hints}}</textarea>
        </div>
        <div class="btn-group">
            <button type="submit" class="btn">저장</button>
            <button type="submit" name="reset" value="1" class="btn btn-reset">초기화</button>
        </div>
    </form>
</div>
```

- `prompt_key`: 프롬프트 키
- `{{.Prompts.키}}`: 템플릿 변수

---

### 8. frontend/index.html - 데스크탑 앱 UI

```html
<div class="form-group">
    <label>주제 힌트 (Topic Hints)</label>
    <textarea id="prompt-topic-hints" class="form-input" rows="2"></textarea>
    <div class="button-group-right">
        <button class="btn-warning btn-reset-prompt" data-key="topic_hints">초기화</button>
        <button class="btn-primary btn-save-prompt" data-key="topic_hints">저장</button>
    </div>
</div>
```

- `id`: `prompt-` + 키 (하이픈 사용)
- `data-key`: 프롬프트 키

---

### 9. frontend/src/main.js - PROMPT_MAP

```javascript
const PROMPT_MAP = {
    // ...
    'topic_hints': 'prompt-topic-hints',
};
```

- 키 → HTML element ID 매핑
- 로드/저장/내보내기/불러오기 자동 지원

---

## 체크리스트

```markdown
- [ ] defaults.go: 기본값 상수 추가
- [ ] llm_service.go: getPromptWithDefault 호출
- [ ] app.go GetPrompt: switch case 추가
- [ ] app.go Export: prompts 맵 + orderedKeys 추가
- [ ] app.go Import: headerToKey 맵 추가
- [ ] server.go: promptKeys + defaults 추가
- [ ] templates_unified.go: 웹 UI 폼 추가
- [ ] index.html: 데스크탑 UI 폼 추가
- [ ] main.js: PROMPT_MAP 추가
- [ ] 빌드 테스트
```

---

## 참고 사항

- **DB 테이블**: `prompts` 테이블에 자동 저장 (key_name, content)
- **초기화 동작**: DB에서 레코드 삭제 → GetPrompt가 기본값 반환
- **txt 내보내기 형식**: `[헤더명]\n내용\n\n` 형태
