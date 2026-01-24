package geometry

import (
	"grinder/pkg/math"
	"testing"
)

func TestCylinder3D_Contains(t *testing.T) {
	cylinder := Cylinder3D{Center: math.Point3D{X: 0, Y: 0, Z: 0}, Radius: 1, Height: 2}

	// Test point inside
	pInside := math.Point3D{X: 0.5, Y: 1, Z: 0.5}
	if !cylinder.Contains(pInside) {
		t.Errorf("Cylinder3D Contains failed: point %v should be inside", pInside)
	}

	// Test point outside radius
	pOutsideRadius := math.Point3D{X: 1.5, Y: 1, Z: 1.5}
	if cylinder.Contains(pOutsideRadius) {
		t.Errorf("Cylinder3D Contains failed: point %v should be outside radius", pOutsideRadius)
	}

	// Test point outside height
	pOutsideHeight := math.Point3D{X: 0.5, Y: 3, Z: 0.5}
	if cylinder.Contains(pOutsideHeight) {
		t.Errorf("Cylinder3D Contains failed: point %v should be outside height", pOutsideHeight)
	}

	// Test point on surface
	pOnSurface := math.Point3D{X: 1, Y: 1, Z: 0}
	if !cylinder.Contains(pOnSurface) {
		t.Errorf("Cylinder3D Contains failed: point %v should be inside (on surface)", pOnSurface)
	}
}

func TestCylinder3D_Intersects(t *testing.T) {
	cylinder := Cylinder3D{Center: math.Point3D{X: 0, Y: 0, Z: 0}, Radius: 1, Height: 2}

	// Test AABB intersecting
	aabbIntersecting := math.AABB3D{Min: math.Point3D{X: 0.5, Y: 0.5, Z: 0.5}, Max: math.Point3D{X: 1.5, Y: 1.5, Z: 1.5}}
	if !cylinder.Intersects(aabbIntersecting) {
		t.Errorf("Cylinder3D Intersects failed: AABB %v should intersect", aabbIntersecting)
	}

	// Test AABB not intersecting
	aabbNotIntersecting := math.AABB3D{Min: math.Point3D{X: 2, Y: 2, Z: 2}, Max: math.Point3D{X: 3, Y: 3, Z: 3}}
	if cylinder.Intersects(aabbNotIntersecting) {
		t.Errorf("Cylinder3D Intersects failed: AABB %v should not intersect", aabbNotIntersecting)
	}

	// Test AABB containing cylinder
	aabbContaining := math.AABB3D{Min: math.Point3D{X: -2, Y: -1, Z: -2}, Max: math.Point3D{X: 2, Y: 3, Z: 2}}
	if !cylinder.Intersects(aabbContaining) {
		t.Errorf("Cylinder3D Intersects failed: AABB %v should intersect (containing)", aabbContaining)
	}
}
