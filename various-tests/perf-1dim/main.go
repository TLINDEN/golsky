package main

import (
	"fmt"
	"log"
	"math/rand"
	"os"
	"runtime/pprof"
	"unsafe"
)

const (
	dim     int  = 1500
	loops   int  = 5000
	density int  = 8
	debug   bool = false
)

var max int

// https://dev.to/chigbeef_77/bool-int-but-stupid-in-go-3jb3
func bool2int(b bool) int {
	return int(*(*byte)(unsafe.Pointer(&b)))
}

func Count(grid []bool, x, y int) int {
	var sum int

	for nbgX := -1; nbgX < 2; nbgX++ {
		for nbgY := -1; nbgY < 2; nbgY++ {
			var col, row int

			if x+nbgX < 0 || x+nbgX >= dim || y+nbgY < 0 || y+nbgY >= dim {
				continue
			}

			col = x + nbgX
			row = y + nbgY

			state := grid[row*col]
			intstate := bool2int(state)
			sum += intstate
		}
	}

	sum -= bool2int(grid[y*x])

	return sum
}

func Init() []bool {
	max = dim * dim

	grid := make([]bool, max)

	for y := 0; y < dim; y++ {
		for x := 0; x < dim; x++ {
			if rand.Intn(density) == 1 {
				grid[y*x] = true
			}
		}
	}

	return grid
}

func Loop(grid []bool) {
	c := 0
	for i := 0; i < loops; i++ {
		for y := 0; y < dim; y++ {
			for x := 0; x < dim; x++ {
				state := grid[y*x]
				neighbors := Count(grid, x, y)
				if state && neighbors > 1 {
					if debug {
						fmt.Printf("Loop %d - cell at %d,%d is %t and has %d living neighbors\n", i, x, y, state, neighbors)
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
	Loop(grid)
}
