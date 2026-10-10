package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"google.golang.org/genai"
	"google.golang.org/genai/interactions/models/interactions"
	"google.golang.org/genai/interactions/models/operations"
)

func TestTrans(req translationRequest) (*InteractionsJSON, error) {
	ctx := context.Background()
	if strings.TrimSpace(req.APIKey) == "" {
		return nil, errors.New("vui lòng nhập Gemini API key")
	}
	if len(req.Document) == 0 {
		return nil, errors.New("vui lòng đính kèm tài liệu")
	}

	client, err := genai.NewClient(ctx, &genai.ClientConfig{
		APIKey:  req.APIKey,
		Backend: genai.BackendGeminiAPI,
	})
	if err != nil {
		return nil, fmt.Errorf("không thể khởi tạo Gemini client: %w", err)
	}

	prompt := BuildTextPrompt(req)
	fileType := req.FileType
	if fileType == "" {
		fileType = mimeTypeForFile(req.FileName)
	}
	localPath, err := saveDocumentTempFile(req.FileName, req.Document)
	if err != nil {
		return nil, fmt.Errorf("không thể lưu tệp đính kèm tạm thời: %w", err)
	}
	defer os.Remove(localPath)

	uploadedFile, err := client.Files.UploadFromPath(ctx, localPath, &genai.UploadFileConfig{
		MIMEType:    fileType,
		DisplayName: safeStoredName(req.FileName),
	})
	if err != nil {
		return nil, fmt.Errorf("không thể tải tệp lên Gemini: %w", err)
	}
	if uploadedFile == nil || uploadedFile.URI == "" {
		return nil, errors.New("Gemini không trả về URI của tệp đã tải lên")
	}

	input := interactions.NewInteractionsInput([]interactions.Content{
		interactions.NewContent(interactions.DocumentContent{
			URI:      genai.Ptr(uploadedFile.URI),
			MimeType: interactions.DocumentContentMimeType(fileType).ToPointer(),
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
		return nil, fmt.Errorf("lỗi gửi yêu cầu interaction: %w", err)
	}
	if res == nil || res.Interaction == nil {
		return nil, errors.New("Gemini không trả về interaction")
	}

	interactionJSON := NewInteractionsJSON(res.Interaction)
	if outputText := interactionJSON.GetOutputText(); outputText != nil {
		fmt.Println(*outputText)
		return interactionJSON, nil
	}
	return nil, errors.New("Gemini không trả về nội dung văn bản trong interaction")
}

func saveDocumentTempFile(fileName string, data []byte) (string, error) {
	ext := filepath.Ext(safeStoredName(fileName))
	file, err := os.CreateTemp("", "haitrans-*"+ext)
	if err != nil {
		return "", err
	}
	path := file.Name()

	if _, err := file.Write(data); err != nil {
		file.Close()
		os.Remove(path)
		return "", err
	}
	if err := file.Close(); err != nil {
		os.Remove(path)
		return "", err
	}
	return path, nil
}
