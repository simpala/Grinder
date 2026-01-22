package main

import (
	"fmt"
	"grinder/pkg/math"
	"grinder/pkg/camera"
	"grinder/pkg/geometry"
	"grinder/pkg/renderer"
	"grinder/pkg/shading"
	"image"
	"image/color"
	"image/draw"
	"image/png"
	"os"
	"sync"
)

// RenderedTile holds the result of a tile rendering operation.
type RenderedTile struct {
	Bounds renderer.ScreenBounds
	Image  *image.RGBA
}

func main() {
	scene := []geometry.Shape{
		geometry.Sphere3D{Center: math.Point3D{X: 0, Y: 0, Z: 0}, Radius: 1.0, Color: color.RGBA{R: 255, G: 80, B: 80, A: 255}},
		geometry.Sphere3D{Center: math.Point3D{X: 1.2, Y: 0.5, Z: -0.5}, Radius: 0.5, Color: color.RGBA{R: 80, G: 255, B: 80, A: 255}},
		geometry.Plane3D{
			Point:  math.Point3D{X: 0, Y: -1.0, Z: 0},
			Normal: math.Normal3D{X: 0, Y: 1, Z: 0},
			Color:  color.RGBA{R: 100, G: 100, B: 100, A: 255},
		},
	}

	cam := camera.NewLookAtCamera(
		math.Point3D{X: 4, Y: 3, Z: 6}, // Eye
		math.Point3D{X: 0, Y: 0, Z: 0}, // Target
		math.Point3D{X: 0, Y: 1, Z: 0}, // Up
		45.0,
		1.0,
	)

	light := shading.Light{Position: math.Point3D{X: 10, Y: 10, Z: 10}, Intensity: 1.3}

	width, height := 512, 512
	rndr := renderer.NewRenderer(cam, scene, light, width, height, 0.004)

	fmt.Println("Rendering...")

	// --- Tiling and Concurrency ---
	const numTilesX, numTilesY = 4, 4
	tileWidth, tileHeight := width/numTilesX, height/numTilesY

	var wg sync.WaitGroup
	renderedTiles := make(chan RenderedTile, numTilesX*numTilesY)

	for y := 0; y < numTilesY; y++ {
		for x := 0; x < numTilesX; x++ {
			wg.Add(1)
			go func(x, y int) {
				defer wg.Done()
				bounds := renderer.ScreenBounds{
					MinX: x * tileWidth,
					MinY: y * tileHeight,
					MaxX: (x + 1) * tileWidth,
					MaxY: (y + 1) * tileHeight,
				}
				tileImg := rndr.Render(bounds)
				renderedTiles <- RenderedTile{Bounds: bounds, Image: tileImg}
			}(x, y)
		}
	}

	// Wait for all rendering to complete, then close the channel.
	go func() {
		wg.Wait()
		close(renderedTiles)
	}()

	// --- Image Assembly ---
	finalImage := image.NewRGBA(image.Rect(0, 0, width, height))
	for tile := range renderedTiles {
		rect := image.Rect(tile.Bounds.MinX, tile.Bounds.MinY, tile.Bounds.MaxX, tile.Bounds.MaxY)
		draw.Draw(finalImage, rect, tile.Image, image.Point{0, 0}, draw.Src)
	}

	f, err := os.Create("render.png")
	if err != nil {
		panic(err)
	}
	defer f.Close()

	if err := png.Encode(f, finalImage); err != nil {
		panic(err)
	}

	fmt.Println("Saved to render.png")
}
