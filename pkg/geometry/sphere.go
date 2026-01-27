package geometry

import (
	"grinder/pkg/math"
	"image/color"
	gomath "math"
)

// Sphere3D represents a sphere in 3D space.
type Sphere3D struct {
	Center            math.Point3D
	Velocity          math.Point3D // Displacement over the shutter window
	Radius            float64
	Color             color.RGBA
	Shininess         float64
	SpecularIntensity float64
	SpecularColor     color.RGBA
}

// GetCenterAt calculates the position for a specific sample's time
func (s Sphere3D) GetCenterAt(t float64) math.Point3D {
	return s.Center.Add(s.Velocity.Mul(t))
}

// Contains checks if a point is inside the sphere.
func (s Sphere3D) Contains(p math.Point3D, t float64) bool {
	center := s.GetCenterAt(t)
	dp := p.Sub(center)
	return dp.Dot(dp) <= s.Radius*s.Radius
}

// Intersects checks if the sphere intersects with an AABB.
func (s Sphere3D) Intersects(aabb math.AABB3D) bool {
	// For simplicity and to correctly handle motion blur, check the full motion block AABB.
	return s.GetAABB().Intersects(aabb)
}

// // NormalAtPoint returns the normal vector at a given point on the sphere's surface.
//
//	func (s Sphere3D) NormalAtPoint(p math.Point3D) math.Normal3D {
//		n := p.Sub(s.Center).Normalize()
//		return math.Normal3D{X: n.X, Y: n.Y, Z: n.Z}
//	}
//
// NormalAtPoint returns the normal vector at a given point on the sphere's surface at time t.
func (s Sphere3D) NormalAtPoint(p math.Point3D, t float64) math.Normal3D {
	center := s.GetCenterAt(t) // Use the helper we talked about
	n := p.Sub(center).Normalize()
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
	// The bounds must encapsulate the sphere at BOTH ends of the motion
	startCenter := s.Center
	endCenter := s.Center.Add(s.Velocity)

	minP := math.Point3D{
		X: gomath.Min(startCenter.X, endCenter.X) - s.Radius,
		Y: gomath.Min(startCenter.Y, endCenter.Y) - s.Radius,
		Z: gomath.Min(startCenter.Z, endCenter.Z) - s.Radius,
	}
	maxP := math.Point3D{
		X: gomath.Max(startCenter.X, endCenter.X) + s.Radius,
		Y: gomath.Max(startCenter.Y, endCenter.Y) + s.Radius,
		Z: gomath.Max(startCenter.Z, endCenter.Z) + s.Radius,
	}
	return math.AABB3D{Min: minP, Max: maxP}
}

// GetCenter returns the sphere's center point.
func (s Sphere3D) GetCenter() math.Point3D {
	return s.Center
}

// IsVolumetric returns false for Sphere3D.
func (s Sphere3D) IsVolumetric() bool { return false }
