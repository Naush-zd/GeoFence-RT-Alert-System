package geospatial

import (
	"errors"
	"strconv"
	"strings"
)

// PolygonWKT converts API coordinates ([latitude, longitude]) to PostGIS WKT.
func PolygonWKT(coordinates [][]float64) (string, error) {
	if len(coordinates) < 4 {
		return "", errors.New("polygon must contain at least four points")
	}

	points := make([]string, 0, len(coordinates))
	for _, coordinate := range coordinates {
		if len(coordinate) != 2 {
			return "", errors.New("polygon coordinates must be latitude/longitude pairs")
		}
		points = append(points, strconv.FormatFloat(coordinate[1], 'f', -1, 64)+" "+strconv.FormatFloat(coordinate[0], 'f', -1, 64))
	}
	return "POLYGON((" + strings.Join(points, ",") + "))", nil
}
