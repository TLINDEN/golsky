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

var Shader string = `
//kage:unit pixels

package main

var Alife int

func Fragment(_ vec4, pos vec2, _ vec4) vec4 {
	if Alife == 1 {
		return vec4(0.0)
	}

	return vec4(1.0)
}
`

func main() {
	config, err := ParseCommandline()
	if err != nil {
		log.Fatal(err)
	}

	if config.ShowVersion {
		fmt.Printf("This is golsky version %s\n", VERSION)
		os.Exit(0)
	}

	shader, err := ebiten.NewShader([]byte(Shader))
	if err != nil {
		fmt.Println(Shader)
		log.Fatalf("failed to compile shader: %s\n", err)
	}

	game := NewGame(config, shader, Play)

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

	fd, err := os.Create("cpu.profile")
	if err != nil {
		log.Fatal(err)
	}
	defer fd.Close()

	pprof.StartCPUProfile(fd)
	defer pprof.StopCPUProfile()
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
