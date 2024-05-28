package main

import (
	"errors"
	"fmt"
	"os"
	"runtime/pprof"
	"strconv"
	"strings"

	"github.com/alecthomas/repr"
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
	UseShader                         bool // to use a shader to render alife cells

	// for internal profiling
	ProfileFile     string
	ProfileDraw     bool
	ProfileMaxLoops int64
}

const (
	VERSION = "v0.0.7"
	Alive   = 1
	Dead    = 0
)

// parse given window geometry and adjust game settings according to it
func (config *Config) ParseGeom(geom string) error {
	if geom == "" {
		config.ScreenWidth = config.Cellsize * config.Width
		config.ScreenHeight = config.Cellsize * config.Height
		return nil
	}

	// force a geom
	geometry := strings.Split(geom, "x")
	if len(geometry) != 2 {
		return errors.New("failed to parse -g parameters, expecting WIDTHxHEIGHT")
	}

	width, err := strconv.Atoi(geometry[0])
	if err != nil {
		return errors.New("failed to parse width, expecting integer")
	}

	height, err := strconv.Atoi(geometry[1])
	if err != nil {
		return errors.New("failed to parse height, expecting integer")
	}

	// adjust dimensions, account for  grid width+height so that cells
	// fit into window
	config.ScreenWidth = width - (width % config.Width)
	config.ScreenHeight = height - (height % config.Height)

	if config.ScreenWidth == 0 || config.ScreenHeight == 0 {
		return errors.New("the number of requested cells don't fit into the requested window size")
	}

	config.Cellsize = config.ScreenWidth / config.Width

	repr.Println(config)
	return nil
}

// check if we have  been given an RLE file to load,  then load it and
// adjust game settings accordingly
func (config *Config) ParseRLE(rlefile string) error {
	if rlefile == "" {
		return nil
	}

	rleobj, err := rle.GetRLE(rlefile)
	if err != nil {
		return err
	}

	if rleobj == nil {
		return errors.New("failed to load RLE file (uncatched module error)")
	}

	config.RLE = rleobj

	// adjust geometry if needed
	if config.RLE.Width > config.Width || config.RLE.Height > config.Height {
		config.Width = config.RLE.Width * 2
		config.Height = config.RLE.Height * 2
		config.Cellsize = config.ScreenWidth / config.Width
	}

	// RLE needs an empty grid
	config.Empty = true

	// it may come with its own rule
	if config.RLE.Rule != "" {
		config.Rule = ParseGameRule(config.RLE.Rule)
	}

	return nil
}

// parse a state file, if given, and adjust game settings accordingly
func (config *Config) ParseStatefile(statefile string) error {
	if config.Statefile == "" {
		return nil
	}

	grid, err := LoadState(config.Statefile)
	if err != nil {
		return fmt.Errorf("failed to load game state: %s", err)
	}

	config.Width = grid.Width
	config.Height = grid.Height
	config.Cellsize = config.ScreenWidth / config.Width
	config.StateGrid = grid

	return nil
}

func (config *Config) EnableCPUProfiling(filename string) error {
	if filename == "" {
		return nil
	}

	fd, err := os.Create(filename)
	if err != nil {
		return err
	}

	pprof.StartCPUProfile(fd)
	defer pprof.StopCPUProfile()

	return nil
}

func ParseCommandline() (*Config, error) {
	config := Config{}

	var (
		rule, rlefile, geom string
	)

	// commandline params, most configure directly config flags
	pflag.IntVarP(&config.Width, "width", "W", 40, "grid width in cells")
	pflag.IntVarP(&config.Height, "height", "H", 40, "grid height in cells")
	pflag.IntVarP(&config.Cellsize, "cellsize", "c", 8, "cell size in pixels")
	pflag.StringVarP(&geom, "geom", "g", "", "window geometry in WxH in pixels, overturns -c")

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
	pflag.BoolVarP(&config.UseShader, "use-shader", "k", false, "use shader for cell rendering")

	pflag.StringVarP(&config.ProfileFile, "profile-file", "", "", "enable profiling")
	pflag.BoolVarP(&config.ProfileDraw, "profile-draw", "", false, "profile draw method (default false)")
	pflag.Int64VarP(&config.ProfileMaxLoops, "profile-max-loops", "", 10, "how many loops to execute (default 10)")

	pflag.Parse()

	err := config.ParseGeom(geom)
	if err != nil {
		return nil, err
	}

	err = config.ParseRLE(rlefile)
	if err != nil {
		return nil, err
	}

	// load  rule from commandline  when no  rule came from  RLE file,
	// default is B3/S23, aka conways game of life
	if config.Rule == nil {
		config.Rule = ParseGameRule(rule)
	}

	return &config, nil
}
