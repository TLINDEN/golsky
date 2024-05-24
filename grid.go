package main

import (
	"bufio"
	"errors"
	"fmt"
	"math/rand"
	"os"
	"strconv"
	"strings"
	"time"
)

type Grid struct {
	Data          [][]int64
	Width, Height int
}

// Create new empty grid and allocate Data according to provided dimensions
func NewGrid(width, height int) *Grid {
	grid := &Grid{
		Height: height,
		Width:  width,
		Data:   make([][]int64, height),
	}

	for y := 0; y < height; y++ {
		grid.Data[y] = make([]int64, width)
	}

	return grid
}

func (grid *Grid) Clone() *Grid {
	newgrid := &Grid{}

	newgrid.Width = grid.Width
	newgrid.Height = grid.Height
	newgrid.Data = grid.Data

	return newgrid
}

func (grid *Grid) Clear() {
	for y := range grid.Data {
		for x := range grid.Data[y] {
			grid.Data[y][x] = 0
		}
	}
}

func (grid *Grid) FillRandom(game *Game) {
	if !game.Empty {
		for y := range grid.Data {
			for x := range grid.Data[y] {
				if rand.Intn(game.Density) == 1 {
					grid.Data[y][x] = 1
				}
			}
		}
	}
}

func GetFilename(generations int64) string {
	now := time.Now()
	return fmt.Sprintf("dump-%s-%d.gol", now.Format("20060102150405"), generations)
}

func (grid *Grid) SaveState(filename string) error {
	file, err := os.Create(filename)
	if err != nil {
		return fmt.Errorf("failed to open state file: %w", err)
	}
	defer file.Close()

	for y, _ := range grid.Data {
		for _, cell := range grid.Data[y] {
			_, err := file.WriteString(strconv.FormatInt(cell, 10))
			if err != nil {
				return fmt.Errorf("failed to write to state file: %w", err)
			}
		}
		file.WriteString("\n")
	}

	return nil
}

func LoadState(filename string) (*Grid, error) {
	fd, err := os.Open(filename)
	if err != nil {
		return nil, err
	}

	scanner := bufio.NewScanner(fd)

	scanner.Split(bufio.ScanLines)

	grid := &Grid{}

	for scanner.Scan() {
		items := strings.Split(scanner.Text(), "")
		row := make([]int64, len(items))

		for idx, item := range items {
			num, err := strconv.ParseInt(item, 10, 64)
			if err != nil {
				return nil, err
			}

			if num > 1 {
				return nil, errors.New("cells must be 0 or 1")
			}

			row[idx] = num
		}

		grid.Data = append(grid.Data, row)
	}

	// sanity check the grid
	explen := 0
	rows := 0
	first := true
	for _, row := range grid.Data {
		length := len(row)

		if first {
			explen = length
			first = false
		}

		if explen != length {
			return nil, fmt.Errorf(fmt.Sprintf("all rows must be in the same length, got: %d, expected: %d",
				length, explen))
		}

		rows++
	}

	grid.Width = explen
	grid.Height = rows

	return grid, nil
}
