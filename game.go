package main

import (
	"github.com/hajimehoshi/ebiten/v2"
)

type Game struct {
	ScreenWidth, ScreenHeight, Cellsize int
	Scenes                              map[SceneName]Scene
	CurrentScene                        SceneName
	Config                              *Config
	Scale                               int
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

	// setup environment
	ebiten.SetWindowSize(game.ScreenWidth, game.ScreenHeight)
	ebiten.SetWindowTitle("golsky - conway's game of life")
	ebiten.SetWindowResizingMode(ebiten.WindowResizingModeEnabled)

	return game
}

func (game *Game) GetCurrentScene() Scene {
	return game.Scenes[game.CurrentScene]
}

func (game *Game) Layout(outsideWidth, outsideHeight int) (int, int) {
	game.Scale = outsideWidth / 100
	return game.ScreenWidth, game.ScreenHeight
}

func (game *Game) Update() error {
	scene := game.GetCurrentScene()
	scene.Update()

	next := scene.GetNext()

	if next != game.CurrentScene {
		// make sure we stay on the selected scene
		scene.ResetNext()

		// finally switch
		game.CurrentScene = next
	}

	return nil
}

func (game *Game) Draw(screen *ebiten.Image) {
	scene := game.GetCurrentScene()

	if scene.Clearscreen() {
		ebiten.SetScreenClearedEveryFrame(true)
	} else {
		ebiten.SetScreenClearedEveryFrame(false)
	}

	scene.Draw(screen)
}
