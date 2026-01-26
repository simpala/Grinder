package geometry

import (
	"grinder/pkg/math"
	"image/color"
	"testing"
)

func TestBilinearQuadContains(t *testing.T) {
	// Create a simple planar quad in the XY plane
	quad := &BilinearQuad{
		P00:       math.Point3D{X: -1, Y: -1, Z: 0},
		P10:       math.Point3D{X: 1, Y: -1, Z: 0},
		P11:       math.Point3D{X: 1, Y: 1, Z: 0},
		P01:       math.Point3D{X: -1, Y: 1, Z: 0},
		Thickness: 0.01,
		Color:     color.RGBA{R: 255, G: 255, B: 0, A: 255},
	}

	// Test a point that should be inside the quad
	pointInside := math.Point3D{X: 0.5, Y: 0.5, Z: 0.005} // slightly above the surface
	if !quad.Contains(pointInside, 0.0) {
		t.Errorf("Point inside quad should be contained")
	}

	// Test a point that should be outside the quad
	pointOutside := math.Point3D{X: 2.0, Y: 2.0, Z: 0.0}
	if quad.Contains(pointOutside, 0.0) {
		t.Errorf("Point outside quad should not be contained")
	}

	// Test a point that should be inside (center)
	pointCenter := math.Point3D{X: 0.0, Y: 0.0, Z: 0.005} // slightly above the surface
	if !quad.Contains(pointCenter, 0.0) {
		t.Errorf("Center point should be contained")
	}

	// Test a point that should be outside (too far from surface)
	pointFar := math.Point3D{X: 0.0, Y: 0.0, Z: 0.1} // too far from surface
	if quad.Contains(pointFar, 0.0) {
		t.Errorf("Point too far from surface should not be contained")
	}
}

func TestBilinearQuadPositionAt(t *testing.T) {
	// Create a simple planar quad in the XY plane
	quad := &BilinearQuad{
		P00:       math.Point3D{X: -1, Y: -1, Z: 0},
		P10:       math.Point3D{X: 1, Y: -1, Z: 0},
		P11:       math.Point3D{X: 1, Y: 1, Z: 0},
		P01:       math.Point3D{X: -1, Y: 1, Z: 0},
		Thickness: 0.01,
		Color:     color.RGBA{R: 255, G: 255, B: 0, A: 255},
	}

	// Test corner positions
	if pos := quad.PositionAt(0, 0); pos != quad.P00 {
		t.Errorf("PositionAt(0,0) should equal P00, got %v", pos)
	}
	if pos := quad.PositionAt(1, 0); pos != quad.P10 {
		t.Errorf("PositionAt(1,0) should equal P10, got %v", pos)
	}
	if pos := quad.PositionAt(1, 1); pos != quad.P11 {
		t.Errorf("PositionAt(1,1) should equal P11, got %v", pos)
	}
	if pos := quad.PositionAt(0, 1); pos != quad.P01 {
		t.Errorf("PositionAt(0,1) should equal P01, got %v", pos)
	}

	// Test center position
	centerExpected := math.Point3D{X: 0, Y: 0, Z: 0}
	centerActual := quad.PositionAt(0.5, 0.5)
	if centerActual.Sub(centerExpected).Length() > 1e-10 {
		t.Errorf("Center position should be (0,0,0), got %v", centerActual)
	}
}
