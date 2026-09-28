// Copyright (c) 2026 Michael D Henderson. All rights reserved.

package hmz2ele

import (
	"bytes"
	"image/png"
	"testing"

	"github.com/maloquacious/dem2hm"
)

const nd = dem2hm.NoDataPixel16

// raster5 returns a 5 × 5 raster whose middle 3 × 3 block is the given
// values, surrounded by no-data.
func raster5(block [9]int16) Raster {
	data := make([]int16, 25)
	for i := range data {
		data[i] = nd
	}
	for i, v := range block {
		data[(1+i/3)*5+1+i%3] = v
	}
	return Raster{Width: 5, Height: 5, Data: data}
}

func TestSample(t *testing.T) {
	for _, tc := range []struct {
		name   string
		block  [9]int16
		want   int16
		wantOK bool
	}{
		{"median of nine", [9]int16{9, 1, 8, 2, 7, 3, 6, 4, 5}, 5, true},
		{"spike ignored", [9]int16{10, 10, 10, 10, 3000, 10, 10, 10, 10}, 10, true},
		{"four no-data, odd count", [9]int16{nd, nd, nd, nd, 1, 2, 3, 4, 5}, 3, true},
		{"even count rounds half up", [9]int16{nd, nd, nd, 1, 2, 3, 4, 5, 6}, 4, true},
		{"even count rounds half away from zero", [9]int16{nd, nd, nd, -1, -2, -3, -4, -5, -6}, -4, true},
		{"five no-data", [9]int16{nd, nd, nd, nd, nd, 2, 3, 4, 5}, 0, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			got, ok := raster5(tc.block).Sample(2.5, 2.5)
			if got != tc.want || ok != tc.wantOK {
				t.Errorf("Sample = %d, %v; want %d, %v", got, ok, tc.want, tc.wantOK)
			}
		})
	}
}

func TestSampleNearestPixelAndEdges(t *testing.T) {
	data := make([]int16, 25)
	for i := range data {
		data[i] = int16(i)
	}
	r := Raster{Width: 5, Height: 5, Data: data}
	// (2.999, 1.0) is in pixel (2, 1); its block is rows 0-2, columns 1-3.
	if got, ok := r.Sample(2.999, 1.0); !ok || got != 7 {
		t.Errorf("Sample(2.999, 1.0) = %d, %v; want 7, true", got, ok)
	}
	// On an edge, three of the nine pixels are outside the raster.
	if got, ok := r.Sample(2.5, 0.5); !ok || got != 5 {
		t.Errorf("Sample on top edge = %d, %v; want 5, true (median of 1,2,3,6,7,8 is 4.5)", got, ok)
	}
	// In a corner, five of the nine pixels are outside the raster.
	if _, ok := r.Sample(0.5, 0.5); ok {
		t.Error("Sample in corner: want no-data")
	}
}

func TestPreviewMapsCentersToTheirHex(t *testing.T) {
	g, err := NewGrid(6, 200, 300)
	if err != nil {
		t.Fatal(err)
	}
	data := make([]int16, 200*300)
	hexes := Build(Raster{Width: 200, Height: 300, Data: data}, g)
	var buf bytes.Buffer
	if err := WritePreview(&buf, g, hexes, 1); err != nil {
		t.Fatal(err)
	}
	if _, err := png.Decode(&buf); err != nil {
		t.Fatal(err)
	}
	// The preview's pixel-to-hex mapping must agree with Grid.Center.
	layout := previewLayout(g)
	for _, h := range hexes {
		x, y := g.Center(h.Col, h.Row)
		col, row := pointToHex(layout, x, y)
		if col != h.Col || row != h.Row {
			t.Errorf("center of (%d, %d) maps to (%d, %d)", h.Col, h.Row, col, row)
		}
	}
}
