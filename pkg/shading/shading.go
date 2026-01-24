package shading

import (
	"grinder/pkg/geometry"
	"grinder/pkg/math"
)

// Atmosphere represents the properties of the atmospheric effect.
type Atmosphere struct {
	Color   math.Point3D `json:"color"`
	Density float64      `json:"density"`
}

// AtmosphereConfig holds the configuration for the atmospheric effect.
type AtmosphereConfig struct {
	Enabled    bool       `json:"enabled"`
	Atmosphere Atmosphere `json:"atmosphere"`
}

// Light represents a light source in the scene.
type Light struct {
	Position  math.Point3D
	Intensity float64
	Radius    float64
	Samples   int // New field
}

// isOccluded checks for shadows by marching small AABBs towards the light source.
func isOccluded(p, lightPos math.Point3D, shapes []geometry.Shape, lightRadius float64) bool {
	const stepSize = 0.5
	vecToLight := lightPos.Sub(p)
	distToLight := vecToLight.Length()
	dirToLight := vecToLight.Normalize()

	// March from the point towards the light.
	for t := stepSize; t < distToLight; t += stepSize {
		samplePoint := p.Add(dirToLight.Mul(t))
		for _, shape := range shapes {
			if _, ok := shape.(geometry.Plane3D); ok {
				continue
			}
			// Use the shape's inherent 'Contains' logic instead of a physical box
			if shape.Contains(samplePoint) {
				return true
			}
		}
	}

	return false // The point is not in shadow.
}
