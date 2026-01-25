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

// AtTime returns the sphere's state at a specific time t.
func (s *Sphere3D) AtTime(t float64, shape Shape) {
	if len(s.Motion) > 0 {
		if sphere, ok := shape.(*Sphere3D); ok {
			sphere.Center = motion.Interpolate(t, s.Motion)
		}
	}
}

// Contains checks if a point is inside the sphere.
func (s *Sphere3D) Contains(p math.Point3D) bool {
	dp := p.Sub(s.Center)
	return dp.Dot(dp) <= s.Radius*s.Radius
}

// Intersects checks if the sphere intersects with an AABB.
func (s *Sphere3D) Intersects(aabb math.AABB3D) bool {
	closestX := gomath.Max(aabb.Min.X, gomath.Min(s.Center.X, aabb.Max.X))
	closestY := gomath.Max(aabb.Min.Y, gomath.Min(s.Center.Y, aabb.Max.Y))
	closestZ := gomath.Max(aabb.Min.Z, gomath.Min(s.Center.Z, aabb.Max.Z))

	p := math.Point3D{X: closestX, Y: closestY, Z: closestZ}
	dp := p.Sub(s.Center)

	return dp.Dot(dp) <= s.Radius*s.Radius
}

// NormalAtPoint returns the normal vector at a given point on the sphere's surface.
func (s *Sphere3D) NormalAtPoint(p math.Point3D) math.Normal3D {
	n := p.Sub(s.Center).Normalize()
	return math.Normal3D{X: n.X, Y: n.Y, Z: n.Z}
}

// GetColor returns the color of the sphere.
func (s *Sphere3D) GetColor() color.RGBA { return s.Color }

// GetShininess returns the shininess of the sphere.
func (s *Sphere3D) GetShininess() float64 { return s.Shininess }

// GetSpecularIntensity returns the specular intensity of the sphere.
func (s *Sphere3D) GetSpecularIntensity() float64 { return s.SpecularIntensity }

// GetSpecularColor returns the specular color of the sphere.
func (s *Sphere3D) GetSpecularColor() color.RGBA { return s.SpecularColor }

// GetAABB returns the bounding box of the sphere.
func (s *Sphere3D) GetAABB() math.AABB3D {
	if len(s.Motion) == 0 {
		return math.AABB3D{
			Min: s.Center.Sub(math.Point3D{X: s.Radius, Y: s.Radius, Z: s.Radius}),
			Max: s.Center.Add(math.Point3D{X: s.Radius, Y: s.Radius, Z: s.Radius}),
		}
	}

	// If there's motion, calculate the AABB that contains the sphere at all keyframes.
	min := math.Point3D{X: gomath.Inf(1), Y: gomath.Inf(1), Z: gomath.Inf(1)}
	max := math.Point3D{X: gomath.Inf(-1), Y: gomath.Inf(-1), Z: gomath.Inf(-1)}

	for _, kf := range s.Motion {
		center := kf.Position
		min.X = gomath.Min(min.X, center.X-s.Radius)
		min.Y = gomath.Min(min.Y, center.Y-s.Radius)
		min.Z = gomath.Min(min.Z, center.Z-s.Radius)
		max.X = gomath.Max(max.X, center.X+s.Radius)
		max.Y = gomath.Max(max.Y, center.Y+s.Radius)
		max.Z = gomath.Max(max.Z, center.Z+s.Radius)
	}

	return math.AABB3D{Min: min, Max: max}
}

// GetCenter returns the sphere's center point.
func (s *Sphere3D) GetCenter() math.Point3D {
	if len(s.Motion) > 0 {
		return motion.Interpolate(0, s.Motion)
	}
	return s.Center
}

// IsVolumetric returns false for Sphere3D.
func (s *Sphere3D) IsVolumetric() bool { return false }
