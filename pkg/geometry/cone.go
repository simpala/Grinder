package geometry

import (
	"grinder/pkg/math"
	"grinder/pkg/motion"
	"image/color"
	gomath "math"
)
// Cone3D represents a cone in 3D space.
type Cone3D struct {
	Center            math.Point3D // Center of the circular base
	Radius            float64
	Height            float64
	Color             color.RGBA
	Shininess         float64
	SpecularIntensity float64
	SpecularColor     color.RGBA
	Motion            []motion.Keyframe
}

func (c *Cone3D) Contains(p math.Point3D, t float64) bool {
	center := c.CenterAt(t)
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

func (c *Cone3D) Intersects(aabb math.AABB3D) bool {
	// 1. Check Y-range
	if aabb.Min.Y > c.Center.Y+c.Height || aabb.Max.Y < c.Center.Y {
		return false
	}

	// 2. Conservative Check: Use the cylinder AABB intersection logic
	// We treat it as a cylinder of its base radius for the broad phase.
	closestX := gomath.Max(aabb.Min.X, gomath.Min(c.Center.X, aabb.Max.X))
	closestZ := gomath.Max(aabb.Min.Z, gomath.Min(c.Center.Z, aabb.Max.Z))

	dx, dz := closestX-c.Center.X, closestZ-c.Center.Z
	// If it doesn't even hit the base cylinder, it doesn't hit the cone
	if (dx*dx + dz*dz) > c.Radius*c.Radius {
		return false
	}

	// 3. Fine-grained check: Check if the closest point on AABB is inside the radius at that Y
	// Use the Y-level of the AABB closest to the base
	targetY := gomath.Max(aabb.Min.Y, c.Center.Y)
	hRatio := (targetY - c.Center.Y) / c.Height
	currentRadius := c.Radius * (1.0 - hRatio)

	return (dx*dx + dz*dz) <= currentRadius*currentRadius
}

func (c *Cone3D) NormalAtPoint(p math.Point3D, t float64) math.Normal3D {
	center := c.CenterAt(t)
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

func (c *Cone3D) CenterAt(t float64) math.Point3D {
	if len(c.Motion) == 0 {
		return c.Center
	}
	_, _, center, _, _ := motion.Interpolate(c.Motion, t)
	return center
}

func (c *Cone3D) GetAABB() math.AABB3D {
	if len(c.Motion) == 0 {
		return math.AABB3D{
			Min: math.Point3D{X: c.Center.X - c.Radius, Y: c.Center.Y, Z: c.Center.Z - c.Radius},
			Max: math.Point3D{X: c.Center.X + c.Radius, Y: c.Center.Y + c.Height, Z: c.Center.Z + c.Radius},
		}
	}
	min := math.Point3D{X: c.Center.X - c.Radius, Y: c.Center.Y, Z: c.Center.Z - c.Radius}
	max := math.Point3D{X: c.Center.X + c.Radius, Y: c.Center.Y + c.Height, Z: c.Center.Z + c.Radius}
	for _, kf := range c.Motion {
		min.X = gomath.Min(min.X, kf.Center.X-c.Radius)
		min.Y = gomath.Min(min.Y, kf.Center.Y)
		min.Z = gomath.Min(min.Z, kf.Center.Z-c.Radius)
		max.X = gomath.Max(max.X, kf.Center.X+c.Radius)
		max.Y = gomath.Max(max.Y, kf.Center.Y+c.Height)
		max.Z = gomath.Max(max.Z, kf.Center.Z+c.Radius)
	}
	return math.AABB3D{Min: min, Max: max}
}

// GetColor returns the color of the cone.
func (s *Cone3D) GetColor() color.RGBA { return s.Color }

// GetShininess returns the shininess of the cone.
func (s *Cone3D) GetShininess() float64 { return s.Shininess }

// GetSpecularIntensity returns the specular intensity of the cone.
func (s *Cone3D) GetSpecularIntensity() float64 { return s.SpecularIntensity }

// GetSpecularColor returns the specular color of the cone.
func (s *Cone3D) GetSpecularColor() color.RGBA { return s.SpecularColor }

// GetCenter returns the geometric center of the cone.
func (c *Cone3D) GetCenter() math.Point3D {
	return math.Point3D{
		X: c.Center.X,
		Y: c.Center.Y + c.Height/4.0, // Center of mass for a solid cone
		Z: c.Center.Z,
	}
}

// AtTime returns a new cone interpolated to time t.
func (c *Cone3D) AtTime(t float64) Shape {
	if len(c.Motion) == 0 {
		return c
	}
	newCone := *c
	newCone.Center = c.CenterAt(t)
	return &newCone
}

// IsVolumetric returns false for Cone3D.
func (c *Cone3D) IsVolumetric() bool { return false }
