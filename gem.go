package main

import (
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"strings"

	"google.golang.org/genai"
	"google.golang.org/genai/interactions/models/interactions"
	"google.golang.org/genai/interactions/models/operations"
)

func TestTrans(req translationRequest) (translationResponse, error) {
	ctx := context.Background()
	if strings.TrimSpace(req.APIKey) == "" {
		return translationResponse{}, errors.New("vui lòng nhập Gemini API key")
	}
	if len(req.Document) == 0 {
		return translationResponse{}, errors.New("vui lòng đính kèm tài liệu")
	}

	client, err := genai.NewClient(ctx, &genai.ClientConfig{
		APIKey:  req.APIKey,
		Backend: genai.BackendGeminiAPI,
	})
	if err != nil {
		return translationResponse{}, fmt.Errorf("không thể khởi tạo Gemini client: %w", err)
	}

	sourceLang := req.SourceLang
	if strings.TrimSpace(sourceLang) == "" {
		sourceLang = "tự động phát hiện"
	}
	prompt := fmt.Sprintf("Hãy dịch toàn bộ nội dung của tài liệu này từ %s sang %s. Giữ nguyên cấu trúc văn bản và định dạng nếu có. Chỉ trả về bản dịch.", sourceLang, req.TargetLang)
	base64Document := base64.StdEncoding.EncodeToString(req.Document)
	input := interactions.NewInteractionsInput([]interactions.Content{
		interactions.NewContent(interactions.DocumentContent{
			Data:     genai.Ptr(base64Document),
			MimeType: interactions.DocumentContentMimeTypeApplicationPdf.ToPointer(),
		}),
		interactions.NewContent(interactions.TextContent{
			Text: prompt,
		}),
	})
	res, err := client.Interactions.Create(ctx, operations.CreateInteractionRequest{
		Body: operations.NewCreateInteractionRequestBody(interactions.CreateModelInteraction{
			Model: interactions.Model(req.Model),
			Input: &input,
		}),
	})
	if err != nil {
		return translationResponse{}, fmt.Errorf("lỗi gửi yêu cầu interaction: %w", err)
	}

	interactionJSON := NewInteractionsJSON(res.Interaction)
	if outputText := interactionJSON.GetOutputText(); outputText != nil {
		return translationResponse{Text: *outputText}, nil
	}
	return translationResponse{}, errors.New("Gemini không trả về nội dung văn bản trong interaction")
}
