package main

import (
	"image/color"
	"log"
	"math/rand"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
)

type Grid struct {
	Data [][]int
}

type Game struct {
	Grids                     []*Grid // 2 grids: one current, one next
	Index                     int     // points to current grid
	Width, Height, Cellsize   int
	ScreenWidth, ScreenHeight int
	Black, White              color.RGBA
}

func (game *Game) Layout(outsideWidth, outsideHeight int) (int, int) {
	return game.ScreenWidth, game.ScreenHeight
}

func (game *Game) Update() error {
	// compute cells
	next := game.Index ^ 1 // next grid index, we just xor 0|1 to 1|0

	for y := 0; y < game.Height; y++ {
		for x := 0; x < game.Width; x++ {
			state := game.Grids[game.Index].Data[y][x] // 0|1 == dead or alive
			neighbors := CountNeighbors(game, x, y)    // alive neighbor count
			var nextstate int

			// the actual game of life rules
			if state == 0 && neighbors == 3 {
				nextstate = 1
			} else if state == 1 && (neighbors < 2 || neighbors > 3) {
				nextstate = 0
			} else {
				nextstate = state
			}

			// change state of current cell in next grid
			game.Grids[next].Data[y][x] = nextstate
		}
	}

	// switch grid for rendering
	game.Index ^= 1

	return nil
}

func (game *Game) Draw(screen *ebiten.Image) {
	for y := 0; y < game.Height; y++ {
		for x := 0; x < game.Width; x++ {
			currentcolor := game.White
			if game.Grids[game.Index].Data[y][x] == 1 {
				currentcolor = game.Black
			}

			vector.DrawFilledRect(screen,
				float32(x*game.Cellsize),
				float32(y*game.Cellsize),
				float32(game.Cellsize),
				float32(game.Cellsize),
				currentcolor, false)

			if currentcolor == game.White {
				// draw black
				vector.DrawFilledRect(screen,
					float32(x*game.Cellsize),
					float32(y*game.Cellsize),
					float32(game.Cellsize),
					float32(game.Cellsize),
					game.Black, false)
				// then fill with 1px lesser rect in white
				// thus creating grid lines
				vector.DrawFilledRect(screen,
					float32(x*game.Cellsize+1),
					float32(y*game.Cellsize+1),
					float32(game.Cellsize-1),
					float32(game.Cellsize-1),
					game.White, false)
			}
		}
	}
}

func (game *Game) Init() {
	// setup the game
	game.ScreenWidth = game.Cellsize * game.Width
	game.ScreenHeight = game.Cellsize * game.Height

	grid := &Grid{Data: make([][]int, game.Height)}
	gridb := &Grid{Data: make([][]int, game.Height)}

	for y := 0; y < game.Height; y++ {
		grid.Data[y] = make([]int, game.Width)
		gridb.Data[y] = make([]int, game.Width)
		for x := 0; x < game.Width; x++ {
			grid.Data[y][x] = rand.Intn(2)
		}
	}

	game.Grids = []*Grid{
		grid,
		gridb,
	}

	game.Black = color.RGBA{0, 0, 0, 0xff}
	game.White = color.RGBA{0xff, 0xff, 0xff, 0xff}

	game.Index = 0
}

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
	game := &Game{Width: 180, Height: 160, Cellsize: 15}
	game.Init()

	ebiten.SetWindowSize(game.ScreenWidth, game.ScreenHeight)
	ebiten.SetWindowTitle("Game of life")
	ebiten.SetTPS(30)

	if err := ebiten.RunGame(game); err != nil {
		log.Fatal(err)
	}

}
