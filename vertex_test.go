// Copyright (c) 2026 Michael D Henderson. All rights reserved.

package hmz2ele

import (
	"math"
	"testing"
)

func TestHexAtAndNearestVertex(t *testing.T) {
	g, err := NewGrid(10, 1000, 1000)
	if err != nil {
		t.Fatal(err)
	}
	for row := range 8 {
		for col := range 8 {
			cx, cy := g.Center(col, row)
			if c, r := g.HexAt(cx, cy); c != col || r != row {
				t.Errorf("HexAt(center of (%d, %d)) = (%d, %d)", col, row, c, r)
			}
			// A point just inside each corner is nearest that corner.
			for corner := range 6 {
				rad := float64(corner) * math.Pi / 3
				x, y := cx+0.9*g.Side*math.Cos(rad), cy+0.9*g.Side*math.Sin(rad)
				if got, want := g.NearestVertex(x, y), CornerOwner(col, row, corner); got != want {
					t.Errorf("hex (%d, %d) corner %d: NearestVertex = %v, want %v", col, row, corner, got, want)
				}
			}
		}
	}
}

func TestVertexGraph(t *testing.T) {
	g, err := NewGrid(10, 1000, 1000)
	if err != nil {
		t.Fatal(err)
	}
	for row := range 6 {
		for col := range 6 {
			for _, corner := range []Corner{WestCorner, EastCorner} {
				k := VertexKey{col + 2, row + 2, corner}
				kx, ky := g.Vertex(k)
				for _, h := range VertexHexes(k) {
					if got := CornerOwner(h.Col, h.Row, h.Corner); got != k {
						t.Errorf("%v: hex (%d, %d) corner %d is %v", k, h.Col, h.Row, h.Corner, got)
					}
				}
				for _, n := range VertexNeighbors(k) {
					nx, ny := g.Vertex(n)
					if d := math.Hypot(nx-kx, ny-ky); math.Abs(d-g.Side) > 1e-9 {
						t.Errorf("%v: neighbor %v is %v away, want %v", k, n, d, g.Side)
					}
					e, ok := EdgeBetween(k, n)
					if !ok {
						t.Fatalf("%v: no edge to neighbor %v", k, n)
					}
					if e2, _ := EdgeBetween(n, k); e2 != e {
						t.Errorf("EdgeBetween(%v, %v) = %v, reversed = %v", k, n, e, e2)
					}
					a, b := EdgeVertices(e)
					if !(a == k && b == n || a == n && b == k) {
						t.Errorf("edge %v has vertices %v, %v; want %v and %v", e, a, b, k, n)
					}
				}
			}
		}
	}
	if _, ok := EdgeBetween(VertexKey{3, 3, WestCorner}, VertexKey{3, 3, EastCorner}); ok {
		t.Error("opposite corners of a hex should not share an edge")
	}
}
