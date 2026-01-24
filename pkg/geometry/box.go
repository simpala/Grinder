package geometry

import (
	"grinder/pkg/math"
	"grinder/pkg/motion"
	"image/color"
	gomath "math"
)
// Box3D represents a solid box (AABB) in 3D space.
type Box3D struct {
	Min, Max          math.Point3D
	Color             color.RGBA
	Shininess         float64
	SpecularIntensity float64
	SpecularColor     color.RGBA
	Motion            []motion.Keyframe
}

func (b *Box3D) Contains(p math.Point3D, t float64) bool {
	center := b.CenterAt(t)
	halfSize := b.Max.Sub(b.Min).Mul(0.5)
	min := center.Sub(halfSize)
	max := center.Add(halfSize)
	return p.X >= min.X && p.X <= max.X &&
		p.Y >= min.Y && p.Y <= max.Y &&
		p.Z >= min.Z && p.Z <= max.Z
}

func (b *Box3D) Intersects(aabb math.AABB3D) bool {
	// Standard AABB-AABB intersection
	return (b.Min.X <= aabb.Max.X && b.Max.X >= aabb.Min.X) &&
		(b.Min.Y <= aabb.Max.Y && b.Max.Y >= aabb.Min.Y) &&
		(b.Min.Z <= aabb.Max.Z && b.Max.Z >= aabb.Min.Z)
}

func (b *Box3D) NormalAtPoint(p math.Point3D, t float64) math.Normal3D {
	center := b.CenterAt(t)
	halfSize := b.Max.Sub(b.Min).Mul(0.5)
	min := center.Sub(halfSize)
	max := center.Add(halfSize)
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

func (b *Box3D) CenterAt(t float64) math.Point3D {
	if len(b.Motion) == 0 {
		return b.GetCenter()
	}
	_, _, center, _, _ := motion.Interpolate(b.Motion, t)
	return center
}

// GetColor returns the color of the box.
func (s *Box3D) GetColor() color.RGBA { return s.Color }

// GetShininess returns the shininess of the box.
func (s *Box3D) GetShininess() float64 { return s.Shininess }

// GetSpecularIntensity returns the specular intensity of the box.
func (s *Box3D) GetSpecularIntensity() float64 { return s.SpecularIntensity }

// GetSpecularColor returns the specular color of the box.
func (s *Box3D) GetSpecularColor() color.RGBA { return s.SpecularColor }

func (b *Box3D) GetAABB() math.AABB3D {
	if len(b.Motion) == 0 {
		return math.AABB3D{Min: b.Min, Max: b.Max}
	}
	min := b.Min
	max := b.Max
	halfSize := b.Max.Sub(b.Min).Mul(0.5)
	for _, kf := range b.Motion {
		min.X = gomath.Min(min.X, kf.Center.X-halfSize.X)
		min.Y = gomath.Min(min.Y, kf.Center.Y-halfSize.Y)
		min.Z = gomath.Min(min.Z, kf.Center.Z-halfSize.Z)
		max.X = gomath.Max(max.X, kf.Center.X+halfSize.X)
		max.Y = gomath.Max(max.Y, kf.Center.Y+halfSize.Y)
		max.Z = gomath.Max(max.Z, kf.Center.Z+halfSize.Z)
	}
	return math.AABB3D{Min: min, Max: max}
}

// GetCenter returns the center of the box.
func (b *Box3D) GetCenter() math.Point3D {
	return b.Min.Add(b.Max).Mul(0.5)
}

// AtTime returns a new box interpolated to time t.
func (b *Box3D) AtTime(t float64) Shape {
	if len(b.Motion) == 0 {
		return b
	}
	newBox := *b
	newBox.Min = b.CenterAt(t).Sub(b.Max.Sub(b.Min).Mul(0.5))
	newBox.Max = b.CenterAt(t).Add(b.Max.Sub(b.Min).Mul(0.5))
	return &newBox
}

// IsVolumetric returns false for Box3D.
func (b *Box3D) IsVolumetric() bool { return false }
