package shading

import (
	"grinder/pkg/geometry"
	"grinder/pkg/math"
	"image/color"
	gomath "math"
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

// ShadedColor computes the color of a point on a surface.
func ShadedColor(p math.Point3D, n math.Normal3D, eye math.Point3D, light Light, s geometry.Shape, shapes []geometry.Shape, t float64) color.RGBA {
	lightDir := light.Position.Sub(p).Normalize()
	// Diffuse
	diffuse := gomath.Max(0, n.Dot(lightDir))
	// Specular
	viewDir := eye.Sub(p).Normalize()
	reflectDir := lightDir.Sub(n.ToVector().Mul(2 * n.Dot(lightDir))).Normalize()
	specular := gomath.Pow(gomath.Max(0, viewDir.Dot(reflectDir)), s.GetShininess())
	// Shadow
	shadow := calculateShadowAttenuation(p.Add(n.ToVector().Mul(0.001)), light.Position, shapes, light.Radius, t)

	// Combine
	c := s.GetColor()
	specColor := s.GetSpecularColor()
	intensity := light.Intensity * shadow
	return color.RGBA{
		R: uint8(gomath.Min(255, float64(c.R)*diffuse*intensity+float64(specColor.R)*specular*s.GetSpecularIntensity()*intensity)),
		G: uint8(gomath.Min(255, float64(c.G)*diffuse*intensity+float64(specColor.G)*specular*s.GetSpecularIntensity()*intensity)),
		B: uint8(gomath.Min(255, float64(c.B)*diffuse*intensity+float64(specColor.B)*specular*s.GetSpecularIntensity()*intensity)),
		A: 255,
	}
}

// ApplyAtmosphere applies the atmospheric effect to a color.
func ApplyAtmosphere(c color.RGBA, depth float64, config AtmosphereConfig) color.RGBA {
	if !config.Enabled {
		return c
	}
	t := gomath.Exp(-config.Atmosphere.Density * depth)
	return color.RGBA{
		R: uint8(float64(c.R)*t + config.Atmosphere.Color.X*(1-t)),
		G: uint8(float64(c.G)*t + config.Atmosphere.Color.Y*(1-t)),
		B: uint8(float64(c.B)*t + config.Atmosphere.Color.Z*(1-t)),
		A: 255,
	}
}

// calculateShadowAttenuation checks for shadows by marching towards the light source.
func calculateShadowAttenuation(p, lightPos math.Point3D, shapes []geometry.Shape, lightRadius float64, t_time float64) float64 {
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
			var tempShape geometry.Shape
			switch v := shape.(type) {
			case *geometry.Sphere3D:
				t := *v
				tempShape = &t
			default:
				tempShape = v
			}
			shape.AtTime(t_time, tempShape)
			if tempShape.Contains(samplePoint) {
				if vol, ok := tempShape.(geometry.VolumetricShape); ok {
					attenuation *= (1.0 - vol.GetDensity()*stepSize)
				} else {
					return 0.0 // Solid object, full occlusion
				}
			}
		}
	}
	return attenuation
}
