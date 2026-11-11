package geospatial

import "testing"

func TestPolygonWKTConvertsLatitudeLongitudeToLongitudeLatitude(t *testing.T) {
	coordinates := [][]float64{{37.7, -122.5}, {37.8, -122.5}, {37.8, -122.4}, {37.7, -122.5}}

	wkt, err := PolygonWKT(coordinates)
	if err != nil {
		t.Fatalf("PolygonWKT returned an error: %v", err)
	}
	want := "POLYGON((-122.5 37.7,-122.5 37.8,-122.4 37.8,-122.5 37.7))"
	if wkt != want {
		t.Fatalf("PolygonWKT = %q, want %q", wkt, want)
	}
}

func TestPolygonWKTRejectsInvalidCoordinates(t *testing.T) {
	tests := []struct {
		name        string
		coordinates [][]float64
	}{
		{name: "too few points", coordinates: [][]float64{{0, 0}, {0, 1}, {1, 1}}},
		{name: "invalid coordinate pair", coordinates: [][]float64{{0, 0}, {0, 1}, {1, 1}, {1}}},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if _, err := PolygonWKT(test.coordinates); err == nil {
				t.Fatal("expected invalid polygon coordinates to return an error")
			}
		})
	}
}
