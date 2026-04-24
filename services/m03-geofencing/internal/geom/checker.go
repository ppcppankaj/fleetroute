package geom

import (
	"math"
)

type Point struct {
	Lat float64 `json:"lat"`
	Lng float64 `json:"lng"`
}

type Polygon []Point

// Ray-casting algorithm to test if a point is inside a polygon
func (p Polygon) Contains(pt Point) bool {
	intersectCount := 0
	for i := 0; i < len(p); i++ {
		p1 := p[i]
		p2 := p[(i+1)%len(p)]

		if pt.Lat > math.Min(p1.Lat, p2.Lat) && pt.Lat <= math.Max(p1.Lat, p2.Lat) {
			if pt.Lng <= math.Max(p1.Lng, p2.Lng) {
				if p1.Lat != p2.Lat {
					xinters := (pt.Lat-p1.Lat)*(p2.Lng-p1.Lng)/(p2.Lat-p1.Lat) + p1.Lng
					if p1.Lng == p2.Lng || pt.Lng <= xinters {
						intersectCount++
					}
				}
			}
		}
	}
	return intersectCount%2 != 0
}

func Distance(p1, p2 Point) float64 {
	const R = 6371e3 // metres
	phi1 := p1.Lat * math.Pi / 180
	phi2 := p2.Lat * math.Pi / 180
	deltaPhi := (p2.Lat - p1.Lat) * math.Pi / 180
	deltaLambda := (p2.Lng - p1.Lng) * math.Pi / 180

	a := math.Sin(deltaPhi/2)*math.Sin(deltaPhi/2) +
		math.Cos(phi1)*math.Cos(phi2)*
			math.Sin(deltaLambda/2)*math.Sin(deltaLambda/2)
	c := 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))

	return R * c
}
