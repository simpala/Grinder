package main

import (
	"flag"
	"fmt"
	"grinder/pkg/camera"
	"grinder/pkg/loader"
	"grinder/pkg/math"
	"grinder/pkg/renderer"
	"image"
	"image/color"
	"image/png"
	gomath "math"
	"os"
	"runtime"
	"sync"
)

func main() {
	scenePath := flag.String("scene", "", "path to scene JSON file (optional, uses header if omitted)")
	bakedPath := flag.String("baked", "final.bin", "path to baked scene binary")
	outPath := flag.String("out", "trace.png", "output image path")
	width := flag.Int("width", 800, "image width")
	height := flag.Int("height", 800, "image height")
	samples := flag.Int("samples", 4, "samples per pixel")
	memLimit := flag.Int64("memlimit", 2048, "memory limit in MB for in-memory loading (default 2GB)")
	flag.Parse()

	scene, err := renderer.LoadBakedScene(*bakedPath, *memLimit*1024*1024)
	if err != nil {
		fmt.Printf("Error loading baked scene: %v\n", err)
		os.Exit(1)
	}
	defer scene.Close()

	var cam camera.Camera
	var near, far float64
	if *scenePath != "" {
		var err error
		cam, _, _, _, near, far, _, err = loader.LoadScene(*scenePath)
		if err != nil {
			fmt.Printf("Error loading scene: %v\n", err)
			os.Exit(1)
		}
	} else {
		// Use camera from header
		bc := scene.Header.BakeCamera
		cam = camera.NewLookAtCamera(
			math.Point3D{X: float64(bc.Eye[0]), Y: float64(bc.Eye[1]), Z: float64(bc.Eye[2])},
			math.Point3D{X: float64(bc.Target[0]), Y: float64(bc.Target[1]), Z: float64(bc.Target[2])},
			math.Point3D{X: float64(bc.Up[0]), Y: float64(bc.Up[1]), Z: float64(bc.Up[2])},
			float64(bc.Fov),
			float64(bc.Aspect),
		)
		near, far = float64(bc.Near), float64(bc.Far)
	}

	if near == 0 {
		near = 0.1
	}
	if far == 0 {
		far = 50.0
	}

	img := image.NewRGBA(image.Rect(0, 0, *width, *height))

	numCPUs := runtime.NumCPU()
	var wg sync.WaitGroup
	wg.Add(numCPUs)

	rowsPerCPU := *height / numCPUs

	for cpu := 0; cpu < numCPUs; cpu++ {
		go func(cpuID int) {
			defer wg.Done()
			startY := cpuID * rowsPerCPU
			endY := startY + rowsPerCPU
			if cpuID == numCPUs-1 {
				endY = *height
			}

			prng := math.NewXorShift32(uint32(cpuID + 1))

			for y := startY; y < endY; y++ {
				for x := 0; x < *width; x++ {
					var colorSum math.Point3D
					for s := 0; s < *samples; s++ {
						fx := (float64(x) + prng.NextFloat64()) / float64(*width)
						fy := (float64(y) + prng.NextFloat64()) / float64(*height)

						pNear := cam.Project(fx, fy, near)
						pFar := cam.Project(fx, fy, far)
						rayDir := pFar.Sub(pNear).Normalize()
						ray := math.Ray{Origin: pNear, Direction: rayDir}

						colorSum = colorSum.Add(trace(ray, scene, 0, prng))
					}
					avg := colorSum.Mul(1.0 / float64(*samples))
					img.Set(x, y, color.RGBA{
						R: uint8(gomath.Min(255, avg.X*255)),
						G: uint8(gomath.Min(255, avg.Y*255)),
						B: uint8(gomath.Min(255, avg.Z*255)),
						A: 255,
					})
				}
			}
		}(cpu)
	}

	wg.Wait()

	f, err := os.Create(*outPath)
	if err != nil {
		fmt.Printf("Error creating output file: %v\n", err)
		os.Exit(1)
	}
	png.Encode(f, img)
	f.Close()
	fmt.Printf("Trace complete. Saved to %s\n", *outPath)
}

func trace(ray math.Ray, scene *renderer.BakedScene, depth int, prng *math.XorShift32) math.Point3D {
	if depth > 2 {
		return math.Point3D{}
	}

	hit, atom := scene.Intersect(ray)
	if !hit {
		return math.Point3D{X: 0.05, Y: 0.05, Z: 0.1} // Dark blue sky
	}

	pos := math.Point3D{X: float64(atom.Pos[0]), Y: float64(atom.Pos[1]), Z: float64(atom.Pos[2])}
	normal := renderer.OctDecode(atom.Normal)
	albedo := math.Point3D{X: float64(atom.Albedo[0]) / 255, Y: float64(atom.Albedo[1]) / 255, Z: float64(atom.Albedo[2]) / 255}

	// Direct Light (baked)
	lDir := renderer.OctDecode(atom.LightDir)
	lCol := math.Point3D{X: float64(atom.LightColor[0]) / 255, Y: float64(atom.LightColor[1]) / 255, Z: float64(atom.LightColor[2]) / 255}
	dot := gomath.Max(0.1, normal.Dot(lDir))
	direct := lCol.Mul(dot)

	// Indirect Bounce
	var indirect math.Point3D
	if depth < 2 {
		nextDir := sampleHemisphere(normal, prng)
		nextRayOrigin := pos.Add(normal.Mul(float64(scene.Header.Epsilon)))
		nextRay := math.Ray{Origin: nextRayOrigin, Direction: nextDir}
		indirect = trace(nextRay, scene, depth+1, prng).Mul(0.5)
	}

	res := direct.Add(indirect)
	return math.Point3D{X: albedo.X * res.X, Y: albedo.Y * res.Y, Z: albedo.Z * res.Z}
}

func sampleHemisphere(n math.Point3D, prng *math.XorShift32) math.Point3D {
	u1 := prng.NextFloat64()
	u2 := prng.NextFloat64()

	r := gomath.Sqrt(1.0 - u1*u1)
	phi := 2.0 * gomath.Pi * u2

	x := r * gomath.Cos(phi)
	y := r * gomath.Sin(phi)
	z := u1

	var up math.Point3D
	if gomath.Abs(n.Y) < 0.9 {
		up = math.Point3D{X: 0, Y: 1, Z: 0}
	} else {
		up = math.Point3D{X: 1, Y: 0, Z: 0}
	}
	tangent := n.Cross(up).Normalize()
	bitangent := n.Cross(tangent).Normalize()

	return tangent.Mul(x).Add(bitangent.Mul(y)).Add(n.Mul(z)).Normalize()
}
