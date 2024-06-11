package main

import (
	"fmt"
	"image/color"
	"os"

	"github.com/ebitenui/ebitenui"
	"github.com/ebitenui/ebitenui/widget"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
)

type SceneMenu struct {
	Game      *Game
	Config    *Config
	Next      SceneName
	Prev      SceneName
	Whoami    SceneName
	Ui        *ebitenui.UI
	FontColor color.RGBA
	First     bool
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

func (scene *SceneMenu) SetPrevious(prev SceneName) {
	scene.Prev = prev
}

func (scene *SceneMenu) ResetNext() {
	scene.Next = scene.Whoami
}

func (scene *SceneMenu) SetNext(next SceneName) {
	scene.Next = next
}

func (scene *SceneMenu) Update() error {
	scene.Ui.Update()

	if inpututil.IsKeyJustPressed(ebiten.KeyEscape) || inpututil.IsKeyJustPressed(ebiten.KeyQ) {
		scene.Config.DelayedStart = false
		scene.Leave()
	}

	return nil

}

func (scene *SceneMenu) IsPrimary() bool {
	return false
}

func (scene *SceneMenu) Draw(screen *ebiten.Image) {
	scene.Ui.Draw(screen)
}

func (scene *SceneMenu) Leave() {
	scene.SetNext(Play)
}

func (scene *SceneMenu) Init() {
	rowContainer := NewRowContainer("Main Menu")

	empty := NewMenuButton("Start with empty grid",
		func(args *widget.ButtonClickedEventArgs) {
			scene.Config.Empty = true
			scene.Config.Restart = true
			scene.Leave()
		})

	random := NewMenuButton("Start with random patterns",
		func(args *widget.ButtonClickedEventArgs) {
			scene.Config.Empty = false
			scene.Config.Restart = true
			scene.Leave()
		})

	copy := NewMenuButton("Save Copy as RLE",
		func(args *widget.ButtonClickedEventArgs) {
			scene.Config.Markmode = true
			scene.Config.Paused = true
			scene.Leave()
		})

	options := NewMenuButton("Options",
		func(args *widget.ButtonClickedEventArgs) {
			fmt.Println("menu => options")
			scene.SetNext(Options)
		})

	bindings := NewMenuButton("Show Key Bindings",
		func(args *widget.ButtonClickedEventArgs) {
			scene.SetNext(Keybindings)
		})

	separator1 := NewSeparator(3)
	separator2 := NewSeparator(3)
	separator3 := NewSeparator(10)

	cancel := NewMenuButton("Back",
		func(args *widget.ButtonClickedEventArgs) {
			scene.Leave()
		})

	quit := NewMenuButton("Exit Golsky",
		func(args *widget.ButtonClickedEventArgs) {
			os.Exit(0)
		})

	rowContainer.AddChild(empty)
	rowContainer.AddChild(random)
	rowContainer.AddChild(separator1)
	rowContainer.AddChild(options)
	rowContainer.AddChild(copy)
	rowContainer.AddChild(bindings)
	rowContainer.AddChild(separator2)
	rowContainer.AddChild(cancel)
	rowContainer.AddChild(separator3)
	rowContainer.AddChild(quit)

	scene.Ui = &ebitenui.UI{
		Container: rowContainer.Container(),
	}

}
