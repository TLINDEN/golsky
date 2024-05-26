# golsky - Conway's game of life written in GO

![Golsky Logo](https://github.com/TLINDEN/golsky/blob/main/.github/assets/golskylogo.png)

[![License](https://img.shields.io/badge/license-GPL-blue.svg)](https://github.com/tlinden/golsky/blob/master/LICENSE)
[![Go Report Card](https://goreportcard.com/badge/github.com/tlinden/golsky)](https://goreportcard.com/report/github.com/tlinden/golsky) 

I wanted to play around a little bit with [**Conways Game of
Life**](https://conwaylife.com/)
in golang and here's the  result. It's a simple game using
[ebitengine](https://github.com/hajimehoshi/ebiten/).

John Conway himself: https://youtu.be/R9Plq-D1gEk?si=yYxs77e9yXxeSNbL

Based on: https://youtu.be/FWSR_7kZuYg?si=ix1dmo76D8AmF25F

# Features

* flexible parameters as grid and cell size
* colors can be inverted
* evolution  tracks can be shown,  with age the cells  color fades and
  old life cells will be drawn in red
* game grid lines can be enabled or disabled
* game speed can be adjusted on startup and in-game
* you can zoom in and out of the canvas and move it around
* game can be paused any time
* it can be run step-wise
* game state can be saved any time and loaded later on startup
* various Life rules can be used, the rule format `B[0-9]+/S[0-9]+` is fully supported
* game patterns can be loaded using RLE files, see https://catagolue.hatsya.com/home
* you can paint your own patterns in the game
* the game can also be started with an empty grid, which is easier to paint patterns
* wrap around grid mode can be enabled
* you can also save rectangles of the grid to RLE files

# Install

In the github releases page you can find ready to use binaries for
your OS. Just download the one you need and use it.

# Build from source

Just execute: `go build .` and use the resulting executable.

You'll need the golang toolchain.

# Usage

The game has a couple of commandline options:

```default
Usage of ./golsky:
  -c, --cellsize int               cell size in pixels (default 8)
  -d, --debug                      show debug info
  -D, --density int                density of random cells (default 10)
  -e, --empty                      start with an empty screen
  -H, --height int                 grid height in cells (default 40)
  -i, --invert                     invert colors (dead cell: black)
  -l, --load-state-file string     game state file
  -n, --nogrid                     do not draw grid lines
  -p, --paused                     do not start simulation (use space to start)
  -f, --rle-file string            RLE pattern file
  -r, --rule string                game rule (default "B3/S23")
  -s, --show-evolution             show evolution tracks
  -t, --ticks-per-generation int   game speed: the higher the slower (default: 10) (default 10)
  -v, --version                    show version
  -W, --width int                  grid width in cells (default 40)
```

While it runs, there are a couple of commands you can use:

* left mouse click: set a cell to alife (also pauses the game)
* right mouse click: set a cell to dead
* space: pause or resume the game
* while game is paused: press n to forward one step
* page up: speed up
* page down: slow down
* Mouse wheel: zoom in or out
* move mouse while middle mouse button pressed: move canvas
* escape: reset to 1:1 zoom
* s: save game state to file (can be loaded with -l)
* c: enter copy mode. Mark a rectangle with the mouse, when you
  release the mous button it is being saved to an RLE file
* q: quit

# Report bugs

[Please open an issue](https://github.com/TLINDEN/golsky/issues). Thanks!

# License

This work is licensed under the terms of the General Public Licens
version 3.

# Author

Copyleft (c) 2024 Thomas von Dein

