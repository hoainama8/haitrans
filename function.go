package main

import (
	"fmt"
	"io"
	"os"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/storage"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

//var maxDocumentSize int = 100 * 1024 * 1024 // 100 MB

func MakeDocumentList(directory string) (fyne.CanvasObject, func()) {
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

func MakeFileAttachmentCard(
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
			".txt", ".md", ".json", ".odt", ".py", ".js", ".txt", ".sh", ".csv", ".pdf", ".docx", ".azw3", ".epub", ".html", ".xml", ".rtf",
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

	return container.NewVBox(
		actionBox,
		statusLabel,
	)
}
