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

Hexes are flat-top, with odd columns shifted down by half a hex, so column 0 sits up. This is Worldographer's `COLUMNS` layout.
Geometry uses continuous raster coordinates: pixel `(i, j)` covers `[i, i+1) × [j, j+1)`.
For apothem `a` and side `s = 2a / sqrt(3)`, the center of the hex in column `c` and row `r` is:

```text
x = s + 1.5 × s × c
y = a + 2a × r + (a if c is odd, else 0)
```

The first column's west corner is at `x = 0`.

Columns and rows are numbered from 0.
A hex is in the grid if its column and row are both at least 0 and its center pixel, `(floor(x), floor(y))`, is inside the raster.
(Without the first condition, the odd-column hexes of row −1, whose centers are at `y = 0`, would count too.)
`columns` counts the columns whose center x satisfies `floor(x) < width`, and `rows` the rows whose center y in column 0 satisfies `floor(y) < height`; even columns sit higher, so they hold the most rows.

The six corners are at these exact offsets from the center, clockwise from east (y down):

| Corner     | Offset       |
| ---------- | ------------ |
| East       | (s, 0)       |
| South-east | (s/2, a)     |
| South-west | (−s/2, a)    |
| West       | (−s, 0)      |
| North-west | (−s/2, −a)   |
| North-east | (s/2, −a)    |

These are `s × (cos θ, sin θ)` for `θ = 0°, 60°, …, 300°`, but an implementation must use the exact values: the corners' y coordinates are whole pixels, so `floor()` of a computed sine can land one pixel row off.

## Sampling

A hex has three sample points: its center, and its west and east corners (the two corners it owns; see [Output](#output)).
Each sample point gets the median of the valid pixels in the 3 × 3 block centered on the pixel nearest the point.
With an even number of valid pixels, it is the mean of the two middle values, rounded half away from zero.
If 5 or more of the 9 pixels are no-data, the sample is no-data.
Block pixels outside the raster count as no-data.

## Output

Elevations are `int16` meters; `null` means no-data.

```json
{
  "hmz2ele_version": "0.4.0",
  "heightmap": { "file_name": "pandemokh.hmz", "metadata": { ... } },
  "grid": { "apothem_px": 48, "side_px": 55.43, "columns": 106, "rows": 222, "hex_count": 23479, ... },
  "sampling": { "window": 3, "statistic": "...", "no_data_majority": 5 },
  "hexes": [ { "col": 49, "row": 0, "center": 398, "west": 505, "east": null }, ... ],
  "boundary_vertices": [ { "col": -1, "row": -1, "corner": "east", "elevation": null }, ... ]
}
```

`heightmap.metadata` holds the heightmap's `dem2hm` metadata with the same keys and values, re-encoded: a number is written in Go's shortest form, so a `0.0` in the heightmap becomes `0`.

Each hex owns its west and east corners, so every vertex is stored exactly once.
A hex's other corners are owned by its neighbors:

| Corner     | Owner                                  |
| ---------- | -------------------------------------- |
| North-east | the north-east neighbor's west corner  |
| South-east | the south-east neighbor's west corner  |
| South-west | the south-west neighbor's east corner  |
| North-west | the north-west neighbor's east corner  |

Corners used by hexes in the grid but owned by hexes outside it are listed in `boundary_vertices`, keyed by their owner's column and row, which may be negative or past the grid's edge, and sampled like any other corner.
They are ordered by owner row, owner column, then corner (`west` before `east`).

Hexes are ordered by row, then column.
Not every column and row within `columns` × `rows` is in the grid: in the last row, odd columns may fall outside the raster.

## Preview

With scale `k` (`-preview-scale`), the preview is `⌈width / k⌉ × ⌈height / k⌉` pixels, and preview pixel `(px, py)` covers raster pixels `[px·k, (px+1)·k) × [py·k, (py+1)·k)`.
Its color comes from the hex containing the raster point at the middle of that square, `((px + 0.5)·k, (py + 0.5)·k)`, found by cube rounding (`Grid.HexAt`); a point in no hex of the grid is white (`#ffffff`).

A hex's color comes from its center elevation `e`:

| Center elevation | Color |
| ---------------- | ----- |
| no-data | `#2b578c` (dark blue) |
| `e ≤ 0` | `#5ea8a7` (teal) |
| `e > 0` | the land ramp below |

The land ramp is a hypsometric ramp with more stops at low elevations, where most of the land is:

| Meters | Color |
| -----: | ----- |
| 0 | `#3a7d44` |
| 100 | `#8cb369` |
| 300 | `#d9d08c` |
| 800 | `#c29b61` |
| 1,500 | `#8f6a4a` |
| 2,500 | `#b5a89f` |
| 3,500 | `#ffffff` |

For `e` between two stops, each channel is interpolated linearly, `lo + (hi − lo)·t` with `t = (e − lo_m) / (hi_m − lo_m)`, and rounded half away from zero; above 3,500 m the color is white.

A preview pixel is an **outline** pixel if its right or lower neighbor, where that neighbor is in the image, belongs to a different hex or to no hex. Outline pixels have each channel of the hex's color multiplied by 3/4, truncated.

## Package

The `hmz2ele` package holds the grid geometry that other pipeline tools share, so they always agree with the sampled elevations:

- `NewGrid`, `Grid.Center`, `Grid.West`, `Grid.East`, `Grid.Contains`, and `Neighbor` lay out the grid.
- `Grid.HexAt` and `Grid.NearestVertex` map a raster point to a hex and to the nearest vertex.
- `CornerOwner`, `VertexHexes`, and `VertexNeighbors` walk the vertex graph; each vertex has three neighbors along hex edges.
- `EdgeBetween` and `EdgeVertices` convert between pairs of vertices and edges, identified by the hex that owns them (`n`, `ne`, or `se`).
- `RenderPreview` draws the preview image so callers can draw over it; `WritePreview` encodes it as PNG.

## License

MIT. See `LICENSE`.
