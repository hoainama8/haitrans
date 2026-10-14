package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"google.golang.org/genai/interactions/models/interactions"
)

type tokenUsageData map[string][]tokenUsageSession

type tokenUsageSession struct {
	SessionID     string    `json:"sesionID"`
	StartedAt     time.Time `json:"started_at"`
	CachedTokens  int       `json:"total_cached_tokens"`
	InputTokens   int       `json:"total_input_tokens"`
	OutputTokens  int       `json:"total_output_tokens"`
	ThoughtTokens int       `json:"total_thought_tokens"`
	TotalTokens   int       `json:"total_tokens"`
	ToolUseTokens int       `json:"total_tool_use_tokens"`
}

type legacyTokenUsageData struct {
	TotalTokens int                                `json:"total_tokens"`
	Sessions    map[string]legacyTokenUsageSession `json:"sessions"`
}

type legacyTokenUsageSession struct {
	StartedAt   time.Time                     `json:"started_at"`
	TotalTokens int                           `json:"total_tokens"`
	ModelTokens map[string]int                `json:"model_tokens"`
	ModelUsage  map[string]interactions.Usage `json:"model_usage"`
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
			if err := tracker.persist(make(tokenUsageData)); err != nil {
				return nil, err
			}
			return tracker, nil
		}
		return nil, fmt.Errorf("không thể đọc thống kê token: %w", err)
	}

	saved, err := decodeTokenUsageData(data)
	if err != nil {
		return nil, err
	}
	for _, sessions := range saved {
		for _, session := range sessions {
			if session.TotalTokens < 0 {
				return nil, fmt.Errorf("số token trong phiên không hợp lệ")
			}
			tracker.totalTokens += session.TotalTokens
		}
	}
	if err := tracker.persist(saved); err != nil {
		return nil, err
	}
	return tracker, nil
}

func decodeTokenUsageData(data []byte) (tokenUsageData, error) {
	var saved tokenUsageData
	if err := json.Unmarshal(data, &saved); err == nil {
		if saved == nil {
			saved = make(tokenUsageData)
		}
		return saved, nil
	}

	var legacy legacyTokenUsageData
	if err := json.Unmarshal(data, &legacy); err != nil {
		return nil, fmt.Errorf("không thể giải mã thống kê token: %w", err)
	}
	converted := make(tokenUsageData)
	for sessionID, session := range legacy.Sessions {
		for model, tokens := range session.ModelTokens {
			usage := session.ModelUsage[model]
			converted[model] = append(converted[model], tokenUsageSession{
				SessionID:     sessionID,
				StartedAt:     session.StartedAt,
				CachedTokens:  valueOrZero(usage.GetTotalCachedTokens()),
				InputTokens:   valueOrZero(usage.GetTotalInputTokens()),
				OutputTokens:  valueOrZero(usage.GetTotalOutputTokens()),
				ThoughtTokens: valueOrZero(usage.GetTotalThoughtTokens()),
				TotalTokens:   tokens,
				ToolUseTokens: valueOrZero(usage.GetTotalToolUseTokens()),
			})
		}
	}
	return converted, nil
}

func valueOrZero(value *int) int {
	if value == nil {
		return 0
	}
	return *value
}

func defaultTokenUsagePath() (string, error) {
	directories, err := defaultApplicationDirectories()
	if err != nil {
		return "", err
	}
	return directories.tokenUsage, nil
}

func (tracker *tokenUsageTracker) Add(model string, tokens int) error {
	return tracker.add(model, tokens, nil)
}

func (tracker *tokenUsageTracker) AddUsage(model string, usage *interactions.Usage) error {
	if usage == nil || usage.GetTotalTokens() == nil {
		return fmt.Errorf("thông tin tổng token không được để trống")
	}
	return tracker.add(model, *usage.GetTotalTokens(), usage)
}

func (tracker *tokenUsageTracker) add(model string, tokens int, usage *interactions.Usage) error {
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
	updated, err := decodeTokenUsageData(data)
	if err != nil {
		return err
	}
	sessions := updated[model]
	sessionIndex := -1
	for index := range sessions {
		if sessions[index].SessionID == tracker.sessionID {
			sessionIndex = index
			break
		}
	}
	if sessionIndex < 0 {
		sessions = append(sessions, tokenUsageSession{
			SessionID: tracker.sessionID,
			StartedAt: time.Now().UTC(),
		})
		sessionIndex = len(sessions) - 1
	}
	session := &sessions[sessionIndex]
	session.TotalTokens += tokens
	if usage != nil {
		session.CachedTokens += valueOrZero(usage.GetTotalCachedTokens())
		session.InputTokens += valueOrZero(usage.GetTotalInputTokens())
		session.OutputTokens += valueOrZero(usage.GetTotalOutputTokens())
		session.ThoughtTokens += valueOrZero(usage.GetTotalThoughtTokens())
		session.ToolUseTokens += valueOrZero(usage.GetTotalToolUseTokens())
	}
	updated[model] = sessions
	if err := tracker.persist(updated); err != nil {
		return err
	}

	tracker.sessionModelToken[model] = modelTokens
	tracker.sessionTokens = sessionTokens
	tracker.totalTokens = totalTokens
	return nil
}

func (tracker *tokenUsageTracker) persist(updated tokenUsageData) error {
	if updated == nil {
		updated = make(tokenUsageData)
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
