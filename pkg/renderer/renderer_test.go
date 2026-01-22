package renderer

import (
	"bytes"
	"grinder/pkg/loader"
	"grinder/pkg/math"
	"image"
	"image/draw"
	"image/png"
	"os"
	"path/filepath"
	"sync"
	"testing"
)

func TestRender(t *testing.T) {
	scenes := []string{"balls", "shadow_test"}

	for _, sceneName := range scenes {
		t.Run(sceneName, func(t *testing.T) {
			// Load scene
			cam, scene, light, err := loader.LoadScene(filepath.Join("..", "..", "scenes", sceneName+".json"))
			if err != nil {
				t.Fatalf("failed to load scene: %v", err)
			}

			// Render
			width, height := 512, 512
			rndr := NewRenderer(cam, scene, *light, width, height, 0.004)

			// Tiling logic must match main.go to ensure deterministic results
			const numTilesX, numTilesY = 4, 4
			tileWidth, tileHeight := width/numTilesX, height/numTilesY

			var wg sync.WaitGroup
			renderedTiles := make(chan struct {
				Bounds ScreenBounds
				Image  *image.RGBA
			}, numTilesX*numTilesY)

			for y := 0; y < numTilesY; y++ {
				for x := 0; x < numTilesX; x++ {
					wg.Add(1)
					go func(x, y int) {
						defer wg.Done()
						bounds := ScreenBounds{
							MinX: x * tileWidth,
							MinY: y * tileHeight,
							MaxX: (x + 1) * tileWidth,
							MaxY: (y + 1) * tileHeight,
						}
						rng := math.NewXorShift32(uint32(y*numTilesX + x))
						tileImg := rndr.Render(bounds, rng)
						renderedTiles <- struct {
							Bounds ScreenBounds
							Image  *image.RGBA
						}{bounds, tileImg}
					}(x, y)
				}
			}

			go func() {
				wg.Wait()
				close(renderedTiles)
			}()

			// Assemble the image
			finalImage := image.NewRGBA(image.Rect(0, 0, width, height))
			for tile := range renderedTiles {
				rect := image.Rect(tile.Bounds.MinX, tile.Bounds.MinY, tile.Bounds.MaxX, tile.Bounds.MaxY)
				draw.Draw(finalImage, rect, tile.Image, image.Point{0, 0}, draw.Src)
			}
			img := finalImage

			// Load golden image
			goldenPath := filepath.Join("..", "..", "testdata", sceneName+"_golden.png")
			goldenFile, err := os.Open(goldenPath)
			if err != nil {
				t.Fatalf("failed to open golden image: %v", err)
			}
			defer goldenFile.Close()

			goldenImg, _, err := image.Decode(goldenFile)
			if err != nil {
				t.Fatalf("failed to decode golden image: %v", err)
			}

			// Compare images
			if !compareImages(img, goldenImg) {
				// Save failing image for debugging
				buf := new(bytes.Buffer)
				if err := png.Encode(buf, img); err != nil {
					t.Fatalf("failed to encode failing image: %v", err)
				}
				if err := os.WriteFile(filepath.Join("..", "..", "testdata", sceneName+"_failed.png"), buf.Bytes(), 0644); err != nil {
					t.Fatalf("failed to save failing image: %v", err)
				}

				t.Errorf("rendered image does not match golden image")
			}
		})
	}
}

func compareImages(a, b image.Image) bool {
	if a.Bounds() != b.Bounds() {
		return false
	}

	for y := a.Bounds().Min.Y; y < a.Bounds().Max.Y; y++ {
		for x := a.Bounds().Min.X; x < a.Bounds().Max.X; x++ {
			if a.At(x, y) != b.At(x, y) {
				return false
			}
		}
	}

	return true
}
