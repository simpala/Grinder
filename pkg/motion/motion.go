package motion

import "grinder/pkg/math"

// Keyframe defines the state of an object's properties at a specific time.
type Keyframe struct {
	Time     float64      `json:"time"`
	Position math.Point3D `json:"position"`
	// Rotation math.Quaternion `json:"rotation"` // Future expansion
	// Scale    math.Point3D   `json:"scale"`    // Future expansion
}

// Interpolate finds the position of an object at a given time `t` by
// linearly interpolating between the two closest keyframes.
func Interpolate(keyframes []Keyframe, t float64) math.Point3D {
	if len(keyframes) == 0 {
		return math.Point3D{} // Return zero vector if no keyframes
	}
	if len(keyframes) == 1 || t <= keyframes[0].Time {
		return keyframes[0].Position
	}
	if t >= keyframes[len(keyframes)-1].Time {
		return keyframes[len(keyframes)-1].Position
	}

	// Find the two keyframes that bracket the time `t`
	var kf1, kf2 Keyframe
	for i := 0; i < len(keyframes)-1; i++ {
		if t >= keyframes[i].Time && t <= keyframes[i+1].Time {
			kf1 = keyframes[i]
			kf2 = keyframes[i+1]
			break
		}
	}

	// Calculate the interpolation factor (alpha)
	alpha := (t - kf1.Time) / (kf2.Time - kf1.Time)

	// Linearly interpolate the position
	return kf1.Position.Lerp(kf2.Position, alpha)
}
