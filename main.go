package main

import (
	"fmt"
	"image/color"
	"io"
	"os"
	"path/filepath"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/storage"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

const applicationID = "vn.hoainam.fynetranslator"

type translationRequest struct {
	APIKey     string
	SourceLang string
	TargetLang string
	Text       string
	FileName   string
	Model      string
	Document   []byte
}

type translationResult struct {
	Text       string
	OutputFile string
}

func TestTranslate(req translationRequest) (translationResult, error) {
	if strings.TrimSpace(req.Text) == "" && len(req.Document) == 0 {
		return translationResult{}, fmt.Errorf("không có nội dung để dịch")
	}

	responseText := req.Text
	if strings.TrimSpace(responseText) == "" && len(req.Document) > 0 {
		responseText = fmt.Sprintf("[Tệp %s đã được nhận. Nội dung đầu vào là tài liệu phụ thuộc định dạng.]", req.FileName)
	}

	return translationResult{
		Text:       fmt.Sprintf("[%s -> %s] %s", req.SourceLang, req.TargetLang, responseText),
		OutputFile: "",
	}, nil
}

// ============================================================================
// HÀM ĐÍNH KÈM FILE TÍCH HỢP QUẢN LÝ FILE
// ============================================================================

func makeFileAttachmentCard(
	win fyne.Window,
	onContentLoaded func(fileName string, content []byte),
) fyne.CanvasObject {

	statusLabel := widget.NewLabel("Chưa có tệp nào được đính kèm")
	statusLabel.TextStyle = fyne.TextStyle{Italic: true}

	// 1. Nút mở ứng dụng File Manager native để CHỌN TỆP
	attachBtn := widget.NewButtonWithIcon("Mở Quản Lý File Chọn Tệp", theme.FolderOpenIcon(), func() {
		fileDialog := dialog.NewFileOpen(func(reader fyne.URIReadCloser, err error) {
			if err != nil {
				fyne.Do(func() {
					dialog.ShowError(err, win)
				})
				return
			}
			if reader == nil {
				return
			}

			fileURI := reader.URI()
			if fileURI == nil {
				_ = reader.Close()
				fyne.Do(func() {
					dialog.ShowError(fmt.Errorf("Tệp được chọn không có đường dẫn hợp lệ"), win)
				})
				return
			}

			data, err := io.ReadAll(io.LimitReader(reader, maxDocumentSize+1))
			closeErr := reader.Close()
			if err != nil {
				fyne.Do(func() {
					dialog.ShowError(fmt.Errorf("Không thể đọc tệp: %v", err), win)
				})
				return
			}
			if len(data) > maxDocumentSize {
				fyne.Do(func() {
					dialog.ShowError(fmt.Errorf("Tệp vượt quá giới hạn 100 MB của ứng dụng"), win)
				})
				return
			}
			if closeErr != nil {
				fyne.Do(func() {
					dialog.ShowError(fmt.Errorf("Không thể đóng tệp: %v", closeErr), win)
				})
				return
			}

			fyne.Do(func() {
				fileName := fileURI.Name()
				statusLabel.SetText(fmt.Sprintf("Đã đính kèm: %s (%.1f KB)", fileName, float64(len(data))/1024.0))
				statusLabel.TextStyle = fyne.TextStyle{Bold: true}

				if onContentLoaded != nil {
					onContentLoaded(fileName, data)
				}
			})
		}, win)

		fileFilter := storage.NewExtensionFileFilter([]string{
			".txt", ".md", ".json", ".go", ".py", ".js", ".html", ".css", ".csv", ".pdf", ".docx",
		})
		fileDialog.SetFilter(fileFilter)
		fileDialog.Show()
	})
	attachBtn.Importance = widget.HighImportance

	// 2. Nút HỦY FILE
	clearBtn := widget.NewButtonWithIcon("", theme.DeleteIcon(), func() {
		statusLabel.SetText("Chưa có tệp nào được đính kèm")
		statusLabel.TextStyle = fyne.TextStyle{Italic: true}
		if onContentLoaded != nil {
			onContentLoaded("", nil)
		}
	})

	actionBox := container.NewHBox(attachBtn, clearBtn)

	return widget.NewCard(
		"Đính Kèm Tài Liệu",
		"Tệp được backend kiểm tra, lưu vào input/ và bản dịch được lưu vào output/",
		container.NewVBox(
			actionBox,
			statusLabel,
		),
	)
}

func makeDocumentList(directory string) (fyne.CanvasObject, func()) {
	var files []string
	fileList := widget.NewList(
		func() int { return len(files) },
		func() fyne.CanvasObject { return widget.NewLabel("Document name") },
		func(index widget.ListItemID, item fyne.CanvasObject) {
			item.(*widget.Label).SetText(files[index])
		},
	)
	emptyLabel := widget.NewLabel("Không có tài liệu")
	listContent := container.NewStack(fileList, emptyLabel)
	fileList.Hide()

	refresh := func() {
		entries, err := os.ReadDir(directory)
		if err != nil {
			files = nil
			emptyLabel.SetText("Không thể đọc thư mục: " + err.Error())
			emptyLabel.Show()
			fileList.Hide()
			fileList.Refresh()
			return
		}

		files = files[:0]
		for _, entry := range entries {
			info, err := entry.Info()
			if err == nil && info.Mode().IsRegular() {
				files = append(files, entry.Name())
			}
		}
		if len(files) == 0 {
			emptyLabel.SetText("Không có tài liệu")
			emptyLabel.Show()
			fileList.Hide()
		} else {
			emptyLabel.Hide()
			fileList.Show()
		}
		fileList.Refresh()
	}

	refreshButton := widget.NewButtonWithIcon("", theme.ViewRefreshIcon(), refresh)
	header := container.NewBorder(nil, nil, nil, refreshButton, widget.NewLabel("Tài liệu trong thư mục"))
	return container.NewBorder(header, nil, nil, nil, listContent), refresh
}

// ============================================================================
// CHƯƠNG TRÌNH CHÍNH
// ============================================================================

func main() {

	myApp := app.NewWithID(applicationID)
	myWindow := myApp.NewWindow("Phần Mềm Dịch Thuật AI Đa Ngôn Ngữ")
	var backend *localBackend

	// CỘT 1: INPUT
	inputEditor := widget.NewMultiLineEntry()
	inputEditor.SetPlaceHolder("Nhập văn bản hoặc đính kèm tệp...")
	inputEditor.Wrapping = fyne.TextWrapWord

	inputStatsLabel := widget.NewLabel("Ký tự: 0 | Từ: 0")
	inputEditor.OnChanged = func(s string) {
		words := len(strings.Fields(s))
		chars := len([]rune(s))
		inputStatsLabel.SetText(fmt.Sprintf("Ký tự: %d | Từ: %d", chars, words))
	}

	leftPanelBg := canvas.NewRectangle(color.NRGBA{R: 225, G: 238, B: 252, A: 255})
	leftColumn := container.NewStack(
		leftPanelBg,
		container.NewBorder(
			widget.NewLabelWithStyle("Tài Liệu Gốc (Input)", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
			inputStatsLabel,
			nil, nil,
			container.NewVScroll(inputEditor),
		),
	)

	// CỘT 3: OUTPUT
	outputEditor := widget.NewMultiLineEntry()
	outputEditor.SetPlaceHolder("Kết quả dịch từ AI sẽ xuất hiện ở đây...")
	outputEditor.Wrapping = fyne.TextWrapWord

	copyBtn := widget.NewButtonWithIcon("Sao Chép", theme.ContentCopyIcon(), func() {
		if outputEditor.Text != "" {
			myWindow.Clipboard().SetContent(outputEditor.Text)
			dialog.ShowInformation("Thành công", "Đã sao chép kết quả vào Clipboard!", myWindow)
		}
	})

	rightPanelBg := canvas.NewRectangle(color.NRGBA{R: 255, G: 248, B: 235, A: 255})
	rightColumn := container.NewStack(
		rightPanelBg,
		container.NewBorder(
			widget.NewLabelWithStyle("Tài Liệu Dịch (Output)", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
			container.NewHBox(copyBtn),
			nil, nil,
			container.NewVScroll(outputEditor),
		),
	)

	inputDocuments, refreshInputDocuments := makeDocumentList(inputDirectory)
	outputDocuments, refreshOutputDocuments := makeDocumentList(outputDirectory)

	// CỘT 2: GIỮA
	type attachedDocument struct {
		name string
		data []byte
	}
	var document *attachedDocument
	var sourceName string
	attachmentSection := makeFileAttachmentCard(myWindow, func(fileName string, content []byte) {
		sourceName = fileName
		document = &attachedDocument{name: fileName, data: content}
		switch strings.ToLower(filepath.Ext(fileName)) {
		case ".pdf", ".docx":
			inputEditor.SetText("")
			inputEditor.SetPlaceHolder("Tài liệu PDF/DOCX đã đính kèm để dịch...")
			inputEditor.Disable()
		default:
			inputEditor.SetPlaceHolder("Nhập văn bản hoặc đính kèm tệp...")
			inputEditor.Enable()
			inputEditor.SetText(string(content))
		}
	})

	sourceLangSelect := widget.NewSelect([]string{"Tự động phát hiện", "Tiếng Việt", "Tiếng Anh", "Tiếng Nhật", "Tiếng Trung"}, func(s string) {})
	sourceLangSelect.SetSelected("Tự động phát hiện")

	targetLangSelect := widget.NewSelect([]string{"Tiếng Việt", "Tiếng Anh", "Tiếng Nhật", "Tiếng Trung", "Tiếng Pháp"}, func(s string) {})
	targetLangSelect.SetSelected("Tiếng Việt")

	modelSelect := widget.NewSelect([]string{"gemini-3.8-flash", "gemma-4-31b-it"}, func(s string) {})
	modelSelect.SetSelected("gemini-3.8-flash")

	apiKeyEntry := widget.NewPasswordEntry()
	apiKeyEntry.SetPlaceHolder("Nhập API key")

	progressBar := widget.NewProgressBarInfinite()
	progressBar.Hide()

	var translateBtn *widget.Button
	translateBtn = widget.NewButtonWithIcon("DỊCH NGAY VIA AI", theme.NavigateNextIcon(), func() {
		if document == nil && strings.TrimSpace(inputEditor.Text) == "" {
			dialog.ShowInformation("Cảnh báo", "Vui lòng nhập văn bản hoặc đính kèm tệp trước!", myWindow)
			return
		}
		if strings.TrimSpace(apiKeyEntry.Text) == "" {
			dialog.ShowInformation("Cảnh báo", "Vui lòng nhập Gemini API key trước!", myWindow)
			return
		}

		request := translationRequest{
			APIKey:     apiKeyEntry.Text,
			SourceLang: sourceLangSelect.Selected,
			TargetLang: targetLangSelect.Selected,
			Text:       inputEditor.Text,
			FileName:   sourceName,
		}
		request.Model = modelSelect.Selected
		if document != nil {
			request.FileName = document.name
			request.Document = append([]byte(nil), document.data...)
			// Backend uses the uploaded bytes as the source of truth and checks
			// its format before forwarding anything to Gemini.
			request.Text = ""
		}
		fmt.Println(request)
		progressBar.Show()
		translateBtn.Disable()

		go func() {
			response, err := TestTrans(request)
			//fmt.Println("req", request)
			//fmt.Println("response:", response)
			fyne.Do(func() {
				progressBar.Hide()
				translateBtn.Enable()
				refreshInputDocuments()
				refreshOutputDocuments()
				if err != nil {
					fmt.Println(err)
					dialog.ShowError(err, myWindow)
					return
				}
				outputEditor.SetText(response.Text)
				showSavedTranslation(myWindow, response.OutputFile)
			})
		}()
	})
	translateBtn.Importance = widget.HighImportance

	optionsForm := widget.NewForm(
		widget.NewFormItem("API key:", apiKeyEntry),
		widget.NewFormItem("Ngôn ngữ nguồn:", sourceLangSelect),
		widget.NewFormItem("Ngôn ngữ đích:", targetLangSelect),
		widget.NewFormItem("Mô hình AI:", modelSelect),
	)

	centerColumn := container.NewVBox(
		attachmentSection,
		widget.NewSeparator(),
		widget.NewLabelWithStyle("Cấu Hình Dịch", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
		optionsForm,
		widget.NewLabel("Cần cài LibreOffice để xử lý tài liệu và tạo PDF."),
		progressBar,
		translateBtn,
		widget.NewCard("Tài liệu gốc", "Danh sách file trong thư mục input/", inputDocuments),
		widget.NewCard("Tài liệu dịch", "Danh sách file trong thư mục output/", outputDocuments),
	)

	mainContent := container.NewGridWithColumns(3,
		leftColumn,
		container.NewVScroll(centerColumn),
		rightColumn,
	)

	var err error
	backend, err = startLocalBackend()
	if err != nil {
		dialog.ShowError(err, myWindow)
		myWindow.ShowAndRun()
		return
	}
	defer backend.close()
	refreshInputDocuments()
	refreshOutputDocuments()

	myWindow.SetContent(container.NewPadded(mainContent))
	myWindow.Resize(fyne.NewSize(1200, 650))
	myWindow.ShowAndRun()
}

func showSavedTranslation(win fyne.Window, outputFile string) {
	message := "Đã lưu bản dịch vào thư mục output."
	if outputFile != "" {
		message = fmt.Sprintf("Đã lưu tệp Gemini trả về: %s", outputFile)
	}
	dialog.ShowInformation("Thành công", message, win)

}

// func main() {
// 	//Boxlayout()
// 	//Borderlayout()
// 	Contain()
// }
