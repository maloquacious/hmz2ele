// Copyright (c) 2026 Michael D Henderson. All rights reserved.

package hmz2ele

import (
	"cmp"
	"slices"

	"github.com/maloquacious/dem2hm"
)

// Elevations is the hex grid's sampled elevations, as written to JSON.
// Elevations are in meters; nil means no-data.
type Elevations struct {
	Hmz2eleVersion   string        `json:"hmz2ele_version"`
	Heightmap        HeightmapInfo `json:"heightmap"`
	Grid             GridInfo      `json:"grid"`
	Sampling         SamplingInfo  `json:"sampling"`
	Hexes            []Hex         `json:"hexes"`
	BoundaryVertices []Vertex      `json:"boundary_vertices"`
}

// HeightmapInfo identifies the heightmap the elevations were sampled from.
type HeightmapInfo struct {
	FileName string          `json:"file_name"`
	Metadata dem2hm.Metadata `json:"metadata"`
}

// GridInfo describes the hex grid. See Grid.
type GridInfo struct {
	Layout     string  `json:"layout"`
	ApothemPx  int     `json:"apothem_px"`
	SidePx     float64 `json:"side_px"`
	Columns    int     `json:"columns"`
	Rows       int     `json:"rows"`
	WidthPx    int     `json:"width_px"`
	HeightPx   int     `json:"height_px"`
	HexCount   int     `json:"hex_count"`
	CornerNote string  `json:"corner_note"`
}

// SamplingInfo describes how each elevation was sampled. See Raster.Sample.
type SamplingInfo struct {
	Window         int    `json:"window"`
	Statistic      string `json:"statistic"`
	NoDataMajority int    `json:"no_data_majority"`
}

// Hex is one hex's center elevation and the elevations of the two corners
// it owns.
type Hex struct {
	Col    int    `json:"col"`
	Row    int    `json:"row"`
	Center *int16 `json:"center"`
	West   *int16 `json:"west"`
	East   *int16 `json:"east"`
}

// Vertex is a corner used by a hex in the grid but owned by a hex outside
// it.
type Vertex struct {
	Col       int    `json:"col"`
	Row       int    `json:"row"`
	Corner    Corner `json:"corner"`
	Elevation *int16 `json:"elevation"`
}

// Build samples the center and owned corners of every hex in the grid, plus
// every corner that a hex in the grid uses but a hex outside the grid owns.
// Hexes are ordered by row, then column.
func Build(r Raster, g Grid) []Hex {
	var hexes []Hex
	for row := range g.Rows {
		for col := range g.Columns {
			if !g.Contains(col, row) {
				continue
			}
			hexes = append(hexes, Hex{
				Col:    col,
				Row:    row,
				Center: r.sampleOrNil(g.Center(col, row)),
				West:   r.sampleOrNil(g.West(col, row)),
				East:   r.sampleOrNil(g.East(col, row)),
			})
		}
	}
	return hexes
}

// BuildBoundaryVertices samples every corner that a hex in the grid uses
// but a hex outside the grid owns. They are ordered by owner row, owner
// column, then corner.
func BuildBoundaryVertices(r Raster, g Grid) []Vertex {
	seen := map[VertexKey]bool{}
	var vertices []Vertex
	for row := range g.Rows {
		for col := range g.Columns {
			if !g.Contains(col, row) {
				continue
			}
			for corner := range 6 {
				k := CornerOwner(col, row, corner)
				if g.Contains(k.Col, k.Row) || seen[k] {
					continue
				}
				seen[k] = true
				vertices = append(vertices, Vertex{
					Col:       k.Col,
					Row:       k.Row,
					Corner:    k.Corner,
					Elevation: r.sampleOrNil(g.Vertex(k)),
				})
			}
		}
	}
	slices.SortFunc(vertices, func(a, b Vertex) int {
		return cmp.Or(
			cmp.Compare(a.Row, b.Row),
			cmp.Compare(a.Col, b.Col),
			cmp.Compare(a.Corner, b.Corner),
		)
	})
	return vertices
}

// NewElevations samples the grid and returns the complete JSON document.
func NewElevations(fileName string, hm *dem2hm.HeightMap16, g Grid) *Elevations {
	r := RasterFromHeightMap(hm)
	hexes := Build(r, g)
	return &Elevations{
		Hmz2eleVersion: Version().String(),
		Heightmap:      HeightmapInfo{FileName: fileName, Metadata: hm.Metadata},
		Grid: GridInfo{
			Layout:     "flat-top, even columns shifted down; center x = s + 1.5·s·col, y = a + 2a·row (+ a if col is even)",
			ApothemPx:  g.Apothem,
			SidePx:     g.Side,
			Columns:    g.Columns,
			Rows:       g.Rows,
			WidthPx:    g.Width,
			HeightPx:   g.Height,
			HexCount:   len(hexes),
			CornerNote: "each hex owns its west and east corners; its NE and SE corners are the west corners of its NE and SE neighbors, its NW and SW corners the east corners of its NW and SW neighbors",
		},
		Sampling: SamplingInfo{
			Window:         SampleWindow,
			Statistic:      "median of valid pixels; mean of the two middle values, rounded half away from zero, when their count is even",
			NoDataMajority: NoDataMajority,
		},
		Hexes:            hexes,
		BoundaryVertices: BuildBoundaryVertices(r, g),
	}
}

func (r Raster) sampleOrNil(x, y float64) *int16 {
	v, ok := r.Sample(x, y)
	if !ok {
		return nil
	}
	return new(v)
}
