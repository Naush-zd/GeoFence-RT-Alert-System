package geospatial

import "testing"

func TestPointInPolygon(t *testing.T) {
	polygon := [][]float64{{0, 0}, {0, 1}, {1, 1}, {1, 0}, {0, 0}}

	if !PointInPolygon(0.5, 0.5, polygon) {
		t.Fatal("expected point inside polygon")
	}
	if PointInPolygon(2, 2, polygon) {
		t.Fatal("expected point outside polygon")
	}
}
