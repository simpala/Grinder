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

// SurfaceSample stores the geometry info for a single sample.
type SurfaceSample struct {
	P     math.Point3D
	N     math.Normal3D
	S     geometry.Shape
	T     float64 // Time of the sample
	Depth float64
}

// VolumeSample stores data for a single sample within a volume.
type VolumeSample struct {
	Shape    geometry.VolumetricShape
	Interval float64 // The length of the ray segment within the volume
	Depth    float64 // The z-depth of the sample
	Time     float64
}

// PixelData holds all the samples for a single pixel.
type PixelData struct {
	SurfaceSamples []SurfaceSample
	VolumeSamples  []VolumeSample
}

// Renderer is a configurable rendering engine.
// Culling/early out is being held off until later when we have more features as its very tricky to get right and breaks with new feature additions.
type Renderer struct {
	Camera           camera.Camera
	Shapes           []geometry.Shape
	Light            shading.Light
	Width            int
	Height           int
	MinSize          float64
	MotionBlurSamples int
	bgColor          color.RGBA
	Near       float64
	Far        float64
	Atmosphere shading.AtmosphereConfig
}

// NewRenderer creates a new renderer with the given configuration.
func NewRenderer(cam camera.Camera, shapes []geometry.Shape, light shading.Light, width, height int, minSize, near, far float64, motionBlurSamples int, atmos shading.AtmosphereConfig) *Renderer {
	if near == 0 {
		near = 0.1
	}
	if far == 0 {
		far = 50.0
	}
	return &Renderer{
		Camera:     cam,
		Atmosphere: atmos,
		Shapes:     shapes,
		Light:      light,
		Width:      width,
		Height:           height,
		MinSize:          minSize,
		MotionBlurSamples: motionBlurSamples,
		bgColor:          color.RGBA{30, 30, 35, 255},
		Near:             near,
		Far:        far,
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

	pixelBuffer := make([][]PixelData, tileHeight)
	for i := range pixelBuffer {
		pixelBuffer[i] = make([]PixelData, tileWidth)
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

	r.subdivide(initialAABB, bounds, pixelBuffer, primaryShapes, r.Shapes)

	// Pass 2: Shading
	for y := 0; y < tileHeight; y++ {
		for x := 0; x < tileWidth; x++ {
			pixelData := pixelBuffer[y][x]
			var rTotal, gTotal, bTotal float64

			if len(pixelData.SurfaceSamples) > 0 {
				for _, sample := range pixelData.SurfaceSamples {
					camAtT := r.Camera.AtTime(sample.T)
					shadedColor := shading.ShadedColor(sample.P, sample.N, camAtT.GetEye(), r.Light, sample.S, r.Shapes, sample.T)
					rTotal += float64(shadedColor.R)
					gTotal += float64(shadedColor.G)
					bTotal += float64(shadedColor.B)
				}
				numSamples := float64(len(pixelData.SurfaceSamples))
				avgColor := color.RGBA{
					R: uint8(rTotal / numSamples),
					G: uint8(gTotal / numSamples),
					B: uint8(bTotal / numSamples),
					A: 255,
				}
				// For now, just use the depth of the first sample for atmosphere.
				// A more advanced implementation might average depths or handle this differently.
				finalColor := shading.ApplyAtmosphere(avgColor, pixelData.SurfaceSamples[0].Depth, r.Atmosphere)
				img.Set(x, y, finalColor)
			} else {
				// No surface hits, just draw the background
				bgColor := shading.ApplyAtmosphere(r.bgColor, r.Far, r.Atmosphere)
				img.Set(x, y, bgColor)
			}

			// TODO: Composite volumetric samples correctly with averaged surface samples.
			// This is a complex step that requires careful blending. For now, we'll skip it
			// to get the primary motion blur effect working.
		}
	}
	return img
}

// subdivide is the core recursive rendering function (Pass 1: Dicing).
func (r *Renderer) subdivide(aabb math.AABB3D, bounds ScreenBounds, pixelBuffer [][]PixelData, primaryShapes []geometry.Shape, fullScene []geometry.Shape) {
	// Don't cull recursively. The primaryShapes list is the definitive set for this tile.
	if len(primaryShapes) == 0 {
		return
	}

	// Base case: If the AABB is small enough, do a fine-grind search for the surface.
	if (aabb.Max.X - aabb.Min.X) < r.MinSize {
		minX, minY := int(aabb.Min.X*float64(r.Width)), int(aabb.Min.Y*float64(r.Height))
		maxX, maxY := int(aabb.Max.X*float64(r.Width)), int(aabb.Max.Y*float64(r.Height))
		prng := math.NewXorShift32(uint32(minX*r.Width + minY))

		for py := minY; py <= maxY; py++ {
			for px := minX; px <= maxX; px++ {
				if px >= bounds.MinX && px < bounds.MaxX && py >= bounds.MinY && py < bounds.MaxY {
					tileX, tileY := px-bounds.MinX, py-bounds.MinY

					// Super-sample for motion blur
					for sample := 0; sample < r.MotionBlurSamples; sample++ {
						t := prng.NextFloat64() * r.Camera.GetShutter()
						camAtT := r.Camera.AtTime(t)
						sx := (float64(px) + prng.NextFloat64()) / float64(r.Width)
						sy := (float64(py) + prng.NextFloat64()) / float64(r.Height)

						// Fine-grind search: find the actual surface within this depth slice
						var closestHit SurfaceSample
						hitFound := false

						for _, s := range primaryShapes {
							shapeAtT := s.AtTime(t)

							if shapeAtT.IsVolumetric() {
								interval := (aabb.Max.Z - aabb.Min.Z) / 7.0
								for i := 0; i < 8; i++ {
									zSample := aabb.Min.Z + (aabb.Max.Z-aabb.Min.Z)*(float64(i)/7.0)
									worldP := camAtT.Project(sx, sy, zSample)
									if shapeAtT.Contains(worldP, t) {
										pixelBuffer[tileY][tileX].VolumeSamples = append(pixelBuffer[tileY][tileX].VolumeSamples, VolumeSample{
											Shape:    shapeAtT.(geometry.VolumetricShape),
											Interval: interval,
											Depth:    zSample,
											Time:     t,
										})
									}
								}
							} else {
								for i := 0; i < 8; i++ {
									zSample := aabb.Min.Z + (aabb.Max.Z-aabb.Min.Z)*(float64(i)/7.0)

									if hitFound && zSample >= closestHit.Depth {
										continue
									}

									worldP := camAtT.Project(sx, sy, zSample)
									if shapeAtT.Contains(worldP, t) {
										closestHit = SurfaceSample{
											P:     worldP,
											N:     shapeAtT.NormalAtPoint(worldP, t),
											S:     shapeAtT,
											Depth: zSample,
											T:     t,
										}
										hitFound = true
										break // Found closest surface for this shape
									}
								}
							}
						}
						if hitFound {
							pixelBuffer[tileY][tileX].SurfaceSamples = append(pixelBuffer[tileY][tileX].SurfaceSamples, closestHit)
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
				}, bounds, pixelBuffer, primaryShapes, fullScene)
			}
		}
	}
}
