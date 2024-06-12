package main

import (
	"fmt"
	"log"
	"math/rand"
	"os"
	"runtime/pprof"

	"github.com/mlange-42/arche/ecs"
	"github.com/mlange-42/arche/generic"
)

const (
	max     int  = 1500
	loops   int  = 5000
	density int  = 8
	debug   bool = false
)

// components
type Pos struct {
	X, Y, GridX, GridY int
}

type Cell struct {
	State     bool
	Neighbors [8]ecs.Entity
}

type ECS struct {
	World  *ecs.World
	Filter *generic.Filter2[Pos, Cell]
	Map    *generic.Map2[Pos, Cell]
}

func (cell *Cell) NeighborCount(ECS *ECS) int {
	sum := 0

	for _, neighbor := range cell.Neighbors {
		if ECS.World.Alive(neighbor) {
			_, cel := ECS.Map.Get(neighbor)
			if cel.State {
				sum++
			}
		}
	}

	return sum
}

func Loop(ECS *ECS) {
	c := 0

	for i := 0; i < loops; i++ {
		query := ECS.Filter.Query(ECS.World)

		for query.Next() {
			_, cel := query.Get()
			if cel.State && cel.NeighborCount(ECS) > 1 {
				c = 1
			}
		}
	}

	if c > 1 {
		c = 0
	}
}

func SetupWorld() *ECS {
	world := ecs.NewWorld()

	builder := generic.NewMap2[Pos, Cell](&world)

	// we need a temporary grid in order to find out neighbors
	grid := [max][max]ecs.Entity{}

	// setup entities
	for y := 0; y < max; y++ {
		for x := 0; x < max; x++ {
			e := builder.New()
			pos, cell := builder.Get(e)
			pos.X = x
			pos.Y = y // pos.GridX = x*cellsize

			cell.State = false
			if rand.Intn(density) == 1 {
				cell.State = true
			}

			// store to tmp grid
			grid[y][x] = e
		}
	}

	// global filter
	filter := generic.NewFilter2[Pos, Cell]()

	query := filter.Query(&world)

	for query.Next() {
		pos, cel := query.Get()

		n := 0
		for x := -1; x < 2; x++ {
			for y := -1; y < 2; y++ {
				XX := pos.X + x
				YY := pos.Y + y
				if XX < 0 || XX >= max || YY < 0 || YY >= max {
					continue
				}

				if pos.X != XX || pos.Y != YY {
					cel.Neighbors[n] = grid[XX][YY]
					n++
				}
			}
		}
	}

	return &ECS{World: &world, Filter: filter, Map: &builder}
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
	fmt.Print("Setup ... ")
	ECS := SetupWorld()
	fmt.Println("done")
	fmt.Println(ECS.World.Stats())

	// main loop
	Loop(ECS)
}
