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

// isOccluded checks for shadows by marching small AABBs towards the light source.
func isOccluded(p, lightPos math.Point3D, shapes []geometry.Shape, self geometry.Shape) bool {
	const stepSize = 0.1
	const boxSize = 0.05 // The size of the AABB at each step.

	vecToLight := lightPos.Sub(p)
	distToLight := vecToLight.Length()
	dirToLight := vecToLight.Normalize()

	// March from the point towards the light.
	for t := stepSize; t < distToLight; t += stepSize {
		samplePoint := p.Add(dirToLight.Mul(t))

		// Create a small AABB around the sample point.
		aabb := math.AABB3D{
			Min: samplePoint.Sub(math.Point3D{X: boxSize, Y: boxSize, Z: boxSize}),
			Max: samplePoint.Add(math.Point3D{X: boxSize, Y: boxSize, Z: boxSize}),
		}

		// Check for intersection with any shape in the scene.
		for _, shape := range shapes {
			if shape == self {
				continue
			}
			// Ignore planes to prevent self-shadowing on the floor.
			if _, ok := shape.(geometry.Plane3D); ok {
				continue
			}
			if shape.Intersects(aabb) {
				return true // The point is in shadow.
			}
		}
	}

	return false // The point is not in shadow.
}

// ShadedColor calculates the color of a point on a surface.
func ShadedColor(p math.Point3D, n math.Normal3D, eye math.Point3D, l Light, shape geometry.Shape, shapes []geometry.Shape) color.RGBA {
	lightVec := l.Position.Sub(p)
	lightDir := lightVec.Normalize()
	base := shape.GetColor()

	// Shadow Check
	shadowBias := 0.001
	checkP := math.Point3D{X: p.X + n.X*shadowBias, Y: p.Y + n.Y*shadowBias, Z: p.Z + n.Z*shadowBias}
	inShadow := isOccluded(checkP, l.Position, shapes, shape)

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
