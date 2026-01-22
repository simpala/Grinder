package math

import "math"

// Point3D represents a point in 3D space.
type Point3D struct{ X, Y, Z float64 }

// AABB3D represents an axis-aligned bounding box in 3D space.
type AABB3D struct{ Min, Max Point3D }

// Normal3D represents a normal vector in 3D space.
type Normal3D struct{ X, Y, Z float64 }

// Normalize returns a unit vector in the same direction as the input vector.
func (p Point3D) Normalize() Point3D {
	d := math.Sqrt(p.X*p.X + p.Y*p.Y + p.Z*p.Z)
	if d == 0 {
		return p
	}
	return Point3D{p.X / d, p.Y / d, p.Z / d}
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

// Distance returns the Euclidean distance between two points.
func (a Point3D) Distance(b Point3D) float64 {
	return math.Sqrt(math.Pow(a.X-b.X, 2) + math.Pow(a.Y-b.Y, 2) + math.Pow(a.Z-b.Z, 2))
}
