package main

import (
	"mime"
	"path/filepath"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

const applicationID = "vn.hoainam.fynetranslator"

type translationRequest struct {
	APIKey     string
	SourceLang string
	TargetLang string
	Text       string
	FileName   string
	FileType   string
	Model      string
	Document   []byte
}

func mimeTypeForFile(fileName string) string {
	switch strings.ToLower(filepath.Ext(fileName)) {
	case ".pdf":
		return "application/pdf"
	case ".docx":
		return "application/vnd.openxmlformats-officedocument.wordprocessingml.document"
	}

	detected := mime.TypeByExtension(strings.ToLower(filepath.Ext(fileName)))
	mediaType, _, err := mime.ParseMediaType(detected)
	if err != nil {
		return "application/octet-stream"
	}
	return mediaType
}

// ============================================================================
// HÀM ĐÍNH KÈM FILE TÍCH HỢP QUẢN LÝ FILE
// ============================================================================

// ============================================================================
// CHƯƠNG TRÌNH CHÍNH
// ============================================================================

func main() {

	myApp := app.NewWithID(applicationID)
	myWindow := myApp.NewWindow("Phần Mềm Dịch Thuật AI Đa Ngôn Ngữ")
	//var backend *localBackend

	// CỘT 1: INPUT
	inputEditor := widget.NewMultiLineEntry()
	inputEditor.TextStyle.Bold = true
	inputEditor.Enable()
	leftColumn := createInputColumn(inputEditor)

	// CỘT 3: OUTPUT
	outputEditor := widget.NewMultiLineEntry()
	rightColumn := createOutputColumn(myApp, outputEditor)

	// CỘT 2: GIỮA
	centerColumn := Control(inputEditor, outputEditor, myWindow)
	mainContent := container.NewGridWithRows(3,
		leftColumn,
		centerColumn,
		rightColumn,
	)

	myWindow.SetContent(container.NewPadded(mainContent))
	myWindow.Resize(fyne.NewSize(1200, 650))
	myWindow.ShowAndRun()
}
