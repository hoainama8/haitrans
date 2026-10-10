package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"time"
)

type tokenUsageData struct {
	TotalTokens int                          `json:"total_tokens"`
	Sessions    map[string]tokenUsageSession `json:"sessions"`
}

type tokenUsageSession struct {
	StartedAt   time.Time      `json:"started_at"`
	TotalTokens int            `json:"total_tokens"`
	ModelTokens map[string]int `json:"model_tokens"`
}

type tokenUsageTracker struct {
	filePath          string
	sessionID         string
	sessionTokens     int
	sessionModelToken map[string]int
	totalTokens       int
}

func newTokenUsageTracker(filePath string) (*tokenUsageTracker, error) {
	tracker := &tokenUsageTracker{
		filePath:          filePath,
		sessionID:         strconv.FormatInt(time.Now().UnixNano(), 10),
		sessionModelToken: make(map[string]int),
	}
	data, err := os.ReadFile(filePath)
	if err != nil {
		if os.IsNotExist(err) {
			if err := tracker.persist(tokenUsageData{Sessions: make(map[string]tokenUsageSession)}); err != nil {
				return nil, err
			}
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
	for _, session := range saved.Sessions {
		if session.TotalTokens < 0 {
			return nil, fmt.Errorf("số token trong phiên không hợp lệ")
		}
		for _, tokens := range session.ModelTokens {
			if tokens < 0 {
				return nil, fmt.Errorf("số token theo mô hình không hợp lệ")
			}
		}
	}
	return tracker, nil
}

func defaultTokenUsagePath() (string, error) {
	directories, err := defaultApplicationDirectories()
	if err != nil {
		return "", err
	}
	return directories.tokenUsage, nil
}

func (tracker *tokenUsageTracker) Add(model string, tokens int) error {
	if model == "" {
		return fmt.Errorf("tên mô hình không được để trống")
	}
	if tokens < 0 {
		return fmt.Errorf("số token không hợp lệ: %d", tokens)
	}
	modelTokens := tracker.sessionModelToken[model] + tokens
	sessionTokens := tracker.sessionTokens + tokens
	totalTokens := tracker.totalTokens + tokens
	data, err := os.ReadFile(tracker.filePath)
	if err != nil {
		return fmt.Errorf("không thể đọc thống kê token: %w", err)
	}
	var updated tokenUsageData
	if err := json.Unmarshal(data, &updated); err != nil {
		return fmt.Errorf("không thể giải mã thống kê token: %w", err)
	}
	if updated.Sessions == nil {
		updated.Sessions = make(map[string]tokenUsageSession)
	}
	session := updated.Sessions[tracker.sessionID]
	if session.ModelTokens == nil {
		session.StartedAt = time.Now().UTC()
		session.ModelTokens = make(map[string]int)
	}
	session.ModelTokens[model] = modelTokens
	session.TotalTokens = sessionTokens
	updated.Sessions[tracker.sessionID] = session
	updated.TotalTokens = totalTokens
	if err := tracker.persist(updated); err != nil {
		return err
	}

	tracker.sessionModelToken[model] = modelTokens
	tracker.sessionTokens = sessionTokens
	tracker.totalTokens = totalTokens
	return nil
}

func (tracker *tokenUsageTracker) persist(updated tokenUsageData) error {
	if updated.Sessions == nil {
		updated.Sessions = make(map[string]tokenUsageSession)
	}
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
