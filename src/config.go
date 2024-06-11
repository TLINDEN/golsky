package main

import (
	"errors"
	"fmt"
	"math"
	"os"
	"runtime/pprof"
	"strconv"
	"strings"

	"github.com/spf13/pflag"
	"github.com/tlinden/golsky/rle"
)

// all the settings comming from commandline, but maybe tweaked later from the UI
type Config struct {
	Width, Height, Cellsize, Density         int // measurements
	ScreenWidth, ScreenHeight                int
	TPG                                      int      // ticks per generation/game speed, 1==max
	Debug, Empty, Paused, Markmode, Drawmode bool     // game modi
	ShowEvolution, ShowGrid, RunOneStep      bool     // flags
	Rule                                     *Rule    // which rule to use, default: B3/S23
	RLE                                      *rle.RLE // loaded GOL pattern from RLE file
	Statefile                                string   // load game state from it if non-nil
	StateGrid                                *Grid    // a grid from a statefile
	Wrap                                     bool     // wether wraparound mode is in place or not
	ShowVersion                              bool
	UseShader                                bool // to use a shader to render alife cells
	Restart, RestartGrid, RestartCache       bool
	StartWithMenu                            bool
	Zoomfactor                               int
	ZoomOutFactor                            int
	InitialCamPos                            []float64
	DelayedStart                             bool // if true game, we wait. like pause but program induced
	Theme                                    string
	ThemeManager                             ThemeManager

	// for internal profiling
	ProfileFile     string
	ProfileDraw     bool
	ProfileMaxLoops int64
}

const (
	VERSION = "v0.0.8"
	Alive   = true
	Dead    = false

	DEFAULT_GRID_WIDTH  = 600
	DEFAULT_GRID_HEIGHT = 400
	DEFAULT_CELLSIZE    = 4
	DEFAULT_ZOOMFACTOR  = 400
	DEFAULT_GEOM        = "640x384"
	DEFAULT_THEME       = "standard" // "light" // inverse => "dark"
)

const KEYBINDINGS string = `
- SPACE: pause or resume the game
- N: while game is paused: forward one step
- PAGE UP: speed up
- PAGE DOWN: slow down
- MOUSE WHEEL: zoom in or out
- LEFT MOUSE BUTTON: use to drag canvas, keep clicked and move mouse
- I: enter "insert" (draw) mode: use left mouse to set cells alife and right
     button to dead. Leave with "space". While in insert mode, use middle mouse
     button to drag grid.
- R: reset to 1:1 zoom
- ESCAPE: open menu, o: open options menu
- S: save game state to file (can be loaded with -l)
- C: enter mark mode. Mark a rectangle with the mouse, when you
     release the mouse buttonx it is being saved to an RLE file
- D: toggle debug output 
- Q: quit game
`

func (config *Config) SetupCamera() {
	config.Zoomfactor = DEFAULT_ZOOMFACTOR / config.Cellsize

	// calculate the initial cam pos. It is negative if the total grid
	// size  is smaller than  the screen  in a centered  position, but
	// it's zero if it's equal or larger than the screen.
	config.InitialCamPos = make([]float64, 2)

	config.InitialCamPos[0] = float64(((config.ScreenWidth - (config.Width * config.Cellsize)) / 2) * -1)
	if config.Width*config.Cellsize >= config.ScreenWidth {
		// must be positive if world wider than screen
		config.InitialCamPos[0] = math.Abs(config.InitialCamPos[0])
	}

	// same for Y
	config.InitialCamPos[1] = float64(((config.ScreenHeight - (config.Height * config.Cellsize)) / 2) * -1)
	if config.Height*config.Cellsize > config.ScreenHeight {
		config.InitialCamPos[1] = math.Abs(config.InitialCamPos[1])
	}

	// Calculate  zoom out factor, which  shows 100% of the  world. We
	// need to reverse math.Pow(1.01,  $zoomfactor) to get the correct
	// percentage of  the world to show.  I.e: with  a ScreenHeight of
	// 384px and a world of 800px the factor to show 100% of the world
	// is  -75: math.Log(384/800) / math.Log(1.01).  The 1.01 constant
	// is being used in camera.go:worldMatrix().

	// FIXME: determine if the diff is larger on width, then calc with
	// width instead of height
	config.ZoomOutFactor = int(
		math.Log(float64(config.ScreenHeight)/(float64(config.Height)*float64(config.Cellsize))) /
			math.Log(1.01))
}

// parse given window geometry and adjust game settings according to it
func (config *Config) ParseGeom(geom string) error {
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

	config.ScreenWidth = width
	config.ScreenHeight = height

	//config.Cellsize = DEFAULT_CELLSIZE

	return nil
}

// check if we have  been given an RLE or LIF file  to load, then load
// it and adjust game settings accordingly
func (config *Config) ParseRLE(rlefile string) error {
	if rlefile == "" {
		return nil
	}

	var rleobj *rle.RLE

	if strings.HasSuffix(rlefile, ".lif") {
		lifobj, err := LoadLIF(rlefile)
		if err != nil {
			return err
		}

		rleobj = lifobj
	} else {
		rleobject, err := rle.GetRLE(rlefile)
		if err != nil {
			return err
		}

		rleobj = rleobject
	}

	if rleobj == nil {
		return errors.New("failed to load pattern file (uncatched module error)")
	}

	config.RLE = rleobj

	// adjust geometry if needed
	if config.RLE.Width > config.Width || config.RLE.Height > config.Height {
		config.Width = config.RLE.Width * 2
		config.Height = config.RLE.Height * 2
		config.Cellsize = config.ScreenWidth / config.Width
	}

	fmt.Printf("width: %d, screenwidth: %d, rlewidth: %d, cellsize: %d\n",
		config.Width, config.ScreenWidth, config.RLE.Width, config.Cellsize)

	// RLE needs an empty grid
	config.Empty = true

	// it may come with its own rule
	if config.RLE.Rule != "" {
		config.Rule = ParseGameRule(config.RLE.Rule)
	}

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
	pflag.IntVarP(&config.Width, "width", "W", DEFAULT_GRID_WIDTH, "grid width in cells")
	pflag.IntVarP(&config.Height, "height", "H", DEFAULT_GRID_HEIGHT, "grid height in cells")
	pflag.IntVarP(&config.Cellsize, "cellsize", "c", 8, "cell size in pixels")
	pflag.StringVarP(&geom, "geom", "G", DEFAULT_GEOM, "window geometry in WxH in pixels, overturns -c")

	pflag.IntVarP(&config.Density, "density", "D", 10, "density of random cells")
	pflag.IntVarP(&config.TPG, "ticks-per-generation", "t", 10,
		"game speed: the higher the slower (default: 10)")

	pflag.StringVarP(&rule, "rule", "r", "B3/S23", "game rule")
	pflag.StringVarP(&rlefile, "pattern-file", "f", "", "RLE or LIF pattern file")

	pflag.BoolVarP(&config.ShowVersion, "version", "v", false, "show version")
	pflag.BoolVarP(&config.ShowGrid, "show-grid", "g", false, "draw grid lines")
	pflag.BoolVarP(&config.ShowEvolution, "show-evolution", "s", false, "show evolution traces")

	pflag.BoolVarP(&config.Paused, "paused", "p", false, "do not start simulation (use space to start)")
	pflag.BoolVarP(&config.Debug, "debug", "d", false, "show debug info")
	pflag.BoolVarP(&config.Empty, "empty", "e", false, "start with an empty screen")

	// style
	pflag.StringVarP(&config.Theme, "theme", "T", DEFAULT_THEME, "color theme: standard, dark, light (default: standard)")

	pflag.BoolVarP(&config.Wrap, "wrap-around", "w", false, "wrap around grid mode")
	pflag.BoolVarP(&config.UseShader, "use-shader", "k", false, "use shader for cell rendering")

	pflag.StringVarP(&config.ProfileFile, "profile-file", "", "", "enable profiling")

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

	config.SetupCamera()

	config.ThemeManager = NewThemeManager(config.Theme, config.Cellsize)

	//repr.Println(config)
	return &config, nil
}

func (config *Config) TogglePaused() {
	config.Paused = !config.Paused
}

func (config *Config) ToggleDebugging() {
	config.Debug = !config.Debug
}

func (config *Config) SwitchTheme(theme string) {
	config.ThemeManager.SetCurrentTheme(theme)
	config.RestartCache = true
}

func (config *Config) ToggleGridlines() {
	config.ShowGrid = !config.ShowGrid
	config.RestartCache = true
}

func (config *Config) ToggleEvolution() {
	config.ShowEvolution = !config.ShowEvolution
}

func (config *Config) ToggleWrap() {
	config.Wrap = !config.Wrap
}
