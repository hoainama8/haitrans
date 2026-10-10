package main

import (
	"fmt"
	"image/color"
	"strconv"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

type attachedDocument struct {
	name string
	data []byte
}

var document *attachedDocument
var sourceName string

func Control(inputEditor *widget.Entry, outputEditor *widget.Entry, myWindow fyne.Window) fyne.CanvasObject {
	tokenUsagePath, pathErr := defaultTokenUsagePath()
	if pathErr != nil {
		tokenUsagePath = "token_usage.json"
	}
	usageTracker, trackerErr := newTokenUsageTracker(tokenUsagePath)
	if trackerErr != nil {
		usageTracker = &tokenUsageTracker{filePath: tokenUsagePath}
	}
	sessionTokensLabel := widget.NewLabel(strconv.Itoa(usageTracker.sessionTokens))
	totalTokensLabel := widget.NewLabel(strconv.Itoa(usageTracker.totalTokens))
	if pathErr != nil {
		dialog.ShowError(pathErr, myWindow)
	}
	if trackerErr != nil {
		dialog.ShowError(trackerErr, myWindow)
	}

	attachmentSection := MakeFileAttachmentCard(myWindow, func(fileName string, content []byte) {
		sourceName = fileName
		if fileName == "" {
			document = nil
			inputEditor.SetText("")
			inputEditor.SetPlaceHolder("Nhập văn bản hoặc đính kèm tệp...")
			inputEditor.Enable()
			return
		}
		document = &attachedDocument{
			name: fileName,
			data: append([]byte(nil), content...),
		}
		sourceName = fileName
		inputEditor.SetText("")
		inputEditor.SetPlaceHolder("Tài liệu " + fileName + " đã đính kèm để dịch...")

	})

	sourceLangSelect := widget.NewSelect([]string{"Tự động phát hiện", "Tiếng Việt", "Tiếng Anh", "Tiếng Nhật", "Tiếng Trung"}, func(s string) {})
	sourceLangSelect.SetSelected("Tự động phát hiện")

	targetLangSelect := widget.NewSelect([]string{"Tiếng Việt", "Tiếng Anh", "Tiếng Nhật", "Tiếng Trung", "Tiếng Pháp", "Tiếng Hàn", "Tiếng Đức", "Tiếng Latin", "Tiếng Nga"}, func(s string) {})
	targetLangSelect.SetSelected("Tiếng Việt")

	modelSelect := widget.NewSelect([]string{"gemini-3.8-flash", "gemini-3.7-flash", "gemini-3.6-flash", "gemma-4-31b-it"}, func(s string) {})
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
			Model:      modelSelect.Selected,
		}
		if document != nil {
			request.FileName = document.name
			request.FileType = mimeTypeForFile(document.name)
			request.Document = append([]byte(nil), document.data...)
		}
		fmt.Println(request.Model, request.SourceLang, request.TargetLang, request.FileName, len(request.Document), len(request.Text))
		progressBar.Show()
		translateBtn.Disable()

		go func() {
			response, err := TestTrans(request)
			fyne.Do(func() {
				progressBar.Hide()
				translateBtn.Enable()
				if err != nil {
					fmt.Println(err)
					dialog.ShowError(err, myWindow)
					return
				}
				if usage := response.GetUsage(); usage != nil && usage.GetTotalTokens() != nil {
					if err := usageTracker.Add(*usage.GetTotalTokens()); err != nil {
						dialog.ShowError(err, myWindow)
					}
					sessionTokensLabel.SetText(strconv.Itoa(usageTracker.sessionTokens))
					totalTokensLabel.SetText(strconv.Itoa(usageTracker.totalTokens))
				}
				outputEditor.SetText(*response.GetOutputText())
				showSavedTranslation(myWindow, request, *response.GetOutputText())
			})
		}()
	})
	translateBtn.Importance = widget.HighImportance

	optionsForm := widget.NewForm(
		widget.NewFormItem("API key:", apiKeyEntry),
		widget.NewFormItem("Ngôn ngữ nguồn:", sourceLangSelect),
		widget.NewFormItem("Ngôn ngữ đích:", targetLangSelect),
		widget.NewFormItem("Mô hình AI:", modelSelect),
		//widget.NewFormItem("Lựa chọn:", radio),
		//widget.NewFormItem("Chọn:", combo),
	)
	bgColor := color.NRGBA{R: 92, G: 51, B: 23, A: 255} // Màu hổ phách // Màu xanh nhạt
	//bg := canvas.NewRectangle(bgColor)
	center1 := container.NewVBox(
		container.NewCenter(attachmentSection),
		optionsForm,
		container.NewCenter(translateBtn),
		progressBar,
	)
	center2 := container.NewVBox(
		widget.NewCard("Token phiên này", "Số token đã dùng từ khi mở ứng dụng", sessionTokensLabel),
		widget.NewCard("Tổng token đã dùng", "Tích lũy qua các lần sử dụng", totalTokensLabel),
	)
	centerColumn := container.NewGridWithColumns(2, center1, center2)
	center_green := container.NewStack(canvas.NewRectangle(bgColor), centerColumn)
	return center_green
}
