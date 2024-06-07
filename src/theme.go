package main

import (
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
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
func NewTheme(life, dead, old, age1, age2, age3, age4, grid color.RGBA, cellsize int, name string) Theme {
	theme := Theme{
		Name: name,
		Colors: map[int]color.RGBA{
			ColLife: life,
			ColDead: dead,
			ColGrid: grid,
			ColAge1: age1,
			ColAge2: age2,
			ColAge3: age3,
			ColAge4: age4,
			ColOld:  old,
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
		color.RGBA{0, 0, 0, 0xff},       // life
		color.RGBA{200, 200, 200, 0xff}, // dead
		color.RGBA{255, 30, 30, 0xff},   // old
		color.RGBA{255, 195, 97, 0xff},  // age 1..4
		color.RGBA{255, 211, 140, 0xff},
		color.RGBA{255, 227, 181, 0xff},
		color.RGBA{255, 240, 224, 0xff},
		color.RGBA{128, 128, 128, 0xff}, // grid
		cellsize,
		"light",
	)

	dark := NewTheme(
		color.RGBA{200, 200, 200, 0xff}, // life
		color.RGBA{0, 0, 0, 0xff},       // dead
		color.RGBA{255, 30, 30, 0xff},   // old
		color.RGBA{82, 38, 0, 0xff},     // age 1..4
		color.RGBA{66, 35, 0, 0xff},
		color.RGBA{43, 27, 0, 0xff},
		color.RGBA{25, 17, 0, 0xff},
		color.RGBA{128, 128, 128, 0xff}, // grid
		cellsize,
		"dark",
	)

	manager := ThemeManager{
		Themes: map[string]Theme{
			"dark":  dark,
			"light": light,
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
