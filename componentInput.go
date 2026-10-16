package main

import (
	"fmt"
	"image/color"
	"os"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

type Price struct {
	Input  float64
	Output float64
	Cached float64
}

var modelPrices = map[string]Price{
	"gemini-3.8-flash":      {Input: 0.75 / 100000000, Output: 3.75 / 100000000, Cached: 0.075 / 100000000},
	"gemini-3.6-flash":      {Input: 0.75 / 100000000, Output: 3.75 / 100000000, Cached: 0.075 / 100000000},
	"gemini-3.5-flash-Lite": {Input: 0.3 / 100000000, Output: 2.5 / 100000000, Cached: 0.03 / 100000000},
	// "gemini-3.6-flash": {Input: ..., Output: ...},
	"gemma-4-31b-it": {Input: 0, Output: 0, Cached: 0},
}

// CỘT 1: INPUT
func createInputColumn(models []string) (*fyne.Container, func()) {

	tabs := createInputTabs(models)
	customLay := layout.CustomPaddedLayout{
		TopPadding:   20,
		LeftPadding:  30,
		RightPadding: 30,
	}
	inputComponents := container.New(customLay, tabs)
	refresh := func() {
		inputComponents.Objects = []fyne.CanvasObject{createInputTabs(models)}
		inputComponents.Refresh()
	}

	return inputComponents, refresh
}

func createInputTabs(models []string) *container.AppTabs {
	text := canvas.NewText("Ngày", color.Black)
	tabs := container.NewAppTabs()
	var tokenData tokenUsageData
	tokenUsagePath, err := defaultTokenUsagePath()
	if err != nil {
		fmt.Println("Error finding token file:", err)
	} else {
		tokenFile, err := os.ReadFile(tokenUsagePath)
		if err != nil {
			fmt.Println("Error reading token file:", err)
		} else if tokenData, err = decodeTokenUsageData(tokenFile); err != nil {
			fmt.Println("Error parsing token file:", err)
		}
	}

	for _, model := range models {
		colDate := container.NewVBox(widget.NewLabelWithStyle(fmt.Sprintf("%s", text.Text), fyne.TextAlignCenter, fyne.TextStyle{Bold: true}))
		colHour := container.NewVBox(widget.NewLabelWithStyle("Giờ", fyne.TextAlignCenter, fyne.TextStyle{Bold: true}))
		colInputToken := container.NewVBox(widget.NewLabelWithStyle("Input Token", fyne.TextAlignCenter, fyne.TextStyle{Bold: true}))
		colOuputToken := container.NewVBox(widget.NewLabelWithStyle("Output Token", fyne.TextAlignCenter, fyne.TextStyle{}))
		colThoughtToken := container.NewVBox(widget.NewLabelWithStyle("Thought Token", fyne.TextAlignCenter, fyne.TextStyle{}))
		colCachedToken := container.NewVBox(widget.NewLabelWithStyle("Cached Token", fyne.TextAlignCenter, fyne.TextStyle{}))
		colCost := container.NewVBox(widget.NewLabelWithStyle("Chi Phi", fyne.TextAlignCenter, fyne.TextStyle{}))

		price := modelPrices[model]
		var totalCost float64
		for _, session := range tokenData[model] {
			colDate.Add(widget.NewLabelWithStyle(session.StartedAt.Format("2006-01-02"), fyne.TextAlignCenter, fyne.TextStyle{}))
			colHour.Add(widget.NewLabelWithStyle(session.StartedAt.Format("15:04:05"), fyne.TextAlignCenter, fyne.TextStyle{}))
			colInputToken.Add(widget.NewLabelWithStyle(fmt.Sprintf("%d", session.InputTokens), fyne.TextAlignCenter, fyne.TextStyle{}))
			colOuputToken.Add(widget.NewLabelWithStyle(fmt.Sprintf("%d", session.OutputTokens), fyne.TextAlignCenter, fyne.TextStyle{}))
			colThoughtToken.Add(widget.NewLabelWithStyle(fmt.Sprintf("%d", session.ThoughtTokens), fyne.TextAlignCenter, fyne.TextStyle{}))
			colCachedToken.Add(widget.NewLabelWithStyle(fmt.Sprintf("%d", session.TotalTokens), fyne.TextAlignCenter, fyne.TextStyle{}))
			sessionCost := (float64(session.InputTokens) * price.Input) + (float64(session.OutputTokens+session.ThoughtTokens) * price.Output)
			totalCost += sessionCost
			colCost.Add(widget.NewLabelWithStyle(fmt.Sprintf("$%.6f", sessionCost), fyne.TextAlignCenter, fyne.TextStyle{}))
		}

		colDate.Add(widget.NewLabelWithStyle("Tổng", fyne.TextAlignCenter, fyne.TextStyle{Bold: true}))
		colHour.Add(widget.NewLabel(""))
		colInputToken.Add(widget.NewLabel(""))
		colOuputToken.Add(widget.NewLabel(""))
		colThoughtToken.Add(widget.NewLabel(""))
		colCachedToken.Add(widget.NewLabel(""))
		colCost.Add(widget.NewLabelWithStyle(fmt.Sprintf("$%.6f", totalCost), fyne.TextAlignCenter, fyne.TextStyle{Bold: true}))

		db := container.NewScroll(container.NewGridWithColumns(7, colDate, colHour, colInputToken, colOuputToken, colThoughtToken, colCachedToken, colCost))
		tabs.Append(container.NewTabItemWithIcon(
			model,
			theme.DocumentIcon(),
			container.NewThemeOverride(
				container.NewStack(db),
				outputEditorTheme{Theme: theme.DefaultTheme()},
			),
		))
	}
	return tabs
}
