package main

import (
	"flag"
	"fmt"
	"grinder/pkg/camera"
	"grinder/pkg/loader"
	"grinder/pkg/math"
	"grinder/pkg/renderer"
	"os"
)

func main() {
	scenePath := flag.String("scene", "scenes/simple.json", "path to scene JSON file")
	tempFile := flag.String("temp", "temp.bin", "temporary atom file")
	outFile := flag.String("out", "final.bin", "output baked scene file")
	minSize := flag.Float64("minsize", 0.05, "minimum voxel size")
	flag.Parse()

	cam, shapes, light, _, near, far, shutter, err := loader.LoadScene(*scenePath)
	if err != nil {
		fmt.Printf("Error loading scene: %v\n", err)
		os.Exit(1)
	}

	// For Near/Far, if they are 0, use defaults
	if near == 0 {
		near = 0.1
	}
	if far == 0 {
		far = 50.0
	}

	fmt.Printf("Baking scene: %s\n", *scenePath)
	fmt.Printf("Voxel MinSize: %f, Near: %f, Far: %f\n", *minSize, near, far)

	// Extract camera info for header
	var target, up math.Point3D
	var fov float64
	if pc, ok := cam.(*camera.PerspectiveCamera); ok {
		target = pc.GetEye().Add(pc.GetForward())
		up = pc.GetUp()
		fov = pc.GetFov()
	}

	engine := renderer.NewBakeEngine(cam, shapes, *light, 1024, 1024, *minSize, near, far, shutter, target, up, fov)
	err = engine.Bake(*tempFile, *outFile)
	if err != nil {
		fmt.Printf("Error during bake: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("Bake completed successfully. Starting verification...")

	err = engine.Verify(*outFile)
	if err != nil {
		fmt.Printf("Error during verification: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("Verification completed.")
}
