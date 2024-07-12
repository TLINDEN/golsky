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

type Cell struct {
	State         uint8
	Neighbors     [8]*Cell
	NeighborCount int
}

func (cell *Cell) Count() uint8 {
	var count uint8

	for idx := 0; idx < cell.NeighborCount; idx++ {
		count += cell.Neighbors[idx].State
	}

	return count
}

type Grid struct {
	Data   [][]*Cell
	Empty  bool
	Config *Config
}

// Create new empty grid and allocate Data according to provided dimensions
func NewGrid(config *Config) *Grid {
	grid := &Grid{
		Data:   make([][]*Cell, config.Height),
		Empty:  config.Empty,
		Config: config,
	}

	// first setup the cells
	for y := 0; y < config.Height; y++ {
		grid.Data[y] = make([]*Cell, config.Width)
		for x := 0; x < config.Width; x++ {
			grid.Data[y][x] = &Cell{}
		}
	}

	// in a second pass, collect pointers to the neighbors of each cell
	for y := 0; y < config.Height; y++ {
		for x := 0; x < config.Width; x++ {
			grid.SetupNeighbors(x, y)
		}
	}

	return grid
}

func (grid *Grid) SetupNeighbors(x, y int) {
	idx := 0

	for nbgY := -1; nbgY < 2; nbgY++ {
		for nbgX := -1; nbgX < 2; nbgX++ {
			var col, row int

			if grid.Config.Wrap {
				// In wrap mode we look at all the 8 neighbors surrounding us.
				// In case we are on an edge we'll look at the neighbor on the
				//  other side  of the  grid, thus  wrapping lookahead  around
				// using the mod() function.
				col = (x + nbgX + grid.Config.Width) % grid.Config.Width
				row = (y + nbgY + grid.Config.Height) % grid.Config.Height

			} else {
				// In traditional grid mode the edges are deadly
				if x+nbgX < 0 || x+nbgX >= grid.Config.Width || y+nbgY < 0 || y+nbgY >= grid.Config.Height {
					continue
				}

				col = x + nbgX
				row = y + nbgY
			}

			if col == x && row == y {
				continue
			}

			grid.Data[y][x].Neighbors[idx] = grid.Data[row][col]
			grid.Data[y][x].NeighborCount++
			idx++
		}
	}
}

// count the living neighbors of a cell
func (grid *Grid) CountNeighbors(x, y int) uint8 {
	return grid.Data[y][x].Count()
}

// Create a new 1:1 instance
func (grid *Grid) Clone() *Grid {
	newgrid := &Grid{}

	newgrid.Config = grid.Config
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
			grid.Data[y][x].State = 0
		}
	}
}

// initialize with random life cells using the given density
func (grid *Grid) FillRandom() {
	if !grid.Empty {
		for y := range grid.Data {
			for x := range grid.Data[y] {
				if rand.Intn(grid.Config.Density) == 1 {
					grid.Data[y][x].State = 1
				}
			}
		}
	}
}

func (grid *Grid) Dump() {
	for y := 0; y < grid.Config.Height; y++ {
		for x := 0; x < grid.Config.Width; x++ {
			if grid.Data[y][x].State == 1 {
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
		startX := (grid.Config.Width / 2) - (pattern.Width / 2)
		startY := (grid.Config.Height / 2) - (pattern.Height / 2)
		var y, x int

		for rowIndex, patternRow := range pattern.Pattern {
			for colIndex := range patternRow {
				if pattern.Pattern[rowIndex][colIndex] > 0 {
					x = colIndex + startX
					y = rowIndex + startY

					grid.Data[y][x].State = 1
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
			row := "."
			if cell.State == 1 {
				row = "o"
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
func GetFilename(generations int64) string {
	now := time.Now()
	return fmt.Sprintf("dump-%s-%d.lif", now.Format("20060102150405"), generations)
}

func GetFilenameRLE(generations int64) string {
	now := time.Now()
	return fmt.Sprintf("rect-%s-%d.rle", now.Format("20060102150405"), generations)
}
