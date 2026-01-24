package geometry

import (
	"grinder/pkg/math"
	"image/color"
	gomath "math"
)

import "grinder/pkg/motion"

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

func (c Cone3D) Contains(p math.Point3D) bool {
	// Check height bounds
	if p.Y < c.Center.Y || p.Y > c.Center.Y+c.Height {
		return false
	}

	// Calculate relative height (0 at base, 1 at tip)
	hRatio := (p.Y - c.Center.Y) / c.Height
	// Radius at this specific height
	currentRadius := c.Radius * (1.0 - hRatio)

	dx, dz := p.X-c.Center.X, p.Z-c.Center.Z
	return (dx*dx + dz*dz) <= currentRadius*currentRadius
}

func (c Cone3D) Intersects(aabb math.AABB3D) bool {
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

func (c Cone3D) NormalAtPoint(p math.Point3D) math.Normal3D {
	eps := 0.0001
	// Bottom cap
	if p.Y <= c.Center.Y+eps {
		return math.Normal3D{X: 0, Y: -1, Z: 0}
	}

	// Side normal: Slanted outward and slightly upward
	dx, dz := p.X-c.Center.X, p.Z-c.Center.Z
	horizontalDist := gomath.Sqrt(dx*dx + dz*dz)

	// The slope of the cone side
	slope := c.Radius / c.Height
	n := math.Point3D{X: dx / horizontalDist, Y: slope, Z: dz / horizontalDist}.Normalize()
	return math.Normal3D{X: n.X, Y: n.Y, Z: n.Z}
}

func (c *Cone3D) GetAABB() math.AABB3D {
	baseAABB := math.AABB3D{
		Min: math.Point3D{X: c.Center.X - c.Radius, Y: c.Center.Y, Z: c.Center.Z - c.Radius},
		Max: math.Point3D{X: c.Center.X + c.Radius, Y: c.Center.Y + c.Height, Z: c.Center.Z + c.Radius},
	}
	if len(c.Motion) == 0 {
		return baseAABB
	}
	for _, kf := range c.Motion {
		kfMin := math.Point3D{
			X: kf.Position.X - c.Radius,
			Y: kf.Position.Y,
			Z: kf.Position.Z - c.Radius,
		}
		kfMax := math.Point3D{
			X: kf.Position.X + c.Radius,
			Y: kf.Position.Y + c.Height,
			Z: kf.Position.Z + c.Radius,
		}
		baseAABB.Min.X = gomath.Min(baseAABB.Min.X, kfMin.X)
		baseAABB.Min.Y = gomath.Min(baseAABB.Min.Y, kfMin.Y)
		baseAABB.Min.Z = gomath.Min(baseAABB.Min.Z, kfMin.Z)
		baseAABB.Max.X = gomath.Max(baseAABB.Max.X, kfMax.X)
		baseAABB.Max.Y = gomath.Max(baseAABB.Max.Y, kfMax.Y)
		baseAABB.Max.Z = gomath.Max(baseAABB.Max.Z, kfMax.Z)
	}
	return baseAABB
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

func (c *Cone3D) AtTime(t float64) Shape {
	if len(c.Motion) == 0 {
		return c
	}
	newCone := *c
	newCone.Center = motion.Interpolate(c.Motion, t)
	return &newCone
}
