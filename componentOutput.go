package main

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

func createOutputColumn(app fyne.App, outputEditor *widget.Entry) fyne.CanvasObject {
	outputEditor.SetPlaceHolder("Kết quả dịch từ AI sẽ xuất hiện ở đây...")
	outputEditor.Wrapping = fyne.TextWrapWord

	copyBtn := widget.NewButtonWithIcon("Sao Chép", theme.ContentCopyIcon(), func() {
		if outputEditor.Text != "" {
			app.Clipboard().SetContent(outputEditor.Text)
			dialog.ShowInformation("Thành công", "Đã sao chép kết quả vào Clipboard!", app.Driver().AllWindows()[0])
		}
	})

	//rightPanelBg := canvas.NewRectangle(color.NRGBA{R: 255, G: 248, B: 235, A: 255})
	rightColumn := container.NewBorder(
		nil,
		container.NewHBox(copyBtn),
		nil, nil,
		container.NewThemeOverride(
			container.NewVScroll(outputEditor),
			outputEditorTheme{Theme: theme.DefaultTheme()},
		),
	)
	return rightColumn

	//inputDocuments, refreshInputDocuments := makeDocumentList(inputDirectory)
	//outputDocuments, refreshOutputDocuments := makeDocumentList(outputDirectory)
}

type outputEditorTheme struct {
	fyne.Theme
}

func (t outputEditorTheme) Size(name fyne.ThemeSizeName) float32 {
	size := t.Theme.Size(name)
	if name == theme.SizeNameText {
		return size * 1.3
	}
	return size
}
