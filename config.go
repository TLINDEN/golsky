package main

import (
	"fmt"
	"log"
	"os"

	"github.com/spf13/pflag"
	"github.com/tlinden/golsky/rle"
)

// all the settings comming from commandline, but maybe tweaked later from the UI
type Config struct {
	Width, Height, Cellsize, Density  int // measurements
	ScreenWidth, ScreenHeight         int
	TPG                               int      // ticks per generation/game speed, 1==max
	Debug, Empty, Invert, Paused      bool     // game modi
	ShowEvolution, NoGrid, RunOneStep bool     // flags
	Rule                              *Rule    // which rule to use, default: B3/S23
	RLE                               *rle.RLE // loaded GOL pattern from RLE file
	Statefile                         string   // load game state from it if non-nil
	StateGrid                         *Grid    // a grid from a statefile
	Wrap                              bool     // wether wraparound mode is in place or not
	ShowVersion                       bool
}

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

func ParseCommandline() *Config {
	config := Config{}

	var rule string
	var rlefile string

	// commandline params, most configure directly config flags
	pflag.IntVarP(&config.Width, "width", "W", 40, "grid width in cells")
	pflag.IntVarP(&config.Height, "height", "H", 40, "grid height in cells")
	pflag.IntVarP(&config.Cellsize, "cellsize", "c", 8, "cell size in pixels")
	pflag.IntVarP(&config.Density, "density", "D", 10, "density of random cells")
	pflag.IntVarP(&config.TPG, "ticks-per-generation", "t", 10,
		"game speed: the higher the slower (default: 10)")

	pflag.StringVarP(&rule, "rule", "r", "B3/S23", "game rule")
	pflag.StringVarP(&rlefile, "rle-file", "f", "", "RLE pattern file")
	pflag.StringVarP(&config.Statefile, "load-state-file", "l", "", "game state file")

	pflag.BoolVarP(&config.ShowVersion, "version", "v", false, "show version")
	pflag.BoolVarP(&config.Paused, "paused", "p", false, "do not start simulation (use space to start)")
	pflag.BoolVarP(&config.Debug, "debug", "d", false, "show debug info")
	pflag.BoolVarP(&config.NoGrid, "nogrid", "n", false, "do not draw grid lines")
	pflag.BoolVarP(&config.Empty, "empty", "e", false, "start with an empty screen")
	pflag.BoolVarP(&config.Invert, "invert", "i", false, "invert colors (dead cell: black)")
	pflag.BoolVarP(&config.ShowEvolution, "show-evolution", "s", false, "show evolution tracks")
	pflag.BoolVarP(&config.Wrap, "wrap-around", "w", false, "wrap around grid mode")

	pflag.Parse()

	// check if we have been given an RLE file to load
	config.RLE = GetRLE(rlefile)
	if config.RLE != nil {
		if config.RLE.Width > config.Width || config.RLE.Height > config.Height {
			config.Width = config.RLE.Width * 2
			config.Height = config.RLE.Height * 2
			fmt.Printf("rlew: %d, rleh: %d, w: %d, h: %d\n",
				config.RLE.Width, config.RLE.Height, config.Width, config.Height)
		}

		// RLE needs an empty grid
		config.Empty = true

		// it may come with its own rule
		if config.RLE.Rule != "" {
			config.Rule = ParseGameRule(config.RLE.Rule)
		}
	} else if config.Statefile != "" {
		grid, err := LoadState(config.Statefile)
		if err != nil {
			log.Fatalf("failed to load game state: %s", err)
		}

		config.Width = grid.Width
		config.Height = grid.Height
		config.StateGrid = grid
	}

	config.ScreenWidth = config.Cellsize * config.Width
	config.ScreenHeight = config.Cellsize * config.Height

	// load  rule from commandline  when no  rule came from  RLE file,
	// default is B3/S23, aka conways game of life
	if config.Rule == nil {
		config.Rule = ParseGameRule(rule)
	}

	return &config
}
