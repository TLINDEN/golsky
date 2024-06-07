package main

import (
	"fmt"
	"log"
	"math/rand"
	"os"
	"runtime/pprof"
	"time"
	"unsafe"
)

const (
	max     int  = 1500
	loops   int  = 5000
	density int  = 8
	debug   bool = false
)

type Cell struct {
	State         bool
	Neighbors     []*Cell
	NeighborCount int
}

// https://dev.to/chigbeef_77/bool-int-but-stupid-in-go-3jb3
func bool2int(b bool) int {
	return int(*(*byte)(unsafe.Pointer(&b)))
}

func (cell *Cell) Count(x, y int) {
	cell.NeighborCount = 0

	for _, neighbor := range cell.Neighbors {
		cell.NeighborCount += bool2int(neighbor.State)
	}
}

func SetNeighbors(grid [][]Cell, x, y int) {
	cells := []*Cell{}

	for nbgX := -1; nbgX < 2; nbgX++ {
		for nbgY := -1; nbgY < 2; nbgY++ {
			var col, row int

			if x+nbgX < 0 || x+nbgX >= max || y+nbgY < 0 || y+nbgY >= max {
				continue
			}

			col = x + nbgX
			row = y + nbgY

			if col == x && row == y {
				continue
			}

			cells = append(cells, &grid[row][col])
		}
	}

	grid[y][x].Neighbors = make([]*Cell, len(cells))
	for idx, cell := range cells {
		grid[y][x].Neighbors[idx] = cell
	}
}

func Init() [][]Cell {
	grid := make([][]Cell, max)
	for y := 0; y < max; y++ {
		grid[y] = make([]Cell, max)
		for x := 0; x < max; x++ {
			if rand.Intn(density) == 1 {
				grid[y][x].State = true
			}
		}
	}

	for y := 0; y < max; y++ {
		for x := 0; x < max; x++ {
			SetNeighbors(grid, x, y)
		}
	}

	return grid
}

func Loop(grid [][]Cell) {
	c := 0
	for i := 0; i < loops; i++ {
		for y := 0; y < max; y++ {
			for x := 0; x < max; x++ {
				cell := &grid[y][x]
				state := cell.State

				cell.Count(x, y)

				if state && cell.NeighborCount > 1 {
					if debug {
						fmt.Printf(
							"Loop %d - cell at %d,%d is %t and has %d living neighbors\n",
							i, x, y, state, cell.NeighborCount)
					}
					c = 1
				}
			}
		}
	}

	if c > 1 {
		c = 0
	}
}

func main() {
	// enable  cpu profiling. Do  NOT use q  to stop the  game but
	// close the window to get a profile
	fd, err := os.Create("cpu.profile")
	if err != nil {
		log.Fatal(err)
	}
	defer fd.Close()

	pprof.StartCPUProfile(fd)
	defer pprof.StopCPUProfile()

	// init
	grid := Init()

	// main loop
	loopstart := time.Now()

	Loop(grid)

	loopend := time.Now()
	diff := loopstart.Sub(loopend)
	fmt.Printf("Loop took %.04f\n", diff.Seconds())
}
