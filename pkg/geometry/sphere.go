package geometry

import (
	"grinder/pkg/math"
	"image/color"
	gomath "math"
)

import "grinder/pkg/motion"

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
func (s Sphere3D) Contains(p math.Point3D) bool {
	dp := p.Sub(s.Center)
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
func (s Sphere3D) NormalAtPoint(p math.Point3D) math.Normal3D {
	n := p.Sub(s.Center).Normalize()
	return math.Normal3D{X: n.X, Y: n.Y, Z: n.Z}
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
	return math.AABB3D{
		Min: s.Center.Sub(math.Point3D{X: s.Radius, Y: s.Radius, Z: s.Radius}),
		Max: s.Center.Add(math.Point3D{X: s.Radius, Y: s.Radius, Z: s.Radius}),
	}
}

// GetCenter returns the sphere's center point.
func (s Sphere3D) GetCenter() math.Point3D {
	return s.Center
}

// IsVolumetric returns false for Sphere3D.
func (s Sphere3D) IsVolumetric() bool { return false }

func (s *Sphere3D) AtTime(t float64) Shape {
	if len(s.Motion) == 0 {
		return s
	}
	newSphere := *s
	newSphere.Center = motion.Interpolate(s.Motion, t)
	return &newSphere
}
