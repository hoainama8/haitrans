package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestTokenUsageTrackerPersistsPerModelAndSession(t *testing.T) {
	filePath := filepath.Join(t.TempDir(), "use_token.json")
	tracker, err := newTokenUsageTracker(filePath)
	if err != nil {
		t.Fatalf("newTokenUsageTracker() error = %v", err)
	}
	if err := tracker.Add("gemini-flash", 120); err != nil {
		t.Fatalf("Add(gemini-flash, 120) error = %v", err)
	}
	if tracker.sessionTokens != 120 || tracker.totalTokens != 120 {
		t.Fatalf("after first Add: session=%d total=%d, want 120/120", tracker.sessionTokens, tracker.totalTokens)
	}
	if err := tracker.Add("gemma", 30); err != nil {
		t.Fatalf("Add(gemma, 30) error = %v", err)
	}

	reloaded, err := newTokenUsageTracker(filePath)
	if err != nil {
		t.Fatalf("reloading tracker error = %v", err)
	}
	if reloaded.sessionTokens != 0 || reloaded.totalTokens != 150 {
		t.Fatalf("reloaded tracker: session=%d total=%d, want 0/150", reloaded.sessionTokens, reloaded.totalTokens)
	}
	if err := reloaded.Add("gemini-flash", 20); err != nil {
		t.Fatalf("Add(gemini-flash, 20) error = %v", err)
	}

	data, err := os.ReadFile(filePath)
	if err != nil {
		t.Fatalf("ReadFile() error = %v", err)
	}
	var saved tokenUsageData
	if err := json.Unmarshal(data, &saved); err != nil {
		t.Fatalf("Unmarshal() error = %v", err)
	}
	if saved.TotalTokens != 170 {
		t.Errorf("saved total_tokens = %d, want 170", saved.TotalTokens)
	}
	if got := saved.Sessions[tracker.sessionID].ModelTokens["gemini-flash"]; got != 120 {
		t.Errorf("first session gemini-flash tokens = %d, want 120", got)
	}
	if got := saved.Sessions[tracker.sessionID].ModelTokens["gemma"]; got != 30 {
		t.Errorf("first session gemma tokens = %d, want 30", got)
	}
	if got := saved.Sessions[reloaded.sessionID].ModelTokens["gemini-flash"]; got != 20 {
		t.Errorf("second session gemini-flash tokens = %d, want 20", got)
	}
}

func TestEnsureApplicationDirectoriesCreatesExpectedFiles(t *testing.T) {
	directories, err := ensureApplicationDirectories(t.TempDir())
	if err != nil {
		t.Fatalf("ensureApplicationDirectories() error = %v", err)
	}
	for _, directory := range []string{directories.root, directories.input, directories.output} {
		info, err := os.Stat(directory)
		if err != nil {
			t.Errorf("expected directory %s: %v", directory, err)
		} else if !info.IsDir() {
			t.Errorf("%s is not a directory", directory)
		}
	}
	if _, err := os.Stat(directories.tokenUsage); err != nil {
		t.Errorf("expected token usage file: %v", err)
	}
}

func TestSaveUploadedInput(t *testing.T) {
	inputDir := filepath.Join(t.TempDir(), "input")
	filePath, err := saveUploadedInput(inputDir, "../document.txt", []byte("uploaded"))
	if err != nil {
		t.Fatalf("saveUploadedInput() error = %v", err)
	}
	if filepath.Dir(filePath) != inputDir {
		t.Fatalf("saved file outside input directory: %s", filePath)
	}
	data, err := os.ReadFile(filePath)
	if err != nil {
		t.Fatalf("ReadFile() error = %v", err)
	}
	if string(data) != "uploaded" {
		t.Errorf("saved content = %q, want uploaded", data)
	}
}
