package shading

import (
	"grinder/pkg/math"
	"image/color"
	gomath "math"
)

func ApplyAtmosphere(surfaceColor color.RGBA, distance float64, config AtmosphereConfig) color.RGBA {
	if !config.Enabled {
		return surfaceColor
	}

	// Convert surface color from RGBA (0-255) to Point3D (0-1)
	surfaceColorVec := math.Point3D{
		X: float64(surfaceColor.R) / 255.0,
		Y: float64(surfaceColor.G) / 255.0,
		Z: float64(surfaceColor.B) / 255.0,
	}

	// Calculate atmosphere blending factor
	factor := 1.0 - gomath.Exp(-distance*config.Atmosphere.Density)

	// Blend surface color with atmosphere color
	finalColorVec := surfaceColorVec.Mul(1.0 - factor).Add(config.Atmosphere.Color.Mul(factor))

	// Convert final color back to RGBA
	return color.RGBA{
		R: uint8(gomath.Min(255, finalColorVec.X*255.0)),
		G: uint8(gomath.Min(255, finalColorVec.Y*255.0)),
		B: uint8(gomath.Min(255, finalColorVec.Z*255.0)),
		A: 255,
	}
}
