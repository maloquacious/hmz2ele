// Copyright (c) 2026 Michael D Henderson. All rights reserved.

package hmz2ele

import (
	"math"
	"testing"
)

const eps = 1e-9

func near(a, b float64) bool { return math.Abs(a-b) < eps }

func TestCenterMatchesReadmeFormula(t *testing.T) {
	g, err := NewGrid(49, 8868, 21307)
	if err != nil {
		t.Fatal(err)
	}
	s := g.Side
	for _, tc := range []struct {
		col, row int
		x, y     float64
	}{
		{0, 0, s, 98},
		{1, 0, 2.5 * s, 49},
		{2, 0, 4 * s, 98},
		{1, 3, 2.5 * s, 49 + 6*49},
		{-1, 0, -0.5 * s, 49},
	} {
		x, y := g.Center(tc.col, tc.row)
		if !near(x, tc.x) || !near(y, tc.y) {
			t.Errorf("Center(%d, %d) = (%v, %v), want (%v, %v)", tc.col, tc.row, x, y, tc.x, tc.y)
		}
	}
	if x, _ := g.West(0, 0); x != 0 {
		t.Errorf("first column's west corner x = %v, want 0", x)
	}
}

func TestPanamaGridSize(t *testing.T) {
	g, err := NewGrid(49, 8868, 21307)
	if err != nil {
		t.Fatal(err)
	}
	if g.Columns != 104 || g.Rows != 217 {
		t.Errorf("grid = %d × %d, want 104 × 217", g.Columns, g.Rows)
	}
	n := 0
	for row := range g.Rows {
		for col := range g.Columns {
			if g.Contains(col, row) {
				n++
			}
		}
	}
	if n != 22568 {
		t.Errorf("hexes in grid = %d, want 22568", n)
	}
}

func TestNeighbors(t *testing.T) {
	g, err := NewGrid(10, 1000, 1000)
	if err != nil {
		t.Fatal(err)
	}
	// Angles of each direction's neighbor, in y-down raster coordinates.
	angles := map[Direction]float64{
		North: 270, NorthEast: 330, SouthEast: 30, South: 90, SouthWest: 150, NorthWest: 210,
	}
	for _, col := range []int{4, 5} {
		cx, cy := g.Center(col, 7)
		for _, d := range Directions {
			nc, nr := Neighbor(col, 7, d)
			nx, ny := g.Center(nc, nr)
			rad := angles[d] * math.Pi / 180
			wx, wy := cx+20*math.Cos(rad), cy+20*math.Sin(rad)
			if math.Abs(nx-wx) > 1e-6 || math.Abs(ny-wy) > 1e-6 {
				t.Errorf("col %d: %v neighbor at (%v, %v), want (%v, %v)", col, d, nx, ny, wx, wy)
			}
			bc, br := Neighbor(nc, nr, Directions[(int(d)+3)%6])
			if bc != col || br != 7 {
				t.Errorf("col %d: going %v and back ends at (%d, %d)", col, d, bc, br)
			}
		}
	}
}

func TestCornerOwnership(t *testing.T) {
	g, err := NewGrid(10, 1000, 1000)
	if err != nil {
		t.Fatal(err)
	}
	type point struct{ x, y int64 }
	round := func(x, y float64) point { return point{int64(math.Round(x * 1e6)), int64(math.Round(y * 1e6))} }
	owners := map[point]map[VertexKey]bool{}
	for row := range 8 {
		for col := range 8 {
			cx, cy := g.Center(col, row)
			for corner := range 6 {
				k := CornerOwner(col, row, corner)
				vx, vy := g.Vertex(k)
				rad := float64(corner) * math.Pi / 3
				wx, wy := cx+g.Side*math.Cos(rad), cy+g.Side*math.Sin(rad)
				if math.Abs(vx-wx) > 1e-6 || math.Abs(vy-wy) > 1e-6 {
					t.Errorf("hex (%d, %d) corner %d: owner %v at (%v, %v), want (%v, %v)", col, row, corner, k, vx, vy, wx, wy)
				}
				p := round(vx, vy)
				if owners[p] == nil {
					owners[p] = map[VertexKey]bool{}
				}
				owners[p][k] = true
			}
		}
	}
	for p, keys := range owners {
		if len(keys) != 1 {
			t.Errorf("vertex at %v has %d owners: %v", p, len(keys), keys)
		}
	}
}
