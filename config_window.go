package main

import (
	"math"
	"strconv"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

func (a *App) NewConfigWindow() {
	if a.configWindow != nil {
		return
	}

	a.configWindow = a.AppHandle.NewWindow("Journalr - Configuration")
	vbox := container.NewVBox()
	updateFuncs := a.generalConfig(vbox, nil)
	updateFuncs = a.themeConfig(vbox, updateFuncs)

	saveButton := container.NewHBox(widget.NewButton("Save", func() {
		for _, update := range updateFuncs {
			update()
		}
	}))

	content := container.NewBorder(nil, saveButton, nil, nil, container.NewVScroll(vbox))

	a.configWindow.SetContent(content)
	a.configWindow.Resize(fyne.NewSize(a.CfgWindowSize.Width, a.CfgWindowSize.Height))
	a.configWindow.CenterOnScreen()
	a.configWindow.SetOnClosed(func() {
		dims := a.configWindow.Canvas().Size()
		a.CfgWindowSize.Width = dims.Width
		a.CfgWindowSize.Height = dims.Height
		a.configWindow = nil
	})

	a.configWindow.Show()
}

func (a *App) generalConfig(ctr *fyne.Container, updateFuncs []func()) []func() {
	lblHeader := widget.NewLabel("General Settings")
	lblHeader.TextStyle.Bold = true
	lblHeader.SizeName = theme.SizeNameSubHeadingText
	ctr.Add(lblHeader)

	saveDir, updateSaveDir := newCfgEntry("Save Directory:", a.DocumentsPath, func(content string) {
		// TODO: Ensure the directory exists
		a.DocumentsPath = content
	})
	ctr.Add(saveDir)

	return append(updateFuncs, updateSaveDir)
}

func (a *App) themeConfig(ctr *fyne.Container, updateFuncs []func()) []func() {
	lblHeader := widget.NewLabel("Theme Settings")
	lblHeader.TextStyle.Bold = true
	lblHeader.SizeName = theme.SizeNameSubHeadingText
	ctr.Add(lblHeader)

	zoom, updateZoom := newCfgEntry(
		"Zoom", strconv.FormatFloat(float64(a.Theme.Zoom), 'f', 2, 32),
		func(content string) {
			newZoom, err := strconv.ParseFloat(content, 32)
			if err != nil {
				return
			}
			a.Theme.Zoom = float32(newZoom)
		},
	)
	ctr.Add(zoom)

	textSize, updateTextSize := a.updateSizeEntry("Text Size", string(theme.SizeNameText))
	ctr.Add(textSize)
	headingSize, updateHeadingSize := a.updateSizeEntry("Header Size", string(theme.SizeNameSubHeadingText))
	ctr.Add(headingSize)

	return append(updateFuncs, updateZoom, updateTextSize, updateHeadingSize)
}

func (a *App) updateSizeEntry(title, sizeName string) (ctr *fyne.Container, saveFunc func()) {
	return newCfgEntry(
		title, strconv.FormatFloat(float64(a.Theme.SizeNoZoom(sizeName)), 'f', 2, 32),
		func(content string) {
			newSize, err := strconv.ParseFloat(content, 32)
			if err != nil {
				return
			}
			if newSize <= 1.0 || newSize >= 100.0 {
				return
			}

			if math.Abs(float64(a.Theme.SizeNoZoom(sizeName))-float64(newSize)) < 0.01 {
				return
			}
			a.Theme.Sizes[sizeName] = float32(newSize)
		},
	)
}

func newCfgEntry(title, initial string, updateFunc func(content string)) (ctr *fyne.Container, saveFunc func()) {
	entry := widget.NewEntry()
	entry.Text = initial
	content := container.NewBorder(nil, nil, widget.NewLabel(title), nil, entry)

	return content, func() {
		updateFunc(entry.Text)
	}
}
