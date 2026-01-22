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

		// Stochastic Surface Normal Bias
		epsilon := 0.01 + (rng.Float64() * 0.02)
		biasedStart := p.Add(n.ToVector().Mul(epsilon))

		// Define the search volume
		searchAABB := math.AABB3D{
			Min: math.Point3D{
				X: gomath.Min(biasedStart.X, targetPos.X),
				Y: gomath.Min(biasedStart.Y, targetPos.Y),
				Z: gomath.Min(biasedStart.Z, targetPos.Z),
			},
			Max: math.Point3D{
				X: gomath.Max(biasedStart.X, targetPos.X),
				Y: gomath.Max(biasedStart.Y, targetPos.Y),
				Z: gomath.Max(biasedStart.Z, targetPos.Z),
			},
		}

		if isOccluded(searchAABB, shapes) {
			shadowHits += 1.0
		}
	}

	return gomath.Max(0.2, 1.0-(shadowHits/float64(numSamples)))
}

func isOccluded(aabb math.AABB3D, shapes []geometry.Shape) bool {
	// Base Case: If the search volume is small enough, check for containment.
	if aabb.Max.X-aabb.Min.X < 0.01 {
		center := aabb.Centroid()
		for _, s := range shapes {
			if _, ok := s.(geometry.Plane3D); ok {
				continue
			}
			if s.Contains(center) {
				return true
			}
		}
		return false
	}

	// Recursive Step: Subdivide the AABB and check for intersections.
	for _, s := range shapes {
		if _, ok := s.(geometry.Plane3D); ok {
			continue
		}
		if s.Intersects(aabb) {
			// Subdivide and recurse
			mx, my, mz := (aabb.Min.X+aabb.Max.X)/2, (aabb.Min.Y+aabb.Max.Y)/2, (aabb.Min.Z+aabb.Max.Z)/2
			xs := [3]float64{aabb.Min.X, mx, aabb.Max.X}
			ys := [3]float64{aabb.Min.Y, my, aabb.Max.Y}
			zs := [3]float64{aabb.Min.Z, mz, aabb.Max.Z}

			for zi := 0; zi < 2; zi++ {
				for xi := 0; xi < 2; xi++ {
					for yi := 0; yi < 2; yi++ {
						childAABB := math.AABB3D{
							Min: math.Point3D{X: xs[xi], Y: ys[yi], Z: zs[zi]},
							Max: math.Point3D{X: xs[xi+1], Y: ys[yi+1], Z: zs[zi+1]},
						}
						if isOccluded(childAABB, shapes) {
							return true
						}
					}
				}
			}
		}
	}

	return false
}
