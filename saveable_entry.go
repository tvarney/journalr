package main

import (
	"os"
	"path/filepath"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/driver/desktop"
	"fyne.io/fyne/v2/widget"
)

type DetailedError struct {
	Message string
	Detail  string
}

func (e *DetailedError) Error() string {
	return e.Message
}

type SaveableEntry struct {
	widget.Entry

	FilePath    string
	LastSave    string
	ErrCallback func(error)
}

func NewSavableEntry(filePath string) *SaveableEntry {
	e := &SaveableEntry{}
	e.ExtendBaseWidget(e)

	e.MultiLine = true
	e.Wrapping = fyne.TextWrapWord
	e.Scroll = fyne.ScrollVerticalOnly
	e.FilePath = filePath
	return e
}

func (e *SaveableEntry) Save() error {
	if e.Text == e.LastSave {
		return nil
	}

	if e.FilePath == "" {
		return nil
	}

	baseDir := filepath.Dir(e.FilePath)
	if _, statErr := os.Stat(baseDir); statErr != nil {
		if err := os.MkdirAll(baseDir, 0o666); err != nil {
			return &DetailedError{
				Message: "Failed to save",
				Detail:  "Unable to create directory: " + err.Error(),
			}
		}
	}

	fp, err := os.Create(e.FilePath)
	if err != nil {
		return &DetailedError{
			Message: "Failed to save",
			Detail:  err.Error(),
		}
	}

	fp.Write([]byte(e.Text))
	fp.Close()
	e.LastSave = e.Text
	return nil
}

func (e *SaveableEntry) TypedShortcut(s fyne.Shortcut) {
	shortcut, ok := s.(*desktop.CustomShortcut)
	if !ok || !(shortcut.KeyName == fyne.KeyS && shortcut.Modifier == fyne.KeyModifierControl) {
		e.Entry.TypedShortcut(s)
		return
	}

	if err := e.Save(); err != nil {
		if e.ErrCallback != nil {
			e.ErrCallback(err)
		}
	}
}
