package main

import (
	"fmt"
	"log"
	"math/rand"
	"os"
	"os/exec"
	"runtime/pprof"
	"unsafe"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
)

type Images struct {
	Black, White *ebiten.Image
}

type Cell struct {
	State         bool
	Neighbors     [8]*Cell
	NeighborCount int
}

func bool2int(b bool) int {
	return int(*(*byte)(unsafe.Pointer(&b)))
}

func (cell *Cell) Count(x, y int) int {
	sum := 0

	for idx := 0; idx < cell.NeighborCount; idx++ {
		sum += bool2int(cell.Neighbors[idx].State)
	}

	return sum
}

func SetNeighbors(grid [][]*Cell, x, y, width, height int) {
	idx := 0
	for nbgX := -1; nbgX < 2; nbgX++ {
		for nbgY := -1; nbgY < 2; nbgY++ {
			var col, row int

			if x+nbgX < 0 || x+nbgX >= width || y+nbgY < 0 || y+nbgY >= height {
				continue
			}

			col = x + nbgX
			row = y + nbgY

			if col == x && row == y {
				continue
			}

			grid[y][x].Neighbors[idx] = grid[row][col]
			grid[y][x].NeighborCount++
			idx++
		}
	}
}

type Grid struct {
	Data                   [][]*Cell
	Width, Height, Density int
}

// Create new empty grid and allocate Data according to provided dimensions
func NewGrid(width, height, density int) *Grid {
	grid := &Grid{
		Height:  height,
		Width:   width,
		Density: density,
		Data:    make([][]*Cell, height),
	}

	for y := 0; y < height; y++ {
		grid.Data[y] = make([]*Cell, width)
		for x := 0; x < width; x++ {
			grid.Data[y][x] = &Cell{}

			if rand.Intn(density) == 1 {
				grid.Data[y][x].State = true
			}
		}
	}

	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			SetNeighbors(grid.Data, x, y, width, height)
		}
	}

	return grid
}

type Game struct {
	Width, Height, Cellsize, Density int
	ScreenWidth, ScreenHeight        int
	Grids                            []*Grid
	Index                            int
	Elapsed                          int64
	TPG                              int64 // adjust game speed independently of TPS
	Pause, Debug, Profile, Gridlines bool
	Pixels                           []byte
	OffScreen                        *ebiten.Image
}

func (game *Game) Layout(outsideWidth, outsideHeight int) (int, int) {
	return game.ScreenWidth, game.ScreenHeight
}

// live console output of the grid
func (game *Game) DebugDump() {
	cmd := exec.Command("clear")
	cmd.Stdout = os.Stdout
	cmd.Run()

	if game.Debug {
		for y := 0; y < game.Height; y++ {
			for x := 0; x < game.Width; x++ {
				if game.Grids[game.Index].Data[y][x].State {
					fmt.Print("XX")
				} else {
					fmt.Print("  ")
				}
			}
			fmt.Println()
		}
	}
	fmt.Printf("FPS: %0.2f\n", ebiten.ActualTPS())
}

func (game *Game) Init() {
	// setup two grids, one for display, one for next state
	grida := NewGrid(game.Width, game.Height, game.Density)
	gridb := NewGrid(game.Width, game.Height, game.Density)

	game.Grids = []*Grid{
		grida,
		gridb,
	}

	game.Pixels = make([]byte, game.ScreenWidth*game.ScreenHeight*4)

	game.OffScreen = ebiten.NewImage(game.ScreenWidth, game.ScreenHeight)
}

// count the living neighbors of a cell
func (game *Game) CountNeighbors(x, y int) int {
	return game.Grids[game.Index].Data[y][x].Count(x, y)
}

// the heart of the game
func (game *Game) CheckRule(state bool, neighbors int) bool {
	var nextstate bool

	if state && neighbors == 3 {
		nextstate = true
	} else if state && (neighbors == 2 || neighbors == 3) {
		nextstate = true
	} else {
		nextstate = false
	}

	return nextstate
}

// we only  update the cells if  we are not  in pause state or  if the
// game timer (TPG) is elapsed.
func (game *Game) UpdateCells() {
	if game.Pause {
		return
	}

	if game.Elapsed < game.TPG {
		game.Elapsed++
		return
	}

	// next grid index. we only have to, so we just xor it
	next := game.Index ^ 1

	// calculate cell life state, this is the actual game of life
	for y := 0; y < game.Height; y++ {
		for x := 0; x < game.Width; x++ {
			state := game.Grids[game.Index].Data[y][x].State
			neighbors := game.CountNeighbors(x, y)

			// actually apply the current rules
			nextstate := game.CheckRule(state, neighbors)

			// change state of current cell in next grid
			game.Grids[next].Data[y][x].State = nextstate
		}
	}

	// switch grid for rendering
	game.Index ^= 1

	game.Elapsed = 0

	game.UpdatePixels()
}

func (game *Game) Update() error {
	game.UpdateCells()

	if inpututil.IsKeyJustPressed(ebiten.KeySpace) {
		game.Pause = !game.Pause
	}

	return nil
}

/*
*
r, g, b := color(it)

	78             p := 4 * (i + j*screenWidth)
	79             gm.offscreenPix[p] = r
	80             gm.offscreenPix[p+1] = g
	81             gm.offscreenPix[p+2] = b
	82             gm.offscreenPix[p+3] = 0xff
*/
func (game *Game) UpdatePixels() {
	var col byte

	gridx := 0
	gridy := 0
	idx := 0

	for y := 0; y < game.ScreenHeight; y++ {
		for x := 0; x < game.ScreenWidth; x++ {
			gridx = x / game.Cellsize
			gridy = y / game.Cellsize

			col = 0xff
			if game.Grids[game.Index].Data[gridy][gridx].State {
				col = 0x0
			}

			if game.Gridlines {
				if x%game.Cellsize == 0 || y%game.Cellsize == 0 {
					col = 128
				}
			}

			idx = 4 * (x + y*game.ScreenWidth)

			game.Pixels[idx] = col
			game.Pixels[idx+1] = col
			game.Pixels[idx+2] = col
			game.Pixels[idx+3] = 0xff

			idx++
		}
	}

	game.OffScreen.WritePixels(game.Pixels)
}

func (game *Game) Draw(screen *ebiten.Image) {
	screen.DrawImage(game.OffScreen, nil)
	game.DebugDump()
}

func main() {
	size := 1500

	game := &Game{
		Width:     size,
		Height:    size,
		Cellsize:  4,
		Density:   8,
		TPG:       10,
		Debug:     false,
		Profile:   true,
		Gridlines: false,
	}

	game.ScreenWidth = game.Width * game.Cellsize
	game.ScreenHeight = game.Height * game.Cellsize

	game.Init()

	ebiten.SetWindowSize(game.ScreenWidth, game.ScreenHeight)
	ebiten.SetWindowTitle("triangle conway's game of life")
	ebiten.SetWindowResizingMode(ebiten.WindowResizingModeEnabled)

	if game.Profile {
		fd, err := os.Create("cpu.profile")
		if err != nil {
			log.Fatal(err)
		}
		defer fd.Close()

		pprof.StartCPUProfile(fd)
		defer pprof.StopCPUProfile()
	}

	if err := ebiten.RunGame(game); err != nil {
		log.Fatal(err)
	}
}
