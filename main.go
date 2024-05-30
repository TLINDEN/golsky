package main

import (
	"fmt"
	"log"
	"os"
	"runtime/pprof"

	_ "net/http/pprof"

	"github.com/hajimehoshi/ebiten/v2"
)

func main() {
	dau := true

	if len(os.Args) > 1 {
		dau = false
	}
	config, err := ParseCommandline()
	if err != nil {
		log.Fatal(err)
	}

	if config.ShowVersion {
		fmt.Printf("This is golsky version %s\n", VERSION)
		os.Exit(0)
	}

	start := Play
	if dau {
		start = Menu
	}
	game := NewGame(config, SceneName(start))

	if config.ProfileFile != "" {
		// enable  cpu profiling. Do  NOT use q  to stop the  game but
		// close the window to get a profile
		fd, err := os.Create(config.ProfileFile)
		if err != nil {
			log.Fatal(err)
		}
		defer fd.Close()

		pprof.StartCPUProfile(fd)
		defer pprof.StopCPUProfile()
	}

	// main loop
	if err := ebiten.RunGame(game); err != nil {
		log.Fatal(err)
	}
}
