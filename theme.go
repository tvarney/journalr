package main

import (
	"image/color"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/theme"
)

type Theme struct {
	Zoom   float32            `json:"zoom"`
	Sizes  map[string]float32 `json:"sizes"`
	Colors map[string]uint32  `json:"colors"`
}

func NewTheme() *Theme {
	return &Theme{
		Zoom:   1.0,
		Sizes:  map[string]float32{},
		Colors: map[string]uint32{},
	}
}

func (t *Theme) Color(name fyne.ThemeColorName, variant fyne.ThemeVariant) color.Color {
	colorInt, ok := t.Colors[string(name)]
	if !ok {
		return theme.DefaultTheme().Color(name, variant)
	}

	return color.RGBA{
		R: uint8((colorInt & 0x00FF0000) >> 16),
		G: uint8((colorInt & 0x0000FF00) >> 8),
		B: uint8((colorInt & 0x000000FF) >> 0),
		A: 0xFF,
	}
}

func (t *Theme) Font(style fyne.TextStyle) fyne.Resource {
	return theme.DefaultTheme().Font(style)
}

func (t *Theme) Icon(name fyne.ThemeIconName) fyne.Resource {
	return theme.DefaultTheme().Icon(name)
}

func (t *Theme) Size(name fyne.ThemeSizeName) float32 {
	if t.Zoom <= 0 || t.Zoom > 10.0 {
		t.Zoom = 1.0
	}

	sizeVal, ok := t.Sizes[string(name)]
	if ok {
		return sizeVal * t.Zoom
	}
	return theme.DefaultTheme().Size(name) * t.Zoom
}

func (t *Theme) SizeNoZoom(name string) float32 {
	sizeVal, ok := t.Sizes[string(name)]
	if ok {
		return sizeVal
	}
	return theme.DefaultTheme().Size(fyne.ThemeSizeName(name))
}
