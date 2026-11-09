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

// PointInPolygon reports whether a latitude/longitude point is inside a
// closed polygon using the ray-casting algorithm.
func PointInPolygon(latitude, longitude float64, polygon [][]float64) bool {
	inside := false
	for i, j := 0, len(polygon)-1; i < len(polygon); j, i = i, i+1 {
		if len(polygon[i]) != 2 || len(polygon[j]) != 2 {
			continue
		}

		latitudeIntersects := (polygon[i][0] > latitude) != (polygon[j][0] > latitude)
		if !latitudeIntersects {
			continue
		}

		intersectionLongitude := (polygon[j][1]-polygon[i][1])*
			(latitude-polygon[i][0])/(polygon[j][0]-polygon[i][0]) + polygon[i][1]
		if longitude < intersectionLongitude {
			inside = !inside
		}
	}
	return inside
}
