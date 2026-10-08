package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/dialog"
)

func showSavedTranslation(win fyne.Window, req translationRequest, translatedText string) {
	textFile, err := saveTranslationText(outputDirectory, req.FileName, req.TargetLang, translatedText)
	if err != nil {
		dialog.ShowError(fmt.Errorf("Không thể lưu tệp dịch: %v", err), win)
		return
	}

	message := fmt.Sprintf("Đã lưu văn bản dịch: %s", textFile)
	dialog.ShowInformation("Thành công", message, win)

}

func saveTranslationText(outputDir, sourceName, targetLang, translatedText string) (string, error) {
	sourceBase := strings.TrimSuffix(safeStoredName(sourceName), filepath.Ext(sourceName))
	if sourceBase == "document" && strings.TrimSpace(sourceName) == "" {
		sourceBase = "translation"
	}
	language := safeStoredName(strings.NewReplacer("/", "-", "\\", "-").Replace(targetLang))
	filename := fmt.Sprintf("%s - %s.txt", sourceBase, language)
	if err := os.MkdirAll(outputDir, 0o700); err != nil {
		return "", fmt.Errorf("không thể tạo thư mục đầu ra: %w", err)
	}
	path := filepath.Join(outputDir, filename)
	if err := os.WriteFile(path, []byte(translatedText), 0o600); err != nil {
		return "", fmt.Errorf("không thể ghi nội dung dịch: %w", err)
	}
	return path, nil
}
