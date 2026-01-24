package geometry

import (
	"grinder/pkg/math"
	"image/color"
	gomath "math"
)

import "grinder/pkg/motion"

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

func (b VolumeBox) Contains(p math.Point3D) bool {
	return p.X >= b.Min.X && p.X <= b.Max.X &&
		p.Y >= b.Min.Y && p.Y <= b.Max.Y &&
		p.Z >= b.Min.Z && p.Z <= b.Max.Z
}

func (b VolumeBox) Intersects(aabb math.AABB3D) bool {
	// Standard AABB-AABB intersection
	return (b.Min.X <= aabb.Max.X && b.Max.X >= aabb.Min.X) &&
		(b.Min.Y <= aabb.Max.Y && b.Max.Y >= aabb.Min.Y) &&
		(b.Min.Z <= aabb.Max.Z && b.Max.Z >= aabb.Min.Z)
}

func (b VolumeBox) NormalAtPoint(p math.Point3D) math.Normal3D {
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
func (s VolumeBox) GetColor() color.RGBA { return s.Color }

// GetShininess returns the shininess of the box.
func (s VolumeBox) GetShininess() float64 { return s.Shininess }

// GetSpecularIntensity returns the specular intensity of the box.
func (s VolumeBox) GetSpecularIntensity() float64 { return s.SpecularIntensity }

// GetSpecularColor returns the specular color of the box.
func (s VolumeBox) GetSpecularColor() color.RGBA { return s.SpecularColor }

func (b *VolumeBox) GetAABB() math.AABB3D {
	baseAABB := math.AABB3D{Min: b.Min, Max: b.Max}
	if len(b.Motion) == 0 {
		return baseAABB
	}
	size := b.Max.Sub(b.Min)
	for _, kf := range b.Motion {
		kfMin := kf.Position.Sub(size.Mul(0.5))
		kfMax := kf.Position.Add(size.Mul(0.5))
		baseAABB.Min.X = gomath.Min(baseAABB.Min.X, kfMin.X)
		baseAABB.Min.Y = gomath.Min(baseAABB.Min.Y, kfMin.Y)
		baseAABB.Min.Z = gomath.Min(baseAABB.Min.Z, kfMin.Z)
		baseAABB.Max.X = gomath.Max(baseAABB.Max.X, kfMax.X)
		baseAABB.Max.Y = gomath.Max(baseAABB.Max.Y, kfMax.Y)
		baseAABB.Max.Z = gomath.Max(baseAABB.Max.Z, kfMax.Z)
	}
	return baseAABB
}

// GetCenter returns the center of the box.
func (b VolumeBox) GetCenter() math.Point3D {
	return b.Min.Add(b.Max).Mul(0.5)
}

// IsVolumetric returns true for VolumeBox.
func (b VolumeBox) IsVolumetric() bool { return true }

// GetDensity returns the density of the volume.
func (b VolumeBox) GetDensity() float64 { return b.Density }

func (b *VolumeBox) AtTime(t float64) Shape {
	if len(b.Motion) == 0 {
		return b
	}
	newBox := *b
	size := b.Max.Sub(b.Min)
	newCenter := motion.Interpolate(b.Motion, t)
	newBox.Min = newCenter.Sub(size.Mul(0.5))
	newBox.Max = newCenter.Add(size.Mul(0.5))
	return &newBox
}
