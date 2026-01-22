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

// isOccluded checks for shadows by marching small AABBs towards the light source.
func isOccluded(p, lightPos math.Point3D, shapes []geometry.Shape, lightRadius float64) bool {
	const stepSize = 0.2
	var boxSize float64

	// Use a small, precise probe for hard shadows and a larger one for soft shadows.
	if lightRadius == 0 {
		boxSize = 0.05
	} else {
		boxSize = 0.21
	}

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
	shadowBias := 1e-4 // Small epsilon to prevent self-shadowing ("shadow acne")
	checkP := math.Point3D{X: p.X + n.X * shadowBias, Y: p.Y + n.Y * shadowBias, Z: p.Z + n.Z * shadowBias}

	// Shadow Culling: Create a bounding box from the surface point to the light
	cullAABB := math.AABB3D{
		Min: math.Point3D{
			X: gomath.Min(checkP.X, l.Position.X-l.Radius),
			Y: gomath.Min(checkP.Y, l.Position.Y-l.Radius),
			Z: gomath.Min(checkP.Z, l.Position.Z-l.Radius),
		},
		Max: math.Point3D{
			X: gomath.Max(checkP.X, l.Position.X+l.Radius),
			Y: gomath.Max(checkP.Y, l.Position.Y+l.Radius),
			Z: gomath.Max(checkP.Z, l.Position.Z+l.Radius),
		},
	}

	// Filter shapes to only those that could possibly cast a shadow.
	occluders := make([]geometry.Shape, 0)
	for _, s := range shapes {
		if s == shape {
			continue // Don't check self
		}
		if s.GetAABB().Intersects(cullAABB) {
			occluders = append(occluders, s)
		}
	}

	inShadow := isOccluded(checkP, l.Position, occluders, l.Radius)

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
