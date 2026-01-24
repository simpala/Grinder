//go:build ebiten

package main

import (
	"image"
	"log"
	"sync"

	"github.com/hajimehoshi/ebiten/v2"
)

// Game holds the Ebitengine game state.
type Game struct {
	MasterImage *image.RGBA
	mu          *sync.Mutex
}

// Update proceeds the game state.
func (g *Game) Update() error {
	return nil
}

// Draw draws the game screen.
func (g *Game) Draw(screen *ebiten.Image) {
	g.mu.Lock()
	defer g.mu.Unlock()
	if g.MasterImage != nil {
		screen.WritePixels(g.MasterImage.Pix)
	}
}

// Layout takes the outside size and returns the (logical) screen size.
func (g *Game) Layout(outsideWidth, outsideHeight int) (screenWidth, screenHeight int) {
	return 512, 512
}

func runGame(finalImage *image.RGBA, mu *sync.Mutex) {
	game := &Game{MasterImage: finalImage, mu: mu}
	ebiten.SetWindowSize(512, 512)
	ebiten.SetWindowTitle("Grinder Live Preview")

	if err := ebiten.RunGame(game); err != nil {
		log.Fatalf("Ebitengine error: %v", err)
	}
}
