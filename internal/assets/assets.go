// Created by DINKIssTyle on 2025. Copyright (C) 2025 DINKI'ssTyle. All rights reserved.

package assets

import (
	"embed"
	"fmt"
	"io/fs"
	"path"
	"strings"
)

//go:embed avarta
var avatarFS embed.FS

// GetAvatarImages 성별에 따른 아바타 이미지 파일명 목록 반환
// gender: "male" or "female" (or "남성", "여성")
func GetAvatarImages(gender string) ([]string, error) {
	var dirName string
	if gender == "male" || gender == "남성" {
		dirName = "male"
	} else if gender == "female" || gender == "여성" {
		dirName = "female"
	} else {
		return []string{}, nil
	}

	dirPath := path.Join("avarta", dirName)
	entries, err := avatarFS.ReadDir(dirPath)
	if err != nil {
		fmt.Printf("[DEBUG] Failed to read dir %s: %v\n", dirPath, err)
		return nil, err
	}

	var images []string
	for _, entry := range entries {
		if !entry.IsDir() {
			name := entry.Name()
			lowerName := strings.ToLower(name)
			if strings.HasSuffix(lowerName, ".png") || strings.HasSuffix(lowerName, ".jpg") || strings.HasSuffix(lowerName, ".jpeg") || strings.HasSuffix(lowerName, ".gif") || strings.HasSuffix(lowerName, ".webp") {
				images = append(images, name)
			}
		}
	}
	fmt.Printf("[DEBUG] Loaded %d avatar images from %s\n", len(images), dirPath)
	return images, nil
}

// GetAvatarFS 아바타 파일 시스템 반환
func GetAvatarFS() fs.FS {
	return avatarFS
}
