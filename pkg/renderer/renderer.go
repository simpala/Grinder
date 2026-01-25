package renderer

import (
	"grinder/pkg/camera"
	"grinder/pkg/geometry"
	"grinder/pkg/math"
	"grinder/pkg/shading"
	"image"
	"image/color"
	gomath "math"
	"sort"
)

// ScreenBounds defines the rectangular region of the screen to be rendered.
type ScreenBounds struct {
	MinX, MinY, MaxX, MaxY int
}

// SurfaceData stores the essential geometry info for a final pixel.
type SurfaceData struct {
	P, N          math.Point3D
	S             geometry.Shape
	Depth         float64
	Hit           bool
	rgba          [4]float64 // Accumulator for R, G, B, and sample count
	volInterval   float64    // Total interval length for volumetric samples
	volDensity    float64    // Accumulated density for volumetric samples
	volColor      color.RGBA // Accumulated color for volumetric samples
	time          float64    // The time at which the sample was taken
	VolumeSamples int        // Number of volume samples accumulated
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
	if motionBlurSamples == 0 {
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

	surfaceBuffer := make([][]SurfaceData, tileHeight)
	for i := range surfaceBuffer {
		surfaceBuffer[i] = make([]SurfaceData, tileWidth)
	}

	// Pass 1: Dicing/Subdivision
	initialAABB := math.AABB3D{
		Min: math.Point3D{X: float64(bounds.MinX) / float64(r.Width), Y: float64(bounds.MinY) / float64(r.Height), Z: r.Near},
		Max: math.Point3D{X: float64(bounds.MaxX) / float64(r.Width), Y: float64(bounds.MaxY) / float64(r.Height), Z: r.Far},
	}

	primaryShapes := make([]geometry.Shape, len(r.Shapes))
	copy(primaryShapes, r.Shapes)
	sort.Slice(primaryShapes, func(i, j int) bool {
		distI := primaryShapes[i].GetCenter().Sub(r.Camera.GetEye()).Length()
		distJ := primaryShapes[j].GetCenter().Sub(r.Camera.GetEye()).Length()
		return distI > distJ
	})

	r.subdivide(initialAABB, bounds, surfaceBuffer, primaryShapes, r.Shapes)

	// Pass 2: Shading
	for y := 0; y < tileHeight; y++ {
		for x := 0; x < tileWidth; x++ {
			surface := &surfaceBuffer[y][x]
			var finalColor color.RGBA

			if surface.Hit {
				avgR := surface.rgba[0] / surface.rgba[3]
				avgG := surface.rgba[1] / surface.rgba[3]
				avgB := surface.rgba[2] / surface.rgba[3]
				finalColor = color.RGBA{R: uint8(avgR), G: uint8(avgG), B: uint8(avgB), A: 255}
				finalColor = shading.ApplyAtmosphere(finalColor, surface.Depth, r.Atmosphere)
			} else {
				finalColor = shading.ApplyAtmosphere(r.bgColor, r.Far, r.Atmosphere)
			}

			if surface.VolumeSamples > 0 {
				avgDensity := surface.volDensity / float64(surface.VolumeSamples)
				blendFactor := gomath.Min(1.0, avgDensity*surface.volInterval)
				volColor := color.RGBA{
					R: uint8(float64(surface.volColor.R) / float64(surface.VolumeSamples)),
					G: uint8(float64(surface.volColor.G) / float64(surface.VolumeSamples)),
					B: uint8(float64(surface.volColor.B) / float64(surface.VolumeSamples)),
				}
				finalColor.R = uint8(float64(finalColor.R)*(1-blendFactor) + float64(volColor.R)*blendFactor)
				finalColor.G = uint8(float64(finalColor.G)*(1-blendFactor) + float64(volColor.G)*blendFactor)
				finalColor.B = uint8(float64(finalColor.B)*(1-blendFactor) + float64(volColor.B)*blendFactor)
			}

			img.Set(x, y, finalColor)
		}
	}
	return img
}

// subdivide is the core recursive rendering function (Pass 1: Dicing).
func (r *Renderer) subdivide(aabb math.AABB3D, bounds ScreenBounds, surfaceBuffer [][]SurfaceData, primaryShapes []geometry.Shape, fullScene []geometry.Shape) {
	// Don't cull recursively. The primaryShapes list is the definitive set for this tile.
	if len(primaryShapes) == 0 {
		return
	}

	// Base case: If the AABB is small enough, do a fine-grind search for the surface.
	if (aabb.Max.X - aabb.Min.X) < r.MinSize {
		minX, minY := int(aabb.Min.X*float64(r.Width)), int(aabb.Min.Y*float64(r.Height))
		maxX, maxY := int(aabb.Max.X*float64(r.Width)), int(aabb.Max.Y*float64(r.Height))

		prng := math.NewXorShift32(uint32(minX*r.Width + minY))
		shutterOpen, shutterClose := r.Camera.GetShutterOpen(), r.Camera.GetShutterClose()

		for py := minY; py <= maxY; py++ {
			for px := minX; px <= maxX; px++ {
				if px >= bounds.MinX && px < bounds.MaxX && py >= bounds.MinY && py < bounds.MaxY {
					tileX, tileY := px-bounds.MinX, py-bounds.MinY
					surface := &surfaceBuffer[tileY][tileX]

					for sample := 0; sample < r.MotionBlurSamples; sample++ {
						// Jitter time for motion blur.
						t := shutterOpen
						if shutterClose > shutterOpen {
							t = shutterOpen + (shutterClose-shutterOpen)*(float64(sample)+prng.NextFloat64())/float64(r.MotionBlurSamples)
						}
						// Jitter for anti-aliasing.
						sx := (float64(px) + prng.NextFloat64()) / float64(r.Width)
						sy := (float64(py) + prng.NextFloat64()) / float64(r.Height)

						// Painter's algorithm for solid surfaces first
						minDepth := 1e9
						var closestShape geometry.Shape
						var hitPoint math.Point3D

						for _, s := range primaryShapes {
							if s.IsVolumetric() {
								continue
							}
							var tempShape geometry.Shape
							switch v := s.(type) {
							case *geometry.Sphere3D:
								t := *v
								tempShape = &t
							default:
								tempShape = v
							}
							s.AtTime(t, tempShape)
							for i := 0; i < 8; i++ {
								zSample := aabb.Min.Z + (aabb.Max.Z-aabb.Min.Z)*(float64(i)/7.0)
								if zSample >= minDepth {
									continue
								}
								worldP := r.Camera.Project(sx, sy, zSample)
								if tempShape.Contains(worldP) {
									minDepth = zSample
									closestShape = tempShape
									hitPoint = worldP
									break // Found the surface for this specific shape
								}
							}
						}
						// Now handle volumetrics between camera and the solid surface
						volInterval := (aabb.Max.Z - aabb.Min.Z) / 7.0
						for _, s := range primaryShapes {
							if !s.IsVolumetric() {
								continue
							}
							var tempShape geometry.Shape
							switch v := s.(type) {
							case *geometry.VolumeBox:
								t := *v
								tempShape = &t
							default:
								tempShape = v
							}
							s.AtTime(t, tempShape)
							if volShape, ok := tempShape.(geometry.VolumetricShape); ok {
								for i := 0; i < 8; i++ {
									zSample := aabb.Min.Z + (aabb.Max.Z-aabb.Min.Z)*(float64(i)/7.0)
									if zSample >= minDepth {
										continue // Don't sample volumes behind solid surfaces
									}
									worldP := r.Camera.Project(sx, sy, zSample)
									if volShape.Contains(worldP) {
										surface.volInterval += volInterval
										surface.volDensity += volShape.GetDensity()
										c := volShape.GetColor()
										surface.volColor.R += c.R
										surface.volColor.G += c.G
										surface.volColor.B += c.B
										surface.VolumeSamples++
									}
								}
							}
						}
						if closestShape != nil {
							surface.Hit = true
							surface.Depth = (surface.Depth*surface.rgba[3] + minDepth) / (surface.rgba[3] + 1)
							surface.time = t
							normal := closestShape.NormalAtPoint(hitPoint)

							shadedColor := shading.ShadedColor(hitPoint, math.Normal3D{X: normal.X, Y: normal.Y, Z: normal.Z}, r.Camera.GetEye(), r.Light, closestShape, fullScene, t)
							surface.rgba[0] += float64(shadedColor.R)
							surface.rgba[1] += float64(shadedColor.G)
							surface.rgba[2] += float64(shadedColor.B)
							surface.rgba[3]++
						}
					}
				}
			}
		}
		return
	}

	// Recursive step: Subdivide the AABB into 8 smaller boxes.
	mx, my, mz := (aabb.Min.X+aabb.Max.X)/2, (aabb.Min.Y+aabb.Max.Y)/2, (aabb.Min.Z+aabb.Max.Z)/2
	xs := [3]float64{aabb.Min.X, mx, aabb.Max.X}
	ys := [3]float64{aabb.Min.Y, my, aabb.Max.Y}
	zs := [3]float64{aabb.Min.Z, mz, aabb.Max.Z}

	for zi := 0; zi < 2; zi++ {
		for xi := 0; xi < 2; xi++ {
			for yi := 0; yi < 2; yi++ {
				r.subdivide(math.AABB3D{
					Min: math.Point3D{X: xs[xi], Y: ys[yi], Z: zs[zi]},
					Max: math.Point3D{X: xs[xi+1], Y: ys[yi+1], Z: zs[zi+1]},
				}, bounds, surfaceBuffer, primaryShapes, fullScene)
			}
		}
	}
}
