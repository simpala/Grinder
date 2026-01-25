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
)

func main() {
	scenePath := flag.String("scene", "", "Path to the scene JSON file")
	flag.Parse()

	if *scenePath == "" {
		fmt.Println("Error: Scene file not provided.")
		os.Exit(1)
	}

	cam, scene, light, atmos, near, far, motionBlurSamples, err := loader.LoadScene(*scenePath)
	if err != nil {
		fmt.Printf("Error loading scene: %v\n", err)
		os.Exit(1)
	}

	width, height := 512, 512
	rndr := renderer.NewRenderer(cam, scene, *light, width, height, 0.004, near, far, atmos, motionBlurSamples)
	rndr.FitDepthPlanes()

	fmt.Println("Rendering...")

	const tileSize = 64
	numTilesX := width / tileSize
	numTilesY := height / tileSize

	type RenderJob struct {
		RenderBounds renderer.ScreenBounds
		DrawBounds   image.Rectangle
	}

	jobs := make(chan RenderJob, numTilesX*numTilesY)
	var wg sync.WaitGroup

	finalImage := image.NewRGBA(image.Rect(0, 0, width, height))
	var mu sync.Mutex

	const overdraw = 1
	worker := func() {
		for job := range jobs {
			tileImg := rndr.Render(job.RenderBounds)
			mu.Lock()
			draw.Draw(finalImage, job.DrawBounds, tileImg, image.Point{overdraw, overdraw}, draw.Src)
			mu.Unlock()
			wg.Done()
		}
	}

	totalTiles := numTilesX * numTilesY
	wg.Add(totalTiles)

	for i := 0; i < runtime.NumCPU(); i++ {
		go worker()
	}

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

	wg.Wait()
	fmt.Println("Render complete. Saving...")
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
