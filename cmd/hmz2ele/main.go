// Copyright (c) 2026 Michael D Henderson. All rights reserved.

// Command hmz2ele samples a dem2hm heightmap (format 1.1) at the centers and
// corners of a flat-top hex grid and writes the elevations as JSON, with an
// optional PNG preview.
package main

import (
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/maloquacious/dem2hm"
	"github.com/maloquacious/hmz2ele"
)

func main() {
	if err := run(os.Args[1:], os.Stdout); err != nil {
		fmt.Fprintf(os.Stderr, "hmz2ele: %v\n", err)
		os.Exit(1)
	}
}

func run(args []string, stdout io.Writer) error {
	fs := flag.NewFlagSet("hmz2ele", flag.ContinueOnError)
	fs.Usage = func() {
		fmt.Fprintf(fs.Output(), "usage: hmz2ele [flags] <input.hmz>\n\n")
		fs.PrintDefaults()
	}
	apothem := fs.Int("apothem", 48, "hex apothem in raster pixels (a hex is twice this tall)")
	output := fs.String("output", "", "JSON file to write (required)")
	preview := fs.String("preview", "", "PNG preview file to write (optional)")
	previewScale := fs.Int("preview-scale", 8, "raster pixels per preview pixel, in each direction")
	showVersion := fs.Bool("version", false, "print the version and exit")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if *showVersion {
		fmt.Fprintln(stdout, hmz2ele.Version())
		return nil
	}
	if fs.NArg() != 1 {
		fs.Usage()
		return fmt.Errorf("expected one input file, got %d", fs.NArg())
	}
	if *output == "" {
		return fmt.Errorf("-output is required")
	}
	input := fs.Arg(0)

	hm, err := readHeightMap(input)
	if err != nil {
		return err
	}
	g, err := hmz2ele.NewGrid(*apothem, hm.Width, hm.Height)
	if err != nil {
		return err
	}
	elev := hmz2ele.NewElevations(filepath.Base(input), hm, g)

	if err := writeFile(*output, func(w io.Writer) error {
		enc := json.NewEncoder(w)
		enc.SetIndent("", "  ")
		return enc.Encode(elev)
	}); err != nil {
		return err
	}
	if *preview != "" {
		if err := writeFile(*preview, func(w io.Writer) error {
			return hmz2ele.WritePreview(w, g, elev.Hexes, *previewScale)
		}); err != nil {
			return err
		}
	}

	s := hmz2ele.Summarize(g, elev.Hexes)
	fmt.Fprintf(stdout, "apothem:           %d px (side %.2f px)\n", g.Apothem, g.Side)
	fmt.Fprintf(stdout, "grid:              %d columns × %d rows\n", g.Columns, g.Rows)
	fmt.Fprintf(stdout, "hexes:             %d\n", s.Hexes)
	fmt.Fprintf(stdout, "land hexes:        %d (center above 0 m)\n", s.Land)
	fmt.Fprintf(stdout, "low hexes:         %d (center at or below 0 m)\n", s.Low)
	fmt.Fprintf(stdout, "no-data hexes:     %d\n", s.NoData)
	fmt.Fprintf(stdout, "coastal sea hexes: %d (not land, next to land)\n", s.CoastalSea)
	fmt.Fprintf(stdout, "boundary vertices: %d\n", len(elev.BoundaryVertices))
	return nil
}

func readHeightMap(path string) (*dem2hm.HeightMap16, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	hm, err := dem2hm.ReadHeightMap16(bufio.NewReader(f))
	if err != nil {
		return nil, fmt.Errorf("%s: %w", path, err)
	}
	return hm, nil
}

func writeFile(path string, write func(io.Writer) error) error {
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	w := bufio.NewWriter(f)
	if err := write(w); err != nil {
		f.Close()
		return fmt.Errorf("%s: %w", path, err)
	}
	if err := w.Flush(); err != nil {
		f.Close()
		return fmt.Errorf("%s: %w", path, err)
	}
	return f.Close()
}
