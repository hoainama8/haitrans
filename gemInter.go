package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"google.golang.org/genai"
	"google.golang.org/genai/interactions/models/interactions"
	"google.golang.org/genai/interactions/models/operations"
)

func Chat(apikey string, text string) {
	ctx := context.Background()
	//apiKey := strings.TrimSpace(os.Getenv("GEMINI_API_KEY"))
	if apikey == "" {
		log.Fatal("GEMINI_API_KEY is not set; export your Gemini API key before running")
	}
	client, err := genai.NewClient(ctx, &genai.ClientConfig{
		APIKey:  apikey,
		Backend: genai.BackendGeminiAPI,
	})

	if err != nil {
		log.Fatal(err)
	}
	res, err := client.Interactions.Create(ctx, operations.CreateInteractionRequest{
		Body: operations.NewCreateInteractionRequestBody(interactions.CreateModelInteraction{
			Model: interactions.Model("gemini-3.8-flash"),
			Input: func() *interactions.InteractionsInput {
				input := interactions.NewInteractionsInput(text)
				return &input
			}(),
		}),
	})
	if err != nil {
		log.Fatal(err)
	}
	interactionJSON := NewInteractionsJSON(res.Interaction)

	formattedInteraction, err := json.MarshalIndent(interactionJSON, "", "  ")
	if err != nil {
		log.Printf("Không thể chuyển Interaction thành JSON: %v", err)
		return
	}
	fmt.Println(string(formattedInteraction))
	if outputText := interactionJSON.GetOutputText(); outputText != nil {
		fmt.Println("2")
		fmt.Println(*outputText)
	}

}
