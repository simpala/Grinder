package geometry

import (
	"grinder/pkg/math"
	"grinder/pkg/motion"
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
	Motion            []motion.Keyframe
}

func (b VolumeBox) Contains(p math.Point3D, t float64) bool {
	min, max := b.MinMaxAt(t)
	return p.X >= min.X && p.X <= max.X &&
		p.Y >= min.Y && p.Y <= max.Y &&
		p.Z >= min.Z && p.Z <= max.Z
}

func (b VolumeBox) Intersects(aabb math.AABB3D) bool {
	// Standard AABB-AABB intersection
	return (b.Min.X <= aabb.Max.X && b.Max.X >= aabb.Min.X) &&
		(b.Min.Y <= aabb.Max.Y && b.Max.Y >= aabb.Min.Y) &&
		(b.Min.Z <= aabb.Max.Z && b.Max.Z >= aabb.Min.Z)
}

func (b VolumeBox) NormalAtPoint(p math.Point3D, t float64) math.Normal3D {
	min, max := b.MinMaxAt(t)
	eps := 0.0001
	if gomath.Abs(p.X-min.X) < eps {
		return math.Normal3D{X: -1, Y: 0, Z: 0}
	}
	if gomath.Abs(p.X-max.X) < eps {
		return math.Normal3D{X: 1, Y: 0, Z: 0}
	}
	if gomath.Abs(p.Y-min.Y) < eps {
		return math.Normal3D{X: 0, Y: -1, Z: 0}
	}
	if gomath.Abs(p.Y-max.Y) < eps {
		return math.Normal3D{X: 0, Y: 1, Z: 0}
	}
	if gomath.Abs(p.Z-min.Z) < eps {
		return math.Normal3D{X: 0, Y: 0, Z: -1}
	}
	return math.Normal3D{X: 0, Y: 0, Z: 1}
}

func (b VolumeBox) MinMaxAt(t float64) (math.Point3D, math.Point3D) {
	if len(b.Motion) == 0 {
		return b.Min, b.Max
	}
	_, _, _, min, max := motion.Interpolate(b.Motion, t)
	return min, max
}

// GetColor returns the color of the box.
func (s VolumeBox) GetColor() color.RGBA { return s.Color }

// GetShininess returns the shininess of the box.
func (s VolumeBox) GetShininess() float64 { return s.Shininess }

// GetSpecularIntensity returns the specular intensity of the box.
func (s VolumeBox) GetSpecularIntensity() float64 { return s.SpecularIntensity }

// GetSpecularColor returns the specular color of the box.
func (s VolumeBox) GetSpecularColor() color.RGBA { return s.SpecularColor }

func (b VolumeBox) GetAABB() math.AABB3D {
	if len(b.Motion) == 0 {
		return math.AABB3D{Min: b.Min, Max: b.Max}
	}
	min := b.Min
	max := b.Max
	for _, kf := range b.Motion {
		min.X = gomath.Min(min.X, kf.Min.X)
		min.Y = gomath.Min(min.Y, kf.Min.Y)
		min.Z = gomath.Min(min.Z, kf.Min.Z)
		max.X = gomath.Max(max.X, kf.Max.X)
		max.Y = gomath.Max(max.Y, kf.Max.Y)
		max.Z = gomath.Max(max.Z, kf.Max.Z)
	}
	return math.AABB3D{Min: min, Max: max}
}

// GetCenter returns the center of the box.
func (b VolumeBox) GetCenter() math.Point3D {
	return b.Min.Add(b.Max).Mul(0.5)
}

// IsVolumetric returns true for VolumeBox.
func (b VolumeBox) IsVolumetric() bool { return true }

// GetDensity returns the density of the volume.
func (b VolumeBox) GetDensity() float64 { return b.Density }
