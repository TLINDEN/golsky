package main

import (
	"fmt"
	"log"
	"os"
	"runtime/pprof"
	"time"

	_ "net/http/pprof"

	"github.com/hajimehoshi/ebiten/v2"
)

func main() {
	config, err := ParseCommandline()
	if err != nil {
		log.Fatal(err)
	}

	if config.ShowVersion {
		fmt.Printf("This is golsky version %s\n", VERSION)
		os.Exit(0)
	}

	game := NewGame(config, Play)

	if config.ProfileFile != "" {
		// enable cpu profiling and use fake game loop
		fd, err := os.Create(config.ProfileFile)
		if err != nil {
			log.Fatal(err)
		}
		defer fd.Close()

		pprof.StartCPUProfile(fd)
		defer pprof.StopCPUProfile()

		Ebitfake(game)

		pprof.StopCPUProfile()
		fd.Close()

		os.Exit(0)
	}

	// main loop
	if err := ebiten.RunGame(game); err != nil {
		log.Fatal(err)
	}
}

// fake game  loop, required to be  able to profile the  program using
// pprof. Otherwise any kind of program  exit leads to an empty profile
// file.
func Ebitfake(game *Game) {
	screen := ebiten.NewImage(game.ScreenWidth, game.ScreenHeight)

	var loops int64

	for {
		err := game.Update()
		if err != nil {
			log.Fatal(err)
		}

		if game.Config.ProfileDraw {
			game.Draw(screen)
		}

		fmt.Print(".")
		time.Sleep(16 * time.Millisecond) // around 60 TPS

		if loops >= game.Config.ProfileMaxLoops {
			break
		}

		loops++
	}
}
