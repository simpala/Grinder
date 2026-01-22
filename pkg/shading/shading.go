package shading

import (
	"grinder/pkg/geometry"
	"grinder/pkg/math"
	"image/color"
	gomath "math"
)

// Light represents a light source in the scene.
type Light struct {
	Position  math.Point3D
	Intensity float64
}

// ShadedColor calculates the color of a point on a surface, including basic shadowing and specular highlights.
func ShadedColor(p math.Point3D, n math.Normal3D, eye math.Point3D, l Light, shape geometry.Shape, shapes []geometry.Shape) color.RGBA {
	lightVec := l.Position.Sub(p)
	lightDir := lightVec.Normalize()
	base := shape.GetColor()

	// Shadow Check: Offset the starting point slightly to avoid self-intersection.
	shadowBias := 0.01
	checkP := math.Point3D{X: p.X + n.X*shadowBias, Y: p.Y + n.Y*shadowBias, Z: p.Z + n.Z*shadowBias}

	inShadow := false
	// Trace towards the light (cheap volumetric shadow check).
	for t := 0.1; t < 5.0; t += 0.2 {
		sampleP := math.Point3D{X: checkP.X + lightDir.X*t, Y: checkP.Y + lightDir.Y*t, Z: checkP.Z + lightDir.Z*t}
		for _, s := range shapes {
			if s == shape {
				continue
			}
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

	// Diffuse (Lambert) component
	dot := n.Dot(lightDir)
	diffuseFactor := gomath.Max(0.15, dot) * l.Intensity // Ambient term is 0.15

	if inShadow {
		diffuseFactor = 0.15 // Ambient only
	}

	// Specular (Phong) component
	var specularR, specularG, specularB float64
	if !inShadow { // No specular highlights in shadow
		viewDir := eye.Sub(p).Normalize()

		// R = 2 * (N . L) * N - L
		dotNL := n.Dot(lightDir)
		reflectDir := n.ToVector().Mul(2 * dotNL).Sub(lightDir)

		specularAngle := gomath.Max(0.0, viewDir.Dot(reflectDir))
		specularFactor := gomath.Pow(specularAngle, shape.GetShininess())
		specularIntensity := shape.GetSpecularIntensity()

		specularColor := shape.GetSpecularColor()
		specularR = float64(specularColor.R) * specularFactor * specularIntensity
		specularG = float64(specularColor.G) * specularFactor * specularIntensity
		specularB = float64(specularColor.B) * specularFactor * specularIntensity
	}

	// Combine components
	finalR := float64(base.R)*diffuseFactor + specularR
	finalG := float64(base.G)*diffuseFactor + specularG
	finalB := float64(base.B)*diffuseFactor + specularB

	return color.RGBA{
		R: uint8(gomath.Min(255, finalR)),
		G: uint8(gomath.Min(255, finalG)),
		B: uint8(gomath.Min(255, finalB)),
		A: 255,
	}
}
