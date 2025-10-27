package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/driver/desktop"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

type Dimensions struct {
	Width  float32 `json:"width"`
	Height float32 `json:"height"`
}

// App is the container for the journalr application.
type App struct {
	ConfigPath    string           `json:"-"`
	DocumentsPath string           `json:"save-dir"`
	AppHandle     fyne.App         `json:"-"`
	Window        fyne.Window      `json:"-"`
	WindowSize    Dimensions       `json:"window"`
	CfgWindowSize Dimensions       `json:"config-window"`
	Saveables     []*SaveableEntry `json:"-"`
	Theme         *Theme           `json:"theme"`

	saveShortcut *desktop.CustomShortcut `json:"-"`
	configWindow fyne.Window             `json:"-"`
}

// New returns a new App instance.
func New() (*App, error) {
	homeDir, confDir, err := GetFolders()
	if err != nil {
		return nil, err
	}
	docsPath := filepath.Join(homeDir, "Documents", "journalr")
	configPath := filepath.Join(confDir, "tvarney", "journalr", "config.json")
	save := &desktop.CustomShortcut{KeyName: fyne.KeyS, Modifier: fyne.KeyModifierControl}
	fyneApp := app.New()
	window := fyneApp.NewWindow("Journalr")

	app := &App{
		ConfigPath:    configPath,
		DocumentsPath: docsPath,
		AppHandle:     fyneApp,
		Window:        window,
		Theme:         NewTheme(),
		WindowSize:    Dimensions{Width: 640, Height: 480.0},
		CfgWindowSize: Dimensions{Width: 640.0, Height: 480.0},
		saveShortcut:  save,
	}

	if _, dneErr := os.Stat(configPath); dneErr == nil {
		data, readErr := os.ReadFile(configPath)
		if readErr == nil {
			_ = json.Unmarshal(data, app)
		}
	}
	if app.Theme == nil {
		fmt.Println("Theme is nil; using new theme")
		app.Theme = NewTheme()
	}

	fyneApp.Settings().SetTheme(app.Theme)

	app.init()
	app.createMenu()
	app.Window.SetOnClosed(func() {
		// Save current size
		dims := app.Window.Canvas().Size()
		app.WindowSize.Width = dims.Width
		app.WindowSize.Height = dims.Height

		// Write out config
		app.WriteConfig()

		// Write out content
		app.Save()
	})

	return app, nil
}

func (a *App) init() {
	now := time.Now()
	year, month, day := now.Year(), now.Month(), now.Day()
	if now.Hour() <= 4 {
		day--
	}
	prefix := fmt.Sprintf("%d-%02d-%02d - ", year, month, day)
	worriesFile := filepath.Join(a.DocumentsPath, prefix+"worries.txt")
	journalFile := filepath.Join(a.DocumentsPath, prefix+"journal.txt")

	wParent, wSaveable := NewTextArea(a.saveShortcut, worriesFile, a.ErrorModal)
	jParent, jSaveable := NewTextArea(a.saveShortcut, journalFile, a.ErrorModal)
	tabs := container.NewAppTabs(
		container.NewTabItem("Worries", wParent),
		container.NewTabItem("Journal", jParent),
	)
	a.Saveables = append(a.Saveables, wSaveable, jSaveable)

	a.Window.SetContent(tabs)
	a.Window.Resize(fyne.NewSize(a.WindowSize.Width, a.WindowSize.Height))
	a.Window.CenterOnScreen()
}

func (a *App) createMenu() {
	a.Window.SetMainMenu(
		fyne.NewMainMenu(
			fyne.NewMenu(
				"File",
				fyne.NewMenuItem("Save", a.Save),
				fyne.NewMenuItemSeparator(),
				fyne.NewMenuItem("Quit", a.Window.Close),
			),
			fyne.NewMenu(
				"Edit",
				fyne.NewMenuItem("Settings", a.NewConfigWindow),
			),
		),
	)
}

func (a *App) Save() {
	for _, item := range a.Saveables {
		if err := item.Save(); err != nil {
			a.ErrorModal(err)
		}
	}
}

// TODO: Test this
func (a *App) ErrorModal(err error) {
	box := container.NewVBox()

	lbl := widget.NewLabel("Error: " + err.Error())
	lbl.SizeName = theme.SizeNameSubHeadingText
	box.Add(lbl)

	if detailedErr, ok := err.(*DetailedError); ok {
		details := widget.NewLabel(detailedErr.Detail)
		details.SizeName = theme.SizeNameText
		box.Add(details)
	}

	popup := widget.NewModalPopUp(box, a.Window.Canvas())
	box.Add(widget.NewButton("Okay", func() { popup.Hide() }))
	popup.Show()
}

// NewTextArea returns a new entry with status bar that can be saved with ctrl+s.
func NewTextArea(
	save *desktop.CustomShortcut, saveLoc string, errCallback func(error),
) (parent *fyne.Container, savable *SaveableEntry) {
	textEntry := NewSavableEntry(saveLoc)
	statusBar := widget.NewLabel("Word Count: 0")
	textEntry.OnChanged = func(content string) {
		statusBar.SetText("Word Count: " + strconv.FormatInt(WordCount(content), 10))
	}
	textEntry.ErrCallback = errCallback

	if _, dneErr := os.Stat(saveLoc); dneErr == nil {
		contents, _ := os.ReadFile(saveLoc)
		textEntry.SetText(string(contents))
		textEntry.LastSave = string(contents)
	}

	ctr := container.NewBorder(nil, statusBar, nil, nil, textEntry)
	return ctr, textEntry
}

// GetFolders returns a set of user folders.
func GetFolders() (homeDir, confDir string, err error) {
	conf, err := os.UserConfigDir()
	if err != nil {
		return "", "", fmt.Errorf("getting user config directory: %w", err)
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return "", "", fmt.Errorf("getting user home directory: %w", err)
	}

	return home, conf, nil
}

func (a *App) WriteConfig() {
	if err := os.MkdirAll(filepath.Dir(a.ConfigPath), 0o666); err != nil {
		a.ErrorModal(&DetailedError{
			Message: "Saving configuration",
			Detail:  "Creating directory: " + err.Error(),
		})
	}

	data, err := json.MarshalIndent(a, "", "  ")
	if err != nil {
		a.ErrorModal(&DetailedError{
			Message: "Saving configuration",
			Detail:  "Marshalling data: " + err.Error(),
		})
	}
	fp, err := os.Create(a.ConfigPath)
	if err != nil {
		a.ErrorModal(&DetailedError{
			Message: "Saving configuration",
			Detail:  "Writing file: " + err.Error(),
		})
		return
	}
	fp.Write(data)
}
