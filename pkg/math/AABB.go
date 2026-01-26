package math

import "math"

type AABB3D struct {
	Min, Max Point3D
}

// Contains checks if a Point3D is inside the bounding box
func (a AABB3D) Contains(p Point3D) bool {
	return p.X >= a.Min.X && p.X <= a.Max.X &&
		p.Y >= a.Min.Y && p.Y <= a.Max.Y &&
		p.Z >= a.Min.Z && p.Z <= a.Max.Z
}

// Center returns the midpoint of the AABB
func (a AABB3D) Center() Point3D {
	return Point3D{
		X: (a.Min.X + a.Max.X) * 0.5,
		Y: (a.Min.Y + a.Max.Y) * 0.5,
		Z: (a.Min.Z + a.Max.Z) * 0.5,
	}
}

// Expand returns a new AABB that includes the given point.
func (a AABB3D) Expand(p Point3D) AABB3D {
	return AABB3D{
		Min: Point3D{
			X: math.Min(a.Min.X, p.X),
			Y: math.Min(a.Min.Y, p.Y),
			Z: math.Min(a.Min.Z, p.Z),
		},
		Max: Point3D{
			X: math.Max(a.Max.X, p.X),
			Y: math.Max(a.Max.Y, p.Y),
			Z: math.Max(a.Max.Z, p.Z),
		},
	}
}

// IntersectRay performs a ray-AABB intersection test using the slab method.
// It returns tmin, tmax, and a boolean indicating if the ray intersects the box.
func (a AABB3D) IntersectRay(r Ray) (float64, float64, bool) {
	tmin := -math.MaxFloat64
	tmax := math.MaxFloat64

	// Use a small epsilon for floating point comparisons to handle axis-parallel rays.
	const epsilon = 1e-6

	// X slab
	if math.Abs(r.Direction.X) < epsilon {
		if r.Origin.X < a.Min.X || r.Origin.X > a.Max.X {
			return 0, 0, false
		}
	} else {
		tx1 := (a.Min.X - r.Origin.X) / r.Direction.X
		tx2 := (a.Max.X - r.Origin.X) / r.Direction.X
		if tx1 > tx2 {
			tx1, tx2 = tx2, tx1
		}
		tmin = math.Max(tmin, tx1)
		tmax = math.Min(tmax, tx2)
	}

	// Y slab
	if math.Abs(r.Direction.Y) < epsilon {
		if r.Origin.Y < a.Min.Y || r.Origin.Y > a.Max.Y {
			return 0, 0, false
		}
	} else {
		ty1 := (a.Min.Y - r.Origin.Y) / r.Direction.Y
		ty2 := (a.Max.Y - r.Origin.Y) / r.Direction.Y
		if ty1 > ty2 {
			ty1, ty2 = ty2, ty1
		}
		tmin = math.Max(tmin, ty1)
		tmax = math.Min(tmax, ty2)
	}

	// Z slab
	if math.Abs(r.Direction.Z) < epsilon {
		if r.Origin.Z < a.Min.Z || r.Origin.Z > a.Max.Z {
			return 0, 0, false
		}
	} else {
		tz1 := (a.Min.Z - r.Origin.Z) / r.Direction.Z
		tz2 := (a.Max.Z - r.Origin.Z) / r.Direction.Z
		if tz1 > tz2 {
			tz1, tz2 = tz2, tz1
		}
		tmin = math.Max(tmin, tz1)
		tmax = math.Min(tmax, tz2)
	}

	// Finally, check if the intersection is valid and not entirely behind the ray origin.
	return tmin, tmax, tmax >= tmin && tmax > 0
}
