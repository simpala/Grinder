package geometry

import (
	"grinder/pkg/math"
	"testing"
)

func TestPlane3D_Contains(t *testing.T) {
	plane := Plane3D{Point: math.Point3D{X: 0, Y: 0, Z: 0}, Normal: math.Normal3D{X: 0, Y: 1, Z: 0}}

	// Test point "under"
	pUnder := math.Point3D{X: 0, Y: -1, Z: 0}
	if !plane.Contains(pUnder) {
		t.Errorf("Plane3D Contains failed: point %v should be under", pUnder)
	}

	// Test point "above"
	pAbove := math.Point3D{X: 0, Y: 1, Z: 0}
	if plane.Contains(pAbove) {
		t.Errorf("Plane3D Contains failed: point %v should be above", pAbove)
	}

	// Test point on plane
	pOnPlane := math.Point3D{X: 1, Y: 0, Z: 1}
	if !plane.Contains(pOnPlane) {
		t.Errorf("Plane3D Contains failed: point %v should be on plane", pOnPlane)
	}
}

func TestPlane3D_Intersects(t *testing.T) {
	plane := Plane3D{Point: math.Point3D{X: 0, Y: 0, Z: 0}, Normal: math.Normal3D{X: 0, Y: 1, Z: 0}}

	// Test AABB intersecting
	aabbIntersecting := math.AABB3D{Min: math.Point3D{X: -1, Y: -1, Z: -1}, Max: math.Point3D{X: 1, Y: 1, Z: 1}}
	if !plane.Intersects(aabbIntersecting) {
		t.Errorf("Plane3D Intersects failed: AABB %v should intersect", aabbIntersecting)
	}

	// Test AABB fully "under"
	aabbUnder := math.AABB3D{Min: math.Point3D{X: -1, Y: -2, Z: -1}, Max: math.Point3D{X: 1, Y: -1, Z: 1}}
	if !plane.Intersects(aabbUnder) {
		t.Errorf("Plane3D Intersects failed: AABB %v should intersect (under)", aabbUnder)
	}

	// Test AABB fully "above"
	aabbAbove := math.AABB3D{Min: math.Point3D{X: -1, Y: 1, Z: -1}, Max: math.Point3D{X: 1, Y: 2, Z: 1}}
	if plane.Intersects(aabbAbove) {
		t.Errorf("Plane3D Intersects failed: AABB %v should not intersect (above)", aabbAbove)
	}
}
