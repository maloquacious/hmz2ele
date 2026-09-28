// Copyright (c) 2026 Michael D Henderson. All rights reserved.

package hmz2ele

import (
	"math"
	"slices"

	"github.com/maloquacious/dem2hm"
)

const (
	// SampleWindow is the side of the square block of pixels sampled around
	// a point.
	SampleWindow = 3
	// NoDataMajority is the number of no-data pixels in the block that makes
	// the sample no-data.
	NoDataMajority = 5
)

// Raster is the elevation data a sample reads: a row-major raster of
// elevations in meters, with dem2hm.NoDataPixel16 for no-data.
type Raster struct {
	Width, Height int
	Data          []int16
}

// RasterFromHeightMap returns a Raster that shares the heightmap's data.
func RasterFromHeightMap(hm *dem2hm.HeightMap16) Raster {
	return Raster{Width: hm.Width, Height: hm.Height, Data: hm.Data}
}

// Sample returns the elevation at (x, y), in meters, taken from the 3 × 3
// block of pixels centered on the pixel nearest the point.
//
// The elevation is the median of the valid pixels in the block. With an even
// number of valid pixels it is the mean of the two middle values, rounded
// half away from zero. Pixels outside the raster count as no-data. If 5 or
// more of the 9 pixels are no-data, Sample returns false.
func (r Raster) Sample(x, y float64) (int16, bool) {
	px, py := int(math.Floor(x)), int(math.Floor(y))
	var buf [SampleWindow * SampleWindow]int16
	valid := buf[:0]
	for dy := -1; dy <= 1; dy++ {
		for dx := -1; dx <= 1; dx++ {
			sx, sy := px+dx, py+dy
			if sx < 0 || sy < 0 || sx >= r.Width || sy >= r.Height {
				continue
			}
			if v := r.Data[sy*r.Width+sx]; v != dem2hm.NoDataPixel16 {
				valid = append(valid, v)
			}
		}
	}
	if len(buf)-len(valid) >= NoDataMajority {
		return 0, false
	}
	slices.Sort(valid)
	n := len(valid)
	if n%2 == 1 {
		return valid[n/2], true
	}
	sum := float64(valid[n/2-1]) + float64(valid[n/2])
	return int16(math.Round(sum / 2)), true
}
