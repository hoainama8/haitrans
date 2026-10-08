package main

import (
	"fmt"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

// CỘT 1: INPUT
func createInputColumn(inputEditor *widget.Entry) fyne.CanvasObject {

	inputEditor.SetPlaceHolder("Nhập văn bản hoặc đính kèm tệp...")
	inputEditor.Wrapping = fyne.TextWrapWord

	inputStatsLabel := widget.NewLabel("Ký tự: 0 | Từ: 0")
	inputEditor.OnChanged = func(s string) {
		words := len(strings.Fields(s))
		chars := len([]rune(s))
		inputStatsLabel.SetText(fmt.Sprintf("Ký tự: %d | Từ: %d", chars, words))
	}

	//leftPanelBg := canvas.NewRectangle(color.NRGBA{R: 225, G: 238, B: 252, A: 255})
	leftColumn := container.NewStack(
		//leftPanelBg,
		container.NewBorder(
			widget.NewLabelWithStyle("Tài Liệu Gốc (Input)", fyne.TextAlignCenter, fyne.TextStyle{Bold: true}),
			inputStatsLabel,
			nil, nil,
			container.NewThemeOverride(
				container.NewScroll(inputEditor),
				outputEditorTheme{Theme: theme.DefaultTheme()},
			),
		),
	)
	return leftColumn
}
