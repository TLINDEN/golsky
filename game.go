package main

import (
	"fmt"

	"github.com/hajimehoshi/ebiten/v2"
)

type Game struct {
	ScreenWidth, ScreenHeight, ReadlWidth, Cellsize int
	Scenes                                          map[SceneName]Scene
	CurrentScene                                    SceneName
	Config                                          *Config
	Scale                                           float32
	Screen                                          *ebiten.Image
}

func NewGame(config *Config, startscene SceneName) *Game {
	game := &Game{
		Config:       config,
		Scenes:       map[SceneName]Scene{},
		ScreenWidth:  config.ScreenWidth,
		ScreenHeight: config.ScreenHeight,
	}

	// setup scene[s]
	game.CurrentScene = startscene
	game.Scenes[Play] = NewPlayScene(game, config)
	game.Scenes[Menu] = NewMenuScene(game, config)
	game.Scenes[Options] = NewOptionsScene(game, config)

	// setup environment
	ebiten.SetWindowSize(game.ScreenWidth, game.ScreenHeight)
	ebiten.SetWindowTitle("golsky - conway's game of life")
	ebiten.SetWindowResizingMode(ebiten.WindowResizingModeEnabled)
	ebiten.SetScreenClearedEveryFrame(true)

	game.Screen = ebiten.NewImage(game.ScreenWidth, game.ScreenHeight)
	return game
}

func (game *Game) GetCurrentScene() Scene {
	return game.Scenes[game.CurrentScene]
}

func (game *Game) Layout(outsideWidth, outsideHeight int) (int, int) {
	game.ReadlWidth = outsideWidth
	game.Scale = float32(game.ScreenWidth) / float32(outsideWidth)
	return game.ScreenWidth, game.ScreenHeight
}

func (game *Game) Update() error {
	scene := game.GetCurrentScene()
	scene.Update()

	fmt.Printf("Clear Screen: %t\n", ebiten.IsScreenClearedEveryFrame())

	return nil
}

func (game *Game) Draw(screen *ebiten.Image) {
	var nextscene Scene
	scene := game.GetCurrentScene()

	next := scene.GetNext()
	if next != game.CurrentScene {
		scene.ResetNext()
		game.CurrentScene = next
		nextscene = game.GetCurrentScene()
		ebiten.SetScreenClearedEveryFrame(nextscene.Clearscreen())
	}

	scene.Draw(screen)

	if nextscene != nil {
		nextscene.Draw(screen)
	}
}
