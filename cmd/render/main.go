package main

import (
	"flag"
	"fmt"
	"grinder/pkg/loader"
	"grinder/pkg/renderer"
	"image"
	"image/draw"
	"log"
	"image/png"
	"os"
	"runtime"
	"sync"

	"github.com/hajimehoshi/ebiten/v2"
)

// Game holds the Ebitengine game state.
type Game struct {
	MasterImage *image.RGBA
	mu          *sync.Mutex
}

// Update proceeds the game state.
// Update is called every tick (1/60 [s] by default).
func (g *Game) Update() error {
	// We don't need to update any state here.
	return nil
}

// Draw draws the game screen.
// Draw is called every frame (typically 1/60[s] for 60Hz display).
func (g *Game) Draw(screen *ebiten.Image) {
	g.mu.Lock()
	defer g.mu.Unlock()
	if g.MasterImage != nil {
		screen.WritePixels(g.MasterImage.Pix)
	}
}

// Layout takes the outside size (e.g., the window size) and returns the (logical) screen size.
// If you don't have to adjust the screen size with the outside size, just return a fixed size.
func (g *Game) Layout(outsideWidth, outsideHeight int) (screenWidth, screenHeight int) {
	return 512, 512 // Should match the image dimensions
}

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


func main() {
	scenePath := flag.String("scene", "", "Path to the scene JSON file")
	fb := flag.Bool("fb", false, "Enable framebuffer preview window")
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
	rndr.FitDepthPlanes()

	fmt.Println("Rendering...")

	// --- Tiling and Concurrency ---
	const tileSize = 64
	numTilesX := width / tileSize
	numTilesY := height / tileSize
	jobs := make(chan renderer.ScreenBounds, numTilesX*numTilesY)
	var wg sync.WaitGroup

	finalImage := image.NewRGBA(image.Rect(0, 0, width, height))

	// --- Worker and Main Logic ---
	var mu sync.Mutex
	worker := func() {
		for bounds := range jobs {
			tileImg := rndr.Render(bounds)
			mu.Lock()
			rect := image.Rect(bounds.MinX, bounds.MinY, bounds.MaxX, bounds.MaxY)
			draw.Draw(finalImage, rect, tileImg, image.Point{0, 0}, draw.Src)
			mu.Unlock()
			wg.Done()
		}
	}

	for i := 0; i < runtime.NumCPU(); i++ {
		go worker()
	}

	// This function sends all the jobs and then closes the channel.
	go func() {
		for y := 0; y < height; y += tileSize {
			for x := 0; x < width; x += tileSize {
				wg.Add(1)
				jobs <- renderer.ScreenBounds{
					MinX: x,
					MinY: y,
					MaxX: x + tileSize,
					MaxY: y + tileSize,
				}
			}
		}
	}()

	// --- Main Control Flow ---
	if *fb {
		// In framebuffer mode, we need to wait for rendering in a separate goroutine
		// so the main thread can run the Ebitengine game loop.
		go func() {
			wg.Wait()
			close(jobs)
		}()

		game := &Game{MasterImage: finalImage, mu: &mu}
		ebiten.SetWindowSize(width, height)
		ebiten.SetWindowTitle("Grinder Live Preview")
		if err := ebiten.RunGame(game); err != nil {
			log.Fatalf("Ebitengine error: %v", err)
		}
	} else {
		// In headless mode, we block the main thread until rendering is complete.
		wg.Wait()
		close(jobs)
	}

	// Save the final image. This runs after the Ebitengine window is closed,
	// or after rendering is complete in headless mode.
	saveImage := func() {
		f, err := os.Create("render.png")
		if err != nil {
			log.Fatalf("Failed to create render.png: %v", err)
		}
		defer f.Close()

		if err := png.Encode(f, finalImage); err != nil {
			log.Fatalf("Failed to encode PNG: %v", err)
		}
		fmt.Println("Saved to render.png")
	}
	saveImage()
}
