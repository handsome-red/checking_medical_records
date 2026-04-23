package main

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
	"fyne.io/fyne/v2/dialog"
	"os"
	"path/filepath"
)

type FileItem struct {
	Name    string
	Path    string
	Content []byte
}

var uploadedFiles []FileItem

func main() {
	myApp := app.New()
	myWindow := myApp.NewWindow("Drag-and-Drop Prep (File Loader)")

	// Список загруженных файлов
	fileList := widget.NewList(
		func() int { return len(uploadedFiles) },
		func() fyne.CanvasObject { return widget.NewLabel("template") },
		func(i int, obj fyne.CanvasObject) {
			obj.(*widget.Label).SetText(uploadedFiles[i].Name)
		},
	)

	// Кнопка загрузки
	loadButton := widget.NewButton("Выбрать файлы", func() {
		dialog.ShowFileOpen(func(reader fyne.URIReadCloser, err error) {
			if err != nil || reader == nil {
				return
			}
			defer reader.Close()

			// Читаем файл
			data := make([]byte, 1024*1024) // ограничим 1MB пока
			n, _ := reader.Read(data)
			
			uploadedFiles = append(uploadedFiles, FileItem{
				Name:    filepath.Base(reader.URI().Path()),
				Path:    reader.URI().Path(),
				Content: data[:n],
			})
			fileList.Refresh()
			
			// Показываем сообщение
			dialog.ShowInformation("Успех", "Файл загружен", myWindow)
		}, myWindow)
	})

	// Зона-подсказка (визуальная, но пока без drag-drop)
	dropHint := widget.NewLabel("Перетащите файлы сюда\n(пока не работает, используйте кнопку ниже)")
	dropHint.Alignment = fyne.TextAlignCenter
	dropZone := container.NewCenter(dropHint)
	
	content := container.NewBorder(
		dropZone,
		loadButton,
		nil, nil,
		container.NewBorder(widget.NewLabel("Загруженные файлы:"), nil, nil, nil, fileList),
	)

	myWindow.SetContent(content)
	myWindow.Resize(fyne.NewSize(600, 400))
	myWindow.ShowAndRun()
}