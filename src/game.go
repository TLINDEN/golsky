package main

import (
	"github.com/hajimehoshi/ebiten/v2"
)

type Game struct {
	ScreenWidth, ScreenHeight, ReadlWidth, Cellsize int
	Scenes                                          map[SceneName]Scene
	CurrentScene                                    SceneName
	Config                                          *Config
	Scale                                           float32
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
	game.Scenes[Toolbar] = NewToolbarScene(game, config)
	game.Scenes[Menu] = NewMenuScene(game, config)
	game.Scenes[Options] = NewOptionsScene(game, config)
	game.Scenes[Keybindings] = NewKeybindingsScene(game, config)

	// setup environment
	ebiten.SetWindowSize(game.ScreenWidth, game.ScreenHeight)
	ebiten.SetWindowTitle("golsky - conway's game of life")
	ebiten.SetWindowResizingMode(ebiten.WindowResizingModeEnabled)
	ebiten.SetScreenClearedEveryFrame(true)

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
	for _, scene := range game.Scenes {
		if scene.IsPrimary() {
			if quit := scene.Update(); quit != nil {
				return quit
			}

		}
	}

	scene := game.GetCurrentScene()
	next := scene.GetNext()

	if next != game.CurrentScene {
		game.Scenes[next].SetPrevious(game.CurrentScene)
		scene.ResetNext()
		game.CurrentScene = next
	}

	return nil
}

func (game *Game) Draw(screen *ebiten.Image) {
	// first draw primary scene[s], although there are only 1
	skip := false
	for current, scene := range game.Scenes {
		if scene.IsPrimary() {
			// primary scenes always draw
			scene.Draw(screen)

			if current == game.CurrentScene {
				// avoid to redraw it in the next step
				skip = true
				break
			}
		}
	}

	if skip {
		return
	}

	scene := game.GetCurrentScene()
	scene.Draw(screen)
}
