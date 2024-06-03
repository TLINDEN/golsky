// this comes from the camera example but I enhanced it a little bit

package main

import (
	"fmt"
	"math"

	"github.com/hajimehoshi/ebiten/v2"
	"golang.org/x/image/math/f64"
)

type Camera struct {
	ViewPort          f64.Vec2
	Position          f64.Vec2
	ZoomFactor        int
	InitialZoomFactor int
	InitialPosition   f64.Vec2
	ZoomOutFactor     int
}

func (c *Camera) String() string {
	return fmt.Sprintf(
		"T: %.1f, S: %d",
		c.Position, c.ZoomFactor,
	)
}

func (c *Camera) viewportCenter() f64.Vec2 {
	return f64.Vec2{
		c.ViewPort[0] * 0.5,
		c.ViewPort[1] * 0.5,
	}
}

func (c *Camera) worldMatrix() ebiten.GeoM {
	m := ebiten.GeoM{}
	m.Translate(-c.Position[0], -c.Position[1])

	viewportCenter := c.viewportCenter()

	// We want to scale and rotate around center of image / screen
	m.Translate(-viewportCenter[0], -viewportCenter[1])

	m.Scale(
		math.Pow(1.01, float64(c.ZoomFactor)),
		math.Pow(1.01, float64(c.ZoomFactor)),
	)

	m.Translate(viewportCenter[0], viewportCenter[1])
	return m
}

func (c *Camera) Render(world, screen *ebiten.Image) {
	screen.DrawImage(world, &ebiten.DrawImageOptions{
		GeoM: c.worldMatrix(),
	})
}

func (c *Camera) ScreenToWorld(posX, posY int) (float64, float64) {
	inverseMatrix := c.worldMatrix()
	if inverseMatrix.IsInvertible() {
		inverseMatrix.Invert()
		return inverseMatrix.Apply(float64(posX), float64(posY))
	} else {
		// When scaling it can happened that matrix is not invertable
		return math.NaN(), math.NaN()
	}
}

func (c *Camera) Setup() {
	c.Position[0] = c.InitialPosition[0]
	c.Position[1] = c.InitialPosition[1]
	c.ZoomFactor = c.InitialZoomFactor
}

func (c *Camera) Reset() {
	c.Position[0] = c.InitialPosition[0]
	c.Position[1] = c.InitialPosition[1]
	c.ZoomFactor = c.ZoomOutFactor
}
