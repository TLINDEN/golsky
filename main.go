package main

import (
	"fmt"
	"image/color"
	"log"
	"math/rand"
	"os"
	"runtime"
	"strconv"
	"strings"

	"github.com/alecthomas/repr"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/hajimehoshi/ebiten/v2/vector"
	"github.com/spf13/pflag"
)

const (
	VERSION = "v0.0.4"
	Alive   = 1
	Dead    = 0
)

type Grid struct {
	Data [][]int
}

type Images struct {
	Black, White, Beige *ebiten.Image
}

type Game struct {
	Grids                             []*Grid // 2 grids: one current, one next
	History                           *Grid   // holds state of past dead cells for evolution tracks
	Index                             int     // points to current grid
	Width, Height, Cellsize, Density  int     // measurements
	ScreenWidth, ScreenHeight         int
	Generations                       int // Stats
	Black, White, Grey, Beige         color.RGBA
	TPG                               int    // ticks per generation/game speed, 1==max
	TicksElapsed                      int    // tick counter for game speed
	Debug, Paused, Empty, Invert      bool   // game modi
	ShowEvolution, NoGrid, RunOneStep bool   // flags
	Rule                              *Rule  // which rule to use, default: B3/S23
	Tiles                             Images // pre-computed tiles for dead and alife cells
}

func (game *Game) Layout(outsideWidth, outsideHeight int) (int, int) {
	return game.ScreenWidth, game.ScreenHeight
}

func (game *Game) CheckRule(state, neighbors int) int {
	var nextstate int

	// The standard Game of Life is symbolized in rule-string notation
	// as B3/S23 (23/3 here).  A cell  is born if it has exactly three
	// neighbors,  survives if it  has two or three  living neighbors,
	// and  dies otherwise. The first  number, or list of  numbers, is
	// what is required for a dead cell to be born.

	if state == 0 && Contains(game.Rule.Birth, neighbors) {
		nextstate = 1
	} else if state == 1 && Contains(game.Rule.Death, neighbors) {
		nextstate = 1
	} else {
		nextstate = 0
	}

	return nextstate
}

// find an item in a list, generic variant
func Contains[E comparable](s []E, v E) bool {
	for _, vs := range s {
		if v == vs {
			return true
		}
	}

	return false
}

// Update all cells according to the current rule
func (game *Game) UpdateCells() {
	// count ticks so we know when to actually run
	game.TicksElapsed++

	if game.TPG > game.TicksElapsed {
		// need to sleep a little more
		return
	}

	// next grid index, we just xor 0|1 to 1|0
	next := game.Index ^ 1

	// compute life status of cells
	for y := 0; y < game.Height; y++ {
		for x := 0; x < game.Width; x++ {
			state := game.Grids[game.Index].Data[y][x] // 0|1 == dead or alive
			neighbors := CountNeighbors(game, x, y)    // alive neighbor count

			// actually apply the current rules
			nextstate := game.CheckRule(state, neighbors)

			// change state of current cell in next grid
			game.Grids[next].Data[y][x] = nextstate

			if state == 1 {
				game.History.Data[y][x] = 1
			}
		}
	}

	// switch grid for rendering
	game.Index ^= 1

	// global stats counter
	game.Generations++

	if game.RunOneStep {
		// setp-wise mode, halt the game
		game.RunOneStep = false
	}

	// reset speed counter
	game.TicksElapsed = 0
}

// a GOL rule
type Rule struct {
	Birth []int
	Death []int
}

// parse one part of a GOL rule into rule slice
func NumbersToList(numbers string) []int {
	list := []int{}

	items := strings.Split(numbers, "")
	for _, item := range items {
		num, err := strconv.Atoi(item)
		if err != nil {
			log.Fatalf("failed to parse game rule part <%s>: %s", numbers, err)
		}

		list = append(list, num)
	}

	return list
}

// parse GOL rule, used in CheckRule()
func ParseGameRule(rule string) *Rule {
	parts := strings.Split(rule, "/")

	if len(parts) < 2 {
		log.Fatalf("Invalid game rule <%s>", rule)
	}

	golrule := &Rule{}

	for _, part := range parts {
		if part[0] == 'B' {
			golrule.Birth = NumbersToList(part[1:])
		} else {
			golrule.Death = NumbersToList(part[1:])
		}
	}

	return golrule
}

// check user input
func (game *Game) CheckInput() {
	if inpututil.IsKeyJustPressed(ebiten.KeyQ) {
		os.Exit(0)
	}

	if inpututil.IsKeyJustPressed(ebiten.KeySpace) || inpututil.IsKeyJustPressed(ebiten.KeyEnter) {
		game.Paused = !game.Paused
	}

	if ebiten.IsMouseButtonPressed(ebiten.MouseButtonLeft) {
		ToggleCell(game, Alive)
		game.Paused = true // drawing while running makes no sense
	}

	if ebiten.IsMouseButtonPressed(ebiten.MouseButtonRight) {
		ToggleCell(game, Dead)
		game.Paused = true // drawing while running makes no sense
	}

	if inpututil.IsKeyJustPressed(ebiten.KeyArrowDown) {
		if game.TPG < 120 {
			game.TPG++
		}
	}

	if inpututil.IsKeyJustPressed(ebiten.KeyArrowUp) {
		if game.TPG > 1 {
			game.TPG--
		}
	}

	if inpututil.IsKeyJustPressed(ebiten.KeyPageDown) {
		if game.TPG <= 115 {
			game.TPG += 5
		}
	}

	if inpututil.IsKeyJustPressed(ebiten.KeyPageUp) {
		switch {
		case game.TPG > 5:
			game.TPG -= 5
		case game.TPG <= 5:
			game.TPG = 1
		}
	}

	if game.Paused {
		if inpututil.IsKeyJustPressed(ebiten.KeyN) {
			game.RunOneStep = true
		}
	}
}

func (game *Game) Update() error {
	game.CheckInput()

	if !game.Paused || game.RunOneStep {
		game.UpdateCells()
	}

	return nil
}

// set a cell to alive or dead
func ToggleCell(game *Game, alive int) {
	xPX, yPX := ebiten.CursorPosition()
	x := xPX / game.Cellsize
	y := yPX / game.Cellsize

	//fmt.Printf("cell at %d,%d\n", x, y)

	game.Grids[game.Index].Data[y][x] = alive
	game.History.Data[y][x] = 1
}

// draw the new grid state
func (game *Game) Draw(screen *ebiten.Image) {
	// we  fill the whole  screen with  a background color,  the cells
	// themselfes will be 1px smaller as their nominal size, producing
	// a nice grey grid with grid lines
	op := &ebiten.DrawImageOptions{}
	if game.NoGrid {
		screen.Fill(game.White)
	} else {
		screen.Fill(game.Grey)
	}

	for y := 0; y < game.Height; y++ {
		for x := 0; x < game.Width; x++ {
			op.GeoM.Reset()
			op.GeoM.Translate(float64(x*game.Cellsize), float64(y*game.Cellsize))

			switch game.Grids[game.Index].Data[y][x] {
			case 1:

				screen.DrawImage(game.Tiles.Black, op)
			case 0:
				if game.History.Data[y][x] == 1 && game.ShowEvolution {
					screen.DrawImage(game.Tiles.Beige, op)
				} else {
					screen.DrawImage(game.Tiles.White, op)
				}
			}
		}
	}

	if game.Debug {
		paused := ""
		if game.Paused {
			paused = "-- paused --"
		}

		ebitenutil.DebugPrint(
			screen,
			fmt.Sprintf("FPS: %0.2f, TPG: %d, Mem: %0.2f MB, Generations: %d   %s",
				ebiten.ActualTPS(), game.TPG, GetMem(), game.Generations, paused),
		)
	}
}

func GetMem() float64 {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	return float64(m.Alloc) / 1024 / 1024
}

func (game *Game) InitGrid() {
	grid := &Grid{Data: make([][]int, game.Height)}
	gridb := &Grid{Data: make([][]int, game.Height)}
	history := &Grid{Data: make([][]int, game.Height)}

	for y := 0; y < game.Height; y++ {
		grid.Data[y] = make([]int, game.Width)
		gridb.Data[y] = make([]int, game.Width)
		history.Data[y] = make([]int, game.Width)
		if !game.Empty {
			for x := 0; x < game.Width; x++ {
				if rand.Intn(game.Density) == 1 {
					history.Data[y][x] = 1
					grid.Data[y][x] = 1
				}
			}
		}
	}

	game.Grids = []*Grid{
		grid,
		gridb,
	}

	game.History = history
}

// fill a cell with the given color
func FillCell(tile *ebiten.Image, cellsize int, col color.RGBA) {
	vector.DrawFilledRect(
		tile,
		float32(1),
		float32(1),
		float32(cellsize-1),
		float32(cellsize-1),
		col, false,
	)
}

// prepare tile images
func (game *Game) InitTiles() {
	game.Black = color.RGBA{0, 0, 0, 0xff}
	game.White = color.RGBA{200, 200, 200, 0xff}
	game.Grey = color.RGBA{128, 128, 128, 0xff}
	game.Beige = color.RGBA{0xff, 0xf8, 0xdc, 0xff}

	if game.Invert {
		game.White = color.RGBA{0, 0, 0, 0xff}
		game.Black = color.RGBA{200, 200, 200, 0xff}
		//game.Beige = color.RGBA{0x8b, 0x1a, 0x1a, 0xff}
		game.Beige = color.RGBA{0x30, 0x1c, 0x11, 0xff}
	}

	game.Tiles.Beige = ebiten.NewImage(game.Cellsize, game.Cellsize)
	game.Tiles.Black = ebiten.NewImage(game.Cellsize, game.Cellsize)
	game.Tiles.White = ebiten.NewImage(game.Cellsize, game.Cellsize)

	cellsize := game.ScreenWidth / game.Cellsize
	FillCell(game.Tiles.Beige, cellsize, game.Beige)
	FillCell(game.Tiles.Black, cellsize, game.Black)
	FillCell(game.Tiles.White, cellsize, game.White)
}

func (game *Game) Init() {
	// setup the game
	game.ScreenWidth = game.Cellsize * game.Width
	game.ScreenHeight = game.Cellsize * game.Height

	game.InitGrid()
	game.InitTiles()

	game.Index = 0
	game.TicksElapsed = 0
}

// count the living neighbors of a cell
func CountNeighbors(game *Game, x, y int) int {
	sum := 0

	// so we look ad all 8 neighbors surrounding us. In case we are on
	// an edge, then  we'll look at the neighbor on  the other side of
	// the grid, thus wrapping lookahead around.
	for i := -1; i < 2; i++ {
		for j := -1; j < 2; j++ {
			col := (x + i + game.Width) % game.Width
			row := (y + j + game.Height) % game.Height
			sum += game.Grids[game.Index].Data[row][col]
		}
	}

	// don't count ourselfes though
	sum -= game.Grids[game.Index].Data[y][x]

	return sum
}

func main() {
	game := &Game{}
	showversion := false
	var rule string

	pflag.IntVarP(&game.Width, "width", "W", 40, "grid width in cells")
	pflag.IntVarP(&game.Height, "height", "H", 40, "grid height in cells")
	pflag.IntVarP(&game.Cellsize, "cellsize", "c", 8, "cell size in pixels")
	pflag.IntVarP(&game.Density, "density", "D", 10, "density of random cells")
	pflag.IntVarP(&game.TPG, "ticks-per-generation", "t", 10, "game speed: the higher the slower (default: 10)")

	pflag.StringVarP(&rule, "rule", "r", "B3/S23", "game rule")

	pflag.BoolVarP(&showversion, "version", "v", false, "show version")
	pflag.BoolVarP(&game.Paused, "paused", "p", false, "do not start simulation (use space to start)")
	pflag.BoolVarP(&game.Debug, "debug", "d", false, "show debug info")
	pflag.BoolVarP(&game.NoGrid, "nogrid", "n", false, "do not draw grid lines")
	pflag.BoolVarP(&game.Empty, "empty", "e", false, "start with an empty screen")
	pflag.BoolVarP(&game.Invert, "invert", "i", false, "invert colors (dead cell: black)")
	pflag.BoolVarP(&game.ShowEvolution, "show-evolution", "s", false, "show evolution tracks")

	pflag.Parse()

	if showversion {
		fmt.Printf("This is gameoflife version %s\n", VERSION)
		os.Exit(0)
	}

	game.Rule = ParseGameRule(rule)

	repr.Print(game.TPG)
	game.Init()

	ebiten.SetWindowSize(game.ScreenWidth, game.ScreenHeight)
	ebiten.SetWindowTitle("Game of life")
	ebiten.SetWindowResizingMode(ebiten.WindowResizingModeEnabled)

	if err := ebiten.RunGame(game); err != nil {
		log.Fatal(err)
	}
}
