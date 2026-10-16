package main

import (
	"fmt"
	"strconv"
	"strings"

	"fyne.io/fyne/v2"
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

func Control(outputEditor *widget.Entry, myWindow fyne.Window, Lang []string, models []string, refreshInputColumn func(), inputDirectory string, outDirectory string) fyne.CanvasObject {
	tokenUsagePath, pathErr := defaultTokenUsagePath()
	if pathErr != nil {
		tokenUsagePath = "token_usage.json"
	}
	usageTracker, trackerErr := newTokenUsageTracker(tokenUsagePath)
	if trackerErr != nil {
		usageTracker = &tokenUsageTracker{filePath: tokenUsagePath}
	}
	sessionTokensLabel := widget.NewLabelWithStyle(strconv.Itoa(usageTracker.sessionTokens), fyne.TextAlignLeading, fyne.TextStyle{})
	totalTokensLabel := widget.NewLabelWithStyle(strconv.Itoa(usageTracker.totalTokens), fyne.TextAlignLeading, fyne.TextStyle{})
	if pathErr != nil {
		dialog.ShowError(pathErr, myWindow)
	}
	if trackerErr != nil {
		dialog.ShowError(trackerErr, myWindow)
	}

	attachmentSection := MakeFileAttachmentCard(myWindow, func(fileName string, content []byte) {
		sourceName = fileName
		document = &attachedDocument{
			name: fileName,
			data: append([]byte(nil), content...),
		}

	})

	sourceLang := append([]string{"Tự động phát hiện"}, Lang...)
	sourceLangSelect := widget.NewSelect(sourceLang, func(s string) {})
	sourceLangSelect.SetSelected("Tự động phát hiện")

	targetLangSelect := widget.NewSelect(Lang, func(s string) {})
	targetLangSelect.SetSelected("Tiếng Việt")

	modelSelect := widget.NewSelect(models, func(s string) {})
	modelSelect.PlaceHolder = "Chọn mô hình AI"

	apiKeyEntry := widget.NewPasswordEntry()
	apiKeyEntry.SetPlaceHolder("Nhập API key")
	progressBar := widget.NewProgressBarInfinite()
	progressBar.Hide()

	var translateBtn *widget.Button

	translateBtn = widget.NewButtonWithIcon("DỊCH NGAY VIA AI", theme.NavigateNextIcon(), func() {
		if document == nil {
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
			//Text:       inputEditor.Text,
			FileName: sourceName,
			Model:    modelSelect.Selected,
		}
		if document != nil {
			request.FileName = document.name
			request.FileType = mimeTypeForFile(document.name)
			request.Document = append([]byte(nil), document.data...)
			// Backend uses the uploaded bytes as the source of truth and checks
			// its format before forwarding anything to Gemini.
			request.Text = string(document.data)

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
				if response.GetUsage() != nil {
					if err := response.SaveUsage(usageTracker); err != nil {
						dialog.ShowError(err, myWindow)
					} else {
						sessionTokensLabel.SetText(strconv.Itoa(usageTracker.sessionTokens))
						totalTokensLabel.SetText(strconv.Itoa(usageTracker.totalTokens))
						refreshInputColumn()
					}
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

	center1 := container.NewVBox(
		container.NewCenter(attachmentSection),
		optionsForm,
		container.NewCenter(translateBtn),
		progressBar,
	)
	center2 := container.NewVBox(container.NewCenter(container.NewHBox(OpenDir(inputDirectory, myWindow), OpenDir(outDirectory, myWindow))),
		widget.NewCard("Token phiên này", "Số token đã dùng từ khi mở ứng dụng", sessionTokensLabel),
		widget.NewCard("Tổng token đã dùng", "Tích lũy qua các lần sử dụng", totalTokensLabel),
	)
	centerColumn := container.NewGridWithColumns(2, center1, center2)
	//center_green := container.NewStack(canvas.NewHorizontalGradient(color.NRGBA{R: 92, G: 51, B: 23, A: 255} , color.White), centerColumn)
	return centerColumn
}
