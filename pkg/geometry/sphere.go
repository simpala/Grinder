package geometry

import (
	"grinder/pkg/math"
	"grinder/pkg/motion"
	"image/color"
	gomath "math"
)
// Sphere3D represents a sphere in 3D space.
type Sphere3D struct {
	Center            math.Point3D
	Radius            float64
	Color             color.RGBA
	Shininess         float64
	SpecularIntensity float64
	SpecularColor     color.RGBA
	Motion            []motion.Keyframe
}

// Contains checks if a point is inside the sphere.
func (s Sphere3D) Contains(p math.Point3D, t float64) bool {
	center := s.CenterAt(t)
	dp := p.Sub(center)
	return dp.Dot(dp) <= s.Radius*s.Radius
}

// Intersects checks if the sphere intersects with an AABB.
func (s Sphere3D) Intersects(aabb math.AABB3D) bool {
	closestX := gomath.Max(aabb.Min.X, gomath.Min(s.Center.X, aabb.Max.X))
	closestY := gomath.Max(aabb.Min.Y, gomath.Min(s.Center.Y, aabb.Max.Y))
	closestZ := gomath.Max(aabb.Min.Z, gomath.Min(s.Center.Z, aabb.Max.Z))

	p := math.Point3D{X: closestX, Y: closestY, Z: closestZ}
	dp := p.Sub(s.Center)

	return dp.Dot(dp) <= s.Radius*s.Radius
}

// NormalAtPoint returns the normal vector at a given point on the sphere's surface.
func (s Sphere3D) NormalAtPoint(p math.Point3D, t float64) math.Normal3D {
	center := s.CenterAt(t)
	n := p.Sub(center).Normalize()
	return math.Normal3D{X: n.X, Y: n.Y, Z: n.Z}
}

func (s Sphere3D) CenterAt(t float64) math.Point3D {
	if len(s.Motion) == 0 {
		return s.Center
	}
	_, _, center, _, _ := motion.Interpolate(s.Motion, t)
	return center
}

// GetColor returns the color of the sphere.
func (s Sphere3D) GetColor() color.RGBA { return s.Color }

// GetShininess returns the shininess of the sphere.
func (s Sphere3D) GetShininess() float64 { return s.Shininess }

// GetSpecularIntensity returns the specular intensity of the sphere.
func (s Sphere3D) GetSpecularIntensity() float64 { return s.SpecularIntensity }

// GetSpecularColor returns the specular color of the sphere.
func (s Sphere3D) GetSpecularColor() color.RGBA { return s.SpecularColor }

// GetAABB returns the bounding box of the sphere.
func (s Sphere3D) GetAABB() math.AABB3D {
	if len(s.Motion) == 0 {
		return math.AABB3D{
			Min: s.Center.Sub(math.Point3D{X: s.Radius, Y: s.Radius, Z: s.Radius}),
			Max: s.Center.Add(math.Point3D{X: s.Radius, Y: s.Radius, Z: s.Radius}),
		}
	}
	min := s.Center.Sub(math.Point3D{X: s.Radius, Y: s.Radius, Z: s.Radius})
	max := s.Center.Add(math.Point3D{X: s.Radius, Y: s.Radius, Z: s.Radius})
	for _, kf := range s.Motion {
		min.X = gomath.Min(min.X, kf.Center.X-s.Radius)
		min.Y = gomath.Min(min.Y, kf.Center.Y-s.Radius)
		min.Z = gomath.Min(min.Z, kf.Center.Z-s.Radius)
		max.X = gomath.Max(max.X, kf.Center.X+s.Radius)
		max.Y = gomath.Max(max.Y, kf.Center.Y+s.Radius)
		max.Z = gomath.Max(max.Z, kf.Center.Z+s.Radius)
	}
	return math.AABB3D{Min: min, Max: max}
}

// GetCenter returns the sphere's center point.
func (s Sphere3D) GetCenter() math.Point3D {
	return s.Center
}

// IsVolumetric returns false for Sphere3D.
func (s Sphere3D) IsVolumetric() bool { return false }
