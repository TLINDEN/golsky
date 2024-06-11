package main

import (
	"fmt"
	"image/color"

	"github.com/ebitenui/ebitenui"
	"github.com/ebitenui/ebitenui/widget"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
)

type SceneToolbar struct {
	Game      *Game
	Config    *Config
	Next      SceneName
	Prev      SceneName
	Whoami    SceneName
	Ui        *ebitenui.UI
	FontColor color.RGBA
}

func NewToolbarScene(game *Game, config *Config) Scene {
	scene := &SceneToolbar{
		Whoami:    Toolbar,
		Game:      game,
		Next:      Toolbar,
		Config:    config,
		FontColor: color.RGBA{255, 30, 30, 0xff},
	}

	scene.Init()

	return scene
}

func (scene *SceneToolbar) GetNext() SceneName {
	return scene.Next
}

func (scene *SceneToolbar) SetPrevious(prev SceneName) {
	scene.Prev = prev
}

func (scene *SceneToolbar) ResetNext() {
	scene.Next = scene.Whoami
}

func (scene *SceneToolbar) SetNext(next SceneName) {
	scene.Next = next
}

func (scene *SceneToolbar) IsPrimary() bool {
	return true
}

func (scene *SceneToolbar) Update() error {
	scene.Ui.Update()

	if inpututil.IsKeyJustPressed(ebiten.KeyEscape) || inpututil.IsKeyJustPressed(ebiten.KeyQ) {
		scene.SetNext(Play)
	}

	return nil

}

func (scene *SceneToolbar) Draw(screen *ebiten.Image) {
	scene.Ui.Draw(screen)
}

func (scene *SceneToolbar) SetInitialValue(w *widget.LabeledCheckbox, value bool) {
	if value {
		w.SetState(
			widget.WidgetChecked,
		)
	}
}

func (scene *SceneToolbar) Init() {
	rowContainer := NewTopRowContainer("Toolbar")

	options := NewToolbarButton(Assets["options"],
		func(args *widget.ButtonClickedEventArgs) {
			fmt.Println("options")
			scene.SetNext(Options)
		})

	rowContainer.AddChild(options)

	scene.Ui = &ebitenui.UI{
		Container: rowContainer.Container(),
	}

}
