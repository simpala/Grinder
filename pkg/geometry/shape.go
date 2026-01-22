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
	GetAABB() math.AABB3D
	GetCenter() math.Point3D
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

// Box3D represents a solid box (AABB) in 3D space.
type Box3D struct {
	Min, Max          math.Point3D
	Color             color.RGBA
	Shininess         float64
	SpecularIntensity float64
	SpecularColor     color.RGBA
}

func (b Box3D) Contains(p math.Point3D) bool {
	return p.X >= b.Min.X && p.X <= b.Max.X &&
		p.Y >= b.Min.Y && p.Y <= b.Max.Y &&
		p.Z >= b.Min.Z && p.Z <= b.Max.Z
}

func (b Box3D) Intersects(aabb math.AABB3D) bool {
	// Standard AABB-AABB intersection
	return (b.Min.X <= aabb.Max.X && b.Max.X >= aabb.Min.X) &&
		(b.Min.Y <= aabb.Max.Y && b.Max.Y >= aabb.Min.Y) &&
		(b.Min.Z <= aabb.Max.Z && b.Max.Z >= aabb.Min.Z)
}

func (b Box3D) NormalAtPoint(p math.Point3D) math.Normal3D {
	// Find which face the point is closest to
	eps := 0.0001
	if gomath.Abs(p.X-b.Min.X) < eps {
		return math.Normal3D{X: -1, Y: 0, Z: 0}
	}
	if gomath.Abs(p.X-b.Max.X) < eps {
		return math.Normal3D{X: 1, Y: 0, Z: 0}
	}
	if gomath.Abs(p.Y-b.Min.Y) < eps {
		return math.Normal3D{X: 0, Y: -1, Z: 0}
	}
	if gomath.Abs(p.Y-b.Max.Y) < eps {
		return math.Normal3D{X: 0, Y: 1, Z: 0}
	}
	if gomath.Abs(p.Z-b.Min.Z) < eps {
		return math.Normal3D{X: 0, Y: 0, Z: -1}
	}
	return math.Normal3D{X: 0, Y: 0, Z: 1}
}

// GetColor returns the color of the sphere.
func (s Box3D) GetColor() color.RGBA { return s.Color }

// GetShininess returns the shininess of the sphere.
func (s Box3D) GetShininess() float64 { return s.Shininess }

// GetSpecularIntensity returns the specular intensity of the sphere.
func (s Box3D) GetSpecularIntensity() float64 { return s.SpecularIntensity }

// GetSpecularColor returns the specular color of the sphere.
func (s Box3D) GetSpecularColor() color.RGBA { return s.SpecularColor }

func (b Box3D) GetAABB() math.AABB3D { return math.AABB3D{Min: b.Min, Max: b.Max} }

// GetCenter returns the center of the box.
func (b Box3D) GetCenter() math.Point3D {
	return b.Min.Add(b.Max).Mul(0.5)
}

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

// GetColor returns the color of the sphere.
func (s Cylinder3D) GetColor() color.RGBA { return s.Color }

// GetShininess returns the shininess of the sphere.
func (s Cylinder3D) GetShininess() float64 { return s.Shininess }

// GetSpecularIntensity returns the specular intensity of the sphere.
func (s Cylinder3D) GetSpecularIntensity() float64 { return s.SpecularIntensity }

// GetSpecularColor returns the specular color of the sphere.
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

type Cone3D struct {
	Center            math.Point3D // Center of the circular base
	Radius            float64
	Height            float64
	Color             color.RGBA
	Shininess         float64
	SpecularIntensity float64
	SpecularColor     color.RGBA
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

func (c Cone3D) GetAABB() math.AABB3D {
	return math.AABB3D{
		Min: math.Point3D{X: c.Center.X - c.Radius, Y: c.Center.Y, Z: c.Center.Z - c.Radius},
		Max: math.Point3D{X: c.Center.X + c.Radius, Y: c.Center.Y + c.Height, Z: c.Center.Z + c.Radius},
	}
}

// ... (Implement other getters)
// GetColor returns the color of the sphere.
func (s Cone3D) GetColor() color.RGBA { return s.Color }

// GetShininess returns the shininess of the sphere.
func (s Cone3D) GetShininess() float64 { return s.Shininess }

// GetSpecularIntensity returns the specular intensity of the sphere.
func (s Cone3D) GetSpecularIntensity() float64 { return s.SpecularIntensity }

// GetSpecularColor returns the specular color of the sphere.
func (s Cone3D) GetSpecularColor() color.RGBA { return s.SpecularColor }

// GetCenter returns the geometric center of the cone.
func (c Cone3D) GetCenter() math.Point3D {
	return math.Point3D{
		X: c.Center.X,
		Y: c.Center.Y + c.Height/4.0, // Center of mass for a solid cone
		Z: c.Center.Z,
	}
}
