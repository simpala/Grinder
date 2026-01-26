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

type SceneConfig struct {
	Camera     CameraConfig             `json:"camera"`
	Shutter    float64                  `json:"shutter,omitempty"` // e.g., 0.5 for 180-degree shutter
	Light      LightConfig              `json:"light"`
	Atmosphere shading.AtmosphereConfig `json:"atmosphere"`
	Shapes     []ShapeConfig            `json:"shapes"`
}
type LightConfig struct {
	Position  math.Point3D `json:"position"`
	Intensity float64      `json:"intensity"`
	Radius    float64      `json:"radius,omitempty"`
	Samples   int          `json:"samples,omitempty"` // New field
}

type ShapeConfig struct {
	Type              string        `json:"type"`
	Center            math.Point3D  `json:"center,omitempty"`
	Destination       math.Point3D  `json:"destination,omitempty"` // New: where motion ends
	Radius            float64       `json:"radius,omitempty"`
	Point             math.Point3D  `json:"point,omitempty"`
	Normal            math.Normal3D `json:"normal,omitempty"`
	Min               math.Point3D  `json:"min,omitempty"`
	Max               math.Point3D  `json:"max,omitempty"`
	Height            float64       `json:"height,omitempty"`
	Density           float64       `json:"density,omitempty"`
	Color             color.RGBA    `json:"color"`
	Shininess         *float64      `json:"shininess,omitempty"`
	SpecularIntensity *float64      `json:"specularIntensity,omitempty"`
	SpecularColor     *color.RGBA   `json:"specularColor,omitempty"`
	P00               math.Point3D  `json:"p00,omitempty"`
	P10               math.Point3D  `json:"p10,omitempty"`
	P11               math.Point3D  `json:"p11,omitempty"`
	P01               math.Point3D  `json:"p01,omitempty"`
	Thickness         float64       `json:"thickness,omitempty"`
	Iterations        int           `json:"iterations"`
}

// Changed return signature: added a float64 before error to hold the shutter value
func LoadScene(filepath string) (camera.Camera, []geometry.Shape, *shading.Light, shading.AtmosphereConfig, float64, float64, float64, error) {
	file, err := os.ReadFile(filepath)
	if err != nil {
		return nil, nil, nil, shading.AtmosphereConfig{}, 0, 0, 0, fmt.Errorf("failed to read scene file: %w", err)
	}

	var config SceneConfig
	if err := json.Unmarshal(file, &config); err != nil {
		return nil, nil, nil, shading.AtmosphereConfig{}, 0, 0, 0, fmt.Errorf("failed to parse scene file: %w", err)
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
		samples = 9
	}

	light := &shading.Light{
		Position:  config.Light.Position,
		Intensity: config.Light.Intensity,
		Radius:    config.Light.Radius,
		Samples:   samples,
	}

	var shapes []geometry.Shape
	for _, shapeConfig := range config.Shapes {
		// ... (your existing shininess/specular logic remains the same) ...
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
			velocity := math.Point3D{X: 0, Y: 0, Z: 0}
			if shapeConfig.Destination != (math.Point3D{}) {
				velocity = shapeConfig.Destination.Sub(shapeConfig.Center)
			}
			shapes = append(shapes, geometry.Sphere3D{
				Center:            shapeConfig.Center,
				Velocity:          velocity,
				Radius:            shapeConfig.Radius,
				Color:             shapeConfig.Color,
				Shininess:         shininess,
				SpecularIntensity: specularIntensity,
				SpecularColor:     specularColor,
			})
		case "box":
			velocity := math.Point3D{X: 0, Y: 0, Z: 0}
			if shapeConfig.Destination != (math.Point3D{}) {
				// For a box, Destination applies to both Min and Max proportionally.
				// We'll calculate velocity based on Min for simplicity, assuming Max moves with Min.
				velocity = shapeConfig.Destination.Sub(shapeConfig.Min)
			}
			shapes = append(shapes, geometry.Box3D{
				Min:               shapeConfig.Min,
				Max:               shapeConfig.Max,
				Velocity:          velocity,
				Color:             shapeConfig.Color,
				Shininess:         shininess,
				SpecularIntensity: specularIntensity,
				SpecularColor:     specularColor,
			})
		case "cylinder":
			velocity := math.Point3D{X: 0, Y: 0, Z: 0}
			if shapeConfig.Destination != (math.Point3D{}) {
				velocity = shapeConfig.Destination.Sub(shapeConfig.Center)
			}
			shapes = append(shapes, geometry.Cylinder3D{
				Center:            shapeConfig.Center,
				Velocity:          velocity,
				Radius:            shapeConfig.Radius,
				Height:            shapeConfig.Height,
				Color:             shapeConfig.Color,
				Shininess:         shininess,
				SpecularIntensity: specularIntensity,
				SpecularColor:     specularColor,
			})
		case "cone":
			velocity := math.Point3D{X: 0, Y: 0, Z: 0}
			if shapeConfig.Destination != (math.Point3D{}) {
				velocity = shapeConfig.Destination.Sub(shapeConfig.Center)
			}
			shapes = append(shapes, geometry.Cone3D{
				Center:            shapeConfig.Center,
				Velocity:          velocity,
				Radius:            shapeConfig.Radius,
				Height:            shapeConfig.Height,
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
		case "quad":
			thickness := shapeConfig.Thickness
			if thickness == 0 {
				thickness = 0.01 // Default tiny thickness so it's not a zero-volume plane
			}
			shapes = append(shapes, &geometry.BilinearQuad{
				P00:               shapeConfig.P00,
				P10:               shapeConfig.P10,
				P11:               shapeConfig.P11,
				P01:               shapeConfig.P01,
				Thickness:         thickness,
				Color:             shapeConfig.Color,
				Shininess:         shininess,
				SpecularIntensity: specularIntensity,
				SpecularColor:     specularColor,
			})
		case "sds_box":
			base := geometry.CreateCubeMesh(shapeConfig.Center, shapeConfig.Radius)
			// Subdivide
			currentMesh := base
			for i := 0; i < shapeConfig.Iterations; i++ { // Now we can handle 3 iterations!
				currentMesh = currentMesh.Subdivide()
			}

			// Calculate the total AABB for the whole mesh
			totalAABB := math.AABB3D{Min: currentMesh.Vertices[0], Max: currentMesh.Vertices[0]}
			for _, v := range currentMesh.Vertices {
				totalAABB = totalAABB.Expand(v)
			}
			// Pad for thickness
			totalAABB.Min = totalAABB.Min.Sub(math.Point3D{X: 0.1, Y: 0.1, Z: 0.1})
			totalAABB.Max = totalAABB.Max.Add(math.Point3D{X: 0.1, Y: 0.1, Z: 0.1})
			// Convert faces to BilinearQuads
			var meshQuads []*geometry.BilinearQuad
			for _, face := range currentMesh.Faces {
				center := shapeConfig.Center
				// For each vertex in the face, calculate the normal pointing away from center

				p00 := currentMesh.Vertices[face[0]]
				p10 := currentMesh.Vertices[face[1]]
				p11 := currentMesh.Vertices[face[2]]
				p01 := currentMesh.Vertices[face[3]]
				n00 := math.Normal3D(currentMesh.Vertices[face[0]].Sub(center).Normalize())
				n10 := math.Normal3D(currentMesh.Vertices[face[1]].Sub(center).Normalize())
				n11 := math.Normal3D(currentMesh.Vertices[face[2]].Sub(center).Normalize())
				n01 := math.Normal3D(currentMesh.Vertices[face[3]].Sub(center).Normalize())

				// Create a local AABB for just this one quad
				quadAABB := math.AABB3D{Min: p00, Max: p00}.
					Expand(p10).
					Expand(p11).
					Expand(p01)

				// Add a tiny bit of padding to close the "shadow holes"
				pad := 0.001
				quadAABB.Min = quadAABB.Min.Sub(math.Point3D{X: pad, Y: pad, Z: pad})
				quadAABB.Max = quadAABB.Max.Add(math.Point3D{X: pad, Y: pad, Z: pad})
				meshQuads = append(meshQuads, &geometry.BilinearQuad{
					N00:               n00,
					N10:               n10,
					N11:               n11,
					N01:               n01,
					P00:               currentMesh.Vertices[face[0]],
					P10:               currentMesh.Vertices[face[1]],
					P11:               currentMesh.Vertices[face[2]],
					P01:               currentMesh.Vertices[face[3]],
					AABB:              quadAABB,
					Thickness:         shapeConfig.Thickness,
					Color:             shapeConfig.Color,
					Shininess:         shininess,
					SpecularIntensity: specularIntensity,
					SpecularColor:     specularColor,
				})
			}

			shapes = append(shapes, &geometry.SDSObject{
				Quads:             meshQuads,
				AABB:              totalAABB,
				Color:             shapeConfig.Color,
				Shininess:         shininess,
				SpecularIntensity: specularIntensity,
				SpecularColor:     specularColor,
			})

		default:
			return nil, nil, nil, shading.AtmosphereConfig{}, 0, 0, 0, fmt.Errorf("unknown shape type: %s", shapeConfig.Type)
		}
	}

	shutter := config.Shutter
	if shutter == 0 {
		shutter = 1.0
	}

	// Returning 8 values now: cam, shapes, light, atmosphere, near, far, SHUTTER, err
	return cam, shapes, light, config.Atmosphere, config.Camera.Near, config.Camera.Far, shutter, nil
}
