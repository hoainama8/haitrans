package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestTokenUsageTrackerPersistsTotalAndResetsSession(t *testing.T) {
	filePath := filepath.Join(t.TempDir(), "token_usage.json")
	tracker, err := newTokenUsageTracker(filePath)
	if err != nil {
		t.Fatalf("newTokenUsageTracker() error = %v", err)
	}
	if err := tracker.Add(120); err != nil {
		t.Fatalf("Add(120) error = %v", err)
	}
	if tracker.sessionTokens != 120 || tracker.totalTokens != 120 {
		t.Fatalf("after first Add: session=%d total=%d, want 120/120", tracker.sessionTokens, tracker.totalTokens)
	}

	reloaded, err := newTokenUsageTracker(filePath)
	if err != nil {
		t.Fatalf("reloading tracker error = %v", err)
	}
	if reloaded.sessionTokens != 0 || reloaded.totalTokens != 120 {
		t.Fatalf("reloaded tracker: session=%d total=%d, want 0/120", reloaded.sessionTokens, reloaded.totalTokens)
	}
	if err := reloaded.Add(30); err != nil {
		t.Fatalf("Add(30) error = %v", err)
	}

	data, err := os.ReadFile(filePath)
	if err != nil {
		t.Fatalf("ReadFile() error = %v", err)
	}
	var saved tokenUsageData
	if err := json.Unmarshal(data, &saved); err != nil {
		t.Fatalf("Unmarshal() error = %v", err)
	}
	if saved.TotalTokens != 150 {
		t.Errorf("saved total_tokens = %d, want 150", saved.TotalTokens)
	}
}
