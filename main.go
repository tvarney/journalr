package main

import (
	"os"

	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/widget"
)

func main() {
	application, err := New()
	if err != nil {
		// TODO: Better way to display this?
		fyneApp := app.New()
		window := fyneApp.NewWindow("Journalr - Error")
		window.SetContent(widget.NewLabel("Error: " + err.Error()))
		window.ShowAndRun()
		os.Exit(1)
	}

	application.Window.ShowAndRun()
}
