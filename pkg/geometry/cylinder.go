package geometry

import (
	"grinder/pkg/math"
	"image/color"
	gomath "math"
)

// Cylinder3D represents a cylinder in 3D space.
type Cylinder3D struct {
	Center            math.Point3D // Center of the base
	Height, Radius    float64
	Color             color.RGBA
	Shininess         float64
	SpecularIntensity float64
	SpecularColor     color.RGBA
}

func (c Cylinder3D) Contains(p math.Point3D) bool {
	if p.Y < c.Center.Y || p.Y > c.Center.Y+c.Height {
		return false
	}
	dx, dz := p.X-c.Center.X, p.Z-c.Center.Z
	return (dx*dx + dz*dz) <= c.Radius*c.Radius
}

func (c Cylinder3D) Intersects(aabb math.AABB3D) bool {
	// Check Y-range first
	if aabb.Min.Y > c.Center.Y+c.Height || aabb.Max.Y < c.Center.Y {
		return false
	}

	// Check circle-AABB intersection in XZ plane
	closestX := gomath.Max(aabb.Min.X, gomath.Min(c.Center.X, aabb.Max.X))
	closestZ := gomath.Max(aabb.Min.Z, gomath.Min(c.Center.Z, aabb.Max.Z))

	dx, dz := closestX-c.Center.X, closestZ-c.Center.Z
	return (dx*dx + dz*dz) <= c.Radius*c.Radius
}

func (c Cylinder3D) NormalAtPoint(p math.Point3D) math.Normal3D {
	eps := 0.0001
	if p.Y >= c.Center.Y+c.Height-eps {
		return math.Normal3D{X: 0, Y: 1, Z: 0}
	}
	if p.Y <= c.Center.Y+eps {
		return math.Normal3D{X: 0, Y: -1, Z: 0}
	}
	n := math.Point3D{X: p.X - c.Center.X, Y: 0, Z: p.Z - c.Center.Z}.Normalize()
	return math.Normal3D{X: n.X, Y: 0, Z: n.Z}
}

// GetColor returns the color of the cylinder.
func (s Cylinder3D) GetColor() color.RGBA { return s.Color }

// GetShininess returns the shininess of the cylinder.
func (s Cylinder3D) GetShininess() float64 { return s.Shininess }

// GetSpecularIntensity returns the specular intensity of the cylinder.
func (s Cylinder3D) GetSpecularIntensity() float64 { return s.SpecularIntensity }

// GetSpecularColor returns the specular color of the cylinder.
func (s Cylinder3D) GetSpecularColor() color.RGBA { return s.SpecularColor }

// GetAABB returns the bounding box of the cylinder.
func (c Cylinder3D) GetAABB() math.AABB3D {
	return math.AABB3D{
		Min: math.Point3D{
			X: c.Center.X - c.Radius,
			Y: c.Center.Y,
			Z: c.Center.Z - c.Radius,
		},
		Max: math.Point3D{
			X: c.Center.X + c.Radius,
			Y: c.Center.Y + c.Height,
			Z: c.Center.Z + c.Radius,
		},
	}
}

// GetCenter returns the geometric center of the cylinder.
func (c Cylinder3D) GetCenter() math.Point3D {
	return math.Point3D{
		X: c.Center.X,
		Y: c.Center.Y + c.Height/2.0,
		Z: c.Center.Z,
	}
}
