// Copyright (c) 2026 Michael D Henderson. All rights reserved.

package hmz2ele

import (
	"fmt"
	"math"
)

// Grid is a flat-top hex grid laid over a raster, with even columns shifted
// down by half a hex. Hex (0, 0) is in the top-left corner and the first
// column's west corner is at x = 0.
//
// Geometry uses continuous raster coordinates: pixel (i, j) covers
// [i, i+1) × [j, j+1).
type Grid struct {
	// Apothem is the distance from a hex's center to the middle of a side,
	// in pixels. A hex is 2 × Apothem pixels tall.
	Apothem int
	// Side is the length of a hex side, which is also the distance from the
	// center to a corner, in pixels.
	Side float64
	// Columns and Rows bound the column and row numbers of hexes in the
	// grid. Not every (column, row) inside those bounds is in the grid; use
	// Contains.
	Columns, Rows int
	// Width and Height are the raster's dimensions in pixels.
	Width, Height int
}

// NewGrid returns the grid of every hex whose center pixel lies inside a
// width × height raster.
func NewGrid(apothem, width, height int) (Grid, error) {
	if apothem < 1 {
		return Grid{}, fmt.Errorf("apothem %d: must be at least 1 pixel", apothem)
	}
	if width < 1 || height < 1 {
		return Grid{}, fmt.Errorf("raster %d × %d: dimensions must be positive", width, height)
	}
	g := Grid{
		Apothem: apothem,
		Side:    2 * float64(apothem) / math.Sqrt(3),
		Width:   width,
		Height:  height,
	}
	for {
		x, _ := g.Center(g.Columns, 0)
		if int(math.Floor(x)) >= width {
			break
		}
		g.Columns++
	}
	if g.Columns == 0 {
		return Grid{}, fmt.Errorf("raster %d × %d: too narrow for apothem %d", width, height, apothem)
	}
	// Odd columns sit higher than even columns, so they hold the most rows.
	rowsCol := min(1, g.Columns-1)
	for {
		_, y := g.Center(rowsCol, g.Rows)
		if int(math.Floor(y)) >= height {
			break
		}
		g.Rows++
	}
	if g.Rows == 0 {
		return Grid{}, fmt.Errorf("raster %d × %d: too short for apothem %d", width, height, apothem)
	}
	return g, nil
}

// Center returns the center of the hex in the given column and row. It is
// defined for any column and row, including hexes outside the grid.
func (g Grid) Center(col, row int) (x, y float64) {
	a := g.Apothem
	x = g.Side + 1.5*g.Side*float64(col)
	yi := a + 2*a*row
	if isEven(col) {
		yi += a
	}
	return x, float64(yi)
}

// West and East return the hex's west (180°) and east (0°) corners, which
// are the two corners each hex owns.
func (g Grid) West(col, row int) (x, y float64) {
	cx, cy := g.Center(col, row)
	return cx - g.Side, cy
}

func (g Grid) East(col, row int) (x, y float64) {
	cx, cy := g.Center(col, row)
	return cx + g.Side, cy
}

// Contains reports whether the hex's center pixel is inside the raster.
func (g Grid) Contains(col, row int) bool {
	if col < 0 || row < 0 {
		return false
	}
	x, y := g.Center(col, row)
	return int(math.Floor(x)) < g.Width && int(math.Floor(y)) < g.Height
}

// Direction names a neighboring hex. Flat-top hexes have no east or west
// neighbors.
type Direction int

const (
	North Direction = iota
	NorthEast
	SouthEast
	South
	SouthWest
	NorthWest
)

// Directions lists every Direction, clockwise from North.
var Directions = [6]Direction{North, NorthEast, SouthEast, South, SouthWest, NorthWest}

// Neighbor returns the column and row of the hex next to (col, row) in the
// given direction. The result may be outside the grid.
func Neighbor(col, row int, d Direction) (int, int) {
	even := isEven(col)
	switch d {
	case North:
		return col, row - 1
	case South:
		return col, row + 1
	case NorthEast, NorthWest:
		dc := 1
		if d == NorthWest {
			dc = -1
		}
		if even {
			return col + dc, row
		}
		return col + dc, row - 1
	case SouthEast, SouthWest:
		dc := 1
		if d == SouthWest {
			dc = -1
		}
		if even {
			return col + dc, row + 1
		}
		return col + dc, row
	}
	panic(fmt.Sprintf("invalid direction %d", d))
}

// Corner names a corner a hex owns.
type Corner int

const (
	WestCorner Corner = iota
	EastCorner
)

func (c Corner) String() string {
	if c == WestCorner {
		return "west"
	}
	return "east"
}

// MarshalText encodes the corner as "west" or "east".
func (c Corner) MarshalText() ([]byte, error) {
	return []byte(c.String()), nil
}

// UnmarshalText decodes "west" or "east" into the corner.
func (c *Corner) UnmarshalText(text []byte) error {
	switch string(text) {
	case "west":
		*c = WestCorner
	case "east":
		*c = EastCorner
	default:
		return fmt.Errorf("invalid corner %q", text)
	}
	return nil
}

// VertexKey identifies a vertex by the hex that owns it and which of that
// hex's two owned corners it is.
type VertexKey struct {
	Col, Row int
	Corner   Corner
}

// CornerOwner returns the owner of the hex's corner that lies between the
// given direction and the next direction clockwise. The north-east corner is
// between NorthEast and SouthEast, and so on. The owner of a hex's own east
// and west corners is the hex itself.
//
// Every vertex touches exactly three hexes, and exactly one of them has the
// vertex as its west or east corner, so each vertex has exactly one owner.
func CornerOwner(col, row int, corner int) VertexKey {
	// Corners are numbered clockwise from east, matching angles
	// 0°, 60°, …, 300° in y-down raster coordinates.
	switch corner {
	case 0:
		return VertexKey{col, row, EastCorner}
	case 1: // south-east corner: the south-east neighbor's west corner
		c, r := Neighbor(col, row, SouthEast)
		return VertexKey{c, r, WestCorner}
	case 2: // south-west corner: the south-west neighbor's east corner
		c, r := Neighbor(col, row, SouthWest)
		return VertexKey{c, r, EastCorner}
	case 3:
		return VertexKey{col, row, WestCorner}
	case 4: // north-west corner: the north-west neighbor's east corner
		c, r := Neighbor(col, row, NorthWest)
		return VertexKey{c, r, EastCorner}
	case 5: // north-east corner: the north-east neighbor's west corner
		c, r := Neighbor(col, row, NorthEast)
		return VertexKey{c, r, WestCorner}
	}
	panic(fmt.Sprintf("invalid corner %d", corner))
}

// Vertex returns the position of the vertex.
func (g Grid) Vertex(k VertexKey) (x, y float64) {
	if k.Corner == WestCorner {
		return g.West(k.Col, k.Row)
	}
	return g.East(k.Col, k.Row)
}

func isEven(n int) bool {
	return n&1 == 0
}
