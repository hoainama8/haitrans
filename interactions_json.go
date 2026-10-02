package main

import (
	"strings"

	"google.golang.org/genai/interactions/models/interactions"
)

// InteractionsJSON is a JSON-friendly view of the complete SDK Interaction.
type InteractionsJSON interactions.Interaction

// NewInteractionsJSON copies an SDK Interaction into InteractionsJSON.
func NewInteractionsJSON(interaction *interactions.Interaction) *InteractionsJSON {
	if interaction == nil {
		return nil
	}
	converted := InteractionsJSON(*interaction)
	return &converted
}

// MarshalJSON preserves the SDK's custom serialization for interaction unions.
func (interaction InteractionsJSON) MarshalJSON() ([]byte, error) {
	return interactions.Interaction(interaction).MarshalJSON()
}

func (interaction *InteractionsJSON) GetAgent() *interactions.AgentOption {
	return (*interactions.Interaction)(interaction).GetAgent()
}

func (interaction *InteractionsJSON) GetAgentConfig() *interactions.InteractionAgentConfig {
	return (*interactions.Interaction)(interaction).GetAgentConfig()
}

func (interaction *InteractionsJSON) GetCachedContent() *string {
	return (*interactions.Interaction)(interaction).GetCachedContent()
}

func (interaction *InteractionsJSON) GetCreated() *string {
	return (*interactions.Interaction)(interaction).GetCreated()
}

func (interaction *InteractionsJSON) GetEnvironment() *interactions.InteractionEnvironment {
	return (*interactions.Interaction)(interaction).GetEnvironment()
}

func (interaction *InteractionsJSON) GetEnvironmentID() *string {
	return (*interactions.Interaction)(interaction).GetEnvironmentID()
}

func (interaction *InteractionsJSON) GetErrors() []interactions.Error {
	return (*interactions.Interaction)(interaction).GetErrors()
}

func (interaction *InteractionsJSON) GetGenerationConfig() *interactions.GenerationConfig {
	return (*interactions.Interaction)(interaction).GetGenerationConfig()
}

func (interaction *InteractionsJSON) GetID() *string {
	if interaction == nil {
		return nil
	}
	return interaction.ID
}

func (interaction *InteractionsJSON) GetInput() *interactions.InteractionsInput {
	return (*interactions.Interaction)(interaction).GetInput()
}

func (interaction *InteractionsJSON) GetLabels() map[string]string {
	return (*interactions.Interaction)(interaction).GetLabels()
}

func (interaction *InteractionsJSON) GetModel() *interactions.Model {
	return (*interactions.Interaction)(interaction).GetModel()
}

func (interaction *InteractionsJSON) GetOutputAudio() *interactions.AudioContent {
	return (*interactions.Interaction)(interaction).GetOutputAudio()
}

func (interaction *InteractionsJSON) GetOutputImage() *interactions.ImageContent {
	return (*interactions.Interaction)(interaction).GetOutputImage()
}

func (interaction *InteractionsJSON) GetOutputText() *string {
	if interaction == nil {
		return nil
	}
	if interaction.OutputText != nil {
		return interaction.OutputText
	}

	var output strings.Builder
	for _, step := range interaction.Steps {
		if step.ModelOutputStep == nil {
			continue
		}
		for _, content := range step.ModelOutputStep.Content {
			if content.TextContent != nil {
				output.WriteString(content.TextContent.Text)
			}
		}
	}
	if output.Len() == 0 {
		return nil
	}

	text := output.String()
	return &text
}

func (interaction *InteractionsJSON) GetOutputVideo() *interactions.VideoContent {
	return (*interactions.Interaction)(interaction).GetOutputVideo()
}

func (interaction *InteractionsJSON) GetPreviousInteractionID() *string {
	return (*interactions.Interaction)(interaction).GetPreviousInteractionID()
}

func (interaction *InteractionsJSON) GetResponseFormat() *interactions.InteractionResponseFormat {
	return (*interactions.Interaction)(interaction).GetResponseFormat()
}

func (interaction *InteractionsJSON) GetResponseMimeType() *string {
	return (*interactions.Interaction)(interaction).GetResponseMimeType()
}

func (interaction *InteractionsJSON) GetResponseModalities() []interactions.ResponseModality {
	return (*interactions.Interaction)(interaction).GetResponseModalities()
}

func (interaction *InteractionsJSON) GetSafetySettings() []interactions.SafetySetting {
	return (*interactions.Interaction)(interaction).GetSafetySettings()
}

func (interaction *InteractionsJSON) GetServiceTier() *interactions.ServiceTier {
	return (*interactions.Interaction)(interaction).GetServiceTier()
}

func (interaction *InteractionsJSON) GetStatus() interactions.InteractionStatus {
	return (*interactions.Interaction)(interaction).GetStatus()
}

func (interaction *InteractionsJSON) GetSteps() []interactions.Step {
	return (*interactions.Interaction)(interaction).GetSteps()
}

func (interaction *InteractionsJSON) GetSystemInstruction() *string {
	return (*interactions.Interaction)(interaction).GetSystemInstruction()
}

func (interaction *InteractionsJSON) GetTools() []interactions.Tool {
	return (*interactions.Interaction)(interaction).GetTools()
}

func (interaction *InteractionsJSON) GetUpdated() *string {
	return (*interactions.Interaction)(interaction).GetUpdated()
}

func (interaction *InteractionsJSON) GetUsage() *interactions.Usage {
	return (*interactions.Interaction)(interaction).GetUsage()
}

func (interaction *InteractionsJSON) GetWebhookConfig() *interactions.WebhookConfig {
	return (*interactions.Interaction)(interaction).GetWebhookConfig()
}
