package geometry

import (
	"grinder/pkg/math"
	"image/color"
	gomath "math"
)

// Shape defines the interface for all geometric objects in the scene.
type Shape interface {
	Contains(p math.Point3D) bool
	Intersects(aabb math.AABB3D) bool
	NormalAtPoint(p math.Point3D) math.Normal3D
	GetColor() color.RGBA
	GetShininess() float64
	GetSpecularIntensity() float64
	GetSpecularColor() color.RGBA
}

// Plane3D represents an infinite plane in 3D space.
type Plane3D struct {
	Point             math.Point3D
	Normal            math.Normal3D
	Color             color.RGBA
	Shininess         float64
	SpecularIntensity float64
	SpecularColor     color.RGBA
}

// Contains checks if a point is "under" the plane (in the direction opposite the normal).
func (pl Plane3D) Contains(p math.Point3D) bool {
	v := p.Sub(pl.Point)
	// Add a tiny epsilon (0.0001) to reduce sampling noise at the surface
	return v.DotNormal(pl.Normal) <= 0.0001
}

// Intersects checks if the plane intersects with an AABB.
func (pl Plane3D) Intersects(aabb math.AABB3D) bool {
	// Check if any of the 8 corners are on opposite sides of the plane.
	points := [8]math.Point3D{
		{aabb.Min.X, aabb.Min.Y, aabb.Min.Z}, {aabb.Max.X, aabb.Min.Y, aabb.Min.Z},
		{aabb.Min.X, aabb.Max.Y, aabb.Min.Z}, {aabb.Max.X, aabb.Max.Y, aabb.Min.Z},
		{aabb.Min.X, aabb.Min.Y, aabb.Max.Z}, {aabb.Max.X, aabb.Min.Y, aabb.Max.Z},
		{aabb.Min.X, aabb.Max.Y, aabb.Max.Z}, {aabb.Max.X, aabb.Max.Y, aabb.Max.Z},
	}

	hasIn, hasOut := false, false
	for _, p := range points {
		v := p.Sub(pl.Point)
		dot := v.DotNormal(pl.Normal)
		if dot <= 0 {
			hasIn = true
		} else {
			hasOut = true
		}
		if hasIn && hasOut {
			return true // Intersects if points are on both sides
		}
	}
	// AABB is fully on one side. It "intersects" if it's on the "in" side.
	return hasIn
}

// NormalAtPoint returns the normal of the plane, which is constant.
func (pl Plane3D) NormalAtPoint(p math.Point3D) math.Normal3D { return pl.Normal }

// GetColor returns the color of the plane.
func (pl Plane3D) GetColor() color.RGBA { return pl.Color }

// GetShininess returns the shininess of the plane.
func (pl Plane3D) GetShininess() float64 { return pl.Shininess }

// GetSpecularIntensity returns the specular intensity of the plane.
func (pl Plane3D) GetSpecularIntensity() float64 { return pl.SpecularIntensity }

// GetSpecularColor returns the specular color of the plane.
func (pl Plane3D) GetSpecularColor() color.RGBA { return pl.SpecularColor }

// Sphere3D represents a sphere in 3D space.
type Sphere3D struct {
	Center            math.Point3D
	Radius            float64
	Color             color.RGBA
	Shininess         float64
	SpecularIntensity float64
	SpecularColor     color.RGBA
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
