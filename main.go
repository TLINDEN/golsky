package main

import (
	"fmt"
	"log"
	"os"

	"github.com/hajimehoshi/ebiten/v2"
)

func main() {
	config := ParseCommandline()

	if config.ShowVersion {
		fmt.Printf("This is golsky version %s\n", VERSION)
		os.Exit(0)
	}

	// grid := [][]int64{
	// 	{0, 1, 1},
	// 	{0, 1, 0},
	// 	{1, 1, 0},
	// }

	// err := rle.StoreGridToRLE(grid, "test.rle", "B3/S23", 3, 3)
	// if err != nil {
	// 	panic(err)
	// }

	// os.Exit(0)

	game := NewGame(config, Play)

	// main loop
	if err := ebiten.RunGame(game); err != nil {
		log.Fatal(err)
	}
}
