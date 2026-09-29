// Copyright (c) 2026 Michael D Henderson. All rights reserved.

package hmz2ele

import (
	"encoding/json"
	"testing"
)

func TestCornerJSONRoundTrip(t *testing.T) {
	for _, want := range []VertexKey{{3, 4, WestCorner}, {5, 6, EastCorner}} {
		data, err := json.Marshal(want)
		if err != nil {
			t.Fatal(err)
		}
		var got VertexKey
		if err := json.Unmarshal(data, &got); err != nil {
			t.Fatalf("Unmarshal(%s): %v", data, err)
		}
		if got != want {
			t.Errorf("round trip of %s = %+v, want %+v", data, got, want)
		}
	}
	var c Corner
	if err := c.UnmarshalText([]byte("north")); err == nil {
		t.Error(`UnmarshalText("north") succeeded, want error`)
	}
}

func TestSideJSONRoundTrip(t *testing.T) {
	for _, want := range []EdgeKey{{1, 2, NorthSide}, {3, 4, NorthEastSide}, {5, 6, SouthEastSide}} {
		data, err := json.Marshal(want)
		if err != nil {
			t.Fatal(err)
		}
		var got EdgeKey
		if err := json.Unmarshal(data, &got); err != nil {
			t.Fatalf("Unmarshal(%s): %v", data, err)
		}
		if got != want {
			t.Errorf("round trip of %s = %+v, want %+v", data, got, want)
		}
	}
	var s Side
	for _, bad := range []string{"s", "sw", "nw", ""} {
		if err := s.UnmarshalText([]byte(bad)); err == nil {
			t.Errorf("UnmarshalText(%q) succeeded, want error", bad)
		}
	}
}
