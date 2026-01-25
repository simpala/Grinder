package motion

import "grinder/pkg/math"

// Keyframe defines a position and time for a shape.
type Keyframe struct {
	Time     float64
	Position math.Point3D
}

// Interpolate returns the interpolated position of a shape at a given time.
func Interpolate(t float64, keyframes []Keyframe) math.Point3D {
	if len(keyframes) == 0 {
		return math.Point3D{}
	}
	if len(keyframes) == 1 || t <= keyframes[0].Time {
		return keyframes[0].Position
	}
	if t >= keyframes[len(keyframes)-1].Time {
		return keyframes[len(keyframes)-1].Position
	}

	for i := 0; i < len(keyframes)-1; i++ {
		if t >= keyframes[i].Time && t <= keyframes[i+1].Time {
			alpha := (t - keyframes[i].Time) / (keyframes[i+1].Time - keyframes[i].Time)
			return keyframes[i].Position.Lerp(keyframes[i+1].Position, alpha)
		}
	}
	return keyframes[len(keyframes)-1].Position
}
