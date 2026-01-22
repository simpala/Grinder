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
	Radius    float64
}

// ShadedColor calculates the color of a point on a surface, including basic shadowing and specular highlights.
func ShadedColor(p math.Point3D, n math.Normal3D, eye math.Point3D, l Light, shape geometry.Shape, shapes []geometry.Shape, rng *math.XorShift32) color.RGBA {
	lightVec := l.Position.Sub(p)
	lightDir := lightVec.Normalize()
	base := shape.GetColor()

	shadowFactor := TraceShadow(p, n, l, shapes, rng)

	// Diffuse (Lambert) component
	dot := n.Dot(lightDir)
	diffuseFactor := gomath.Max(0.15, dot*shadowFactor) * l.Intensity


	// Specular (Phong) component
	var specularR, specularG, specularB float64
	if shadowFactor > 0.2 { // No specular highlights in shadow
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

func TraceShadow(p math.Point3D, n math.Normal3D, l Light, shapes []geometry.Shape, rng *math.XorShift32) float64 {
	const numSamples = 8
	shadowHits := 0.0

	for i := 0; i < numSamples; i++ {
		// Jitter the light position within its radius
		jitter := math.Point3D{
			X: (rng.Float64() - 0.5) * l.Radius,
			Y: (rng.Float64() - 0.5) * l.Radius,
			Z: (rng.Float64() - 0.5) * l.Radius,
		}
		targetPos := l.Position.Add(jitter)

		dir := targetPos.Sub(p).Normalize()
		distToLight := p.Distance(targetPos)

		// Bias to prevent self-intersection
		currentP := p.Add(n.ToVector().Mul(0.01))

		hit := false
		for t := 0.0; t < distToLight; t += 0.2 {
			sampleP := currentP.Add(dir.Mul(t))
			for _, s := range shapes {
				if _, ok := s.(geometry.Plane3D); ok {
					continue
				}
				if s.Contains(sampleP) {
					hit = true
					break
				}
			}
			if hit {
				break
			}
		}
		if hit {
			shadowHits += 1.0
		}
	}

	return gomath.Max(0.2, 1.0-(shadowHits/float64(numSamples)))
}
