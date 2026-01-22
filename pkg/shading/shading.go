package shading

import (
	"grinder/pkg/math"
	"grinder/pkg/geometry"
	"image/color"
	gomath "math"
)

// Light represents a light source in the scene.
type Light struct {
	Position  math.Point3D
	Intensity float64
}

// ShadedColor calculates the color of a point on a surface, including basic shadowing.
func ShadedColor(p math.Point3D, n math.Normal3D, l Light, base color.RGBA, shapes []geometry.Shape) color.RGBA {
	lightVec := l.Position.Sub(p)
	lightDir := lightVec.Normalize()

	// Shadow Check: Offset the starting point slightly to avoid self-intersection.
	shadowBias := 0.01
	checkP := math.Point3D{X: p.X + n.X*shadowBias, Y: p.Y + n.Y*shadowBias, Z: p.Z + n.Z*shadowBias}

	inShadow := false
	// Trace towards the light (cheap volumetric shadow check).
	for t := 0.1; t < 5.0; t += 0.2 {
		sampleP := math.Point3D{X: checkP.X + lightDir.X*t, Y: checkP.Y + lightDir.Y*t, Z: checkP.Z + lightDir.Z*t}
		for _, s := range shapes {
			// Ignore planes for this specific simple shadow check to prevent infinite floor shadows.
			if _, ok := s.(geometry.Plane3D); ok {
				continue
			}
			if s.Contains(sampleP) {
				inShadow = true
				break
			}
		}
		if inShadow {
			break
		}
	}

	dot := n.X*lightDir.X + n.Y*lightDir.Y + n.Z*lightDir.Z
	factor := gomath.Max(0.15, dot) * l.Intensity

	if inShadow {
		factor = 0.15 // Ambient only
	}

	return color.RGBA{
		R: uint8(gomath.Min(255, float64(base.R)*factor)),
		G: uint8(gomath.Min(255, float64(base.G)*factor)),
		B: uint8(gomath.Min(255, float64(base.B)*factor)),
		A: 255,
	}
}
