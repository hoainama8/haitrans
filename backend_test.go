package main

import (
	"os"
	"strings"
	"testing"
)

func TestBuildTextPromptRequestsDocumentOCRAndTranslation(t *testing.T) {
	request := translationRequest{
		SourceLang: "Tiếng Anh",
		TargetLang: "Tiếng Việt",
		Text:       "Không được đưa nội dung nhị phân hoặc văn bản này vào prompt.",
	}

	prompt := BuildTextPrompt(request)
	for _, expected := range []string{
		"đọc toàn bộ tài liệu đính kèm",
		"PDF scan hoặc PDF lai",
		"nhận dạng chữ trong ảnh (OCR)",
		"Tiếng Anh sang Tiếng Việt",
	} {
		if !strings.Contains(prompt, expected) {
			t.Errorf("BuildTextPrompt() không chứa %q", expected)
		}
	}
	if strings.Contains(prompt, request.Text) {
		t.Error("BuildTextPrompt() không được thêm văn bản đầu vào vào prompt tài liệu")
	}
}

func TestSaveDocumentTempFile(t *testing.T) {
	want := []byte("%PDF-1.7 test content")
	path, err := saveDocumentTempFile("../exam.pdf", want)
	if err != nil {
		t.Fatalf("saveDocumentTempFile() error = %v", err)
	}
	defer os.Remove(path)

	if got := path[len(path)-len(".pdf"):]; got != ".pdf" {
		t.Errorf("tệp tạm có phần mở rộng %q, muốn .pdf", got)
	}
	got, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("os.ReadFile() error = %v", err)
	}
	if string(got) != string(want) {
		t.Errorf("nội dung tệp tạm = %q, muốn %q", got, want)
	}
}
