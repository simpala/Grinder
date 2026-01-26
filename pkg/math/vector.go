package math

import "math"

// Point3D represents a point in 3D space.
type Point3D struct{ X, Y, Z float64 }

// Intersects checks if two AABBs overlap.
func (a AABB3D) Intersects(b AABB3D) bool {
	return (a.Min.X <= b.Max.X && a.Max.X >= b.Min.X) &&
		(a.Min.Y <= b.Max.Y && a.Max.Y >= b.Min.Y) &&
		(a.Min.Z <= b.Max.Z && a.Max.Z >= b.Min.Z)
}
func (aabb AABB3D) GetCorners() [8]Point3D {
	return [8]Point3D{
		{aabb.Min.X, aabb.Min.Y, aabb.Min.Z}, {aabb.Max.X, aabb.Min.Y, aabb.Min.Z},
		{aabb.Min.X, aabb.Max.Y, aabb.Min.Z}, {aabb.Max.X, aabb.Max.Y, aabb.Min.Z},
		{aabb.Min.X, aabb.Min.Y, aabb.Max.Z}, {aabb.Max.X, aabb.Min.Y, aabb.Max.Z},
		{aabb.Min.X, aabb.Max.Y, aabb.Max.Z}, {aabb.Max.X, aabb.Max.Y, aabb.Max.Z},
	}
}

// Normal3D represents a normal vector in 3D space.
type Normal3D struct{ X, Y, Z float64 }

// Normalize returns a unit normal.
func (n Normal3D) Normalize() Normal3D {
	d := math.Sqrt(n.X*n.X + n.Y*n.Y + n.Z*n.Z)
	if d == 0 {
		return n
	}
	return Normal3D{n.X / d, n.Y / d, n.Z / d}
}

// Mul scalar multiplication for normals (useful for blending/interpolation).
func (n Normal3D) Mul(s float64) Normal3D {
	return Normal3D{n.X * s, n.Y * s, n.Z * s}
}

// Add for normals.
func (n Normal3D) Add(other Normal3D) Normal3D {
	return Normal3D{n.X + other.X, n.Y + other.Y, n.Z + other.Z}
}

// Normalize returns a unit vector in the same direction as the input vector.
func (p Point3D) Normalize() Point3D {
	d := math.Sqrt(p.X*p.X + p.Y*p.Y + p.Z*p.Z)
	if d == 0 {
		return p
	}
	return Point3D{p.X / d, p.Y / d, p.Z / d}
}

func (p Point3D) LengthSquared() float64 {
	return p.X*p.X + p.Y*p.Y + p.Z*p.Z
}

// Cross returns the cross product of two vectors.
func (a Point3D) Cross(b Point3D) Point3D {
	return Point3D{a.Y*b.Z - a.Z*b.Y, a.Z*b.X - a.X*b.Z, a.X*b.Y - a.Y*b.X}
}

// Sub returns the vector difference between two points.
func (a Point3D) Sub(b Point3D) Point3D {
	return Point3D{a.X - b.X, a.Y - b.Y, a.Z - b.Z}
}

// Dot returns the dot product of two vectors.
func (a Point3D) Dot(b Point3D) float64 {
	return a.X*b.X + a.Y*b.Y + a.Z*b.Z
}

// DotNormal returns the dot product of a Point3D and a Normal3D.
func (a Point3D) DotNormal(b Normal3D) float64 {
	return a.X*b.X + a.Y*b.Y + a.Z*b.Z
}

// Dot returns the dot product of two normal vectors.
func (a Normal3D) Dot(b Point3D) float64 {
	return a.X*b.X + a.Y*b.Y + a.Z*b.Z
}

// ToVector converts a Normal3D to a Point3D.
func (n Normal3D) ToVector() Point3D {
	return Point3D{n.X, n.Y, n.Z}
}

// Mul returns the scalar product of a vector.
func (a Point3D) Mul(s float64) Point3D {
	return Point3D{a.X * s, a.Y * s, a.Z * s}
}

// Add returns the vector sum of two points.
func (a Point3D) Add(b Point3D) Point3D {
	return Point3D{a.X + b.X, a.Y + b.Y, a.Z + b.Z}
}

// Length returns the magnitude of the vector.
func (a Point3D) Length() float64 {
	return math.Sqrt(a.X*a.X + a.Y*a.Y + a.Z*a.Z)
}

func (p Point3D) DistanceToPlane(planePoint Point3D, planeNormal Normal3D) float64 {
	return math.Abs(p.Sub(planePoint).DotNormal(planeNormal))
}

// // IntersectRay performs a ray-AABB intersection test using the slab method.
// // It returns tmin, tmax, and a boolean indicating if the ray intersects the box.
// func (a AABB3D) IntersectRay(r Ray) (float64, float64, bool) {
// 	tmin := -math.MaxFloat64
// 	tmax := math.MaxFloat64

// 	// Use a small epsilon for floating point comparisons to handle axis-parallel rays.
// 	const epsilon = 1e-6

// 	// X slab
// 	if math.Abs(r.Direction.X) < epsilon {
// 		if r.Origin.X < a.Min.X || r.Origin.X > a.Max.X {
// 			return 0, 0, false
// 		}
// 	} else {
// 		tx1 := (a.Min.X - r.Origin.X) / r.Direction.X
// 		tx2 := (a.Max.X - r.Origin.X) / r.Direction.X
// 		if tx1 > tx2 {
// 			tx1, tx2 = tx2, tx1
// 		}
// 		tmin = math.Max(tmin, tx1)
// 		tmax = math.Min(tmax, tx2)
// 	}

// 	// Y slab
// 	if math.Abs(r.Direction.Y) < epsilon {
// 		if r.Origin.Y < a.Min.Y || r.Origin.Y > a.Max.Y {
// 			return 0, 0, false
// 		}
// 	} else {
// 		ty1 := (a.Min.Y - r.Origin.Y) / r.Direction.Y
// 		ty2 := (a.Max.Y - r.Origin.Y) / r.Direction.Y
// 		if ty1 > ty2 {
// 			ty1, ty2 = ty2, ty1
// 		}
// 		tmin = math.Max(tmin, ty1)
// 		tmax = math.Min(tmax, ty2)
// 	}

// 	// Z slab
// 	if math.Abs(r.Direction.Z) < epsilon {
// 		if r.Origin.Z < a.Min.Z || r.Origin.Z > a.Max.Z {
// 			return 0, 0, false
// 		}
// 	} else {
// 		tz1 := (a.Min.Z - r.Origin.Z) / r.Direction.Z
// 		tz2 := (a.Max.Z - r.Origin.Z) / r.Direction.Z
// 		if tz1 > tz2 {
// 			tz1, tz2 = tz2, tz1
// 		}
// 		tmin = math.Max(tmin, tz1)
// 		tmax = math.Min(tmax, tz2)
// 	}

// 	// Finally, check if the intersection is valid and not entirely behind the ray origin.
// 	return tmin, tmax, tmax >= tmin && tmax > 0
// }
