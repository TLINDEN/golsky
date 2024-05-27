// original source: https://github.com/nhoffmann/life by N.Hoffmann 2020.
package rle

import (
	"fmt"
	"os"
	"regexp"
	"strconv"
	"strings"
)

type RLE struct {
	Rule    string  // rule
	Width   int     // x
	Height  int     // y
	Pattern [][]int // The actual pattern

	inputLines       []string
	headerLineIndex  int
	patternLineIndex int
}

// wrapper to load a RLE file
func GetRLE(filename string) (*RLE, error) {
	if filename == "" {
		return nil, nil
	}

	content, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	parsedRle, err := Parse(string(content))
	if err != nil {
		return nil, fmt.Errorf("failed to load RLE pattern file: %s", err)
	}

	return &parsedRle, nil
}

func Parse(input string) (RLE, error) {
	rle := RLE{
		inputLines: strings.Split(input, "\n"),
	}

	rle.partitionFile()

	err := rle.parseComments()
	if err != nil {
		return RLE{}, err
	}
	err = rle.parseHeader()
	if err != nil {
		return RLE{}, err
	}
	err = rle.parsePattern()
	if err != nil {
		return RLE{}, err
	}

	return rle, nil
}

func (rle *RLE) partitionFile() error {
	for index, line := range rle.inputLines {
		cleanLine := removeWhitespace(line)
		if strings.HasPrefix(cleanLine, "x=") {
			rle.headerLineIndex = index
			rle.patternLineIndex = index + 1
			return nil
		}
	}

	return fmt.Errorf("invalid input: Header is missing")
}

func (rle *RLE) parseComments() error {
	return nil
}

func (rle *RLE) parseHeader() (err error) {
	headerLine := removeWhitespace(rle.inputLines[rle.headerLineIndex])

	headerElements := strings.SplitN(headerLine, ",", 3)

	rle.Width, err = strconv.Atoi(strings.TrimPrefix(headerElements[0], "x="))
	if err != nil {
		return err
	}
	rle.Height, err = strconv.Atoi(strings.TrimPrefix(headerElements[1], "y="))
	if err != nil {
		return err
	}

	rle.Pattern = make([][]int, rle.Width)

	// check wehter a rule is present, since it's optional
	if len(headerElements) == 3 {
		rle.Rule = strings.TrimPrefix(headerElements[2], "rule=")
	}

	return nil
}

func (rle *RLE) parsePattern() error {
	patternString := strings.Join(rle.inputLines[rle.patternLineIndex:], "")

	l := NewLexer(patternString)
	pp := NewParser(l)

	rle.Pattern = pp.ParsePattern(rle.Width, rle.Height)

	return nil
}

func removeWhitespace(input string) string {
	re := regexp.MustCompile(` *\t*\r*\n*`)
	return re.ReplaceAllString(input, "")
}

// Store a grid to an RLE file
func StoreGridToRLE(grid [][]int64, filename, rule string, width, height int) error {
	fd, err := os.Create(filename)
	if err != nil {
		return err
	}

	var pattern string

	for y := 0; y < height; y++ {
		line := ""
		for x := 0; x < width; x++ {
			switch grid[y][x] {
			case 0:
				line += "b"
			case 1:
				line += "o"
			}
		}

		// if first row is: 001011110, then line is now:
		//                  bboboooob

		encoded := RunLengthEncode(line)

		// and now its:     2bob4ob
		pattern += encoded

		if y != height-1 {
			pattern += "$"
		}
	}

	pattern += "!"

	wrapped := ""
	for idx, char := range pattern {
		if idx%70 == 0 && idx != 0 {
			wrapped += "\n"
		}
		wrapped += string(char)
	}

	_, err = fmt.Fprintf(fd, "#N %s\nx = %d, y = %d, rule = %s\n%s\n",
		filename, width, height, rule, wrapped)

	if err != nil {
		return err
	}

	return nil
}

// by peterSO on
// https://codereview.stackexchange.com/questions/238893/run-length-encoding-in-golang
func RunLengthEncode(s string) string {
	e := make([]byte, 0, len(s))
	for i := 0; i < len(s); i++ {
		c := s[i]
		j := i + 1
		for ; j <= len(s); j++ {
			if j < len(s) && s[j] == c {
				continue
			}
			if j-i > 1 {
				e = strconv.AppendInt(e, int64(j-i), 10)
			}
			e = append(e, c)
			break
		}
		i = j - 1
	}
	return string(e)
}
