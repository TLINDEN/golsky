package main

import "github.com/hajimehoshi/ebiten/v2"

// Wrapper for  different screens  to be  shown, as  Welcome, Options,
// About, Menu Level and of course the actual game
// Scenes are responsible for screen clearing! That way a scene is able
// to render its content onto the running level, e.g. the options scene
// etc.

type SceneName int

type Scene interface {
	SetNext(SceneName)
	GetNext() SceneName
	ResetNext()
	Update() error
	Draw(screen *ebiten.Image)
	IsPrimary() bool // if true, this scene will be always drawn
}

const (
	Menu = iota // main top level menu
	Play        // actual playing happens here
	Options
)
