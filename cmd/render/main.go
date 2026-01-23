package main

import (
	"flag"
	"fmt"
	"grinder/pkg/loader"
	"grinder/pkg/renderer"
	"image"
	"image/draw"
	"image/png"
	"os"
	"sync"
)

const sampleScene = `{
  "camera": {
    "eye": {"x": 4, "y": 3, "z": 6},
    "target": {"x": 0, "y": 0, "z": 0},
    "up": {"x": 0, "y": 1, "z": 0},
    "fov": 45,
    "aspect": 1
  },
  "light": {
    "position": {"x": 10, "y": 10, "z": 10},
    "intensity": 1.3
  },
  "shapes": [
    {
      "type": "sphere",
      "center": {"x": 0, "y": 0, "z": 0},
      "radius": 1,
      "color": {"R": 255, "G": 80, "B": 80, "A": 255}
    },
    {
      "type": "sphere",
      "center": {"x": 1.2, "y": 0.5, "z": -0.5},
      "radius": 0.5,
      "color": {"R": 80, "G": 255, "B": 80, "A": 255}
    },
    {
      "type": "plane",
      "point": {"x": 0, "y": -1, "z": 0},
      "normal": {"x": 0, "y": 1, "z": 0},
      "color": {"R": 100, "G": 100, "B": 100, "A": 255}
    }
  ]
}`

// RenderedTile holds the result of a tile rendering operation.
type RenderedTile struct {
	Bounds renderer.ScreenBounds
	Image  *image.RGBA
}

func main() {
	scenePath := flag.String("scene", "", "Path to the scene JSON file")
	flag.Parse()

	if *scenePath == "" {
		fmt.Println("Error: Scene file not provided.")
		fmt.Println("Usage: go run . -scene=<path_to_scene.json>")
		fmt.Println("\nSample Scene JSON:")
		fmt.Println(sampleScene)
		os.Exit(1)
	}

	cam, scene, light, atmos, near, far, err := loader.LoadScene(*scenePath)
	if err != nil {
		fmt.Printf("Error loading scene: %v\n", err)
		os.Exit(1)
	}

	width, height := 512, 512
	// Pass 'atmos' as the final argument to NewRenderer
	rndr := renderer.NewRenderer(cam, scene, *light, width, height, 0.004, near, far, atmos)

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
