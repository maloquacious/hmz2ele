// Copyright (c) 2026 Michael D Henderson. All rights reserved.

package hmz2ele

// Summary counts hexes by what their center sample shows.
type Summary struct {
	// Hexes is the number of hexes in the grid.
	Hexes int
	// Land hexes have a center elevation above sea level (0 m).
	Land int
	// Low hexes have a valid center elevation at or below sea level.
	Low int
	// NoData hexes have a no-data center: sea, or foreign land at a border
	// cut.
	NoData int
	// CoastalSea hexes are Low or NoData hexes next to at least one Land hex.
	CoastalSea int
}

// IsLand reports whether the hex's center elevation is above sea level.
func (h Hex) IsLand() bool {
	return h.Center != nil && *h.Center > 0
}

// Summarize counts the hexes.
func Summarize(g Grid, hexes []Hex) Summary {
	land := make(map[[2]int]bool, len(hexes))
	for _, h := range hexes {
		if h.IsLand() {
			land[[2]int{h.Col, h.Row}] = true
		}
	}
	s := Summary{Hexes: len(hexes), Land: len(land)}
	for _, h := range hexes {
		if h.IsLand() {
			continue
		}
		if h.Center == nil {
			s.NoData++
		} else {
			s.Low++
		}
		for _, d := range Directions {
			c, r := Neighbor(h.Col, h.Row, d)
			if land[[2]int{c, r}] {
				s.CoastalSea++
				break
			}
		}
	}
	return s
}
