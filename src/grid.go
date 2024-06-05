package main

import (
	"bufio"
	"errors"
	"fmt"
	"math/rand"
	"os"
	"strings"
	"time"

	"github.com/tlinden/golsky/rle"
)

type Grid struct {
	Data                   [][]uint8
	Width, Height, Density int
	Empty                  bool
}

// Create new empty grid and allocate Data according to provided dimensions
func NewGrid(width, height, density int, empty bool) *Grid {
	grid := &Grid{
		Height:  height,
		Width:   width,
		Density: density,
		Data:    make([][]uint8, height),
		Empty:   empty,
	}

	for y := 0; y < height; y++ {
		grid.Data[y] = make([]uint8, width)
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

// load a lif file parameters like R and P are not supported yet
func LoadLIF(filename string) (*rle.RLE, error) {
	fd, err := os.Open(filename)
	if err != nil {
		return nil, err
	}

	scanner := bufio.NewScanner(fd)

	scanner.Split(bufio.ScanLines)

	gothead := false

	grid := &rle.RLE{}

	for scanner.Scan() {
		line := scanner.Text()
		items := strings.Split(line, "")

		if len(items) < 0 {
			continue
		}

		if strings.Contains(line, "# r") {
			parts := strings.Split(line, " ")
			if len(parts) == 2 {
				grid.Rule = parts[1]
			}

			continue
		}

		if items[0] == "#" {
			if gothead {
				break
			}

			continue
		}

		gothead = true

		row := make([]int, len(items))

		for idx, item := range items {
			switch item {
			case ".":
				row[idx] = 0
			case "o":
				fallthrough
			case "*":
				row[idx] = 1
			default:
				return nil, errors.New("cells must be . or o")
			}
		}

		grid.Pattern = append(grid.Pattern, row)
	}

	// sanity check the grid
	explen := 0
	rows := 0
	first := true
	for _, row := range grid.Pattern {
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

// save  the contents  of  the  whole grid  as  a  simple lif  alike
// file. One line per row, 0 for dead and 1 for life cell.
// file format: https://conwaylife.com/wiki/Life_1.05
func (grid *Grid) SaveState(filename, rule string) error {
	file, err := os.Create(filename)
	if err != nil {
		return fmt.Errorf("failed to open state file: %w", err)
	}
	defer file.Close()

	fmt.Fprintf(file, "#Life 1.05\n#R %s\n#D golsky state file\n#P -1 -1\n", rule)

	for y := range grid.Data {
		for _, cell := range grid.Data[y] {
			row := ""
			switch cell {
			case 1:
				row += "o"
			case 0:
				row += "."
			}

			_, err := file.WriteString(row)
			if err != nil {
				return fmt.Errorf("failed to write to state file: %w", err)
			}
		}
		file.WriteString("\n")
	}

	return nil
}

// generate filenames for dumps
func GetFilename(generations uint64) string {
	now := time.Now()
	return fmt.Sprintf("dump-%s-%d.lif", now.Format("20060102150405"), generations)
}

func GetFilenameRLE(generations uint64) string {
	now := time.Now()
	return fmt.Sprintf("rect-%s-%d.rle", now.Format("20060102150405"), generations)
}
