package main

import (
	"fmt"
	"image/color"

	"github.com/alecthomas/repr"
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
	if value {
		w.SetState(
			widget.WidgetChecked,
		)
	}
}

func (scene *SceneOptions) Init() {
	rowContainer := NewRowContainer("Options")

	pause := NewCheckbox("Pause",
		scene.Config.Paused,
		func(args *widget.CheckboxChangedEventArgs) {
			scene.Config.TogglePaused()
		})

	debugging := NewCheckbox("Debugging",
		scene.Config.Debug,
		func(args *widget.CheckboxChangedEventArgs) {
			scene.Config.ToggleDebugging()
		})

	gridlines := NewCheckbox("Show grid lines",
		scene.Config.ShowGrid,
		func(args *widget.CheckboxChangedEventArgs) {
			fmt.Println("CHECKBOX CALLED")
			repr.Println(args.State)
			scene.Config.ToggleGridlines()
		})

	evolution := NewCheckbox("Show evolution traces",
		scene.Config.ShowEvolution,
		func(args *widget.CheckboxChangedEventArgs) {
			scene.Config.ToggleEvolution()
		})

	wrap := NewCheckbox("Wrap around edges",
		scene.Config.Wrap,
		func(args *widget.CheckboxChangedEventArgs) {
			scene.Config.ToggleWrap()
		})

	themenames := make([]string, len(THEMES))
	i := 0
	for name := range THEMES {
		themenames[i] = name
		i++
	}

	themes := NewCombobox(
		themenames,
		scene.Config.Theme,
		func(args *widget.ListComboButtonEntrySelectedEventArgs) {
			scene.Config.SwitchTheme(args.Entry.(ListEntry).Name)
		})

	themelabel := NewLabel("Themes")
	combocontainer := NewColumnContainer()
	combocontainer.AddChild(themes)
	combocontainer.AddChild(themelabel)

	separator := NewSeparator(3)
	separator2 := NewSeparator(3)

	cancel := NewMenuButton("Close",
		func(args *widget.ButtonClickedEventArgs) {
			scene.SetNext(scene.Prev)
		})

	rowContainer.AddChild(pause)
	rowContainer.AddChild(debugging)
	rowContainer.AddChild(gridlines)
	rowContainer.AddChild(evolution)
	rowContainer.AddChild(wrap)

	rowContainer.AddChild(separator)

	rowContainer.AddChild(combocontainer)

	rowContainer.AddChild(separator2)

	rowContainer.AddChild(cancel)

	scene.Ui = &ebitenui.UI{
		Container: rowContainer.Container(),
	}

}
