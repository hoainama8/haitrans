package main

import (
	"encoding/json"
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
