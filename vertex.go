// Copyright (c) 2026 Michael D Henderson. All rights reserved.

package hmz2ele

import (
	"fmt"
	"math"

	"github.com/maloquacious/hexg"
)

// HexAt returns the column and row of the hex containing the point. The hex
// may be outside the grid.
func (g Grid) HexAt(x, y float64) (col, row int) {
	// hexg's odd-q layout shifts odd columns down, matching the grid, once
	// hex (0, 0) is centered where Center puts it.
	x0, y0 := g.Center(0, 0)
	l := hexg.NewLayout(hexg.OddQ, hexg.Point{X: g.Side, Y: g.Side}, hexg.Point{X: x0, Y: y0})
	oc := l.CubeToOffset(l.PixelToHexRounded(hexg.Point{X: x, Y: y}))
	return oc.Col, oc.Row
}

// NearestVertex returns the vertex nearest the point. The nearest vertex is
// always a corner of the hex containing the point.
func (g Grid) NearestVertex(x, y float64) VertexKey {
	col, row := g.HexAt(x, y)
	var best VertexKey
	bestD := math.Inf(1)
	for corner := range 6 {
		k := CornerOwner(col, row, corner)
		vx, vy := g.Vertex(k)
		if d := math.Hypot(vx-x, vy-y); d < bestD {
			best, bestD = k, d
		}
	}
	return best
}

// HexCorner is a hex and one of its corners, numbered clockwise from east
// (0°) in y-down raster coordinates: 0 east, 1 south-east, 2 south-west,
// 3 west, 4 north-west, 5 north-east.
type HexCorner struct {
	Col, Row, Corner int
}

// VertexHexes returns the three hexes that meet at the vertex, with the
// vertex's corner number in each.
func VertexHexes(k VertexKey) [3]HexCorner {
	if k.Corner == EastCorner {
		nc, nr := Neighbor(k.Col, k.Row, NorthEast)
		sc, sr := Neighbor(k.Col, k.Row, SouthEast)
		return [3]HexCorner{{k.Col, k.Row, 0}, {nc, nr, 2}, {sc, sr, 4}}
	}
	nc, nr := Neighbor(k.Col, k.Row, NorthWest)
	sc, sr := Neighbor(k.Col, k.Row, SouthWest)
	return [3]HexCorner{{k.Col, k.Row, 3}, {nc, nr, 1}, {sc, sr, 5}}
}

// VertexNeighbors returns the three vertices joined to the vertex by a hex
// edge.
func VertexNeighbors(k VertexKey) [3]VertexKey {
	var out [3]VertexKey
	n := 0
	for _, h := range VertexHexes(k) {
		for _, step := range [2]int{1, 5} {
			v := CornerOwner(h.Col, h.Row, (h.Corner+step)%6)
			dup := false
			for _, o := range out[:n] {
				dup = dup || o == v
			}
			if !dup {
				out[n] = v
				n++
			}
		}
	}
	if n != 3 {
		panic(fmt.Sprintf("vertex %v: found %d neighbors, want 3", k, n))
	}
	return out
}

// Side names an edge a hex owns. Each hex owns its north, north-east, and
// south-east edges; its south, south-west, and north-west edges are the
// north, north-east, and south-east edges of its south, south-west, and
// north-west neighbors.
type Side int

const (
	NorthSide Side = iota
	NorthEastSide
	SouthEastSide
)

func (s Side) String() string {
	switch s {
	case NorthSide:
		return "n"
	case NorthEastSide:
		return "ne"
	case SouthEastSide:
		return "se"
	}
	return fmt.Sprintf("Side(%d)", int(s))
}

// MarshalText encodes the side as "n", "ne", or "se".
func (s Side) MarshalText() ([]byte, error) {
	return []byte(s.String()), nil
}

// UnmarshalText decodes "n", "ne", or "se" into the side.
func (s *Side) UnmarshalText(text []byte) error {
	switch string(text) {
	case "n":
		*s = NorthSide
	case "ne":
		*s = NorthEastSide
	case "se":
		*s = SouthEastSide
	default:
		return fmt.Errorf("invalid side %q", text)
	}
	return nil
}

// EdgeKey identifies an edge by the hex that owns it and which of that hex's
// three owned sides it is.
type EdgeKey struct {
	Col, Row int
	Side     Side
}

// edgeOwner returns the owner of the edge between corners c and c+1 of the
// hex (col, row).
func edgeOwner(col, row, c int) EdgeKey {
	switch c {
	case 4: // north: corners 4 and 5
		return EdgeKey{col, row, NorthSide}
	case 5: // north-east: corners 5 and 0
		return EdgeKey{col, row, NorthEastSide}
	case 0: // south-east: corners 0 and 1
		return EdgeKey{col, row, SouthEastSide}
	case 1: // south: the south neighbor's north edge
		nc, nr := Neighbor(col, row, South)
		return EdgeKey{nc, nr, NorthSide}
	case 2: // south-west: the south-west neighbor's north-east edge
		nc, nr := Neighbor(col, row, SouthWest)
		return EdgeKey{nc, nr, NorthEastSide}
	case 3: // north-west: the north-west neighbor's south-east edge
		nc, nr := Neighbor(col, row, NorthWest)
		return EdgeKey{nc, nr, SouthEastSide}
	}
	panic(fmt.Sprintf("invalid corner %d", c))
}

// EdgeBetween returns the edge joining two vertices, or false if they are
// not neighbors.
func EdgeBetween(u, v VertexKey) (EdgeKey, bool) {
	for _, h := range VertexHexes(u) {
		if CornerOwner(h.Col, h.Row, (h.Corner+1)%6) == v {
			return edgeOwner(h.Col, h.Row, h.Corner), true
		}
		if CornerOwner(h.Col, h.Row, (h.Corner+5)%6) == v {
			return edgeOwner(h.Col, h.Row, (h.Corner+5)%6), true
		}
	}
	return EdgeKey{}, false
}

// EdgeVertices returns the edge's two vertices, in clockwise order around
// the hex that owns it.
func EdgeVertices(e EdgeKey) (VertexKey, VertexKey) {
	var c int
	switch e.Side {
	case NorthSide:
		c = 4
	case NorthEastSide:
		c = 5
	case SouthEastSide:
		c = 0
	default:
		panic(fmt.Sprintf("invalid side %d", e.Side))
	}
	return CornerOwner(e.Col, e.Row, c), CornerOwner(e.Col, e.Row, (c+1)%6)
}
