package renderer

import (
	"grinder/pkg/camera"
	"grinder/pkg/geometry"
	"grinder/pkg/math"
	"grinder/pkg/shading"
	"image"
	"image/color"
	gomath "math"
)

// ScreenBounds defines the rectangular region of the screen to be rendered.
type ScreenBounds struct {
	MinX, MinY, MaxX, MaxY int
}

// SurfaceSample stores the geometry info for a single sample in time.
type SurfaceSample struct {
	P     math.Point3D
	N     math.Normal3D
	S     geometry.Shape
	T     float64 // Time of the sample
	Depth float64
}

// SurfaceData stores all the samples for a final pixel.
type SurfaceData struct {
	AccumulatedColor [3]float64 // R, G, B
	SampleCount      int
	VolumeSamples    []VolumeSample
}

// VolumeSample stores data for a single sample within a volume.
type VolumeSample struct {
	Shape    geometry.VolumetricShape
	Interval float64 // The length of the ray segment within the volume
	Depth    float64 // The z-depth of the sample
}

// Renderer is a configurable rendering engine.
// Culling/early out is being held off until later when we have more features as its very tricky to get right and breaks with new feature additions.
type Renderer struct {
	Camera            camera.Camera
	Shapes            []geometry.Shape
	Light             shading.Light
	Width             int
	Height            int
	MinSize           float64
	bgColor           color.RGBA
	Near              float64
	Far               float64
	Atmosphere        shading.AtmosphereConfig
	MotionBlurSamples int
}

// NewRenderer creates a new renderer with the given configuration.
func NewRenderer(cam camera.Camera, shapes []geometry.Shape, light shading.Light, width, height int, minSize, near, far float64, atmos shading.AtmosphereConfig, motionBlurSamples int) *Renderer {
	if near == 0 {
		near = 0.1
	}
	if far == 0 {
		far = 50.0
	}
	if motionBlurSamples <= 0 {
		motionBlurSamples = 1
	}
	return &Renderer{
		Camera:            cam,
		Atmosphere:        atmos,
		Shapes:            shapes,
		Light:             light,
		Width:             width,
		Height:            height,
		MinSize:           minSize,
		bgColor:           color.RGBA{30, 30, 35, 255},
		Near:              near,
		Far:               far,
		MotionBlurSamples: motionBlurSamples,
	}
}

// Calculate the tightest possible Near/Far for the current camera view
func (r *Renderer) FitDepthPlanes() { // this should fix banding on ill fitting scenes, Implement depth jitter if they return.
	eye := r.Camera.GetEye()
	minDist := 1e9
	maxDist := -1e9
	foundFinite := false

	for _, shape := range r.Shapes {
		aabb := shape.GetAABB()

		// Skip infinite planes for depth fitting
		if gomath.IsInf(aabb.Min.X, -1) {
			continue
		}
		foundFinite = true

		for _, corner := range aabb.GetCorners() {
			// Distance from camera eye to the corner
			dist := corner.Sub(eye).Length()
			if dist < minDist {
				minDist = dist
			}
			if dist > maxDist {
				maxDist = dist
			}
		}
	}

	if !foundFinite {
		// Fallback defaults if scene only has planes
		r.Near = 0.1
		r.Far = 100.0
		return
	}

	// Apply a 10% buffer so we don't accidentally clip the front or back
	r.Near = gomath.Max(0.1, minDist*0.9)
	r.Far = maxDist * 1.1
}

func (r *Renderer) Render(bounds ScreenBounds) *image.RGBA {
	tileWidth := bounds.MaxX - bounds.MinX
	tileHeight := bounds.MaxY - bounds.MinY
	img := image.NewRGBA(image.Rect(0, 0, tileWidth, tileHeight))
	prng := math.NewXorShift32(uint32(bounds.MinX*r.Width + bounds.MinY))

	surfaceBuffer := make([][]SurfaceData, tileHeight)
	for i := range surfaceBuffer {
		surfaceBuffer[i] = make([]SurfaceData, tileWidth)
	}

	// Pass 1: Dicing/Subdivision (Visibility)
	initialAABB := math.AABB3D{
		Min: math.Point3D{X: float64(bounds.MinX) / float64(r.Width), Y: float64(bounds.MinY) / float64(r.Height), Z: r.Near},
		Max: math.Point3D{X: float64(bounds.MaxX) / float64(r.Width), Y: float64(bounds.MaxY) / float64(r.Height), Z: r.Far},
	}

	// The primaryShapes list is now static for the entire tile render.
	// The GetAABB() methods on the shapes now return a temporal AABB.
	r.subdivide(initialAABB, bounds, surfaceBuffer, r.Shapes, prng)

	// Pass 2: Shading
	for y := 0; y < tileHeight; y++ {
		for x := 0; x < tileWidth; x++ {
			surface := surfaceBuffer[y][x]
			if surface.SampleCount > 0 {
				avgR := surface.AccumulatedColor[0] / float64(surface.SampleCount)
				avgG := surface.AccumulatedColor[1] / float64(surface.SampleCount)
				avgB := surface.AccumulatedColor[2] / float64(surface.SampleCount)
				finalColor := color.RGBA{R: uint8(avgR), G: uint8(avgG), B: uint8(avgB), A: 255}
				img.Set(x, y, finalColor)
			} else {
				// --- Background / Volumetrics ---
				bgColor := shading.ApplyAtmosphere(r.bgColor, r.Far, r.Atmosphere)
				img.Set(x, y, bgColor)
			}
		}
	}
	return img
}

// subdivide is the core recursive rendering function (Pass 1: Dicing).
func (r *Renderer) subdivide(aabb math.AABB3D, bounds ScreenBounds, surfaceBuffer [][]SurfaceData, shapes []geometry.Shape, prng *math.XorShift32) {
	// Broad phase cull
	intersectingShapes := make([]geometry.Shape, 0)
	for _, s := range shapes {
		if s.GetAABB().Intersects(aabb) {
			intersectingShapes = append(intersectingShapes, s)
		}
	}
	if len(intersectingShapes) == 0 {
		return
	}

	// Base case: If the AABB is small enough, do a fine-grind search for the surface.
	if (aabb.Max.X-aabb.Min.X) < r.MinSize || (aabb.Max.Y-aabb.Min.Y) < r.MinSize {
		minX, minY := int(aabb.Min.X*float64(r.Width)), int(aabb.Min.Y*float64(r.Height))
		maxX, maxY := int(aabb.Max.X*float64(r.Width)), int(aabb.Max.Y*float64(r.Height))

		for py := minY; py <= maxY; py++ {
			for px := minX; px <= maxX; px++ {
				if px >= bounds.MinX && px < bounds.MaxX && py >= bounds.MinY && py < bounds.MaxY {
					tileX, tileY := px-bounds.MinX, py-bounds.MinY
					sx, sy := (float64(px)+prng.NextFloat64())/float64(r.Width), (float64(py)+prng.NextFloat64())/float64(r.Height)

					// Perform multiple time samples for this pixel
					for i := 0; i < r.MotionBlurSamples; i++ {
						t := r.Camera.GetShutterOpen() + (r.Camera.GetShutterClose()-r.Camera.GetShutterOpen())*prng.NextFloat64()

						ray := r.Camera.GetRay(sx, sy)

						// Find the closest surface at this time 't'
						var closestSample *SurfaceSample
						minT := r.Far

						for _, s := range intersectingShapes {
							if s.IsVolumetric() {
								// Volumetric handling needs to be re-evaluated with ray marching
							} else {
								movedShape := s.AtTime(t)
								// Fine-grind search from near plane to closest hit
								// Increased steps for more precision since we aren't using the subdivided Z anymore
								for j := 0; j < 256; j++ {
									tSample := r.Near + (minT-r.Near)*(float64(j)/255.0)
									worldP := ray.At(tSample)
									if movedShape.Contains(worldP) {
										minT = tSample
										closestSample = &SurfaceSample{
											P:     worldP,
											N:     movedShape.NormalAtPoint(worldP),
											S:     s, // Store the original shape, not the temporary one
											T:     t,
											Depth: tSample, // Use t as depth
										}
										break // Found closest point for this shape, check next shape
									}
								}
							}
						}
						if closestSample != nil {
							c := shading.ShadedColor(closestSample.P, closestSample.N, r.Camera.GetEye(), r.Light, closestSample.S, shapes, t)
							surfaceBuffer[tileY][tileX].AccumulatedColor[0] += float64(c.R)
							surfaceBuffer[tileY][tileX].AccumulatedColor[1] += float64(c.G)
							surfaceBuffer[tileY][tileX].AccumulatedColor[2] += float64(c.B)
							surfaceBuffer[tileY][tileX].SampleCount++
						}
					}
				}
			}
		}
		return
	}

	// Recursive step: Subdivide the AABB into 8 smaller boxes.
	mx, my, mz := (aabb.Min.X+aabb.Max.X)/2, (aabb.Min.Y+aabb.Max.Y)/2, (aabb.Min.Z+aabb.Max.Z)/2
	subdivisions := []*math.AABB3D{
		{Min: math.Point3D{X: aabb.Min.X, Y: aabb.Min.Y, Z: aabb.Min.Z}, Max: math.Point3D{X: mx, Y: my, Z: mz}}, // Bottom-left-front
		{Min: math.Point3D{X: mx, Y: aabb.Min.Y, Z: aabb.Min.Z}, Max: math.Point3D{X: aabb.Max.X, Y: my, Z: mz}}, // Bottom-right-front
		{Min: math.Point3D{X: aabb.Min.X, Y: my, Z: aabb.Min.Z}, Max: math.Point3D{X: mx, Y: aabb.Max.Y, Z: mz}}, // Top-left-front
		{Min: math.Point3D{X: mx, Y: my, Z: aabb.Min.Z}, Max: math.Point3D{X: aabb.Max.X, Y: aabb.Max.Y, Z: mz}}, // Top-right-front
		{Min: math.Point3D{X: aabb.Min.X, Y: aabb.Min.Y, Z: mz}, Max: math.Point3D{X: mx, Y: my, Z: aabb.Max.Z}}, // Bottom-left-back
		{Min: math.Point3D{X: mx, Y: aabb.Min.Y, Z: mz}, Max: math.Point3D{X: aabb.Max.X, Y: my, Z: aabb.Max.Z}}, // Bottom-right-back
		{Min: math.Point3D{X: aabb.Min.X, Y: my, Z: mz}, Max: math.Point3D{X: mx, Y: aabb.Max.Y, Z: aabb.Max.Z}}, // Top-left-back
		{Min: math.Point3D{X: mx, Y: my, Z: mz}, Max: math.Point3D{X: aabb.Max.X, Y: aabb.Max.Y, Z: aabb.Max.Z}}, // Top-right-back
	}

	for _, subAABB := range subdivisions {
		r.subdivide(*subAABB, bounds, surfaceBuffer, intersectingShapes, prng)
	}
}
