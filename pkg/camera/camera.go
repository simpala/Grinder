package camera

import (
	"grinder/pkg/math"
	gomath "math"
)

// Camera defines the interface for a camera system.
type Camera interface {
	Project(sx, sy, z float64) math.Point3D
}

// PerspectiveCamera represents a camera with perspective projection.
type PerspectiveCamera struct {
	Position, Forward, Right, Up math.Point3D
	FovScale, Aspect             float64
}

// NewLookAtCamera creates a new camera that looks at a target from a given position.
func NewLookAtCamera(pos, target, up math.Point3D, fov, aspect float64) *PerspectiveCamera {
	f := target.Sub(pos).Normalize()
	r := f.Cross(up).Normalize()
	u := r.Cross(f)
	return &PerspectiveCamera{
		Position: pos, Forward: f, Right: r, Up: u,
		FovScale: gomath.Tan(fov * 0.5 * gomath.Pi / 180.0),
		Aspect:   aspect,
	}
}

// Project transforms a screen-space coordinate (sx, sy) and a depth (z) to a 3D world point.
func (c *PerspectiveCamera) Project(sx, sy, z float64) math.Point3D {
	nx := (2.0*sx - 1.0) * c.Aspect * c.FovScale * z
	ny := (1.0 - 2.0*sy) * c.FovScale * z
	return math.Point3D{
		X: c.Position.X + c.Forward.X*z + c.Right.X*nx + c.Up.X*ny,
		Y: c.Position.Y + c.Forward.Y*z + c.Right.Y*nx + c.Up.Y*ny,
		Z: c.Position.Z + c.Forward.Z*z + c.Right.Z*nx + c.Up.Z*ny,
	}
}
