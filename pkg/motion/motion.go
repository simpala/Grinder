package motion

import "grinder/pkg/math"

// Keyframe defines a point in time for an object's animation.
type Keyframe struct {
	Time   float64
	Eye    math.Point3D
	Target math.Point3D
	Center math.Point3D
	Min    math.Point3D
	Max    math.Point3D
}

func Interpolate(keyframes []Keyframe, t float64) (math.Point3D, math.Point3D, math.Point3D, math.Point3D, math.Point3D) {
	if len(keyframes) == 0 {
		return math.Point3D{}, math.Point3D{}, math.Point3D{}, math.Point3D{}, math.Point3D{}
	}
	var kf1, kf2 Keyframe
	for i := range keyframes {
		if keyframes[i].Time >= t {
			kf2 = keyframes[i]
			if i > 0 {
				kf1 = keyframes[i-1]
			} else {
				kf1 = kf2
			}
			break
		}
		kf1 = keyframes[i]
		kf2 = kf1
	}
	if kf1.Time == kf2.Time {
		return kf1.Eye, kf1.Target, kf1.Center, kf1.Min, kf1.Max
	}
	alpha := (t - kf1.Time) / (kf2.Time - kf1.Time)
	eye := kf1.Eye.Lerp(kf2.Eye, alpha)
	target := kf1.Target.Lerp(kf2.Target, alpha)
	center := kf1.Center.Lerp(kf2.Center, alpha)
	min := kf1.Min.Lerp(kf2.Min, alpha)
	max := kf1.Max.Lerp(kf2.Max, alpha)
	return eye, target, center, min, max
}
