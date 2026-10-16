package main

import (
	"fmt"
	"image/color"
	"mime"
	"os"
	"path/filepath"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/driver/desktop"
	"fyne.io/fyne/v2/layout"
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
	Lang := []string{"Tiếng Việt", "Tiếng Anh", "Tiếng Nhật", "Tiếng Trung", "Tiếng Pháp", "Tiếng Hàn", "Tiếng Đức", "Tiếng Latin", "Tiếng Nga"}
	models := []string{"gemini-3.8-flash", "gemini-3.7-flash", "gemini-3.6-flash", "gemma-4-31b-it"}
	homeDir, err := os.UserHomeDir()
	if err != nil {
		fmt.Fprintf(os.Stderr, "không thể tìm thư mục home người dùng: %v\n", err)
		return
	}
	directories, err := applicationDirectoriesFor(homeDir)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return
	}
	for _, directory := range []string{directories.root, directories.input, directories.output} {
		if err := os.MkdirAll(directory, 0o700); err != nil {
			fmt.Fprintf(os.Stderr, "không thể tạo thư mục %s: %v\n", directory, err)
			return
		}
	}
	if _, err := newTokenUsageTracker(directories.tokenUsage); err != nil {
		fmt.Fprintln(os.Stderr, err)
		return
	}
	myApp := app.NewWithID(applicationID)
	myWindow := myApp.NewWindow("Phần Mềm Dịch Thuật AI Đa Ngôn Ngữ")
	//var backend *localBackend

	// CỘT 1: INPUT
	leftColumn, refreshInputColumn := createInputColumn(models)

	// CỘT 3: OUTPUT
	outputEditor := widget.NewMultiLineEntry()
	rightColumn := createOutputColumn(myApp, outputEditor)

	// CỘT 2: GIỮA
	centerColumn := Control(outputEditor, myWindow, Lang, models, refreshInputColumn, directories.input, directories.output)
	if desk, ok := myApp.(desktop.App); ok {
		m := fyne.NewMenu("MyApp",
			fyne.NewMenuItem("Show", func() {
				myWindow.Show()
			}))
		desk.SetSystemTrayMenu(m)
	}
	mainContent := container.NewGridWithRows(3,
		leftColumn,
		centerColumn,
		rightColumn,
	)
	lgCustom := layout.CustomPaddedLayout{
		TopPadding:   20,
		LeftPadding:  20,
		RightPadding: 20,
	}
	myWindow.SetContent(container.New(lgCustom, container.NewStack(
		canvas.NewRadialGradient(
			color.NRGBA{R: 92, G: 51, B: 23, A: 250},
			color.NRGBA{R: 90, G: 70, B: 65, A: 220}),
		mainContent)))
	myWindow.Resize(fyne.NewSize(1200, 650))
	// myWindow.SetCloseIntercept(func() {
	// 	myWindow.Hide()
	// })
	myWindow.ShowAndRun()
}
