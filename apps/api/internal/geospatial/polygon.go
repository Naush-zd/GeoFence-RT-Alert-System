package geospatial

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
