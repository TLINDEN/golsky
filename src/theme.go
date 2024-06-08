package main

import (
	"fmt"
	"image/color"
	"log"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
)

// Color definitions. ColLife could be black or white depending on theme
const (
	ColLife = iota
	ColDead
	ColOld
	ColAge1
	ColAge2
	ColAge3
	ColAge4
	ColGrid
)

// A Theme defines  how the grid and the cells  are colored. We define
// the  colors and  the  actual tile  images here,  so  that they  are
// readily available from play.go
type Theme struct {
	Tiles  map[int]*ebiten.Image
	Colors map[int]color.RGBA
	Name   string
}

// create a new theme
func NewTheme(life, dead, old, age1, age2, age3, age4, grid string,
	cellsize int, name string) Theme {
	theme := Theme{
		Name: name,
		Colors: map[int]color.RGBA{
			ColLife: HexColor2RGBA(life),
			ColDead: HexColor2RGBA(dead),
			ColGrid: HexColor2RGBA(grid),
			ColAge1: HexColor2RGBA(age1),
			ColAge2: HexColor2RGBA(age2),
			ColAge3: HexColor2RGBA(age3),
			ColAge4: HexColor2RGBA(age4),
			ColOld:  HexColor2RGBA(old),
		},
	}

	theme.Tiles = make(map[int]*ebiten.Image, 6)

	for cid, col := range theme.Colors {
		theme.Tiles[cid] = ebiten.NewImage(cellsize, cellsize)
		FillCell(theme.Tiles[cid], cellsize, col)
	}

	return theme
}

// return  the tile  image  for  the requested  color  type. panic  if
// unknown type is being used, which is ok, since the code is the only
// user anyway
func (theme *Theme) Tile(col int) *ebiten.Image {
	return theme.Tiles[col]
}

func (theme *Theme) Color(col int) color.RGBA {
	return theme.Colors[col]
}

type ThemeManager struct {
	Theme  string
	Themes map[string]Theme
}

// Manager is used to easily switch themes from cli or menu
func NewThemeManager(initial string, cellsize int) ThemeManager {
	light := NewTheme(
		"000000", // life
		"c8c8c8", // dead
		"ff1e1e", // old
		"ffc361", // age 1..4
		"ffd38c",
		"ffe3b5",
		"fff0e0",
		"808080", // grid
		cellsize,
		"light",
	)

	dark := NewTheme(
		"c8c8c8", // life
		"000000", // dead
		"ff1e1e", // old
		"522600", // age 1..4
		"422300",
		"2b1b00",
		"191100",
		"808080", // grid
		cellsize,
		"dark",
	)

	standard := NewTheme(
		"e15f0b", // life
		"5a5a5a", // dead
		"e13c0b", // old
		"7b5e4b", // age 1..4
		"735f52",
		"6c6059",
		"635d59",
		"808080", // grid
		cellsize,
		"dark",
	)

	manager := ThemeManager{
		Themes: map[string]Theme{
			"dark":     dark,
			"light":    light,
			"standard": standard,
		},
		Theme: initial,
	}

	return manager
}

func (manager *ThemeManager) GetCurrentTheme() Theme {
	return manager.Themes[manager.Theme]
}

func (manager *ThemeManager) GetCurrentThemeName() string {
	return manager.Theme
}

func (manager *ThemeManager) SetCurrentTheme(theme string) {
	if Exists(manager.Themes, theme) {
		manager.Theme = theme
	}
}

// fill a cell with the given color
func FillCell(tile *ebiten.Image, cellsize int, col color.RGBA) {
	vector.DrawFilledRect(
		tile,
		float32(1),
		float32(1),
		float32(cellsize),
		float32(cellsize),
		col, false,
	)
}

func HexColor2RGBA(hex string) color.RGBA {
	var r, g, b uint8

	_, err := fmt.Sscanf(hex, "%02x%02x%02x", &r, &g, &b)
	if err != nil {
		log.Fatalf("failed to parse hex color: %s", err)
	}

	return color.RGBA{r, g, b, 255}
}
