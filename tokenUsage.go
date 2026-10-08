package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

type tokenUsageData struct {
	TotalTokens int `json:"total_tokens"`
}

type tokenUsageTracker struct {
	filePath      string
	sessionTokens int
	totalTokens   int
}

func newTokenUsageTracker(filePath string) (*tokenUsageTracker, error) {
	tracker := &tokenUsageTracker{filePath: filePath}
	data, err := os.ReadFile(filePath)
	if err != nil {
		if os.IsNotExist(err) {
			return tracker, nil
		}
		return nil, fmt.Errorf("không thể đọc thống kê token: %w", err)
	}

	var saved tokenUsageData
	if err := json.Unmarshal(data, &saved); err != nil {
		return nil, fmt.Errorf("không thể giải mã thống kê token: %w", err)
	}
	if saved.TotalTokens < 0 {
		return nil, fmt.Errorf("tổng token trong file không hợp lệ")
	}
	tracker.totalTokens = saved.TotalTokens
	return tracker, nil
}

func defaultTokenUsagePath() (string, error) {
	configDir, err := os.UserConfigDir()
	if err != nil {
		return "", fmt.Errorf("không thể tìm thư mục cấu hình người dùng: %w", err)
	}
	return filepath.Join(configDir, applicationID, "token_usage.json"), nil
}

func (tracker *tokenUsageTracker) Add(tokens int) error {
	if tokens < 0 {
		return fmt.Errorf("số token không hợp lệ: %d", tokens)
	}
	tracker.sessionTokens += tokens
	tracker.totalTokens += tokens
	updated := tokenUsageData{TotalTokens: tracker.totalTokens}
	encoded, err := json.MarshalIndent(updated, "", "  ")
	if err != nil {
		return fmt.Errorf("không thể mã hóa thống kê token: %w", err)
	}
	if err := os.MkdirAll(filepath.Dir(tracker.filePath), 0o700); err != nil {
		return fmt.Errorf("không thể tạo thư mục thống kê token: %w", err)
	}
	if err := os.WriteFile(tracker.filePath, encoded, 0o600); err != nil {
		return fmt.Errorf("không thể lưu thống kê token: %w", err)
	}

	return nil
}
