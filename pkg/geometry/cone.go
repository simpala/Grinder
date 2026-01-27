package geometry

import (
	"grinder/pkg/math"
	"image/color"
	gomath "math"
)

// Cone3D represents a cone in 3D space.
type Cone3D struct {
	Center            math.Point3D // Center of the circular base
	Velocity          math.Point3D // Displacement over the shutter window
	Radius            float64
	Height            float64
	Color             color.RGBA
	Shininess         float64
	SpecularIntensity float64
	SpecularColor     color.RGBA
}

// GetCenterAt calculates the position for a specific sample's time
func (c Cone3D) GetCenterAt(t float64) math.Point3D {
	return c.Center.Add(c.Velocity.Mul(t))
}

func (c Cone3D) Contains(p math.Point3D, t float64) bool {
	center := c.GetCenterAt(t)
	// Check height bounds
	if p.Y < center.Y || p.Y > center.Y+c.Height {
		return false
	}

	// Calculate relative height (0 at base, 1 at tip)
	hRatio := (p.Y - center.Y) / c.Height
	// Radius at this specific height
	currentRadius := c.Radius * (1.0 - hRatio)

	dx, dz := p.X-center.X, p.Z-center.Z
	return (dx*dx + dz*dz) <= currentRadius*currentRadius
}

func (c Cone3D) Intersects(aabb math.AABB3D) bool {
	// Account for motion by using the full motion-expanded AABB
	return c.GetAABB().Intersects(aabb)
}

func (c Cone3D) NormalAtPoint(p math.Point3D, t float64) math.Normal3D {
	center := c.GetCenterAt(t)
	eps := 0.0001
	// Bottom cap
	if p.Y <= center.Y+eps {
		return math.Normal3D{X: 0, Y: -1, Z: 0}
	}

	// Side normal: Slanted outward and slightly upward
	dx, dz := p.X-center.X, p.Z-center.Z
	horizontalDist := gomath.Sqrt(dx*dx + dz*dz)

	// The slope of the cone side
	slope := c.Radius / c.Height
	n := math.Point3D{X: dx / horizontalDist, Y: slope, Z: dz / horizontalDist}.Normalize()
	return math.Normal3D{X: n.X, Y: n.Y, Z: n.Z}
}

func (c Cone3D) GetAABB() math.AABB3D {
	startCenter := c.GetCenterAt(0)
	endCenter := c.GetCenterAt(1)

	minP := math.Point3D{
		X: gomath.Min(startCenter.X, endCenter.X) - c.Radius,
		Y: gomath.Min(startCenter.Y, endCenter.Y),
		Z: gomath.Min(startCenter.Z, endCenter.Z) - c.Radius,
	}
	maxP := math.Point3D{
		X: gomath.Max(startCenter.X, endCenter.X) + c.Radius,
		Y: gomath.Max(startCenter.Y, endCenter.Y) + c.Height,
		Z: gomath.Max(startCenter.Z, endCenter.Z) + c.Radius,
	}
	return math.AABB3D{Min: minP, Max: maxP}
}

// GetColor returns the color of the cone.
func (s Cone3D) GetColor() color.RGBA { return s.Color }

// GetShininess returns the shininess of the cone.
func (s Cone3D) GetShininess() float64 { return s.Shininess }

// GetSpecularIntensity returns the specular intensity of the cone.
func (s Cone3D) GetSpecularIntensity() float64 { return s.SpecularIntensity }

// GetSpecularColor returns the specular color of the cone.
func (s Cone3D) GetSpecularColor() color.RGBA { return s.SpecularColor }

// GetCenter returns the geometric center of the cone.
func (c Cone3D) GetCenter() math.Point3D {
	return math.Point3D{
		X: c.Center.X,
		Y: c.Center.Y + c.Height/4.0, // Center of mass for a solid cone
		Z: c.Center.Z,
	}
}

// IsVolumetric returns false for Cone3D.
func (c Cone3D) IsVolumetric() bool { return false }
