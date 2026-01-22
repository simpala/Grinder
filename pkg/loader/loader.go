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
}

type LightConfig struct {
	Position  math.Point3D `json:"position"`
	Intensity float64      `json:"intensity"`
}

type ShapeConfig struct {
	Type   string         `json:"type"`
	Center math.Point3D `json:"center,omitempty"`
	Radius float64        `json:"radius,omitempty"`
	Point  math.Point3D `json:"point,omitempty"`
	Normal math.Normal3D  `json:"normal,omitempty"`
	Color  color.RGBA     `json:"color"`
}

type SceneConfig struct {
	Camera CameraConfig  `json:"camera"`
	Light  LightConfig   `json:"light"`
	Shapes []ShapeConfig `json:"shapes"`
}

func LoadScene(filepath string) (*camera.Camera, []geometry.Shape, *shading.Light, error) {
	file, err := os.ReadFile(filepath)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("failed to read scene file: %w", err)
	}

	var config SceneConfig
	if err := json.Unmarshal(file, &config); err != nil {
		return nil, nil, nil, fmt.Errorf("failed to parse scene file: %w", err)
	}

	cam := camera.NewLookAtCamera(
		config.Camera.Eye,
		config.Camera.Target,
		config.Camera.Up,
		config.Camera.Fov,
		config.Camera.Aspect,
	)

	light := &shading.Light{
		Position:  config.Light.Position,
		Intensity: config.Light.Intensity,
	}

	var shapes []geometry.Shape
	for _, shapeConfig := range config.Shapes {
		switch shapeConfig.Type {
		case "sphere":
			shapes = append(shapes, geometry.Sphere3D{
				Center: shapeConfig.Center,
				Radius: shapeConfig.Radius,
				Color:  shapeConfig.Color,
			})
		case "plane":
			shapes = append(shapes, geometry.Plane3D{
				Point:  shapeConfig.Point,
				Normal: shapeConfig.Normal,
				Color:  shapeConfig.Color,
			})
		default:
			return nil, nil, nil, fmt.Errorf("unknown shape type: %s", shapeConfig.Type)
		}
	}

	return cam, shapes, light, nil
}
