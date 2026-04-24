package geom_test

import (
	"math"
	"testing"

	"gpsgo/services/m03-geofencing/internal/geom"
)

// ── Point-in-Polygon (ray-casting) ───────────────────────────────────────────

func squarePolygon() geom.Polygon {
	return geom.Polygon{
		{Lat: 0, Lng: 0},
		{Lat: 0, Lng: 10},
		{Lat: 10, Lng: 10},
		{Lat: 10, Lng: 0},
	}
}

func TestPolygon_Contains_Inside(t *testing.T) {
	if !squarePolygon().Contains(geom.Point{Lat: 5, Lng: 5}) {
		t.Fatal("centre point must be inside square")
	}
}

func TestPolygon_Contains_Outside(t *testing.T) {
	if squarePolygon().Contains(geom.Point{Lat: 15, Lng: 15}) {
		t.Fatal("far point must be outside square")
	}
}

func TestPolygon_Contains_OnEdge_NoError(t *testing.T) {
	// Edge handling is implementation-defined; just must not panic
	_ = squarePolygon().Contains(geom.Point{Lat: 0, Lng: 5})
}

func TestPolygon_Contains_Triangle_Inside(t *testing.T) {
	tri := geom.Polygon{
		{Lat: 0, Lng: 5},
		{Lat: 10, Lng: 0},
		{Lat: 10, Lng: 10},
	}
	if !tri.Contains(geom.Point{Lat: 8, Lng: 5}) {
		t.Fatal("centroid-ish point must be inside triangle")
	}
}

func TestPolygon_Contains_Triangle_Outside(t *testing.T) {
	tri := geom.Polygon{
		{Lat: 0, Lng: 5},
		{Lat: 10, Lng: 0},
		{Lat: 10, Lng: 10},
	}
	// Clearly outside: far negative coords
	if tri.Contains(geom.Point{Lat: -5, Lng: 5}) {
		t.Fatal("negative-lat point is clearly outside triangle")
	}
}

func TestPolygon_Contains_Empty_ReturnsFalse(t *testing.T) {
	if (geom.Polygon{}).Contains(geom.Point{Lat: 1, Lng: 1}) {
		t.Fatal("empty polygon must always return false")
	}
}

// ── Haversine Distance ────────────────────────────────────────────────────────

func TestDistance_SamePoint(t *testing.T) {
	p := geom.Point{Lat: 19.076, Lng: 72.877}
	if d := geom.Distance(p, p); d != 0 {
		t.Fatalf("distance from point to itself must be 0, got %v", d)
	}
}

func TestDistance_KnownPair(t *testing.T) {
	// Mumbai CST to Dadar is ~4.5 km
	cst   := geom.Point{Lat: 18.9400, Lng: 72.8347}
	dadar := geom.Point{Lat: 19.0178, Lng: 72.8478}
	d := geom.Distance(cst, dadar)
	// Allow ±20% tolerance around 8.8km (haversine)
	if d < 7000 || d > 10000 {
		t.Fatalf("expected ~8.8 km between CST and Dadar, got %.0f m", d)
	}
}

func TestDistance_Symmetry(t *testing.T) {
	a := geom.Point{Lat: 28.6139, Lng: 77.2090} // Delhi
	b := geom.Point{Lat: 19.0760, Lng: 72.8777} // Mumbai
	if math.Abs(geom.Distance(a, b)-geom.Distance(b, a)) > 1.0 {
		t.Fatal("distance must be symmetric")
	}
}
