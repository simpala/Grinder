package geometry

import (
	"grinder/pkg/math"
	"image/color"
	gomath "math"
)

// Box3D represents a solid box (AABB) in 3D space.
type Box3D struct {
	Min, Max          math.Point3D
	Velocity          math.Point3D // Displacement over the shutter window
	Color             color.RGBA
	Shininess         float64
	SpecularIntensity float64
	SpecularColor     color.RGBA
}

// GetBoxAt returns the box at a specific time t
func (b Box3D) GetBoxAt(t float64) Box3D {
	displacement := b.Velocity.Mul(t)
	return Box3D{
		Min: b.Min.Add(displacement),
		Max: b.Max.Add(displacement),
	}
}

func (b Box3D) Contains(p math.Point3D, t float64) bool {
	boxAtT := b.GetBoxAt(t)
	return p.X >= boxAtT.Min.X && p.X <= boxAtT.Max.X &&
		p.Y >= boxAtT.Min.Y && p.Y <= boxAtT.Max.Y &&
		p.Z >= boxAtT.Min.Z && p.Z <= boxAtT.Max.Z
}

func (b Box3D) Intersects(aabb math.AABB3D) bool {
	// Standard AABB-AABB intersection
	return (b.Min.X <= aabb.Max.X && b.Max.X >= aabb.Min.X) &&
		(b.Min.Y <= aabb.Max.Y && b.Max.Y >= aabb.Min.Y) &&
		(b.Min.Z <= aabb.Max.Z && b.Max.Z >= aabb.Min.Z)
}

func (b Box3D) NormalAtPoint(p math.Point3D, t float64) math.Normal3D {
	boxAtT := b.GetBoxAt(t)
	// Find which face the point is closest to
	eps := 0.0001
	if gomath.Abs(p.X-boxAtT.Min.X) < eps {
		return math.Normal3D{X: -1, Y: 0, Z: 0}
	}
	if gomath.Abs(p.X-boxAtT.Max.X) < eps {
		return math.Normal3D{X: 1, Y: 0, Z: 0}
	}
	if gomath.Abs(p.Y-boxAtT.Min.Y) < eps {
		return math.Normal3D{X: 0, Y: -1, Z: 0}
	}
	if gomath.Abs(p.Y-boxAtT.Max.Y) < eps {
		return math.Normal3D{X: 0, Y: 1, Z: 0}
	}
	if gomath.Abs(p.Z-boxAtT.Min.Z) < eps {
		return math.Normal3D{X: 0, Y: 0, Z: -1}
	}
	return math.Normal3D{X: 0, Y: 0, Z: 1}
}

// GetColor returns the color of the box.
func (s Box3D) GetColor() color.RGBA { return s.Color }

// GetShininess returns the shininess of the box.
func (s Box3D) GetShininess() float64 { return s.Shininess }

// GetSpecularIntensity returns the specular intensity of the box.
func (s Box3D) GetSpecularIntensity() float64 { return s.SpecularIntensity }

// GetSpecularColor returns the specular color of the box.
func (s Box3D) GetSpecularColor() color.RGBA { return s.SpecularColor }

func (b Box3D) GetAABB() math.AABB3D {
	startBox := b.GetBoxAt(0)
	endBox := b.GetBoxAt(1)

	minP := math.Point3D{
		X: gomath.Min(startBox.Min.X, endBox.Min.X),
		Y: gomath.Min(startBox.Min.Y, endBox.Min.Y),
		Z: gomath.Min(startBox.Min.Z, endBox.Min.Z),
	}
	maxP := math.Point3D{
		X: gomath.Max(startBox.Max.X, endBox.Max.X),
		Y: gomath.Max(startBox.Max.Y, endBox.Max.Y),
		Z: gomath.Max(startBox.Max.Z, endBox.Max.Z),
	}
	return math.AABB3D{Min: minP, Max: maxP}
}

// GetCenter returns the center of the box.
func (b Box3D) GetCenter() math.Point3D {
	return b.Min.Add(b.Max).Mul(0.5)
}

// IsVolumetric returns false for Box3D.
func (b Box3D) IsVolumetric() bool { return false }
