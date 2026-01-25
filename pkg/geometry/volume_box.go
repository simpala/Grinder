package geometry

import (
	"grinder/pkg/math"
	"image/color"
	gomath "math"
)

// VolumeBox represents a volumetric box in 3D space.
type VolumeBox struct {
	Min, Max          math.Point3D
	Color             color.RGBA
	Shininess         float64
	SpecularIntensity float64
	SpecularColor     color.RGBA
	Density           float64
}

// AtTime returns the volume's state at a specific time t.
func (b *VolumeBox) AtTime(t float64, shape Shape) {
	// Volumes are static, so nothing to do.
}

func (b *VolumeBox) Contains(p math.Point3D) bool {
	return p.X >= b.Min.X && p.X <= b.Max.X &&
		p.Y >= b.Min.Y && p.Y <= b.Max.Y &&
		p.Z >= b.Min.Z && p.Z <= b.Max.Z
}

func (b *VolumeBox) Intersects(aabb math.AABB3D) bool {
	// Standard AABB-AABB intersection
	return (b.Min.X <= aabb.Max.X && b.Max.X >= aabb.Min.X) &&
		(b.Min.Y <= aabb.Max.Y && b.Max.Y >= aabb.Min.Y) &&
		(b.Min.Z <= aabb.Max.Z && b.Max.Z >= aabb.Min.Z)
}

func (b *VolumeBox) NormalAtPoint(p math.Point3D) math.Normal3D {
	// Find which face the point is closest to
	eps := 0.0001
	if gomath.Abs(p.X-b.Min.X) < eps {
		return math.Normal3D{X: -1, Y: 0, Z: 0}
	}
	if gomath.Abs(p.X-b.Max.X) < eps {
		return math.Normal3D{X: 1, Y: 0, Z: 0}
	}
	if gomath.Abs(p.Y-b.Min.Y) < eps {
		return math.Normal3D{X: 0, Y: -1, Z: 0}
	}
	if gomath.Abs(p.Y-b.Max.Y) < eps {
		return math.Normal3D{X: 0, Y: 1, Z: 0}
	}
	if gomath.Abs(p.Z-b.Min.Z) < eps {
		return math.Normal3D{X: 0, Y: 0, Z: -1}
	}
	return math.Normal3D{X: 0, Y: 0, Z: 1}
}

// GetColor returns the color of the box.
func (s *VolumeBox) GetColor() color.RGBA { return s.Color }

// GetShininess returns the shininess of the box.
func (s *VolumeBox) GetShininess() float64 { return s.Shininess }

// GetSpecularIntensity returns the specular intensity of the box.
func (s *VolumeBox) GetSpecularIntensity() float64 { return s.SpecularIntensity }

// GetSpecularColor returns the specular color of the box.
func (s *VolumeBox) GetSpecularColor() color.RGBA { return s.SpecularColor }

func (b *VolumeBox) GetAABB() math.AABB3D { return math.AABB3D{Min: b.Min, Max: b.Max} }

// GetCenter returns the center of the box.
func (b *VolumeBox) GetCenter() math.Point3D {
	return b.Min.Add(b.Max).Mul(0.5)
}

// IsVolumetric returns true for VolumeBox.
func (b *VolumeBox) IsVolumetric() bool { return true }

// GetDensity returns the density of the volume.
func (b *VolumeBox) GetDensity() float64 { return b.Density }
