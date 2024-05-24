package main

import (
	"fmt"
	"image/color"
	"log"
	"math/rand"
	"os"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/hajimehoshi/ebiten/v2/vector"
	"github.com/tlinden/golsky/rle"
	"golang.org/x/image/math/f64"
)

type Images struct {
	Black, White, Age1, Age2, Age3, Age4, Old *ebiten.Image
}

type Game struct {
	Grids                                      []*Grid // 2 grids: one current, one next
	History                                    *Grid   // holds state of past dead cells for evolution tracks
	Index                                      int     // points to current grid
	Width, Height, Cellsize, Density           int     // measurements
	ScreenWidth, ScreenHeight                  int
	Generations                                int64 // Stats
	Black, White, Grey, Old                    color.RGBA
	AgeColor1, AgeColor2, AgeColor3, AgeColor4 color.RGBA
	TPG                                        int           // ticks per generation/game speed, 1==max
	TicksElapsed                               int           // tick counter for game speed
	Debug, Paused, Empty, Invert               bool          // game modi
	ShowEvolution, NoGrid, RunOneStep          bool          // flags
	Rule                                       *Rule         // which rule to use, default: B3/S23
	Tiles                                      Images        // pre-computed tiles for dead and alife cells
	RLE                                        *rle.RLE      // loaded GOL pattern from RLE file
	Camera                                     Camera        // for zoom+move
	World                                      *ebiten.Image // actual image we render to
	WheelTurned                                bool          // when user turns wheel multiple times, zoom faster
	Dragging                                   bool          // middle mouse is pressed, move canvas
	LastCursorPos                              []int         // used to check if the user is dragging
	Statefile                                  string        // load game state from it if non-nil
}

func (game *Game) Layout(outsideWidth, outsideHeight int) (int, int) {
	return game.ScreenWidth, game.ScreenHeight
}

func (game *Game) CheckRule(state int64, neighbors int64) int64 {
	var nextstate int64

	// The standard Game of Life is symbolized in rule-string notation
	// as B3/S23 (23/3 here).  A cell  is born if it has exactly three
	// neighbors,  survives if it  has two or three  living neighbors,
	// and  dies otherwise. The first  number, or list of  numbers, is
	// what is required for a dead cell to be born.

	if state == 0 && Contains(game.Rule.Birth, neighbors) {
		nextstate = Alive
	} else if state == 1 && Contains(game.Rule.Death, neighbors) {
		nextstate = Alive
	} else {
		nextstate = Dead
	}

	return nextstate
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
			neighbors := game.CountNeighbors(x, y)     // alive neighbor count

			// actually apply the current rules
			nextstate := game.CheckRule(state, neighbors)

			// change state of current cell in next grid
			game.Grids[next].Data[y][x] = nextstate

			// set history  to current generation so we  can infer the
			// age of the cell's state  during rendering and use it to
			// deduce the color to use if evolution tracking is enabled
			if state != nextstate {
				game.History.Data[y][x] = game.Generations
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

// check user input
func (game *Game) CheckInput() {
	if inpututil.IsKeyJustPressed(ebiten.KeyQ) {
		os.Exit(0)
	}

	if inpututil.IsKeyJustPressed(ebiten.KeySpace) || inpututil.IsKeyJustPressed(ebiten.KeyEnter) {
		game.Paused = !game.Paused
	}

	if ebiten.IsMouseButtonPressed(ebiten.MouseButtonLeft) {
		game.ToggleCellOnCursorPos(Alive)
		game.Paused = true // drawing while running makes no sense
	}

	if ebiten.IsMouseButtonPressed(ebiten.MouseButtonRight) {
		game.ToggleCellOnCursorPos(Dead)
		game.Paused = true // drawing while running makes no sense
	}

	if ebiten.IsKeyPressed(ebiten.KeyPageDown) {
		if game.TPG < 120 {
			game.TPG++
		}
	}

	if ebiten.IsKeyPressed(ebiten.KeyPageUp) {
		if game.TPG > 1 {
			game.TPG--
		}
	}

	if inpututil.IsKeyJustPressed(ebiten.KeyS) {
		filename := GetFilename(game.Generations)
		err := game.Grids[game.Index].SaveState(filename)
		if err != nil {
			log.Printf("failed to save game state to %s: %s", filename, err)
		}
		log.Printf("saved game state to %s at generation %d\n", filename, game.Generations)
	}

	if game.Paused {
		if inpututil.IsKeyJustPressed(ebiten.KeyN) {
			game.RunOneStep = true
		}
	}
}

// Check dragging input.  move the canvas with the  mouse while pressing
// the middle mouse button, zoom in and out using the wheel.
func (game *Game) CheckDraggingInput() {
	// move canvas
	if game.Dragging && !ebiten.IsMouseButtonPressed(ebiten.MouseButton1) {
		// release
		game.Dragging = false
	}

	if !game.Dragging && ebiten.IsMouseButtonPressed(ebiten.MouseButton1) {
		// start dragging
		game.Dragging = true
		game.LastCursorPos[0], game.LastCursorPos[1] = ebiten.CursorPosition()
	}

	if game.Dragging {
		x, y := ebiten.CursorPosition()

		if x != game.LastCursorPos[0] || y != game.LastCursorPos[1] {
			// actually drag by mouse cursor pos diff to last cursor pos
			game.Camera.Position[0] -= float64(x - game.LastCursorPos[0])
			game.Camera.Position[1] -= float64(y - game.LastCursorPos[1])
		}

		game.LastCursorPos[0], game.LastCursorPos[1] = ebiten.CursorPosition()
	}

	// also support the arrow keys to move the canvas
	if ebiten.IsKeyPressed(ebiten.KeyArrowLeft) {
		game.Camera.Position[0] -= 1
	}
	if ebiten.IsKeyPressed(ebiten.KeyArrowRight) {
		game.Camera.Position[0] += 1
	}
	if ebiten.IsKeyPressed(ebiten.KeyArrowUp) {
		game.Camera.Position[1] -= 1
	}
	if ebiten.IsKeyPressed(ebiten.KeyArrowDown) {
		game.Camera.Position[1] += 1
	}

	// Zoom
	_, dy := ebiten.Wheel()
	step := 1

	if game.WheelTurned {
		// if keep scrolling the wheel, zoom faster
		step = 50
	} else {
		game.WheelTurned = false
	}

	if dy < 0 {
		if game.Camera.ZoomFactor > -2400 {
			game.Camera.ZoomFactor -= step
		}
	}

	if dy > 0 {
		if game.Camera.ZoomFactor < 2400 {
			game.Camera.ZoomFactor += step
		}
	}

	if inpututil.IsKeyJustPressed(ebiten.KeyEscape) {
		game.Camera.Reset()
	}

}

func (game *Game) Update() error {
	game.CheckInput()
	game.CheckDraggingInput()

	if !game.Paused || game.RunOneStep {
		game.UpdateCells()
	}

	return nil
}

// set a cell to alive or dead
func (game *Game) ToggleCellOnCursorPos(alive int64) {
	// use cursor pos relative to the world
	worldX, worldY := game.Camera.ScreenToWorld(ebiten.CursorPosition())
	x := int(worldX) / game.Cellsize
	y := int(worldY) / game.Cellsize

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
		game.World.Fill(game.White)
	} else {
		game.World.Fill(game.Grey)
	}

	for y := 0; y < game.Height; y++ {
		for x := 0; x < game.Width; x++ {
			op.GeoM.Reset()
			op.GeoM.Translate(float64(x*game.Cellsize), float64(y*game.Cellsize))

			age := game.Generations - game.History.Data[y][x]

			switch game.Grids[game.Index].Data[y][x] {
			case 1:
				if age > 50 && game.ShowEvolution {
					game.World.DrawImage(game.Tiles.Old, op)
				} else {
					game.World.DrawImage(game.Tiles.Black, op)
				}
			case 0:
				if game.History.Data[y][x] > 1 && game.ShowEvolution {
					switch {
					case age < 10:
						game.World.DrawImage(game.Tiles.Age1, op)
					case age < 20:
						game.World.DrawImage(game.Tiles.Age2, op)
					case age < 30:
						game.World.DrawImage(game.Tiles.Age3, op)
					default:
						game.World.DrawImage(game.Tiles.Age4, op)
					}
				} else {
					game.World.DrawImage(game.Tiles.White, op)
				}
			}
		}
	}

	game.Camera.Render(game.World, screen)

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

// FIXME: move these into Grid
// load a pre-computed pattern from RLE file
func (game *Game) InitPattern() {
	if game.RLE != nil {
		startX := (game.Width / 2) - (game.RLE.Width / 2)
		startY := (game.Height / 2) - (game.RLE.Height / 2)
		var y, x int

		for rowIndex, patternRow := range game.RLE.Pattern {
			for colIndex := range patternRow {
				if game.RLE.Pattern[rowIndex][colIndex] > 0 {
					x = colIndex + startX
					y = rowIndex + startY

					game.History.Data[y][x] = 1
					game.Grids[0].Data[y][x] = 1
				}
			}
		}
	}
}

// initialize playing field/grid
func (game *Game) InitGrid(grid *Grid) {
	if grid != nil {
		// use pre-loaded grid
		game.Grids = []*Grid{
			grid,
			NewGrid(grid.Width, grid.Height),
		}

		game.History = NewGrid(grid.Width, grid.Height)

		return
	}

	grida := NewGrid(game.Width, game.Height)
	gridb := NewGrid(game.Width, game.Height)
	history := NewGrid(game.Width, game.Height)

	for y := 0; y < game.Height; y++ {
		if !game.Empty {
			for x := 0; x < game.Width; x++ {
				if rand.Intn(game.Density) == 1 {
					history.Data[y][x] = 1
					grida.Data[y][x] = 1
				}
			}
		}
	}

	game.Grids = []*Grid{
		grida,
		gridb,
	}

	game.History = history
}

// prepare tile images
func (game *Game) InitTiles() {
	game.Grey = color.RGBA{128, 128, 128, 0xff}
	game.Old = color.RGBA{255, 30, 30, 0xff}

	game.Black = color.RGBA{0, 0, 0, 0xff}
	game.White = color.RGBA{200, 200, 200, 0xff}
	game.AgeColor1 = color.RGBA{255, 195, 97, 0xff} // FIXME: use slice!
	game.AgeColor2 = color.RGBA{255, 211, 140, 0xff}
	game.AgeColor3 = color.RGBA{255, 227, 181, 0xff}
	game.AgeColor4 = color.RGBA{255, 240, 224, 0xff}

	if game.Invert {
		game.White = color.RGBA{0, 0, 0, 0xff}
		game.Black = color.RGBA{200, 200, 200, 0xff}

		game.AgeColor1 = color.RGBA{82, 38, 0, 0xff}
		game.AgeColor2 = color.RGBA{66, 35, 0, 0xff}
		game.AgeColor3 = color.RGBA{43, 27, 0, 0xff}
		game.AgeColor4 = color.RGBA{25, 17, 0, 0xff}
	}

	game.Tiles.Black = ebiten.NewImage(game.Cellsize, game.Cellsize)
	game.Tiles.White = ebiten.NewImage(game.Cellsize, game.Cellsize)
	game.Tiles.Old = ebiten.NewImage(game.Cellsize, game.Cellsize)
	game.Tiles.Age1 = ebiten.NewImage(game.Cellsize, game.Cellsize)
	game.Tiles.Age2 = ebiten.NewImage(game.Cellsize, game.Cellsize)
	game.Tiles.Age3 = ebiten.NewImage(game.Cellsize, game.Cellsize)
	game.Tiles.Age4 = ebiten.NewImage(game.Cellsize, game.Cellsize)

	cellsize := game.ScreenWidth / game.Cellsize

	FillCell(game.Tiles.Black, cellsize, game.Black)
	FillCell(game.Tiles.White, cellsize, game.White)
	FillCell(game.Tiles.Old, cellsize, game.Old)
	FillCell(game.Tiles.Age1, cellsize, game.AgeColor1)
	FillCell(game.Tiles.Age2, cellsize, game.AgeColor2)
	FillCell(game.Tiles.Age3, cellsize, game.AgeColor3)
	FillCell(game.Tiles.Age4, cellsize, game.AgeColor4)
}

func (game *Game) Init() {
	// setup the game
	var grid *Grid

	if game.Statefile != "" {
		g, err := LoadState(game.Statefile)
		if err != nil {
			log.Fatalf("failed to load game state: %s", err)
		}

		grid = g

		game.Width = grid.Width
		game.Height = grid.Height
	}

	game.ScreenWidth = game.Cellsize * game.Width
	game.ScreenHeight = game.Cellsize * game.Height

	game.Camera = Camera{
		ViewPort: f64.Vec2{
			float64(game.ScreenWidth),
			float64(game.ScreenHeight),
		},
	}

	game.World = ebiten.NewImage(game.ScreenWidth, game.ScreenHeight)

	game.InitGrid(grid)
	game.InitPattern()
	game.InitTiles()

	game.Index = 0
	game.TicksElapsed = 0

	game.LastCursorPos = make([]int, 2)
}

// count the living neighbors of a cell
func (game *Game) CountNeighbors(x, y int) int64 {
	var sum int64

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
