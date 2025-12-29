// Created by DINKIssTyle on 2025. Copyright (C) 2025 DINKI'ssTyle. All rights reserved.

package services

import (
	"aibbs/internal/database"
	"aibbs/internal/models"
	"errors"
	"fmt"
	"math/rand"
	"time"
)

// CharacterService AI 캐릭터 서비스
type CharacterService struct {
	db *database.Database
}

// NewCharacterService 새 캐릭터 서비스 생성
func NewCharacterService(db *database.Database) *CharacterService {
	return &CharacterService{db: db}
}

// GenerateCharacters AI 캐릭터 자동 생성
func (s *CharacterService) GenerateCharacters(count int) ([]models.AICharacter, error) {
	if count < 1 || count > 100 {
		return nil, errors.New("캐릭터 수는 1~100명 사이여야 합니다")
	}

	db := s.db.GetDB()
	if db == nil {
		return nil, errors.New("데이터베이스에 연결되지 않았습니다")
	}

	rng := rand.New(rand.NewSource(time.Now().UnixNano()))
	characters := make([]models.AICharacter, 0, count)
	genders := []string{"남성", "여성"}

	// 현재 최대 인덱스 찾기 (활동전AI 이름 충돌 방지용)
	// 간단히 현재 시간 기반이나 루프에서 체크

	for i := 0; i < count; i++ {
		// 모델 인덱스 할당 (1, 2, 3 순환)
		modelIdx := (i % 3) + 1

		// 초기 닉네임 생성
		nickname := fmt.Sprintf("활동전AI%d", i+1)

		// 중복 확인 및 재시도 (활동전AI + 난수)
		for retry := 0; retry < 100; retry++ {
			var cnt int
			err := db.QueryRow("SELECT COUNT(*) FROM ai_characters WHERE nickname = ?", nickname).Scan(&cnt)
			if err != nil || cnt == 0 {
				break
			}
			nickname = fmt.Sprintf("활동전AI%d_%d", i+1, rng.Intn(1000))
		}

		character := models.AICharacter{
			Nickname:           nickname,
			Gender:             genders[rng.Intn(2)],
			Age:                rng.Intn(50) + 15, // 15~64세
			Hobby:              models.Hobbies[rng.Intn(len(models.Hobbies))],
			JobCategory:        models.JobCategories[rng.Intn(len(models.JobCategories))],
			MBTI:               models.MBTITypes[rng.Intn(len(models.MBTITypes))],
			AggressionLevel:    rng.Intn(11), // 0~10
			FormalityLevel:     rng.Intn(11),
			RoleplayLevel:      rng.Intn(11),
			AssignedModelIndex: modelIdx,
			IsActive:           true,
		}

		// 인격 요약 생성
		character.PersonaSummary = fmt.Sprintf("%s 출신의 %s 캐릭터입니다. %s인 성격이며 취미는 %s입니다.",
			character.Region, character.JobCategory, character.MBTI, character.Hobby)

		// 생년월일 생성 (나이를 바탕으로 역산)
		currentYear := time.Now().Year()
		birthYear := currentYear - character.Age
		birthMonth := rng.Intn(12) + 1
		birthDay := rng.Intn(28) + 1 // 간단히 1~28일
		character.Birthdate = fmt.Sprintf("%04d-%02d-%02d", birthYear, birthMonth, birthDay)

		// 지역 무작위 선택
		character.Region = models.KoreaRegions[rng.Intn(len(models.KoreaRegions))]

		result, err := db.Exec(`
			INSERT INTO ai_characters (nickname, gender, age, birthdate, region, hobby, job_category, mbti, 
				aggression_level, formality_level, roleplay_level, persona_summary, assigned_model_index, is_active,
				post_count, comment_count)
			VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, 0, 0)`,
			character.Nickname, character.Gender, character.Age, character.Birthdate, character.Region,
			character.Hobby, character.JobCategory, character.MBTI, character.AggressionLevel,
			character.FormalityLevel, character.RoleplayLevel, character.PersonaSummary, character.AssignedModelIndex, character.IsActive,
		)
		if err != nil {
			continue // 중복 등 오류 시 스킵
		}

		id, _ := result.LastInsertId()
		character.ID = int(id)
		character.CreatedAt = time.Now()
		characters = append(characters, character)
	}

	return characters, nil
}

// GetAllCharacters 모든 캐릭터 조회
func (s *CharacterService) GetAllCharacters() ([]models.AICharacter, error) {
	db := s.db.GetDB()
	if db == nil {
		return nil, errors.New("데이터베이스에 연결되지 않았습니다")
	}

	rows, err := db.Query(`
		SELECT id, nickname, gender, age, COALESCE(birthdate, ''), COALESCE(region, ''),
		       hobby, job_category, mbti, 
		       aggression_level, formality_level, roleplay_level, assigned_model_index,
		       COALESCE(post_count, 0), COALESCE(comment_count, 0), COALESCE(persona_summary, ''),
		       is_active, created_at
		FROM ai_characters
		ORDER BY id DESC
	`)
	if err != nil {
		return nil, fmt.Errorf("캐릭터 조회 실패: %w", err)
	}
	defer rows.Close()

	var characters []models.AICharacter
	for rows.Next() {
		var c models.AICharacter
		err := rows.Scan(&c.ID, &c.Nickname, &c.Gender, &c.Age, &c.Birthdate, &c.Region,
			&c.Hobby, &c.JobCategory, &c.MBTI, &c.AggressionLevel, &c.FormalityLevel,
			&c.RoleplayLevel, &c.AssignedModelIndex, &c.PostCount, &c.CommentCount,
			&c.PersonaSummary, &c.IsActive, &c.CreatedAt)
		if err != nil {
			continue
		}
		characters = append(characters, c)
	}

	return characters, nil
}

// GetCharacter 캐릭터 조회
func (s *CharacterService) GetCharacter(id int) (*models.AICharacter, error) {
	db := s.db.GetDB()
	if db == nil {
		return nil, errors.New("데이터베이스에 연결되지 않았습니다")
	}

	c := &models.AICharacter{}
	err := db.QueryRow(`
		SELECT id, nickname, gender, age, COALESCE(birthdate, ''), COALESCE(region, ''),
		       hobby, job_category, mbti,
		       aggression_level, formality_level, roleplay_level, assigned_model_index,
		       COALESCE(post_count, 0), COALESCE(comment_count, 0), COALESCE(persona_summary, ''),
		       is_active, created_at
		FROM ai_characters WHERE id = ?
	`, id).Scan(&c.ID, &c.Nickname, &c.Gender, &c.Age, &c.Birthdate, &c.Region,
		&c.Hobby, &c.JobCategory, &c.MBTI, &c.AggressionLevel, &c.FormalityLevel,
		&c.RoleplayLevel, &c.AssignedModelIndex, &c.PostCount, &c.CommentCount,
		&c.PersonaSummary, &c.IsActive, &c.CreatedAt)

	if err != nil {
		return nil, fmt.Errorf("캐릭터 조회 실패: %w", err)
	}

	return c, nil
}

// GetRandomActiveCharacter 활성화된 랜덤 캐릭터 조회
func (s *CharacterService) GetRandomActiveCharacter() (*models.AICharacter, error) {
	db := s.db.GetDB()
	if db == nil {
		return nil, errors.New("데이터베이스에 연결되지 않았습니다")
	}

	c := &models.AICharacter{}
	err := db.QueryRow(`
		SELECT id, nickname, gender, age, COALESCE(birthdate, ''), COALESCE(region, ''),
		       hobby, job_category, mbti,
		       aggression_level, formality_level, roleplay_level, assigned_model_index,
		       COALESCE(post_count, 0), COALESCE(comment_count, 0), COALESCE(persona_summary, ''),
		       is_active, created_at
		FROM ai_characters WHERE is_active = 1
		ORDER BY RANDOM() LIMIT 1
	`).Scan(&c.ID, &c.Nickname, &c.Gender, &c.Age, &c.Birthdate, &c.Region,
		&c.Hobby, &c.JobCategory, &c.MBTI, &c.AggressionLevel, &c.FormalityLevel,
		&c.RoleplayLevel, &c.AssignedModelIndex, &c.PostCount, &c.CommentCount,
		&c.PersonaSummary, &c.IsActive, &c.CreatedAt)

	if err != nil {
		return nil, fmt.Errorf("랜덤 캐릭터 조회 실패: %w", err)
	}

	return c, nil
}

// UpdateCharacter 캐릭터 정보 수정
func (s *CharacterService) UpdateCharacter(character models.AICharacter) error {
	db := s.db.GetDB()
	if db == nil {
		return errors.New("데이터베이스에 연결되지 않았습니다")
	}

	// assigned_model_index 업데이트 포함
	_, err := db.Exec(`
		UPDATE ai_characters 
		SET nickname = ?, gender = ?, age = ?, birthdate = ?, region = ?,
		    hobby = ?, job_category = ?, mbti = ?, 
		    aggression_level = ?, formality_level = ?, roleplay_level = ?,
		    persona_summary = ?, assigned_model_index = ?,
		    is_active = ?
		WHERE id = ?`,
		character.Nickname, character.Gender, character.Age, character.Birthdate, character.Region,
		character.Hobby, character.JobCategory, character.MBTI,
		character.AggressionLevel, character.FormalityLevel, character.RoleplayLevel,
		character.PersonaSummary, character.AssignedModelIndex,
		character.IsActive, character.ID,
	)

	if err != nil {
		return fmt.Errorf("캐릭터 수정 실패: %w", err)
	}

	return nil
}

// DeleteCharacter 캐릭터 삭제
func (s *CharacterService) DeleteCharacter(id int) error {
	db := s.db.GetDB()
	if db == nil {
		return errors.New("데이터베이스에 연결되지 않았습니다")
	}

	_, err := db.Exec("DELETE FROM ai_characters WHERE id = ?", id)
	if err != nil {
		return fmt.Errorf("캐릭터 삭제 실패: %w", err)
	}

	return nil
}

// GetActiveCharacterCount 활성 캐릭터 수 조회
func (s *CharacterService) GetActiveCharacterCount() (int, error) {
	db := s.db.GetDB()
	if db == nil {
		return 0, errors.New("데이터베이스에 연결되지 않았습니다")
	}

	var count int
	err := db.QueryRow("SELECT COUNT(*) FROM ai_characters WHERE is_active = 1").Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("캐릭터 수 조회 실패: %w", err)
	}

	return count, nil
}

// GetActiveCharacters 활성화된 모든 캐릭터 조회
func (s *CharacterService) GetActiveCharacters() ([]models.AICharacter, error) {
	// 위 GetAllCharacters와 로직 거의 동일 (WHERE is_active = 1)
	// 중복 코드 줄일 수 있지만 명시적으로 작성
	db := s.db.GetDB()
	if db == nil {
		return nil, errors.New("데이터베이스에 연결되지 않았습니다")
	}

	rows, err := db.Query(`
		SELECT id, nickname, gender, age, COALESCE(birthdate, ''), COALESCE(region, ''),
		       hobby, job_category, mbti, 
		       aggression_level, formality_level, roleplay_level, assigned_model_index,
		       COALESCE(post_count, 0), COALESCE(comment_count, 0), COALESCE(persona_summary, ''),
		       is_active, created_at
		FROM ai_characters WHERE is_active = 1
		ORDER BY id DESC
	`)
	if err != nil {
		return nil, fmt.Errorf("캐릭터 조회 실패: %w", err)
	}
	defer rows.Close()

	var characters []models.AICharacter
	for rows.Next() {
		var c models.AICharacter
		err := rows.Scan(&c.ID, &c.Nickname, &c.Gender, &c.Age, &c.Birthdate, &c.Region,
			&c.Hobby, &c.JobCategory, &c.MBTI, &c.AggressionLevel, &c.FormalityLevel,
			&c.RoleplayLevel, &c.AssignedModelIndex, &c.PostCount, &c.CommentCount,
			&c.PersonaSummary, &c.IsActive, &c.CreatedAt)
		if err != nil {
			continue
		}
		characters = append(characters, c)
	}

	return characters, nil
}

// IncrementPostCount AI 캐릭터의 글 작성 횟수 증가
func (s *CharacterService) IncrementPostCount(characterID int) error {
	db := s.db.GetDB()
	if db == nil {
		return errors.New("데이터베이스에 연결되지 않았습니다")
	}

	_, err := db.Exec("UPDATE ai_characters SET post_count = COALESCE(post_count, 0) + 1 WHERE id = ?", characterID)
	return err
}

// IncrementCommentCount AI 캐릭터의 댓글 작성 횟수 증가
func (s *CharacterService) IncrementCommentCount(characterID int) error {
	db := s.db.GetDB()
	if db == nil {
		return errors.New("데이터베이스에 연결되지 않았습니다")
	}

	_, err := db.Exec("UPDATE ai_characters SET comment_count = COALESCE(comment_count, 0) + 1 WHERE id = ?", characterID)
	return err
}

// GetActivityStats AI 캐릭터의 활동 통계 조회
func (s *CharacterService) GetActivityStats(characterID int) (postCount, commentCount int, err error) {
	db := s.db.GetDB()
	if db == nil {
		return 0, 0, errors.New("데이터베이스에 연결되지 않았습니다")
	}

	err = db.QueryRow("SELECT COALESCE(post_count, 0), COALESCE(comment_count, 0) FROM ai_characters WHERE id = ?", characterID).Scan(&postCount, &commentCount)
	return
}

// UpdatePersonaSummary AI 캐릭터의 인격 요약 저장
func (s *CharacterService) UpdatePersonaSummary(characterID int, summary string) error {
	db := s.db.GetDB()
	if db == nil {
		return errors.New("데이터베이스에 연결되지 않았습니다")
	}

	_, err := db.Exec("UPDATE ai_characters SET persona_summary = ? WHERE id = ?", summary, characterID)
	return err
}

// HasPersonaSummary AI 캐릭터의 인격 요약 존재 여부 확인
func (s *CharacterService) HasPersonaSummary(characterID int) (bool, error) {
	db := s.db.GetDB()
	if db == nil {
		return false, errors.New("데이터베이스에 연결되지 않았습니다")
	}

	var summary string
	err := db.QueryRow("SELECT COALESCE(persona_summary, '') FROM ai_characters WHERE id = ?", characterID).Scan(&summary)
	if err != nil {
		return false, err
	}
	return summary != "", nil
}

// GetCharacterWithPendingNickname "활동전AI"로 시작하는 닉네임을 가진 캐릭터 1명 조회
func (s *CharacterService) GetCharacterWithPendingNickname() (*models.AICharacter, error) {
	db := s.db.GetDB()
	if db == nil {
		return nil, errors.New("데이터베이스에 연결되지 않았습니다")
	}

	var c models.AICharacter
	err := db.QueryRow(`
		SELECT id, nickname, gender, age, birthdate, region, hobby, 
		       job_category, mbti, aggression_level, formality_level, roleplay_level, 
		       persona_summary, assigned_model_index, is_active, post_count, comment_count
		FROM ai_characters 
		WHERE nickname LIKE '활동전AI%' AND is_active = 1
		ORDER BY RANDOM()
		LIMIT 1
	`).Scan(
		&c.ID, &c.Nickname, &c.Gender, &c.Age, &c.Birthdate, &c.Region, &c.Hobby,
		&c.JobCategory, &c.MBTI, &c.AggressionLevel, &c.FormalityLevel, &c.RoleplayLevel,
		&c.PersonaSummary, &c.AssignedModelIndex, &c.IsActive, &c.PostCount, &c.CommentCount,
	)
	if err != nil {
		return nil, err // 변경 대상 없음
	}
	return &c, nil
}

// GetCharacterNeedingPersonaUpdate 인격 갱신이 필요한 캐릭터 1명 조회
// 조건: 활동량(글+댓글) >= 3이고, 인격이 없거나 활동량이 3의 배수에 도달한 캐릭터
func (s *CharacterService) GetCharacterNeedingPersonaUpdate() (*models.AICharacter, error) {
	db := s.db.GetDB()
	if db == nil {
		return nil, errors.New("데이터베이스에 연결되지 않았습니다")
	}

	var c models.AICharacter
	// 인격이 없고 활동량 >= 3인 캐릭터 또는
	// 인격이 있고 활동량이 6 이상이며 3의 배수인 캐릭터
	err := db.QueryRow(`
		SELECT id, nickname, gender, age, birthdate, region, hobby, 
		       job_category, mbti, aggression_level, formality_level, roleplay_level, 
		       persona_summary, assigned_model_index, is_active, post_count, comment_count
		FROM ai_characters 
		WHERE is_active = 1 AND (
			(COALESCE(persona_summary, '') = '' AND (post_count + comment_count) >= 3)
			OR
			(COALESCE(persona_summary, '') != '' AND (post_count + comment_count) >= 6 
			 AND (post_count + comment_count) % 3 = 0)
		)
		ORDER BY RANDOM()
		LIMIT 1
	`).Scan(
		&c.ID, &c.Nickname, &c.Gender, &c.Age, &c.Birthdate, &c.Region, &c.Hobby,
		&c.JobCategory, &c.MBTI, &c.AggressionLevel, &c.FormalityLevel, &c.RoleplayLevel,
		&c.PersonaSummary, &c.AssignedModelIndex, &c.IsActive, &c.PostCount, &c.CommentCount,
	)
	if err != nil {
		return nil, err // 갱신 대상 없음
	}
	return &c, nil
}
