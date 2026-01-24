//go:build !ebiten

package main

import (
	"image"
	"sync"
)

func runGame(finalImage *image.RGBA, mu *sync.Mutex) {
	// No-op for headless builds
}
