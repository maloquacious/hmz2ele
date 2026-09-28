// Copyright (c) 2026 Michael D Henderson. All rights reserved.

package hmz2ele

import (
	"fmt"
	"image"
	"image/color"
	"image/png"
	"io"
	"math"
)

var (
	outsideColor = color.RGBA{0xff, 0xff, 0xff, 0xff}
	noDataColor  = color.RGBA{0x2b, 0x57, 0x8c, 0xff} // sea or foreign land
	lowColor     = color.RGBA{0x5e, 0xa8, 0xa7, 0xff} // valid, at or below sea level
)

// landRamp is a hypsometric color ramp for elevations above sea level, with
// more stops at low elevations, where most of the land is.
var landRamp = []struct {
	meters float64
	color  color.RGBA
}{
	{0, color.RGBA{0x3a, 0x7d, 0x44, 0xff}},
	{100, color.RGBA{0x8c, 0xb3, 0x69, 0xff}},
	{300, color.RGBA{0xd9, 0xd0, 0x8c, 0xff}},
	{800, color.RGBA{0xc2, 0x9b, 0x61, 0xff}},
	{1500, color.RGBA{0x8f, 0x6a, 0x4a, 0xff}},
	{2500, color.RGBA{0xb5, 0xa8, 0x9f, 0xff}},
	{3500, color.RGBA{0xff, 0xff, 0xff, 0xff}},
}

// WritePreview writes a PNG of the grid with each hex filled by the color of
// its center elevation and hex outlines drawn darker. Each preview pixel
// covers scale × scale raster pixels.
func WritePreview(w io.Writer, g Grid, hexes []Hex, scale int) error {
	img, err := RenderPreview(g, hexes, scale)
	if err != nil {
		return err
	}
	return png.Encode(w, img)
}

// RenderPreview draws the preview that WritePreview writes, so callers can
// draw over it. Preview pixel (px, py) covers raster pixels
// [px × scale, (px+1) × scale) × [py × scale, (py+1) × scale).
func RenderPreview(g Grid, hexes []Hex, scale int) (*image.RGBA, error) {
	if scale < 1 {
		return nil, fmt.Errorf("preview scale %d: must be at least 1", scale)
	}
	index := make([]int, g.Columns*g.Rows)
	for i := range index {
		index[i] = -1
	}
	for i, h := range hexes {
		index[h.Row*g.Columns+h.Col] = i
	}

	hexAt := func(px, py int) int {
		col, row := g.HexAt((float64(px)+0.5)*float64(scale), (float64(py)+0.5)*float64(scale))
		if col < 0 || row < 0 || col >= g.Columns || row >= g.Rows {
			return -1
		}
		return index[row*g.Columns+col]
	}

	pw, ph := (g.Width+scale-1)/scale, (g.Height+scale-1)/scale
	owner := make([]int, pw*ph)
	for py := range ph {
		for px := range pw {
			owner[py*pw+px] = hexAt(px, py)
		}
	}

	img := image.NewRGBA(image.Rect(0, 0, pw, ph))
	for py := range ph {
		for px := range pw {
			i := owner[py*pw+px]
			if i < 0 {
				img.SetRGBA(px, py, outsideColor)
				continue
			}
			c := elevationColor(hexes[i].Center)
			edge := (px+1 < pw && owner[py*pw+px+1] != i) || (py+1 < ph && owner[(py+1)*pw+px] != i)
			if edge {
				c = darken(c)
			}
			img.SetRGBA(px, py, c)
		}
	}
	return img, nil
}

func elevationColor(e *int16) color.RGBA {
	if e == nil {
		return noDataColor
	}
	m := float64(*e)
	if m <= 0 {
		return lowColor
	}
	for i := 1; i < len(landRamp); i++ {
		lo, hi := landRamp[i-1], landRamp[i]
		if m <= hi.meters || i == len(landRamp)-1 {
			t := min(1, (m-lo.meters)/(hi.meters-lo.meters))
			return color.RGBA{
				R: lerp(lo.color.R, hi.color.R, t),
				G: lerp(lo.color.G, hi.color.G, t),
				B: lerp(lo.color.B, hi.color.B, t),
				A: 0xff,
			}
		}
	}
	panic("unreachable")
}

func lerp(a, b uint8, t float64) uint8 {
	return uint8(math.Round(float64(a) + (float64(b)-float64(a))*t))
}

func darken(c color.RGBA) color.RGBA {
	scale := func(v uint8) uint8 { return uint8(uint16(v) * 3 / 4) }
	return color.RGBA{R: scale(c.R), G: scale(c.G), B: scale(c.B), A: c.A}
}
