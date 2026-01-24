package geometry

import (
	"grinder/pkg/math"
	"testing"
)

func TestSphere3D_Contains(t *testing.T) {
	sphere := Sphere3D{Center: math.Point3D{X: 0, Y: 0, Z: 0}, Radius: 1}

	// Test point inside
	pInside := math.Point3D{X: 0.5, Y: 0.5, Z: 0.5}
	if !sphere.Contains(pInside, 0.0) {
		t.Errorf("Sphere3D Contains failed: point %v should be inside", pInside)
	}

	// Test point outside
	pOutside := math.Point3D{X: 1, Y: 1, Z: 1}
	if sphere.Contains(pOutside, 0.0) {
		t.Errorf("Sphere3D Contains failed: point %v should be outside", pOutside)
	}

	// Test point on surface
	pOnSurface := math.Point3D{X: 1, Y: 0, Z: 0}
	if !sphere.Contains(pOnSurface, 0.0) {
		t.Errorf("Sphere3D Contains failed: point %v should be inside (on surface)", pOnSurface)
	}
}

func TestSphere3D_Intersects(t *testing.T) {
	sphere := Sphere3D{Center: math.Point3D{X: 0, Y: 0, Z: 0}, Radius: 1}

	// Test AABB intersecting
	aabbIntersecting := math.AABB3D{Min: math.Point3D{X: 0.5, Y: 0.5, Z: 0.5}, Max: math.Point3D{X: 1.5, Y: 1.5, Z: 1.5}}
	if !sphere.Intersects(aabbIntersecting) {
		t.Errorf("Sphere3D Intersects failed: AABB %v should intersect", aabbIntersecting)
	}

	// Test AABB not intersecting
	aabbNotIntersecting := math.AABB3D{Min: math.Point3D{X: 2, Y: 2, Z: 2}, Max: math.Point3D{X: 3, Y: 3, Z: 3}}
	if sphere.Intersects(aabbNotIntersecting) {
		t.Errorf("Sphere3D Intersects failed: AABB %v should not intersect", aabbNotIntersecting)
	}

	// Test AABB containing sphere
	aabbContaining := math.AABB3D{Min: math.Point3D{X: -2, Y: -2, Z: -2}, Max: math.Point3D{X: 2, Y: 2, Z: 2}}
	if !sphere.Intersects(aabbContaining) {
		t.Errorf("Sphere3D Intersects failed: AABB %v should intersect (containing)", aabbContaining)
	}
}
