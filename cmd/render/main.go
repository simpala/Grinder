package main

import (
	"flag"
	"fmt"
	"grinder/pkg/loader"
	"grinder/pkg/renderer"
	"image"
	"image/draw"
	"image/png"
	"log"
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

	cam, scene, light, atmos, near, far, shutter, err := loader.LoadScene(*scenePath)
	if err != nil {
		fmt.Printf("Error loading scene: %v\n", err)
		os.Exit(1)
	}

	width, height := 512, 512
	rndr := renderer.NewRenderer(cam, scene, *light, width, height, 0.004, near, far, atmos, shutter)
	rndr.FitDepthPlanes()

	fmt.Println("Rendering...")

	// --- Tiling and Concurrency ---
	const tileSize = 64
	const overdraw = 1
	numTilesX := width / tileSize
	numTilesY := height / tileSize

	type RenderJob struct {
		RenderBounds renderer.ScreenBounds
		DrawBounds   image.Rectangle
	}

	// Create a channel with enough buffer for all jobs
	jobs := make(chan RenderJob, numTilesX*numTilesY)
	var wg sync.WaitGroup

	finalImage := image.NewRGBA(image.Rect(0, 0, width, height))
	var mu sync.Mutex

	// Define the save function early so it's in scope for all blocks
	saveImage := func() {
		mu.Lock() // Ensure we aren't saving while a worker is mid-draw
		defer mu.Unlock()

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

	// --- WORKER POOL ---
	worker := func() {
		for job := range jobs {
			tileImg := rndr.Render(job.RenderBounds)
			mu.Lock()
			draw.Draw(finalImage, job.DrawBounds, tileImg, image.Point{overdraw, overdraw}, draw.Src)
			mu.Unlock()
			wg.Done()
		}
	}

	// 1. IMPORTANT: Pre-add to the WaitGroup so wg.Wait() doesn't bypass early
	totalTiles := numTilesX * numTilesY
	wg.Add(totalTiles)

	// 2. Start the workers
	for i := 0; i < runtime.NumCPU(); i++ {
		go worker()
	}

	// 3. Feed the jobs
	go func() {
		for y := 0; y < height; y += tileSize {
			for x := 0; x < width; x += tileSize {
				jobs <- RenderJob{
					RenderBounds: renderer.ScreenBounds{
						MinX: x - overdraw,
						MinY: y - overdraw,
						MaxX: x + tileSize + overdraw,
						MaxY: y + tileSize + overdraw,
					},
					DrawBounds: image.Rect(x, y, x+tileSize, y+tileSize),
				}
			}
		}
		close(jobs)
	}()

	// --- MAIN CONTROL FLOW ---
	if *fb {
		// FB Mode: Save in background when done, but keep window open
		go func() {
			wg.Wait()
			fmt.Println("Render complete. Saving auto-snapshot...")
			saveImage()
		}()

		game := &Game{MasterImage: finalImage, mu: &mu}
		ebiten.SetWindowSize(width, height)
		ebiten.SetWindowTitle("Grinder Live Preview")

		if err := ebiten.RunGame(game); err != nil {
			log.Fatalf("Ebitengine error: %v", err)
		}
	} else {
		// Headless Mode: Block here until all workers call wg.Done()
		wg.Wait()
		fmt.Println("Render complete. Saving...")
		saveImage()
	}
}
