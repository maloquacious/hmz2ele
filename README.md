# hmz2ele

`hmz2ele` samples a `dem2hm` heightmap (format 1.1) at the centers and corners of a flat-top hex grid and writes the elevations as JSON, with an optional PNG preview.

## Usage

```text
go run ./cmd/hmz2ele [flags] <input.hmz>
```

Flags:

- `-apothem <px>` sets the hex apothem in raster pixels. A hex is twice this tall. The default is `48`, the campaign grid's apothem.
- `-output <file>` is the JSON file to write. Required.
- `-preview <file>` is a PNG preview to write. Optional.
- `-preview-scale <n>` sets how many raster pixels, in each direction, one preview pixel covers. The default is `8`.
- `-version` prints the version.

The command prints a summary: grid size, hex count, and the number of land, low, no-data, and coastal sea hexes.

## Grid

Hexes are flat-top, with even columns shifted down by half a hex.
Geometry uses continuous raster coordinates: pixel `(i, j)` covers `[i, i+1) × [j, j+1)`.
For apothem `a` and side `s = 2a / sqrt(3)`, the center of the hex in column `c` and row `r` is:

```text
x = s + 1.5 × s × c
y = a + 2a × r + (a if c is even, else 0)
```

The first column's west corner is at `x = 0`.
A hex is in the grid if its center pixel, `(floor(x), floor(y))`, is inside the raster.

## Sampling

Each sample point gets the median of the valid pixels in the 3 × 3 block centered on the pixel nearest the point.
With an even number of valid pixels, it is the mean of the two middle values, rounded half away from zero.
If 5 or more of the 9 pixels are no-data, the sample is no-data.
Block pixels outside the raster count as no-data.

## Output

Elevations are `int16` meters; `null` means no-data.

```json
{
  "hmz2ele_version": "0.1.0",
  "heightmap": { "file_name": "pandemokh.hmz", "metadata": { ... } },
  "grid": { "apothem_px": 48, "side_px": 55.43, "columns": 106, "rows": 222, "hex_count": 23479, ... },
  "sampling": { "window": 3, "statistic": "...", "no_data_majority": 5 },
  "hexes": [ { "col": 18, "row": 47, "center": 81, "west": 148, "east": null }, ... ],
  "boundary_vertices": [ { "col": 0, "row": -1, "corner": "east", "elevation": null }, ... ]
}
```

`heightmap.metadata` is the heightmap's `dem2hm` metadata, copied unchanged.

Each hex owns its west and east corners, so every vertex is stored exactly once.
A hex's other corners are owned by its neighbors:

| Corner     | Owner                                  |
| ---------- | -------------------------------------- |
| North-east | the north-east neighbor's west corner  |
| South-east | the south-east neighbor's west corner  |
| South-west | the south-west neighbor's east corner  |
| North-west | the north-west neighbor's east corner  |

Corners used by hexes in the grid but owned by hexes outside it are listed in `boundary_vertices`, keyed by their owner's column and row, which may be negative or past the grid's edge.

Hexes are ordered by row, then column.
Not every column and row within `columns` × `rows` is in the grid: in the last row, even columns may fall outside the raster.

## Preview

The preview fills each hex with the color of its center elevation, using a hypsometric ramp with more color stops at low elevations, and darkens hex outlines.
No-data hexes are dark blue, hexes at or below 0 m are teal, and areas outside the grid are white.

## License

MIT. See `LICENSE`.
