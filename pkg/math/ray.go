package math

// Ray represents a ray with an origin and a direction.
type Ray struct {
	Origin    Point3D
	Direction Point3D
}

func (r Ray) At(t float64) Point3D {
	return r.Origin.Add(r.Direction.Multiply(t))
}
