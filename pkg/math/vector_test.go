package math

import (
	"math"
	"testing"
)

func TestPoint3D_Add(t *testing.T) {
	p1 := Point3D{X: 1, Y: 2, Z: 3}
	p2 := Point3D{X: 4, Y: 5, Z: 6}
	result := p1.Add(p2)
	expected := Point3D{X: 5, Y: 7, Z: 9}
	if result != expected {
		t.Errorf("Add failed: got %v, want %v", result, expected)
	}
}

func TestPoint3D_Sub(t *testing.T) {
	p1 := Point3D{X: 1, Y: 2, Z: 3}
	p2 := Point3D{X: 4, Y: 5, Z: 6}
	result := p1.Sub(p2)
	expected := Point3D{X: -3, Y: -3, Z: -3}
	if result != expected {
		t.Errorf("Sub failed: got %v, want %v", result, expected)
	}
}

func TestPoint3D_Mul(t *testing.T) {
	p := Point3D{X: 1, Y: 2, Z: 3}
	scalar := 2.0
	result := p.Mul(scalar)
	expected := Point3D{X: 2, Y: 4, Z: 6}
	if result != expected {
		t.Errorf("Mul failed: got %v, want %v", result, expected)
	}
}

func TestPoint3D_Dot(t *testing.T) {
	p1 := Point3D{X: 1, Y: 2, Z: 3}
	p2 := Point3D{X: 4, Y: 5, Z: 6}
	result := p1.Dot(p2)
	expected := 32.0
	if result != expected {
		t.Errorf("Dot failed: got %v, want %v", result, expected)
	}
}

func TestPoint3D_Length(t *testing.T) {
	p := Point3D{X: 3, Y: 4, Z: 0}
	result := p.Length()
	expected := 5.0
	if result != expected {
		t.Errorf("Length failed: got %v, want %v", result, expected)
	}
}

func TestPoint3D_Normalize(t *testing.T) {
	p := Point3D{X: 3, Y: 4, Z: 0}
	result := p.Normalize()
	expected := Point3D{X: 0.6, Y: 0.8, Z: 0}
	if math.Abs(result.X-expected.X) > 1e-9 || math.Abs(result.Y-expected.Y) > 1e-9 || math.Abs(result.Z-expected.Z) > 1e-9 {
		t.Errorf("Normalize failed: got %v, want %v", result, expected)
	}
}

func TestAABB3D_Intersects(t *testing.T) {
	aabb1 := AABB3D{Min: Point3D{X: 0, Y: 0, Z: 0}, Max: Point3D{X: 2, Y: 2, Z: 2}}
	aabb2 := AABB3D{Min: Point3D{X: 1, Y: 1, Z: 1}, Max: Point3D{X: 3, Y: 3, Z: 3}}
	aabb3 := AABB3D{Min: Point3D{X: 3, Y: 3, Z: 3}, Max: Point3D{X: 4, Y: 4, Z: 4}}

	if !aabb1.Intersects(aabb2) {
		t.Errorf("AABB3D Intersects failed: aabb1 should intersect aabb2")
	}

	if aabb1.Intersects(aabb3) {
		t.Errorf("AABB3D Intersects failed: aabb1 should not intersect aabb3")
	}
}
