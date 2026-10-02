package main

import (
	"context"
	"encoding/base64"
	"fmt"
	"log"
	"os"

	"google.golang.org/genai"
	"google.golang.org/genai/interactions/models/interactions"
	"google.golang.org/genai/interactions/models/operations"
)

func TestTrans(apikey string) {
	ctx := context.Background()
	// 1. Khởi tạo GenAI Client (đảm bảo đã cài đặt GEMINI_API_KEY trong biến môi trường)
	client, err := genai.NewClient(ctx, &genai.ClientConfig{
		APIKey:  apikey,
		Backend: genai.BackendGeminiAPI,
	})
	if err != nil {
		log.Fatal("Lỗi khởi tạo client:", err)
	}

	// 2. Đọc file tài liệu (PDF) và mã hóa sang Base64
	pdfPath := "exam.pdf" // Thay đường dẫn tới file tài liệu của bạn
	pdfBytes, err := os.ReadFile(pdfPath)
	if err != nil {
		log.Fatal("Lỗi đọc file:", err)
	}
	base64Pdf := base64.StdEncoding.EncodeToString(pdfBytes)
	input := interactions.NewInteractionsInput([]interactions.Content{
		// Nạp tài liệu PDF
		interactions.NewContent(interactions.DocumentContent{
			Data:     genai.Ptr(base64Pdf),
			MimeType: interactions.DocumentContentMimeTypeApplicationPdf.ToPointer(),
		}),
		// Prompt yêu cầu dịch thuật
		interactions.NewContent(interactions.TextContent{
			Text: "Hãy dịch toàn bộ nội dung của tài liệu này sang Tiếng Anh. Giữ nguyên cấu trúc văn bản và định dạng nếu có.",
		}),
	})
	// 3. Gửi yêu cầu dịch tài liệu tới Gemini API
	res, err := client.Interactions.Create(ctx, operations.CreateInteractionRequest{
		Body: operations.NewCreateInteractionRequestBody(interactions.CreateModelInteraction{
			Model: interactions.Model("gemini-3.7-flash"),
			Input: &input,
		}),
	})
	if err != nil {
		fmt.Println(err)
		log.Fatal("Lỗi gửi yêu cầu interaction:", err)
	}

	// 4. In kết quả tài liệu đã được dịch
	interactionJSON := NewInteractionsJSON(res.Interaction)
	if outputText := interactionJSON.GetOutputText(); outputText != nil {
		fmt.Println("=== NỘI DUNG TÀI LIỆU SAU KHI DỊCH ===")
		fmt.Println(*outputText)
	} else {
		log.Println("Gemini không trả về nội dung văn bản trong interaction")
	}
}
