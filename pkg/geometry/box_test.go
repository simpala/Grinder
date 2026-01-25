package geometry

import (
	"grinder/pkg/math"
	"testing"
)

func TestBox3D_Contains(t *testing.T) {
	box := Box3D{Min: math.Point3D{X: -1, Y: -1, Z: -1}, Max: math.Point3D{X: 1, Y: 1, Z: 1}}

	// Test point inside
	pInside := math.Point3D{X: 0, Y: 0, Z: 0}
	if !box.Contains(pInside, 0.0) {
		t.Errorf("Box3D Contains failed: point %v should be inside", pInside)
	}

	// Test point outside
	pOutside := math.Point3D{X: 2, Y: 2, Z: 2}
	if box.Contains(pOutside, 0.0) {
		t.Errorf("Box3D Contains failed: point %v should be outside", pOutside)
	}

	// Test point on surface
	pOnSurface := math.Point3D{X: 1, Y: 0, Z: 0}
	if !box.Contains(pOnSurface, 0.0) {
		t.Errorf("Box3D Contains failed: point %v should be inside (on surface)", pOnSurface)
	}
}

func TestBox3D_Intersects(t *testing.T) {
	box := Box3D{Min: math.Point3D{X: -1, Y: -1, Z: -1}, Max: math.Point3D{X: 1, Y: 1, Z: 1}}

	// Test AABB intersecting
	aabbIntersecting := math.AABB3D{Min: math.Point3D{X: 0, Y: 0, Z: 0}, Max: math.Point3D{X: 2, Y: 2, Z: 2}}
	if !box.Intersects(aabbIntersecting) {
		t.Errorf("Box3D Intersects failed: AABB %v should intersect", aabbIntersecting)
	}

	// Test AABB not intersecting
	aabbNotIntersecting := math.AABB3D{Min: math.Point3D{X: 2, Y: 2, Z: 2}, Max: math.Point3D{X: 3, Y: 3, Z: 3}}
	if box.Intersects(aabbNotIntersecting) {
		t.Errorf("Box3D Intersects failed: AABB %v should not intersect", aabbNotIntersecting)
	}

	// Test AABB containing box
	aabbContaining := math.AABB3D{Min: math.Point3D{X: -2, Y: -2, Z: -2}, Max: math.Point3D{X: 2, Y: 2, Z: 2}}
	if !box.Intersects(aabbContaining) {
		t.Errorf("Box3D Intersects failed: AABB %v should intersect (containing)", aabbContaining)
	}
}
