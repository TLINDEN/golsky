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

type ThemeDef struct {
	life, dead, grid, old, age1, age2, age3, age4 string
}

var THEMES = map[string]ThemeDef{
	"standard": {
		life: "e15f0b",
		dead: "5a5a5a",
		grid: "e13c0b",
		old:  "7b5e4b",
		age1: "735f52",
		age2: "6c6059",
		age3: "635d59",
		age4: "808080",
	},
	"dark": {
		life: "c8c8c8",
		dead: "000000",
		grid: "ff1e1e",
		old:  "522600",
		age1: "422300",
		age2: "2b1b00",
		age3: "191100",
		age4: "808080",
	},
	"light": {
		life: "000000",
		dead: "c8c8c8",
		grid: "ff1e1e",
		old:  "ffc361",
		age1: "ffd38c",
		age2: "ffe3b5",
		age3: "fff0e0",
		age4: "808080",
	},
}

// create a new theme
func NewTheme(def ThemeDef, cellsize int, name string) Theme {
	theme := Theme{
		Name: name,
		Colors: map[int]color.RGBA{
			ColLife: HexColor2RGBA(def.life),
			ColDead: HexColor2RGBA(def.dead),
			ColGrid: HexColor2RGBA(def.grid),
			ColAge1: HexColor2RGBA(def.age1),
			ColAge2: HexColor2RGBA(def.age2),
			ColAge3: HexColor2RGBA(def.age3),
			ColAge4: HexColor2RGBA(def.age4),
			ColOld:  HexColor2RGBA(def.old),
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
	manager := ThemeManager{
		Theme: initial,
	}

	manager.Themes = make(map[string]Theme, len(THEMES))

	for name, def := range THEMES {
		manager.Themes[name] = NewTheme(def, cellsize, name)
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
