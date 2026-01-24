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

// calculateShadowAttenuation checks for shadows by marching towards the light source.
func calculateShadowAttenuation(p, lightPos math.Point3D, shapes []geometry.Shape, lightRadius float64, time float64) float64 {
	const stepSize = 0.25
	vecToLight := lightPos.Sub(p)
	distToLight := vecToLight.Length()
	dirToLight := vecToLight.Normalize()
	attenuation := 1.0

	// March from the point towards the light.
	for t := stepSize; t < distToLight; t += stepSize {
		samplePoint := p.Add(dirToLight.Mul(t))
		for _, shape := range shapes {
			if _, ok := shape.(*geometry.Plane3D); ok {
				continue
			}
			movedShape := shape.AtTime(time)
			if movedShape.Contains(samplePoint) {
				if vol, ok := movedShape.(geometry.VolumetricShape); ok {
					attenuation *= (1.0 - vol.GetDensity()*stepSize)
				} else {
					return 0.0 // Solid object, full occlusion
				}
			}
		}
	}

	return attenuation
}
