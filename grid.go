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

	"github.com/tlinden/golsky/rle"
)

type Grid struct {
	Data                   [][]int64
	Width, Height, Density int
	Empty                  bool
}

// Create new empty grid and allocate Data according to provided dimensions
func NewGrid(width, height, density int, empty bool) *Grid {
	grid := &Grid{
		Height:  height,
		Width:   width,
		Density: density,
		Data:    make([][]int64, height),
		Empty:   empty,
	}

	for y := 0; y < height; y++ {
		grid.Data[y] = make([]int64, width)
	}

	return grid
}

// Create a new 1:1 instance
func (grid *Grid) Clone() *Grid {
	newgrid := &Grid{}

	newgrid.Width = grid.Width
	newgrid.Height = grid.Height
	newgrid.Data = grid.Data

	return newgrid
}

// copy data
func (grid *Grid) Copy(other *Grid) {
	for y := range grid.Data {
		for x := range grid.Data[y] {
			other.Data[y][x] = grid.Data[y][x]
		}
	}
}

// delete all contents
func (grid *Grid) Clear() {
	for y := range grid.Data {
		for x := range grid.Data[y] {
			grid.Data[y][x] = 0
		}
	}
}

// initialize with random life cells using the given density
func (grid *Grid) FillRandom() {
	if !grid.Empty {
		for y := range grid.Data {
			for x := range grid.Data[y] {
				if rand.Intn(grid.Density) == 1 {
					grid.Data[y][x] = 1
				}
			}
		}
	}
}

func (grid *Grid) Dump() {
	for y := 0; y < grid.Height; y++ {
		for x := 0; x < grid.Width; x++ {
			if grid.Data[y][x] == 1 {
				fmt.Print("XX")
			} else {
				fmt.Print("  ")
			}
		}
		fmt.Println()
	}
}

// initialize using a given RLE pattern
func (grid *Grid) LoadRLE(pattern *rle.RLE) {
	if pattern != nil {
		startX := (grid.Width / 2) - (pattern.Width / 2)
		startY := (grid.Height / 2) - (pattern.Height / 2)
		var y, x int

		for rowIndex, patternRow := range pattern.Pattern {
			for colIndex := range patternRow {
				if pattern.Pattern[rowIndex][colIndex] > 0 {
					x = colIndex + startX
					y = rowIndex + startY

					grid.Data[y][x] = 1
				}
			}
		}

		//grid.Dump()
	}
}

// save  the contents  of  the  whole grid  as  a  simple mcell  alike
// file. One line per row, 0 for dead and 1 for life cell.
func (grid *Grid) SaveState(filename string) error {
	file, err := os.Create(filename)
	if err != nil {
		return fmt.Errorf("failed to open state file: %w", err)
	}
	defer file.Close()

	for y := range grid.Data {
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

// the reverse of the above, load a mcell file
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
			return nil, fmt.Errorf(
				fmt.Sprintf("all rows must be in the same length, got: %d, expected: %d",
					length, explen))
		}

		rows++
	}

	grid.Width = explen
	grid.Height = rows

	return grid, nil
}

// generate filenames for dumps
func GetFilename(generations int64) string {
	now := time.Now()
	return fmt.Sprintf("dump-%s-%d.gol", now.Format("20060102150405"), generations)
}

func GetFilenameRLE(generations int64) string {
	now := time.Now()
	return fmt.Sprintf("rect-%s-%d.rle", now.Format("20060102150405"), generations)
}
