package geometry

import (
	"grinder/pkg/math"
	"image/color"
	gomath "math"
)

// Plane3D represents an infinite plane in 3D space.
type Plane3D struct {
	Point             math.Point3D
	Normal            math.Normal3D
	Color             color.RGBA
	Shininess         float64
	SpecularIntensity float64
	SpecularColor     color.RGBA
}

// Contains checks if a point is "under" a plane, ignoring time for static planes.
func (pl Plane3D) Contains(p math.Point3D, t float64) bool {
	v := p.Sub(pl.Point)
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
func (pl Plane3D) NormalAtPoint(p math.Point3D, t float64) math.Normal3D { return pl.Normal }

// GetColor returns the color of the plane.
func (pl Plane3D) GetColor() color.RGBA { return pl.Color }

// GetShininess returns the shininess of the plane.
func (pl Plane3D) GetShininess() float64 { return pl.Shininess }

// GetSpecularIntensity returns the specular intensity of the plane.
func (pl Plane3D) GetSpecularIntensity() float64 { return pl.SpecularIntensity }

// GetSpecularColor returns the specular color of the plane.
func (pl Plane3D) GetSpecularColor() color.RGBA { return pl.SpecularColor }

// GetAABB for a plane is infinite, so we return a huge box.
func (pl Plane3D) GetAABB() math.AABB3D {
	inf := gomath.Inf(1)
	return math.AABB3D{
		Min: math.Point3D{X: gomath.Inf(-1), Y: gomath.Inf(-1), Z: gomath.Inf(-1)},
		Max: math.Point3D{X: inf, Y: inf, Z: inf},
	}
}

// GetCenter returns the plane's reference point.
func (pl Plane3D) GetCenter() math.Point3D {
	return pl.Point
}

// IsVolumetric returns false for Plane3D.
func (pl Plane3D) IsVolumetric() bool { return false }
