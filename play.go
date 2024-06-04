package main

import (
	"fmt"
	"image"
	"image/color"
	"log"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/hajimehoshi/ebiten/v2/vector"
	"github.com/tlinden/golsky/rle"
	"golang.org/x/image/math/f64"
)

type Images struct {
	Black, White, Age1, Age2, Age3, Age4, Old *ebiten.Image
}

const (
	DEBUG_FORMAT = "FPS: %0.2f, TPG: %d, M: %0.2fMB, Generations: %d\nScale: %.02f, Zoom: %d, Cam: %.02f,%.02f Cursor: %d,%d  %s"
)

type ScenePlay struct {
	Game   *Game
	Config *Config
	Next   SceneName
	Prev   SceneName
	Whoami SceneName

	Clear bool

	Grids                                      []*Grid // 2 grids: one current, one next
	History                                    *Grid   // holds state of past dead cells for evolution traces
	Index                                      int     // points to current grid
	Generations                                int64   // Stats
	Black, White, Grey, Old                    color.RGBA
	AgeColor1, AgeColor2, AgeColor3, AgeColor4 color.RGBA
	TicksElapsed                               int           // tick counter for game speed
	Tiles                                      Images        // pre-computed tiles for dead and alife cells
	Camera                                     Camera        // for zoom+move
	World, Cache                               *ebiten.Image // actual image we render to
	WheelTurned                                bool          // when user turns wheel multiple times, zoom faster
	Dragging                                   bool          // middle mouse is pressed, move canvas
	LastCursorPos                              []int         // used to check if the user is dragging
	MarkTaken                                  bool          // true when mouse1 pressed
	MarkDone                                   bool          // true when mouse1 released, copy cells between Mark+Point
	Mark, Point                                image.Point   // area to marks+save
	RunOneStep                                 bool          // mutable flags from config
	TPG                                        int
}

func NewPlayScene(game *Game, config *Config) Scene {
	scene := &ScenePlay{
		Whoami:     Play,
		Game:       game,
		Next:       Play,
		Config:     config,
		TPG:        config.TPG,
		RunOneStep: config.RunOneStep,
	}

	scene.Init()

	return scene
}

func (scene *ScenePlay) IsPrimary() bool {
	return true
}

func (scene *ScenePlay) GetNext() SceneName {
	return scene.Next
}

func (scene *ScenePlay) SetPrevious(prev SceneName) {
	scene.Prev = prev
}

func (scene *ScenePlay) ResetNext() {
	scene.Next = scene.Whoami
}

func (scene *ScenePlay) SetNext(next SceneName) {
	scene.Next = next
}

func (scene *ScenePlay) CheckRule(state int64, neighbors int64) int64 {
	var nextstate int64

	// The standard Scene of Life is symbolized in rule-string notation
	// as B3/S23 (23/3 here).  A cell  is born if it has exactly three
	// neighbors,  survives if it  has two or three  living neighbors,
	// and  dies otherwise. The first  number, or list of  numbers, is
	// what is required for a dead cell to be born.

	if state == 0 && Contains(scene.Config.Rule.Birth, neighbors) {
		nextstate = Alive
	} else if state == 1 && Contains(scene.Config.Rule.Death, neighbors) {
		nextstate = Alive
	} else {
		nextstate = Dead
	}

	return nextstate
}

// Update all cells according to the current rule
func (scene *ScenePlay) UpdateCells() {
	// count ticks so we know when to actually run
	scene.TicksElapsed++

	if scene.TPG > scene.TicksElapsed {
		// need to sleep a little more
		return
	}

	// next grid index, we just xor 0|1 to 1|0
	next := scene.Index ^ 1

	// compute life status of cells
	for y := 0; y < scene.Config.Height; y++ {
		for x := 0; x < scene.Config.Width; x++ {
			state := scene.Grids[scene.Index].Data[y][x] // 0|1 == dead or alive
			neighbors := scene.CountNeighbors(x, y)      // alive neighbor count

			// actually apply the current rules
			nextstate := scene.CheckRule(state, neighbors)

			// change state of current cell in next grid
			scene.Grids[next].Data[y][x] = nextstate

			// set history  to current generation so we  can infer the
			// age of the cell's state  during rendering and use it to
			// deduce the color to use if evolution tracing is enabled
			if state != nextstate {
				scene.History.Data[y][x] = scene.Generations
			}
		}
	}

	// switch grid for rendering
	scene.Index ^= 1

	// global stats counter
	scene.Generations++

	if scene.Config.RunOneStep {
		// setp-wise mode, halt the game
		scene.Config.RunOneStep = false
	}

	// reset speed counter
	scene.TicksElapsed = 0
}

func (scene *ScenePlay) Reset() {
	scene.Config.Paused = true
	scene.InitGrid()
	scene.Config.Paused = false
}

// check user input
func (scene *ScenePlay) CheckExit() error {
	if inpututil.IsKeyJustPressed(ebiten.KeyQ) {
		return ebiten.Termination
	}

	return nil
}

func (scene *ScenePlay) CheckInput() {
	if inpututil.IsKeyJustPressed(ebiten.KeyEscape) {
		scene.SetNext(Menu)
	}

	if inpututil.IsKeyJustPressed(ebiten.KeyO) {
		scene.SetNext(Options)
	}

	if inpututil.IsKeyJustPressed(ebiten.KeyC) {
		scene.Config.Markmode = true
		scene.Config.Paused = true
	}

	if scene.Config.Markmode {
		return
	}

	if inpututil.IsKeyJustPressed(ebiten.KeySpace) || inpututil.IsKeyJustPressed(ebiten.KeyEnter) {
		scene.Config.TogglePaused()
	}

	if ebiten.IsMouseButtonPressed(ebiten.MouseButtonLeft) {
		scene.ToggleCellOnCursorPos(Alive)
		scene.Config.Paused = true // drawing while running makes no sense
	}

	if ebiten.IsMouseButtonPressed(ebiten.MouseButtonRight) {
		scene.ToggleCellOnCursorPos(Dead)
		scene.Config.Paused = true // drawing while running makes no sense
	}

	if inpututil.IsKeyJustPressed(ebiten.KeyPageDown) {
		if scene.TPG < 120 {
			scene.TPG++
		}
	}

	if inpututil.IsKeyJustPressed(ebiten.KeyPageUp) {
		if scene.TPG >= 1 {
			scene.TPG--
		}
	}

	if inpututil.IsKeyJustPressed(ebiten.KeyS) {
		scene.SaveState()
	}

	if inpututil.IsKeyJustPressed(ebiten.KeyD) {
		scene.Config.Debug = !scene.Config.Debug
	}

	if scene.Config.Paused {
		if inpututil.IsKeyJustPressed(ebiten.KeyN) {
			scene.Config.RunOneStep = true
		}
	}
}

// Check dragging input.  move the canvas with the  mouse while pressing
// the middle mouse button, zoom in and out using the wheel.
func (scene *ScenePlay) CheckDraggingInput() {
	if scene.Config.Markmode {
		return
	}

	// move canvas
	if scene.Dragging && !ebiten.IsMouseButtonPressed(ebiten.MouseButton1) {
		// release
		scene.Dragging = false
	}

	if !scene.Dragging && ebiten.IsMouseButtonPressed(ebiten.MouseButton1) {
		// start dragging
		scene.Dragging = true
		scene.LastCursorPos[0], scene.LastCursorPos[1] = ebiten.CursorPosition()
	}

	if scene.Dragging {
		x, y := ebiten.CursorPosition()

		if x != scene.LastCursorPos[0] || y != scene.LastCursorPos[1] {
			// actually drag by mouse cursor pos diff to last cursor pos
			scene.Camera.Position[0] -= float64(x - scene.LastCursorPos[0])
			scene.Camera.Position[1] -= float64(y - scene.LastCursorPos[1])
		}

		scene.LastCursorPos[0], scene.LastCursorPos[1] = ebiten.CursorPosition()
	}

	// also support the arrow keys to move the canvas
	if ebiten.IsKeyPressed(ebiten.KeyArrowLeft) {
		scene.Camera.Position[0] -= 1
	}
	if ebiten.IsKeyPressed(ebiten.KeyArrowRight) {
		scene.Camera.Position[0] += 1
	}
	if ebiten.IsKeyPressed(ebiten.KeyArrowUp) {
		scene.Camera.Position[1] -= 1
	}
	if ebiten.IsKeyPressed(ebiten.KeyArrowDown) {
		scene.Camera.Position[1] += 1
	}

	// Zoom
	_, dy := ebiten.Wheel()

	if dy != 0 {
		scene.Camera.ZoomFactor += (int(dy) * 5)
	}

	if inpututil.IsKeyJustPressed(ebiten.KeyR) {
		scene.Camera.Reset()
	}

}

func (scene *ScenePlay) GetWorldCursorPos() image.Point {
	worldX, worldY := scene.Camera.ScreenToWorld(ebiten.CursorPosition())
	return image.Point{
		X: int(worldX) / scene.Config.Cellsize,
		Y: int(worldY) / scene.Config.Cellsize,
	}
}

func (scene *ScenePlay) CheckMarkInput() {
	if !scene.Config.Markmode {
		return
	}

	if inpututil.IsKeyJustPressed(ebiten.KeyEscape) {
		scene.Config.Markmode = false
	}

	if ebiten.IsMouseButtonPressed(ebiten.MouseButton0) {
		if !scene.MarkTaken {
			scene.Mark = scene.GetWorldCursorPos()
			scene.MarkTaken = true
			scene.MarkDone = false
		}

		scene.Point = scene.GetWorldCursorPos()
		//fmt.Printf("Mark: %v, Point: %v\n", scene.Mark, scene.Point)
	} else if inpututil.IsMouseButtonJustReleased(ebiten.MouseButton0) {
		scene.Config.Markmode = false
		scene.MarkTaken = false
		scene.MarkDone = true

		scene.SaveRectRLE()
	}
}

func (scene *ScenePlay) SaveState() {
	filename := GetFilename(scene.Generations)
	err := scene.Grids[scene.Index].SaveState(filename, scene.Config.Rule.Definition)
	if err != nil {
		log.Printf("failed to save game state to %s: %s", filename, err)
	}
	log.Printf("saved game state to %s at generation %d\n", filename, scene.Generations)
}

func (scene *ScenePlay) SaveRectRLE() {
	filename := GetFilenameRLE(scene.Generations)

	if scene.Mark.X == scene.Point.X || scene.Mark.Y == scene.Point.Y {
		log.Printf("can't save non-rectangle\n")
		return
	}

	var width int
	var height int
	var startx int
	var starty int

	if scene.Mark.X < scene.Point.X {
		// mark left point
		startx = scene.Mark.X
		width = scene.Point.X - scene.Mark.X
	} else {
		// mark right point
		startx = scene.Point.X
		width = scene.Mark.X - scene.Point.X
	}

	if scene.Mark.Y < scene.Point.Y {
		// mark above point
		starty = scene.Mark.Y
		height = scene.Point.Y - scene.Mark.Y
	} else {
		// mark below point
		starty = scene.Point.Y
		height = scene.Mark.Y - scene.Point.Y
	}

	grid := make([][]int64, height)

	for y := 0; y < height; y++ {
		grid[y] = make([]int64, width)

		for x := 0; x < width; x++ {
			grid[y][x] = scene.Grids[scene.Index].Data[y+starty][x+startx]
		}
	}

	err := rle.StoreGridToRLE(grid, filename, scene.Config.Rule.Definition, width, height)
	if err != nil {
		log.Printf("failed to save rect to %s: %s\n", filename, err)
	} else {
		log.Printf("saved selected rect to %s at generation %d\n", filename, scene.Generations)
	}

}

func (scene *ScenePlay) Update() error {
	if scene.Config.Restart {
		scene.Config.Restart = false
		scene.InitGrid()
		scene.InitCache()
		return nil
	}

	if scene.Config.RestartCache {
		scene.Config.RestartCache = false
		scene.InitTiles()
		scene.InitCache()
		return nil
	}

	if quit := scene.CheckExit(); quit != nil {
		return quit
	}

	scene.CheckInput()
	scene.CheckDraggingInput()
	scene.CheckMarkInput()

	if !scene.Config.Paused || scene.RunOneStep {
		scene.UpdateCells()
	}

	return nil
}

// set a cell to alive or dead
func (scene *ScenePlay) ToggleCellOnCursorPos(alive int64) {
	// use cursor pos relative to the world
	worldX, worldY := scene.Camera.ScreenToWorld(ebiten.CursorPosition())
	x := int(worldX) / scene.Config.Cellsize
	y := int(worldY) / scene.Config.Cellsize

	if x > -1 && y > -1 && x < scene.Config.Width && y < scene.Config.Height {
		scene.Grids[scene.Index].Data[y][x] = alive
		scene.History.Data[y][x] = 1
	}
}

// draw the new grid state
func (scene *ScenePlay) Draw(screen *ebiten.Image) {
	// we  fill the whole  screen with  a background color,  the cells
	// themselfes will be 1px smaller as their nominal size, producing
	// a nice grey grid with grid lines
	op := &ebiten.DrawImageOptions{}

	op.GeoM.Translate(0, 0)
	scene.World.DrawImage(scene.Cache, op)

	var age int64

	for y := 0; y < scene.Config.Height; y++ {
		for x := 0; x < scene.Config.Width; x++ {
			op.GeoM.Reset()
			op.GeoM.Translate(
				float64(x*scene.Config.Cellsize),
				float64(y*scene.Config.Cellsize),
			)

			age = scene.Generations - scene.History.Data[y][x]

			switch scene.Grids[scene.Index].Data[y][x] {
			case Alive:
				if age > 50 && scene.Config.ShowEvolution {
					scene.World.DrawImage(scene.Tiles.Old, op)

				} else {
					scene.World.DrawImage(scene.Tiles.Black, op)
				}
			case Dead:
				// only draw dead cells in case evolution trace is enabled
				if scene.History.Data[y][x] > 1 && scene.Config.ShowEvolution {
					switch {
					case age < 10:
						scene.World.DrawImage(scene.Tiles.Age1, op)
					case age < 20:
						scene.World.DrawImage(scene.Tiles.Age2, op)
					case age < 30:
						scene.World.DrawImage(scene.Tiles.Age3, op)
					default:
						scene.World.DrawImage(scene.Tiles.Age4, op)
					}
				}
			}
		}
	}

	scene.DrawMark(scene.World)

	scene.Camera.Render(scene.World, screen)

	scene.DrawDebug(screen)

	op.GeoM.Reset()
	op.GeoM.Translate(0, 0)

	scene.Game.Screen.DrawImage(screen, op)
}

func (scene *ScenePlay) DrawMark(screen *ebiten.Image) {
	if scene.Config.Markmode && scene.MarkTaken {
		x := float32(scene.Mark.X * scene.Config.Cellsize)
		y := float32(scene.Mark.Y * scene.Config.Cellsize)
		w := float32((scene.Point.X - scene.Mark.X) * scene.Config.Cellsize)
		h := float32((scene.Point.Y - scene.Mark.Y) * scene.Config.Cellsize)

		vector.StrokeRect(
			scene.World,
			x+1, y+1,
			w, h,
			1.0, scene.Old, false,
		)
	}
}

func (scene *ScenePlay) DrawDebug(screen *ebiten.Image) {
	if scene.Config.Debug {
		paused := ""
		if scene.Config.Paused {
			paused = "-- paused --"
		}

		if scene.Config.Markmode {
			paused = "-- mark --"
		}

		x, y := ebiten.CursorPosition()
		debug := fmt.Sprintf(
			DEBUG_FORMAT,
			ebiten.ActualTPS(), scene.TPG, GetMem(), scene.Generations,
			scene.Game.Scale, scene.Camera.ZoomFactor,
			scene.Camera.Position[0], scene.Camera.Position[1],
			x, y,
			paused)

		FontRenderer.Renderer.SetSizePx(10 + int(scene.Game.Scale*10))
		FontRenderer.Renderer.SetTarget(screen)

		FontRenderer.Renderer.SetColor(scene.Black)
		FontRenderer.Renderer.Draw(debug, 31, 31)

		FontRenderer.Renderer.SetColor(scene.Old)
		FontRenderer.Renderer.Draw(debug, 30, 30)

		fmt.Println(debug)
	}

}

// load a pre-computed pattern from RLE file
func (scene *ScenePlay) InitPattern() {
	scene.Grids[0].LoadRLE(scene.Config.RLE)
	scene.History.LoadRLE(scene.Config.RLE)
}

// pre-render offscreen cache image
func (scene *ScenePlay) InitCache() {
	op := &ebiten.DrawImageOptions{}

	if scene.Config.ShowGrid {
		scene.Cache.Fill(scene.Grey)
	} else {
		scene.Cache.Fill(scene.White)
	}

	for y := 0; y < scene.Config.Height; y++ {
		for x := 0; x < scene.Config.Width; x++ {
			op.GeoM.Reset()
			op.GeoM.Translate(
				float64(x*scene.Config.Cellsize),
				float64(y*scene.Config.Cellsize),
			)

			scene.Cache.DrawImage(scene.Tiles.White, op)
		}
	}
}

// initialize grid[s], either using pre-computed from state or rle file, or random
func (scene *ScenePlay) InitGrid() {
	grida := NewGrid(scene.Config.Width, scene.Config.Height, scene.Config.Density, scene.Config.Empty)
	gridb := NewGrid(scene.Config.Width, scene.Config.Height, scene.Config.Density, scene.Config.Empty)
	history := NewGrid(scene.Config.Width, scene.Config.Height, scene.Config.Density, scene.Config.Empty)

	// startup is delayed until user has selected options
	grida.FillRandom()
	grida.Copy(history)

	scene.Grids = []*Grid{
		grida,
		gridb,
	}

	scene.History = history
}

// prepare tile images
func (scene *ScenePlay) InitTiles() {
	scene.Grey = color.RGBA{128, 128, 128, 0xff}
	scene.Old = color.RGBA{255, 30, 30, 0xff}

	scene.Black = color.RGBA{0, 0, 0, 0xff}
	scene.White = color.RGBA{200, 200, 200, 0xff}
	scene.AgeColor1 = color.RGBA{255, 195, 97, 0xff} // FIXME: use slice!
	scene.AgeColor2 = color.RGBA{255, 211, 140, 0xff}
	scene.AgeColor3 = color.RGBA{255, 227, 181, 0xff}
	scene.AgeColor4 = color.RGBA{255, 240, 224, 0xff}

	if scene.Config.Invert {
		scene.White = color.RGBA{0, 0, 0, 0xff}
		scene.Black = color.RGBA{200, 200, 200, 0xff}

		scene.AgeColor1 = color.RGBA{82, 38, 0, 0xff}
		scene.AgeColor2 = color.RGBA{66, 35, 0, 0xff}
		scene.AgeColor3 = color.RGBA{43, 27, 0, 0xff}
		scene.AgeColor4 = color.RGBA{25, 17, 0, 0xff}
	}

	scene.Tiles.Black = ebiten.NewImage(scene.Config.Cellsize, scene.Config.Cellsize)
	scene.Tiles.White = ebiten.NewImage(scene.Config.Cellsize, scene.Config.Cellsize)
	scene.Tiles.Old = ebiten.NewImage(scene.Config.Cellsize, scene.Config.Cellsize)
	scene.Tiles.Age1 = ebiten.NewImage(scene.Config.Cellsize, scene.Config.Cellsize)
	scene.Tiles.Age2 = ebiten.NewImage(scene.Config.Cellsize, scene.Config.Cellsize)
	scene.Tiles.Age3 = ebiten.NewImage(scene.Config.Cellsize, scene.Config.Cellsize)
	scene.Tiles.Age4 = ebiten.NewImage(scene.Config.Cellsize, scene.Config.Cellsize)

	cellsize := scene.Config.ScreenWidth / scene.Config.Cellsize

	FillCell(scene.Tiles.Black, cellsize, scene.Black)
	FillCell(scene.Tiles.White, cellsize, scene.White)
	FillCell(scene.Tiles.Old, cellsize, scene.Old)
	FillCell(scene.Tiles.Age1, cellsize, scene.AgeColor1)
	FillCell(scene.Tiles.Age2, cellsize, scene.AgeColor2)
	FillCell(scene.Tiles.Age3, cellsize, scene.AgeColor3)
	FillCell(scene.Tiles.Age4, cellsize, scene.AgeColor4)
}

func (scene *ScenePlay) Init() {
	// setup the scene
	scene.Camera = Camera{
		ViewPort: f64.Vec2{
			float64(scene.Config.ScreenWidth),
			float64(scene.Config.ScreenHeight),
		},
		InitialZoomFactor: scene.Config.Zoomfactor,
		InitialPosition: f64.Vec2{
			scene.Config.InitialCamPos[0],
			scene.Config.InitialCamPos[1],
		},
		ZoomOutFactor: scene.Config.ZoomOutFactor,
	}

	scene.World = ebiten.NewImage(
		scene.Config.Width*scene.Config.Cellsize,
		scene.Config.Height*scene.Config.Cellsize,
	)

	scene.Cache = ebiten.NewImage(
		scene.Config.Width*scene.Config.Cellsize,
		scene.Config.Height*scene.Config.Cellsize,
	)

	scene.InitTiles()
	scene.InitCache()

	if scene.Config.DelayedStart && !scene.Config.Empty {
		scene.Config.Empty = true
		scene.InitGrid()
		scene.Config.Empty = false
	} else {
		scene.InitGrid()
	}

	scene.InitPattern()

	scene.Index = 0
	scene.TicksElapsed = 0

	scene.LastCursorPos = make([]int, 2)

	if scene.Config.Zoomfactor < 0 || scene.Config.Zoomfactor > 0 {
		scene.Camera.ZoomFactor = scene.Config.Zoomfactor
	}

	scene.Camera.Setup()
}

// count the living neighbors of a cell
func (scene *ScenePlay) CountNeighbors(x, y int) int64 {
	var sum int64

	for nbgX := -1; nbgX < 2; nbgX++ {
		for nbgY := -1; nbgY < 2; nbgY++ {
			var col, row int
			if scene.Config.Wrap {
				// In wrap mode we look at all the 8 neighbors surrounding us.
				// In case we are on an edge we'll look at the neighbor on the
				//  other side  of the  grid, thus  wrapping lookahead  around
				// using the mod() function.
				col = (x + nbgX + scene.Config.Width) % scene.Config.Width
				row = (y + nbgY + scene.Config.Height) % scene.Config.Height

			} else {
				// In traditional grid mode the edges are deadly
				if x+nbgX < 0 || x+nbgX >= scene.Config.Width || y+nbgY < 0 || y+nbgY >= scene.Config.Height {
					continue
				}
				col = x + nbgX
				row = y + nbgY
			}

			sum += scene.Grids[scene.Index].Data[row][col]
		}
	}

	// don't count ourselfes though
	sum -= scene.Grids[scene.Index].Data[y][x]

	return sum
}

// fill a cell with the given color
func FillCell(tile *ebiten.Image, cellsize int, col color.RGBA) {
	vector.DrawFilledRect(
		tile,
		float32(1),
		float32(1),
		float32(cellsize-1),
		float32(cellsize-1),
		col, false,
	)
}
