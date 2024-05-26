package main

import (
	"fmt"
	"log"
	"os"

	"github.com/tlinden/golsky/rle"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/spf13/pflag"
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
	game := &Game{}
	showversion := false
	var rule string
	var rlefile string

	// commandline params, most configure directly game flags
	pflag.IntVarP(&game.Width, "width", "W", 40, "grid width in cells")
	pflag.IntVarP(&game.Height, "height", "H", 40, "grid height in cells")
	pflag.IntVarP(&game.Cellsize, "cellsize", "c", 8, "cell size in pixels")
	pflag.IntVarP(&game.Density, "density", "D", 10, "density of random cells")
	pflag.IntVarP(&game.TPG, "ticks-per-generation", "t", 10,
		"game speed: the higher the slower (default: 10)")

	pflag.StringVarP(&rule, "rule", "r", "B3/S23", "game rule")
	pflag.StringVarP(&rlefile, "rle-file", "f", "", "RLE pattern file")
	pflag.StringVarP(&game.Statefile, "load-state-file", "l", "", "game state file")

	pflag.BoolVarP(&showversion, "version", "v", false, "show version")
	pflag.BoolVarP(&game.Paused, "paused", "p", false, "do not start simulation (use space to start)")
	pflag.BoolVarP(&game.Debug, "debug", "d", false, "show debug info")
	pflag.BoolVarP(&game.NoGrid, "nogrid", "n", false, "do not draw grid lines")
	pflag.BoolVarP(&game.Empty, "empty", "e", false, "start with an empty screen")
	pflag.BoolVarP(&game.Invert, "invert", "i", false, "invert colors (dead cell: black)")
	pflag.BoolVarP(&game.ShowEvolution, "show-evolution", "s", false, "show evolution tracks")
	pflag.BoolVarP(&game.Wrap, "wrap-around", "w", false, "wrap around grid mode")

	pflag.Parse()

	if showversion {
		fmt.Printf("This is golsky version %s\n", VERSION)
		os.Exit(0)
	}

	// check if we have been given an RLE file to load
	game.RLE = GetRLE(rlefile)
	if game.RLE != nil {
		if game.RLE.Width > game.Width || game.RLE.Height > game.Height {
			game.Width = game.RLE.Width * 2
			game.Height = game.RLE.Height * 2
			fmt.Printf("rlew: %d, rleh: %d, w: %d, h: %d\n",
				game.RLE.Width, game.RLE.Height, game.Width, game.Height)
		}

		// RLE needs an empty grid
		game.Empty = true

		// it may come with its own rule
		if game.RLE.Rule != "" {
			game.Rule = ParseGameRule(game.RLE.Rule)
		}
	}

	// load  rule from commandline  when no  rule came from  RLE file,
	// default is B3/S23, aka conways game of life
	if game.Rule == nil {
		game.Rule = ParseGameRule(rule)
	}

	// bootstrap the game
	game.Init()

	// setup environment
	ebiten.SetWindowSize(game.ScreenWidth, game.ScreenHeight)
	ebiten.SetWindowTitle("Game of life")
	ebiten.SetWindowResizingMode(ebiten.WindowResizingModeEnabled)

	// main loop
	if err := ebiten.RunGame(game); err != nil {
		log.Fatal(err)
	}
}
