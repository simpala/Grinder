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

func calculateShadowAttenuation(p, lightPos math.Point3D, occluders []geometry.Shape, lightRadius float64, tSample float64) float64 {
	const stepSize = 0.5 // Double the step size (0.5 instead of 0.25) for 2x speed
	vecToLight := lightPos.Sub(p)
	distToLight := vecToLight.Length()
	dirToLight := vecToLight.Normalize()
	attenuation := 1.0

	// March towards the light
	for t := stepSize; t < distToLight; t += stepSize {
		samplePoint := p.Add(dirToLight.Mul(t))

		for _, shape := range occluders {
			// 1. TEMPORAL CHECK: This is what makes the shadow follow the sphere
			if shape.Contains(samplePoint, tSample) {

				// 2. VOLUME CHECK
				if vol, ok := shape.(geometry.VolumetricShape); ok {
					attenuation *= (1.0 - vol.GetDensity()*stepSize)
				} else {
					// 3. SOLID HIT: Return immediately
					return 0.0
				}
			}
		}

		// OPTIMIZATION: If we are almost fully dark, stop marching
		if attenuation < 0.01 {
			return 0.0
		}
	}
	return attenuation
}
