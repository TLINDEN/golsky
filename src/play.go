package main

import (
	"fmt"
	"image"
	"log"
	"sync"
	"unsafe"

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

type History struct {
	Age [][]int64
}

func NewHistory(height, width int) History {
	hist := History{}

	hist.Age = make([][]int64, height)
	for y := 0; y < height; y++ {
		hist.Age[y] = make([]int64, width)
	}

	return hist
}

type ScenePlay struct {
	Game   *Game
	Config *Config
	Next   SceneName
	Prev   SceneName
	Whoami SceneName

	Clear bool

	Grids         []*Grid       // 2 grids: one current, one next
	History       History       // holds state of past dead cells for evolution traces
	Index         int           // points to current grid
	Generations   int64         // Stats
	TicksElapsed  int           // tick counter for game speed
	Camera        Camera        // for zoom+move
	World, Cache  *ebiten.Image // actual image we render to
	WheelTurned   bool          // when user turns wheel multiple times, zoom faster
	Dragging      bool          // middle mouse is pressed, move canvas
	LastCursorPos []float64     // used to check if the user is dragging
	MarkTaken     bool          // true when mouse1 pressed
	MarkDone      bool          // true when mouse1 released, copy cells between Mark+Point
	Mark, Point   image.Point   // area to marks+save
	RunOneStep    bool          // mutable flags from config
	TPG           int           // current game speed (ticks per game)
	Theme         Theme
	RuleCheckFunc func(uint8, uint8) uint8
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

/* The standard Scene of Life is symbolized in rule-string notation
 * as B3/S23 (23/3 here).  A cell  is born if it has exactly three
 * neighbors,  survives if it  has two or three  living neighbors,
 * and  dies otherwise.
 * we  abbreviate the calculation: if  state is 0 and  3 neighbors
 * are a life, check will be just  3. If the cell is alive, 9 will
 * be added  to the life neighbors (to avoid  a collision with the
 * result 3), which will be 11|12 in case of 2|3 life neighbors.
 */
func (scene *ScenePlay) CheckRuleB3S23(state uint8, neighbors uint8) uint8 {
	switch (9 * state) + neighbors {
	case 11:
		fallthrough
	case 12:
		fallthrough
	case 3:
		return Alive
	}

	return Dead
}

/*
 * The generic  rule checker is able  to calculate cell state  for any
 * GOL rul, including B3/S23.
 */
func (scene *ScenePlay) CheckRuleGeneric(state uint8, neighbors uint8) uint8 {
	var nextstate uint8

	if state != 1 && Contains(scene.Config.Rule.Birth, neighbors) {
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

	var wg sync.WaitGroup
	wg.Add(scene.Config.Height)

	// compute life status of cells
	for y := 0; y < scene.Config.Height; y++ {

		go func() {
			defer wg.Done()

			for x := 0; x < scene.Config.Width; x++ {
				state := scene.Grids[scene.Index].Data[y][x].State // 0|1 == dead or alive
				neighbors := scene.Grids[scene.Index].CountNeighbors(x, y)

				// actually apply the current rules
				nextstate := scene.RuleCheckFunc(state, neighbors)

				// change state of current cell in next grid
				scene.Grids[next].Data[y][x].State = nextstate

				if scene.Config.ShowEvolution {
					// set history  to current generation so we  can infer the
					// age of the cell's state  during rendering and use it to
					// deduce the color to use if evolution tracing is enabled
					// 60FPS:
					if state != nextstate {
						scene.History.Age[y][x] = scene.Generations
					}
				}
			}
		}()
	}

	wg.Wait()

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
	// primary functions, always available
	switch {
	case inpututil.IsKeyJustPressed(ebiten.KeyEscape):
		scene.SetNext(Menu)
	case inpututil.IsKeyJustPressed(ebiten.KeyO):
		scene.SetNext(Options)
	case inpututil.IsKeyJustPressed(ebiten.KeyC):
		scene.Config.Markmode = true
		scene.Config.Drawmode = false
		scene.Config.Paused = true
	case inpututil.IsKeyJustPressed(ebiten.KeyI):
		scene.Config.Drawmode = true
		scene.Config.Paused = true
	}

	if scene.Config.Markmode {
		// no need to check any more input in mark mode
		return
	}

	switch {
	case inpututil.IsKeyJustPressed(ebiten.KeySpace) || inpututil.IsKeyJustPressed(ebiten.KeyEnter):
		scene.Config.TogglePaused()
	case inpututil.IsKeyJustPressed(ebiten.KeyPageDown):
		if scene.TPG < 120 {
			scene.TPG++
		}
	case inpututil.IsKeyJustPressed(ebiten.KeyPageUp):
		if scene.TPG >= 1 {
			scene.TPG--
		}
	case inpututil.IsKeyJustPressed(ebiten.KeyS):
		scene.SaveState()
	case inpututil.IsKeyJustPressed(ebiten.KeyD):
		scene.Config.Debug = !scene.Config.Debug
	}

	if scene.Config.Paused {
		if inpututil.IsKeyJustPressed(ebiten.KeyN) {
			scene.Config.RunOneStep = true
		}
	}
}

func (scene *ScenePlay) CheckDrawingInput() {
	if scene.Config.Drawmode {
		switch {
		case inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft):
			scene.ToggleCellOnCursorPos()
		case inpututil.IsKeyJustPressed(ebiten.KeyEscape):
			scene.Config.Drawmode = false
		}
	}
}

// Check dragging input.  move the canvas with the  mouse while pressing
// the middle mouse button, zoom in and out using the wheel.
func (scene *ScenePlay) CheckDraggingInput() {
	if scene.Config.Markmode {
		return
	}

	dragbutton := ebiten.MouseButtonLeft

	if scene.Config.Drawmode {
		dragbutton = ebiten.MouseButtonMiddle
	}

	// move canvas
	if scene.Dragging && !ebiten.IsMouseButtonPressed(dragbutton) {
		// release
		scene.Dragging = false
	}

	if !scene.Dragging && ebiten.IsMouseButtonPressed(dragbutton) {
		// start dragging
		scene.Dragging = true
		scene.LastCursorPos[0], scene.LastCursorPos[1] = scene.Camera.ScreenToWorld(ebiten.CursorPosition())
	}

	if scene.Dragging {
		x, y := scene.Camera.ScreenToWorld(ebiten.CursorPosition())

		if x != scene.LastCursorPos[0] || y != scene.LastCursorPos[1] {
			// actually drag by mouse cursor pos diff to last cursor pos
			scene.Camera.Position[0] -= float64(x - scene.LastCursorPos[0])
			scene.Camera.Position[1] -= float64(y - scene.LastCursorPos[1])
		}

		scene.LastCursorPos[0], scene.LastCursorPos[1] = scene.Camera.ScreenToWorld(ebiten.CursorPosition())
	}

	// also support the arrow keys to move the canvas
	switch {
	case ebiten.IsKeyPressed(ebiten.KeyArrowLeft):
		scene.Camera.Position[0] -= 1
	case ebiten.IsKeyPressed(ebiten.KeyArrowRight):
		scene.Camera.Position[0] += 1
	case ebiten.IsKeyPressed(ebiten.KeyArrowUp):
		scene.Camera.Position[1] -= 1
	case ebiten.IsKeyPressed(ebiten.KeyArrowDown):
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

	grid := make([][]uint8, height)

	for y := 0; y < height; y++ {
		grid[y] = make([]uint8, width)

		for x := 0; x < width; x++ {
			grid[y][x] = scene.Grids[scene.Index].Data[y+starty][x+startx].State
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
		scene.Generations = 0
		scene.InitGrid()
		scene.InitCache()
		return nil
	}

	if scene.Config.RestartCache {
		scene.Config.RestartCache = false
		scene.Theme = scene.Config.ThemeManager.GetCurrentTheme()
		scene.InitCache()
		return nil
	}

	if quit := scene.CheckExit(); quit != nil {
		return quit
	}

	scene.CheckInput()
	scene.CheckDrawingInput()
	scene.CheckDraggingInput()
	scene.CheckMarkInput()

	if !scene.Config.Paused || scene.RunOneStep {
		scene.UpdateCells()
	}

	return nil
}

// set a cell to alive or dead
func (scene *ScenePlay) ToggleCellOnCursorPos() {
	// use cursor pos relative to the world
	worldX, worldY := scene.Camera.ScreenToWorld(ebiten.CursorPosition())
	x := int(worldX) / scene.Config.Cellsize
	y := int(worldY) / scene.Config.Cellsize

	if x > -1 && y > -1 && x < scene.Config.Width && y < scene.Config.Height {
		scene.Grids[scene.Index].Data[y][x].State ^= 1
		scene.History.Age[y][x] = 1
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

	for y := 0; y < scene.Config.Height; y++ {
		for x := 0; x < scene.Config.Width; x++ {
			op.GeoM.Reset()
			op.GeoM.Translate(
				float64(x*scene.Config.Cellsize),
				float64(y*scene.Config.Cellsize),
			)

			if scene.Config.ShowEvolution {
				scene.DrawEvolution(screen, x, y, op)
			} else {
				if scene.Grids[scene.Index].Data[y][x].State == 1 {
					scene.World.DrawImage(scene.Theme.Tile(ColLife), op)
				}
			}
		}
	}

	scene.DrawMark(scene.World)

	scene.Camera.Render(scene.World, screen)

	scene.DrawDebug(screen)
}

func (scene *ScenePlay) DrawEvolution(screen *ebiten.Image, x, y int, op *ebiten.DrawImageOptions) {
	age := scene.Generations - scene.History.Age[y][x]

	switch scene.Grids[scene.Index].Data[y][x].State {
	case Alive:
		if age > 50 && scene.Config.ShowEvolution {
			scene.World.DrawImage(scene.Theme.Tile(ColOld), op)
		} else {
			scene.World.DrawImage(scene.Theme.Tile(ColLife), op)
		}
	case Dead:
		// only draw dead cells in case evolution trace is enabled
		if scene.History.Age[y][x] > 1 && scene.Config.ShowEvolution {
			switch {
			case age < 10:
				scene.World.DrawImage(scene.Theme.Tile(ColAge1), op)
			case age < 20:
				scene.World.DrawImage(scene.Theme.Tile(ColAge2), op)
			case age < 30:
				scene.World.DrawImage(scene.Theme.Tile(ColAge3), op)
			default:
				scene.World.DrawImage(scene.Theme.Tile(ColAge4), op)
			}
		}
	}
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
			1.0, scene.Theme.Color(ColOld), false,
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

		if scene.Config.Drawmode {
			paused = "-- insert --"
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

		FontRenderer.Renderer.SetColor(scene.Theme.Color(ColLife))
		FontRenderer.Renderer.Draw(debug, 31, 31)

		FontRenderer.Renderer.SetColor(scene.Theme.Color(ColOld))
		FontRenderer.Renderer.Draw(debug, 30, 30)

		fmt.Println(debug)
	}

}

// load a pre-computed pattern from RLE file
func (scene *ScenePlay) InitPattern() {
	scene.Grids[0].LoadRLE(scene.Config.RLE)

	// rule might have changed
	scene.InitRuleCheckFunc()
}

// pre-render offscreen cache image
func (scene *ScenePlay) InitCache() {
	// setup theme
	scene.Theme.SetGrid(scene.Config.ShowGrid)

	if !scene.Config.ShowGrid {
		scene.Cache.Fill(scene.Theme.Color(ColDead))
		return
	}

	op := &ebiten.DrawImageOptions{}

	scene.Cache.Fill(scene.Theme.Color(ColGrid))

	for y := 0; y < scene.Config.Height; y++ {
		for x := 0; x < scene.Config.Width; x++ {
			op.GeoM.Reset()
			op.GeoM.Translate(
				float64(x*scene.Config.Cellsize),
				float64(y*scene.Config.Cellsize),
			)

			scene.Cache.DrawImage(scene.Theme.Tile(ColDead), op)
		}
	}
}

// initialize grid[s], either using pre-computed from state or rle file, or random
func (scene *ScenePlay) InitGrid() {
	grida := NewGrid(scene.Config)
	gridb := NewGrid(scene.Config)

	// startup is delayed until user has selected options
	grida.FillRandom()

	scene.Grids = []*Grid{
		grida,
		gridb,
	}

	scene.History = NewHistory(scene.Config.Height, scene.Config.Width)

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

	scene.Theme = scene.Config.ThemeManager.GetCurrentTheme()
	scene.InitCache()

	if scene.Config.DelayedStart && !scene.Config.Empty {
		// do not fill the grid when the main menu comes up first, the
		// user decides interactively what to do
		scene.Config.Empty = true
		scene.InitGrid()
		scene.Config.Empty = false
	} else {
		scene.InitGrid()
	}

	scene.InitPattern()

	scene.Index = 0
	scene.TicksElapsed = 0

	scene.LastCursorPos = make([]float64, 2)

	if scene.Config.Zoomfactor < 0 || scene.Config.Zoomfactor > 0 {
		scene.Camera.ZoomFactor = scene.Config.Zoomfactor
	}

	scene.Camera.Setup()
}

func bool2int(b bool) int {
	return int(*(*byte)(unsafe.Pointer(&b)))
}

func (scene *ScenePlay) InitRuleCheckFunc() {
	if scene.Config.Rule.Definition == "B3/S23" {
		scene.RuleCheckFunc = scene.CheckRuleB3S23
	} else {
		scene.RuleCheckFunc = scene.CheckRuleGeneric
	}
}
