package main

import (
	"image/color"

	"github.com/ebitenui/ebitenui"
	"github.com/ebitenui/ebitenui/widget"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
)

type SceneOptions struct {
	Game      *Game
	Config    *Config
	Next      SceneName
	Prev      SceneName
	Whoami    SceneName
	Ui        *ebitenui.UI
	FontColor color.RGBA
}

func NewOptionsScene(game *Game, config *Config) Scene {
	scene := &SceneOptions{
		Whoami:    Options,
		Game:      game,
		Next:      Options,
		Config:    config,
		FontColor: color.RGBA{255, 30, 30, 0xff},
	}

	scene.Init()

	return scene
}

func (scene *SceneOptions) GetNext() SceneName {
	return scene.Next
}

func (scene *SceneOptions) SetPrevious(prev SceneName) {
	scene.Prev = prev
}

func (scene *SceneOptions) ResetNext() {
	scene.Next = scene.Whoami
}

func (scene *SceneOptions) SetNext(next SceneName) {
	scene.Next = next
}

func (scene *SceneOptions) IsPrimary() bool {
	return false
}

func (scene *SceneOptions) Update() error {
	scene.Ui.Update()

	if inpututil.IsKeyJustPressed(ebiten.KeyEscape) || inpututil.IsKeyJustPressed(ebiten.KeyQ) {
		scene.SetNext(Play)
	}

	return nil

}

func (scene *SceneOptions) Draw(screen *ebiten.Image) {
	scene.Ui.Draw(screen)
}

func (scene *SceneOptions) SetInitialValue(w *widget.LabeledCheckbox, value bool) {
	var intval int
	if value {
		intval = 1
	}

	w.SetState(
		widget.WidgetState(intval),
	)
}

func (scene *SceneOptions) Init() {
	rowContainer := NewRowContainer("Options")

	pause := NewCheckbox("Pause",
		func(args *widget.CheckboxChangedEventArgs) {
			scene.Config.TogglePaused()
		})

	debugging := NewCheckbox("Debugging",
		func(args *widget.CheckboxChangedEventArgs) {
			scene.Config.ToggleDebugging()
		})
	scene.SetInitialValue(debugging, scene.Config.Debug)

	invert := NewCheckbox("Invert",
		func(args *widget.CheckboxChangedEventArgs) {
			scene.Config.ToggleInvert()
		})
	scene.SetInitialValue(invert, scene.Config.Invert)

	gridlines := NewCheckbox("Show grid lines",
		func(args *widget.CheckboxChangedEventArgs) {
			scene.Config.ToggleGridlines()
		})
	scene.SetInitialValue(gridlines, scene.Config.ShowGrid)

	evolution := NewCheckbox("Show evolution traces",
		func(args *widget.CheckboxChangedEventArgs) {
			scene.Config.ToggleEvolution()
		})
	scene.SetInitialValue(evolution, scene.Config.ShowEvolution)

	separator := NewSeparator(3)

	cancel := NewMenuButton("Close",
		func(args *widget.ButtonClickedEventArgs) {
			scene.SetNext(scene.Prev)
		})

	rowContainer.AddChild(pause)
	rowContainer.AddChild(debugging)
	rowContainer.AddChild(invert)
	rowContainer.AddChild(gridlines)
	rowContainer.AddChild(evolution)
	rowContainer.AddChild(separator)
	rowContainer.AddChild(cancel)

	scene.Ui = &ebitenui.UI{
		Container: rowContainer.Container(),
	}

}
