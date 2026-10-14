package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"google.golang.org/genai/interactions/models/interactions"
)

func TestNewInteractionsJSON(t *testing.T) {
	id := "interaction-123"
	model := interactions.Model("gemini-test")
	status := interactions.InteractionStatus("completed")
	outputText := "hello"
	source := &interactions.Interaction{
		ID:         &id,
		Model:      &model,
		Status:     status,
		OutputText: &outputText,
	}

	converted := NewInteractionsJSON(source)
	if converted == nil {
		t.Fatal("NewInteractionsJSON() returned nil")
	}
	if got := *converted.GetID(); got != id {
		t.Errorf("GetID() = %q, want %q", got, id)
	}
	if got := *converted.GetModel(); string(got) != string(model) {
		t.Errorf("GetModel() = %q, want %q", got, model)
	}
	if got := converted.GetStatus(); got != status {
		t.Errorf("GetStatus() = %q, want %q", got, status)
	}
	if got := *converted.GetOutputText(); got != outputText {
		t.Errorf("GetOutputText() = %q, want %q", got, outputText)
	}

	encoded, err := json.MarshalIndent(converted, "", "  ")
	if err != nil {
		t.Fatalf("MarshalIndent() error = %v", err)
	}
	for _, expected := range []string{`"id": "interaction-123"`, `"model": "gemini-test"`, `"output_text": "hello"`} {
		if !strings.Contains(string(encoded), expected) {
			t.Errorf("JSON does not contain %q: %s", expected, encoded)
		}
	}
}

func TestNewInteractionsJSONNil(t *testing.T) {
	if converted := NewInteractionsJSON(nil); converted != nil {
		t.Fatalf("NewInteractionsJSON(nil) = %v, want nil", converted)
	}
}

func TestInteractionsJSONGettersNilReceiver(t *testing.T) {
	var interaction *InteractionsJSON
	if got := interaction.GetID(); got != nil {
		t.Errorf("GetID() = %v, want nil", got)
	}
	if got := interaction.GetOutputText(); got != nil {
		t.Errorf("GetOutputText() = %v, want nil", got)
	}
}

func TestInteractionsJSONGetOutputTextFromModelOutputStep(t *testing.T) {
	source := &interactions.Interaction{
		Steps: []interactions.Step{
			{
				ModelOutputStep: &interactions.ModelOutputStep{
					Content: []interactions.Content{
						{TextContent: &interactions.TextContent{Text: "answer from model output"}},
					},
				},
			},
		},
	}

	got := NewInteractionsJSON(source).GetOutputText()
	if got == nil || *got != "answer from model output" {
		t.Fatalf("GetOutputText() = %v, want %q", got, "answer from model output")
	}
}

func TestInteractionsJSONSaveUsageUsesResponseModel(t *testing.T) {
	filePath := filepath.Join(t.TempDir(), "use_token.json")
	tracker, err := newTokenUsageTracker(filePath)
	if err != nil {
		t.Fatalf("newTokenUsageTracker() error = %v", err)
	}

	model := interactions.Model("gemini-test")
	totalTokens := 42
	totalInputTokens := 25
	totalOutputTokens := 10
	totalThoughtTokens := 7
	interaction := &InteractionsJSON{
		Model: &model,
		Usage: &interactions.Usage{
			TotalInputTokens:   &totalInputTokens,
			TotalOutputTokens:  &totalOutputTokens,
			TotalThoughtTokens: &totalThoughtTokens,
			TotalTokens:        &totalTokens,
		},
	}
	if err := interaction.SaveUsage(tracker); err != nil {
		t.Fatalf("SaveUsage() error = %v", err)
	}

	data, err := os.ReadFile(filePath)
	if err != nil {
		t.Fatalf("ReadFile() error = %v", err)
	}
	var saved tokenUsageData
	if err := json.Unmarshal(data, &saved); err != nil {
		t.Fatalf("Unmarshal() error = %v", err)
	}
	savedSession := findTokenUsageSession(saved["gemini-test"], tracker.sessionID)
	if got := savedSession.TotalTokens; got != totalTokens {
		t.Errorf("saved model tokens = %d, want %d", got, totalTokens)
	}
	if savedSession.SessionID != tracker.sessionID || savedSession.TotalTokens != totalTokens {
		t.Errorf("saved session = %+v", savedSession)
	}
	if savedSession.InputTokens != totalInputTokens || savedSession.OutputTokens != totalOutputTokens || savedSession.ThoughtTokens != totalThoughtTokens {
		t.Errorf("saved usage breakdown = %+v", savedSession)
	}
	for _, key := range []string{`"gemini-test": [`, `"sesionID":`, `"started_at":`, `"total_cached_tokens":`, `"total_input_tokens":`, `"total_output_tokens":`, `"total_thought_tokens":`, `"total_tokens":`, `"total_tool_use_tokens":`} {
		if !strings.Contains(string(data), key) {
			t.Errorf("saved JSON does not contain %q: %s", key, data)
		}
	}
}
