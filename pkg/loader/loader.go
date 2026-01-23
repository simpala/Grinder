package loader

import (
	"encoding/json"
	"fmt"
	"grinder/pkg/camera"
	"grinder/pkg/geometry"
	"grinder/pkg/math"
	"grinder/pkg/shading"
	"image/color"
	"os"
)

type CameraConfig struct {
	Eye    math.Point3D `json:"eye"`
	Target math.Point3D `json:"target"`
	Up     math.Point3D `json:"up"`
	Fov    float64      `json:"fov"`
	Aspect float64      `json:"aspect"`
	Near   float64      `json:"near,omitempty"`
	Far    float64      `json:"far,omitempty"`
}

//	type AtmosphereConfig struct {
//		Type    string     `json:"type"`    // e.g., "linear_fog"
//		Color   color.RGBA `json:"color"`   // The color of the "haze"
//		Density float64    `json:"density"` // Strength of the effect (0.0 to 1.0)
//	}
type LightConfig struct {
	Position  math.Point3D `json:"position"`
	Intensity float64      `json:"intensity"`
	Radius    float64      `json:"radius,omitempty"`
	Samples   int          `json:"samples,omitempty"` // New field
}

type ShapeConfig struct {
	Type   string        `json:"type"`
	Center math.Point3D  `json:"center,omitempty"`
	Radius float64       `json:"radius,omitempty"`
	Point  math.Point3D  `json:"point,omitempty"`
	Normal math.Normal3D `json:"normal,omitempty"`
	// New fields for Box and Cylinder
	Min    math.Point3D `json:"min,omitempty"`
	Max    math.Point3D `json:"max,omitempty"`
	Height float64      `json:"height,omitempty"`
	// ---
	Color             color.RGBA  `json:"color"`
	Shininess         *float64    `json:"shininess,omitempty"`
	SpecularIntensity *float64    `json:"specularIntensity,omitempty"`
	SpecularColor     *color.RGBA `json:"specularColor,omitempty"`
}

type SceneConfig struct {
	Camera     CameraConfig             `json:"camera"`
	Light      LightConfig              `json:"light"`
	Atmosphere shading.AtmosphereConfig `json:"atmosphere"` // Use the shared one
	Shapes     []ShapeConfig            `json:"shapes"`
}

func LoadScene(filepath string) (camera.Camera, []geometry.Shape, *shading.Light, shading.AtmosphereConfig, float64, float64, error) {
	file, err := os.ReadFile(filepath)
	if err != nil {
		return nil, nil, nil, shading.AtmosphereConfig{}, 0, 0, fmt.Errorf("failed to read scene file: %w", err)
	}

	var config SceneConfig
	if err := json.Unmarshal(file, &config); err != nil {
		return nil, nil, nil, shading.AtmosphereConfig{}, 0, 0, fmt.Errorf("failed to parse scene file: %w", err)
	}

	cam := camera.NewLookAtCamera(
		config.Camera.Eye,
		config.Camera.Target,
		config.Camera.Up,
		config.Camera.Fov,
		config.Camera.Aspect,
	)

	samples := config.Light.Samples
	if samples <= 0 {
		samples = 9 // Default fallback
	}

	light := &shading.Light{
		Position:  config.Light.Position,
		Intensity: config.Light.Intensity,
		Radius:    config.Light.Radius,
		Samples:   samples, // Pass it through
	}

	var shapes []geometry.Shape
	for _, shapeConfig := range config.Shapes {
		shininess := 32.0
		if shapeConfig.Shininess != nil {
			shininess = *shapeConfig.Shininess
		}

		specularIntensity := 0.5
		if shapeConfig.SpecularIntensity != nil {
			specularIntensity = *shapeConfig.SpecularIntensity
		}

		specularColor := color.RGBA{R: 255, G: 255, B: 255, A: 255}
		if shapeConfig.SpecularColor != nil {
			specularColor = *shapeConfig.SpecularColor
		}

		switch shapeConfig.Type {
		case "sphere":
			shapes = append(shapes, geometry.Sphere3D{
				Center:            shapeConfig.Center,
				Radius:            shapeConfig.Radius,
				Color:             shapeConfig.Color,
				Shininess:         shininess,
				SpecularIntensity: specularIntensity,
				SpecularColor:     specularColor,
			})
		case "plane":
			shapes = append(shapes, geometry.Plane3D{
				Point:             shapeConfig.Point,
				Normal:            shapeConfig.Normal,
				Color:             shapeConfig.Color,
				Shininess:         shininess,
				SpecularIntensity: specularIntensity,
				SpecularColor:     specularColor,
			})
		case "box":
			shapes = append(shapes, geometry.Box3D{
				Min:               shapeConfig.Min,
				Max:               shapeConfig.Max,
				Color:             shapeConfig.Color,
				Shininess:         shininess,
				SpecularIntensity: specularIntensity,
				SpecularColor:     specularColor,
			})

		case "cylinder":
			shapes = append(shapes, geometry.Cylinder3D{
				Center:            shapeConfig.Center, // Base center
				Radius:            shapeConfig.Radius,
				Height:            shapeConfig.Height,
				Color:             shapeConfig.Color,
				Shininess:         shininess,
				SpecularIntensity: specularIntensity,
				SpecularColor:     specularColor,
			})
		case "cone":
			shapes = append(shapes, geometry.Cone3D{
				Center:            shapeConfig.Center,
				Radius:            shapeConfig.Radius,
				Height:            shapeConfig.Height,
				Color:             shapeConfig.Color,
				Shininess:         shininess,
				SpecularIntensity: specularIntensity,
				SpecularColor:     specularColor,
			})
		default:
			return nil, nil, nil, shading.AtmosphereConfig{}, 0, 0, fmt.Errorf("unknown shape type: %s", shapeConfig.Type)
		}
	}

	return cam, shapes, light, config.Atmosphere, config.Camera.Near, config.Camera.Far, nil
}
