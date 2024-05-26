package main

import (
	"fmt"
	"log"
	"os"

	"github.com/tlinden/golsky/rle"

	"github.com/hajimehoshi/ebiten/v2"
)

const (
	VERSION = "v0.0.6"
	Alive   = 1
	Dead    = 0
)

func GetRLE(filename string) *rle.RLE {
	if filename == "" {
		return nil
	}

	content, err := os.ReadFile(filename)
	if err != nil {
		log.Fatal(err)
	}

	parsedRle, err := rle.Parse(string(content))
	if err != nil {
		log.Fatalf("failed to load RLE pattern file: %s", err)
	}

	return &parsedRle
}

func main() {
	config := ParseCommandline()

	if config.ShowVersion {
		fmt.Printf("This is golsky version %s\n", VERSION)
		os.Exit(0)
	}

	game := NewGame(config, Play)

	// setup environment
	ebiten.SetWindowSize(game.ScreenWidth, game.ScreenHeight)
	ebiten.SetWindowTitle("Game of life")
	ebiten.SetWindowResizingMode(ebiten.WindowResizingModeEnabled)

	// main loop
	if err := ebiten.RunGame(game); err != nil {
		log.Fatal(err)
	}
}
