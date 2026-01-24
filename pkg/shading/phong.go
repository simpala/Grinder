package shading

import (
	"grinder/pkg/geometry"
	"grinder/pkg/math"
	"image/color"
	gomath "math"
)

// ShadedColor calculates the color of a point on a surface using the Phong reflection model.
func ShadedColor(p math.Point3D, n math.Normal3D, eye math.Point3D, l Light, shape geometry.Shape, shapes []geometry.Shape, time float64) color.RGBA {
	lightVec := l.Position.Sub(p)
	lightDir := lightVec.Normalize()
	base := shape.GetColor()

	// Shadow Check
	shadowBias := 1e-4 // Small epsilon to prevent self-shadowing ("shadow acne")
	checkP := math.Point3D{X: p.X + n.X*shadowBias, Y: p.Y + n.Y*shadowBias, Z: p.Z + n.Z*shadowBias}

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
		shapeAtT := s.AtTime(time)
		if shapeAtT == shape {
			continue // Don't check self
		}
		if shapeAtT.GetAABB().Intersects(cullAABB) {
			occluders = append(occluders, shapeAtT)
		}
	}

	shadowAttenuation := calculateShadowAttenuation(checkP, l.Position, occluders, l.Radius, time)

	// Diffuse (Lambert) component
	dot := n.Dot(lightDir)
	diffuseFactor := gomath.Max(0.15, dot*l.Intensity*shadowAttenuation) // Ambient term is 0.15

	// Specular (Phong) component
	var specularR, specularG, specularB float64
	if shadowAttenuation > 0 { // No specular highlights in full shadow
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
