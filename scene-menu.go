package main

import (
	"image/color"

	"github.com/ebitenui/ebitenui"
	"github.com/ebitenui/ebitenui/widget"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
)

type SceneMenu struct {
	Game      *Game
	Config    *Config
	Next      SceneName
	Whoami    SceneName
	Ui        *ebitenui.UI
	FontColor color.RGBA
}

func NewMenuScene(game *Game, config *Config) Scene {
	scene := &SceneMenu{
		Whoami:    Menu,
		Game:      game,
		Next:      Menu,
		Config:    config,
		FontColor: color.RGBA{255, 30, 30, 0xff},
	}

	scene.Init()

	return scene
}

func (scene *SceneMenu) GetNext() SceneName {
	return scene.Next
}

func (scene *SceneMenu) ResetNext() {
	scene.Next = scene.Whoami
}

func (scene *SceneMenu) SetNext(next SceneName) {
	scene.Next = next
}

func (scene *SceneMenu) Clearscreen() bool {
	return false
}

func (scene *SceneMenu) Update() error {
	scene.Ui.Update()

	if inpututil.IsKeyJustPressed(ebiten.KeyEscape) || inpututil.IsKeyJustPressed(ebiten.KeyQ) {
		scene.SetNext(Play)
	}

	return nil

}

func (scene *SceneMenu) Draw(screen *ebiten.Image) {
	scene.Ui.Draw(screen)
}

func (scene *SceneMenu) Init() {
	rowContainer := NewRowContainer()

	pause := NewCheckbox("Pause", *FontRenderer.FontSmall,
		func(args *widget.CheckboxChangedEventArgs) {
			scene.Config.TogglePaused()
		})

	copy := NewMenuButton("Save Copy as RLE", *FontRenderer.FontSmall,
		func(args *widget.ButtonClickedEventArgs) {
			scene.Config.Markmode = true
			scene.Config.Paused = true
			scene.SetNext(Play)
		})

	label := widget.NewText(
		widget.TextOpts.Text("Menu", *FontRenderer.FontNormal, scene.FontColor),
		widget.TextOpts.Position(widget.TextPositionCenter, widget.TextPositionCenter),
	)

	rowContainer.AddChild(label)
	rowContainer.AddChild(pause)
	rowContainer.AddChild(copy)

	scene.Ui = &ebitenui.UI{
		Container: rowContainer.Container(),
	}

}
