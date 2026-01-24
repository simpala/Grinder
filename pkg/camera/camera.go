package camera

import (
	"grinder/pkg/math"
	"grinder/pkg/motion"
	gomath "math"
)

// Camera defines the interface for a camera system.
type Camera interface {
	Project(sx, sy, z float64) math.Point3D
	GetEye() math.Point3D
	GetShutter() float64
	AtTime(t float64) Camera
}

// PerspectiveCamera represents a camera with perspective projection.
type PerspectiveCamera struct {
	Position, Forward, Right, Up math.Point3D
	FovScale, Aspect             float64
	Shutter                      float64
	Motion                       []motion.Keyframe
	target, up                   math.Point3D
	fov                          float64
}

// NewLookAtCamera creates a new camera that looks at a target from a given position.
func NewLookAtCamera(pos, target, up math.Point3D, fov, aspect, shutter float64, motion []motion.Keyframe) *PerspectiveCamera {
	f := target.Sub(pos).Normalize()
	r := f.Cross(up).Normalize()
	u := r.Cross(f)
	return &PerspectiveCamera{
		Position: pos, Forward: f, Right: r, Up: u,
		FovScale: gomath.Tan(fov * 0.5 * gomath.Pi / 180.0),
		Aspect:   aspect,
		Shutter:  shutter,
		Motion:   motion,
		target:   target,
		up:       up,
		fov:      fov,
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

// GetEye returns the position of the camera.
func (c *PerspectiveCamera) GetEye() math.Point3D {
	return c.Position
}

// GetShutter returns the shutter time of the camera.
func (c *PerspectiveCamera) GetShutter() float64 {
	return c.Shutter
}

// AtTime returns a new camera with its position and target interpolated.
func (c *PerspectiveCamera) AtTime(t float64) Camera {
	if len(c.Motion) == 0 {
		return c
	}
	eye, target, _, _, _ := motion.Interpolate(c.Motion, t)
	return NewLookAtCamera(eye, target, c.up, c.fov, c.Aspect, c.Shutter, c.Motion)
}
