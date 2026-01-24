package shading

import (
	"grinder/pkg/geometry"
	"grinder/pkg/math"
	"image/color"
	gomath "math"
)

type AtmosphereConfig struct {
	Type    string     `json:"type"`
	Color   color.RGBA `json:"color"`
	Density float64    `json:"density"`
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

// shading/atmosphere.go

func ApplyAtmosphere(surfaceColor color.RGBA, eyePos, worldP math.Point3D, near, far float64, config AtmosphereConfig) color.RGBA {
	if config.Type == "" || config.Density <= 0 {
		return surfaceColor
	}

	dist := worldP.Sub(eyePos).Length()

	var f float64
	if config.Type == "exp" {
		// Exponential fog: feels like thick soup
		// Higher density = thicker fog
		f = 1.0 - gomath.Exp(-dist*config.Density*0.1)
	} else {
		// Linear fog: classic 90s look
		f = (dist - near) / (far - near)
	}

	// Clamp and Blend
	if f < 0 {
		f = 0
	}
	if f > 1 {
		f = 1
	}

	return color.RGBA{
		R: uint8(float64(surfaceColor.R)*(1-f) + float64(config.Color.R)*f),
		G: uint8(float64(surfaceColor.G)*(1-f) + float64(config.Color.G)*f),
		B: uint8(float64(surfaceColor.B)*(1-f) + float64(config.Color.B)*f),
		A: 255,
	}
}

