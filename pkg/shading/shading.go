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

// ShadowMinBoxSize is the smallest AABB size used in the recursive shadow check.
const ShadowMinBoxSize = 0.1

// isOccluded performs a recursive AABB search to determine if a point is in shadow.
func isOccluded(p, lightPos math.Point3D, shapes []geometry.Shape, self geometry.Shape) bool {
	// Initial AABB for the search
	aabb := math.AABB3D{Min: p, Max: p}
	aabb.Min.X = gomath.Min(p.X, lightPos.X)
	aabb.Min.Y = gomath.Min(p.Y, lightPos.Y)
	aabb.Min.Z = gomath.Min(p.Z, lightPos.Z)
	aabb.Max.X = gomath.Max(p.X, lightPos.X)
	aabb.Max.Y = gomath.Max(p.Y, lightPos.Y)
	aabb.Max.Z = gomath.Max(p.Z, lightPos.Z)

	return checkOcclusionRecursive(p, lightPos, shapes, self, aabb, 0)
}

func checkOcclusionRecursive(p, lightPos math.Point3D, shapes []geometry.Shape, self geometry.Shape, aabb math.AABB3D, depth int) bool {
	if depth > 10 { // Max recursion depth
		return false
	}

	for _, shape := range shapes {
		if shape == self {
			continue
		}
		if _, ok := shape.(geometry.Plane3D); ok {
			continue
		}
		if shape.Intersects(aabb) {
			// If the box is small enough, consider it a hit
			if (aabb.Max.X-aabb.Min.X) < ShadowMinBoxSize && (aabb.Max.Y-aabb.Min.Y) < ShadowMinBoxSize && (aabb.Max.Z-aabb.Min.Z) < ShadowMinBoxSize {
				return true
			}

			// Determine which sub-box the light ray passes through.
			nextAABB := getNextAABB(aabb, lightPos)

			if checkOcclusionRecursive(p, lightPos, shapes, self, nextAABB, depth+1) {
				return true
			}
		}
	}
	return false
}

func getNextAABB(currentAABB math.AABB3D, lightPos math.Point3D) math.AABB3D {
	mid := currentAABB.Min.Add(currentAABB.Max).Mul(0.5)
	nextMin := currentAABB.Min
	nextMax := mid

	if lightPos.X > mid.X {
		nextMin.X = mid.X
		nextMax.X = currentAABB.Max.X
	}
	if lightPos.Y > mid.Y {
		nextMin.Y = mid.Y
		nextMax.Y = currentAABB.Max.Y
	}
	if lightPos.Z > mid.Z {
		nextMin.Z = mid.Z
		nextMax.Z = currentAABB.Max.Z
	}
	return math.AABB3D{Min: nextMin, Max: nextMax}
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
