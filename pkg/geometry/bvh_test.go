package geometry

import (
	"grinder/pkg/math"
	"testing"
)

func TestBVH(t *testing.T) {
	shapes := []Shape{
		Sphere3D{Center: math.Point3D{X: 0, Y: 0, Z: 0}, Radius: 1},
		Sphere3D{Center: math.Point3D{X: 10, Y: 10, Z: 10}, Radius: 1},
	}

	bvh := NewBVH(shapes)

	if bvh.Root == nil {
		t.Fatal("BVH root is nil")
	}

	// Test intersecting shapes
	aabb1 := math.AABB3D{
		Min: math.Point3D{X: -0.5, Y: -0.5, Z: -0.5},
		Max: math.Point3D{X: 0.5, Y: 0.5, Z: 0.5},
	}
	res1 := bvh.IntersectsShapes(aabb1)
	if len(res1) != 1 {
		t.Errorf("Expected 1 shape, got %d", len(res1))
	}

	aabb2 := math.AABB3D{
		Min: math.Point3D{X: 9.5, Y: 9.5, Z: 9.5},
		Max: math.Point3D{X: 10.5, Y: 10.5, Z: 10.5},
	}
	res2 := bvh.IntersectsShapes(aabb2)
	if len(res2) != 1 {
		t.Errorf("Expected 1 shape, got %d", len(res2))
	}

	aabb3 := math.AABB3D{
		Min: math.Point3D{X: -20, Y: -20, Z: -20},
		Max: math.Point3D{X: 20, Y: 20, Z: 20},
	}
	res3 := bvh.IntersectsShapes(aabb3)
	if len(res3) != 2 {
		t.Errorf("Expected 2 shapes, got %d", len(res3))
	}

    aabb4 := math.AABB3D{
		Min: math.Point3D{X: 5, Y: 5, Z: 5},
		Max: math.Point3D{X: 6, Y: 6, Z: 6},
	}
	res4 := bvh.IntersectsShapes(aabb4)
	if len(res4) != 0 {
		t.Errorf("Expected 0 shapes, got %d", len(res4))
	}
}

func TestBVHWithInfiniteShape(t *testing.T) {
	shapes := []Shape{
		Sphere3D{Center: math.Point3D{X: 0, Y: 0, Z: 0}, Radius: 1},
		Plane3D{Point: math.Point3D{X: 0, Y: -1, Z: 0}, Normal: math.Normal3D{X: 0, Y: 1, Z: 0}},
	}

	bvh := NewBVH(shapes)

	if len(bvh.InfiniteShapes) != 1 {
		t.Errorf("Expected 1 infinite shape, got %d", len(bvh.InfiniteShapes))
	}

	aabb := math.AABB3D{
		Min: math.Point3D{X: 5, Y: 5, Z: 5},
		Max: math.Point3D{X: 6, Y: 6, Z: 6},
	}
	res := bvh.IntersectsShapes(aabb)
	// Should include the plane because it's infinite and always considered intersecting for simplicity or filtered correctly by Plane3D.Intersects.
	// Actually Plane3D.Intersects should check.
	foundPlane := false
	for _, s := range res {
		if _, ok := s.(Plane3D); ok {
			foundPlane = true
		}
	}
	if !foundPlane {
		t.Error("Expected to find plane in results")
	}
}
