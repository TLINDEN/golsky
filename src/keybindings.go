package main

import (
	"image/color"

	"github.com/ebitenui/ebitenui"
	"github.com/ebitenui/ebitenui/widget"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
)

type SceneKeybindings struct {
	Game      *Game
	Config    *Config
	Next      SceneName
	Prev      SceneName
	Whoami    SceneName
	Ui        *ebitenui.UI
	FontColor color.RGBA
	First     bool
}

func NewKeybindingsScene(game *Game, config *Config) Scene {
	scene := &SceneKeybindings{
		Whoami:    Keybindings,
		Game:      game,
		Next:      Keybindings,
		Config:    config,
		FontColor: color.RGBA{255, 30, 30, 0xff},
	}

	scene.Init()

	return scene
}

func (scene *SceneKeybindings) GetNext() SceneName {
	return scene.Next
}

func (scene *SceneKeybindings) SetPrevious(prev SceneName) {
	scene.Prev = prev
}

func (scene *SceneKeybindings) ResetNext() {
	scene.Next = scene.Whoami
}

func (scene *SceneKeybindings) SetNext(next SceneName) {
	scene.Next = next
}

func (scene *SceneKeybindings) Update() error {
	scene.Ui.Update()

	if inpututil.IsKeyJustPressed(ebiten.KeyEscape) || inpututil.IsKeyJustPressed(ebiten.KeyQ) {
		scene.Config.DelayedStart = false
		scene.Leave()
	}

	return nil

}

func (scene *SceneKeybindings) IsPrimary() bool {
	return false
}

func (scene *SceneKeybindings) Draw(screen *ebiten.Image) {
	scene.Ui.Draw(screen)
}

func (scene *SceneKeybindings) Leave() {
	scene.SetNext(Play)
}

func (scene *SceneKeybindings) Init() {
	rowContainer := NewRowContainer("Key Bindings")

	bindings := widget.NewText(
		widget.TextOpts.WidgetOpts(widget.WidgetOpts.LayoutData(widget.RowLayoutData{
			Stretch: true,
		})),
		widget.TextOpts.Text(KEYBINDINGS, *FontRenderer.FontSmall, color.NRGBA{0xdf, 0xf4, 0xff, 0xff}))

	cancel := NewMenuButton("Back",
		func(args *widget.ButtonClickedEventArgs) {
			scene.Leave()
		})

	rowContainer.AddChild(bindings)
	rowContainer.AddChild(cancel)

	scene.Ui = &ebitenui.UI{
		Container: rowContainer.Container(),
	}

}
