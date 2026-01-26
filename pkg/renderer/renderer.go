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
	P             math.Point3D
	N             math.Normal3D
	S             geometry.Shape
	TSample       float64 // <--- Ensure this is here!
	Depth         float64
	Hit           bool
	VolumeSamples []VolumeSample
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
	Camera     camera.Camera
	Shapes     []geometry.Shape
	BVH        *geometry.BVH
	Light      shading.Light
	Width      int
	Height     int
	MinSize    float64
	bgColor    color.RGBA
	Near       float64
	Far        float64
	Atmosphere shading.AtmosphereConfig
	Shutter    float64 // Add this!
}

// NewRenderer creates a new renderer with the given configuration.
func NewRenderer(cam camera.Camera, shapes []geometry.Shape, light shading.Light, width, height int, minSize, near, far float64, atmos shading.AtmosphereConfig, shutter float64) *Renderer {
	if near == 0 {
		near = 0.1
	}
	if far == 0 {
		far = 50.0
	}
	bvh := geometry.NewBVH(shapes)
	// Add BVH to the scene shapes list so it can be used by other systems as a Shape
	allShapes := append([]geometry.Shape{}, shapes...)
	allShapes = append(allShapes, bvh)

	return &Renderer{
		Camera:     cam,
		Shutter:    shutter,
		Atmosphere: atmos,
		Shapes:     allShapes,
		BVH:        bvh,
		Light:      light,
		Width:      width,
		Height:     height,
		MinSize:    minSize,
		bgColor:    color.RGBA{30, 30, 35, 255},
		Near:       near,
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

func (r *Renderer) computeTileAABB(bounds ScreenBounds) math.AABB3D {
	sx := []float64{float64(bounds.MinX) / float64(r.Width), float64(bounds.MaxX) / float64(r.Width)}
	sy := []float64{float64(bounds.MinY) / float64(r.Height), float64(bounds.MaxY) / float64(r.Height)}
	sz := []float64{r.Near, r.Far}

	first := true
	var result math.AABB3D
	for _, z := range sz {
		for _, y := range sy {
			for _, x := range sx {
				p := r.Camera.Project(x, y, z)
				if first {
					result = math.AABB3D{Min: p, Max: p}
					first = false
				} else {
					result = result.Expand(p)
				}
			}
		}
	}
	return result
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

	tileAABB := r.computeTileAABB(bounds)
	primaryShapes := r.BVH.IntersectsShapes(tileAABB)

	sort.Slice(primaryShapes, func(i, j int) bool {
		distI := primaryShapes[i].GetCenter().Sub(r.Camera.GetEye()).Length()
		distJ := primaryShapes[j].GetCenter().Sub(r.Camera.GetEye()).Length()
		return distI > distJ
	})

	r.subdivide(initialAABB, bounds, surfaceBuffer, primaryShapes, r.Shapes)

	// Pass 2: Shading with Stratified Light Sampling
	prng := math.NewXorShift32(uint32(bounds.MinX*r.Width + bounds.MinY))

	// Restored Pixel Loops
	for y := 0; y < tileHeight; y++ {
		for x := 0; x < tileWidth; x++ {
			surface := surfaceBuffer[y][x]

			// 1. Determine the background color (either a solid surface or the scene background)
			var bgColor color.RGBA
			if surface.Hit {
				var rTotal, gTotal, bTotal float64
				gridSize := int(gomath.Sqrt(float64(r.Light.Samples)))
				if gridSize < 1 {
					gridSize = 1
				}
				totalSamples := float64(gridSize * gridSize)

				lightVec := r.Light.Position.Sub(surface.P)
				lightDir := lightVec.Normalize()

				var up math.Point3D
				if gomath.Abs(lightDir.Y) < 0.9 {
					up = math.Point3D{X: 0, Y: 1, Z: 0}
				} else {
					up = math.Point3D{X: 1, Y: 0, Z: 0}
				}
				right := lightDir.Cross(up).Normalize()
				vUp := lightDir.Cross(right).Normalize()

				for gy := 0; gy < gridSize; gy++ {
					for gx := 0; gx < gridSize; gx++ {
						sx := (float64(bounds.MinX+x) + prng.NextFloat64()) / float64(r.Width)
						sy := (float64(bounds.MinY+y) + prng.NextFloat64()) / float64(r.Height)
						worldP := r.Camera.Project(sx, sy, surface.Depth)

						var jitteredLight shading.Light
						if r.Light.Radius > 0 {
							u := (float64(gx) + prng.NextFloat64()) / float64(gridSize)
							v := (float64(gy) + prng.NextFloat64()) / float64(gridSize)
							offU := (u*2 - 1) * r.Light.Radius
							offV := (v*2 - 1) * r.Light.Radius
							jitteredPos := r.Light.Position.Add(right.Mul(offU)).Add(vUp.Mul(offV))

							jitteredLight = shading.Light{
								Position:  jitteredPos,
								Intensity: r.Light.Intensity,
								Radius:    r.Light.Radius,
							}
						} else {
							jitteredLight = r.Light
						}

						shadedColor := shading.ShadedColor(worldP, surface.N, r.Camera.GetEye(), jitteredLight, surface.S, r.Shapes, surface.TSample)
						rTotal += float64(shadedColor.R)
						gTotal += float64(shadedColor.G)
						bTotal += float64(shadedColor.B)
					}
				}

				surfaceColor := color.RGBA{
					R: uint8(rTotal / totalSamples),
					G: uint8(gTotal / totalSamples),
					B: uint8(bTotal / totalSamples),
					A: 255,
				}
				bgColor = shading.ApplyAtmosphere(surfaceColor, surface.Depth, r.Atmosphere)
			} else {
				bgColor = shading.ApplyAtmosphere(r.bgColor, r.Far, r.Atmosphere)
			}

			// 2. Composite Volumetric Samples
			finalColor := bgColor
			if len(surface.VolumeSamples) > 0 {
				for _, sample := range surface.VolumeSamples {
					// Only composite samples that are in front of the solid surface
					if !surface.Hit || sample.Depth < surface.Depth {
						volColor := sample.Shape.GetColor()
						density := sample.Shape.GetDensity()
						blendFactor := gomath.Min(1.0, density*sample.Interval)

						finalColor.R = uint8(float64(finalColor.R)*(1-blendFactor) + float64(volColor.R)*blendFactor)
						finalColor.G = uint8(float64(finalColor.G)*(1-blendFactor) + float64(volColor.G)*blendFactor)
						finalColor.B = uint8(float64(finalColor.B)*(1-blendFactor) + float64(volColor.B)*blendFactor)
					}
				}
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

		blockSeed := uint32(aabb.Min.X*10000 + aabb.Min.Y*1000 + aabb.Min.Z*100)
		prng := math.NewXorShift32(blockSeed)

		for py := minY; py <= maxY; py++ {
			for px := minX; px <= maxX; px++ {
				if px >= bounds.MinX && px < bounds.MaxX && py >= bounds.MinY && py < bounds.MaxY {
					tileX, tileY := px-bounds.MinX, py-bounds.MinY
					jitterX := (prng.NextFloat64() - 0.5) / float64(r.Width)
					jitterY := (prng.NextFloat64() - 0.5) / float64(r.Height)
					sx, sy := (float64(px)/float64(r.Width))+jitterX, (float64(py)/float64(r.Height))+jitterY
					interval := (aabb.Max.Z - aabb.Min.Z) / 7.0
					// Inside the px/py loops, before you iterate over shapes:
					pixelNoise := float64((px*127+py*431)%1000) / 1000.0
					// Every pixel gets a consistent time sample for the whole depth stack
					tSampleForPixel := gomath.Mod(prng.NextFloat64()+pixelNoise, 1.0) * r.Shutter
					//sx, sy := float64(px)/float64(r.Width), float64(py)/float64(r.Height)

					// Fine-grind search: find the actual surface within this depth slice
					for _, s := range primaryShapes {
						// Determine sample count: "The Density Drop"
						// For motion blur, we can use fewer depth samples (e.g., 4 instead of 8)
						// to get the speed-up you wanted, since it's blurred anyway.
						steps := 1
						isMoving := false
						if sphere, ok := s.(geometry.Sphere3D); ok {
							// Use a small epsilon to check for actual motion
							if sphere.Velocity.Length() > 0.001 {
								isMoving = true
							}
						}
						// 3. Scale steps accordingly
						if !isMoving {
							// Static objects get the full 8-16 samples to prevent banding
							steps = 8
						} else {
							// Moving objects use your "90% drop" strategy
							// This is where you get the speed boost!
							steps = 2
						}
						if s.IsVolumetric() {
							// interval := (aabb.Max.Z - aabb.Min.Z) / 7.0
							// // Inside the px/py loops, before you iterate over shapes:
							// pixelNoise := float64((px*127+py*431)%1000) / 1000.0
							// // Every pixel gets a consistent time sample for the whole depth stack
							// tSampleForPixel := gomath.Mod(prng.NextFloat64()+pixelNoise, 1.0) * r.Shutter
							for i := 0; i < steps; i++ {
								tSample := gomath.Mod(prng.NextFloat64()+pixelNoise, 1.0) * r.Shutter
								// Use the consistent pixel time
								zThickness := aabb.Max.Z - aabb.Min.Z
								zJitter := prng.NextFloat64() * (zThickness / float64(steps))
								zSample := aabb.Min.Z + (zThickness * (float64(i) / float64(steps))) + zJitter

								worldP := r.Camera.Project(sx, sy, zSample)
								if s.Contains(worldP, tSample) {
									surfaceBuffer[tileY][tileX].VolumeSamples = append(surfaceBuffer[tileY][tileX].VolumeSamples, VolumeSample{
										Shape:    s.(geometry.VolumetricShape),
										Interval: interval,
										Depth:    zSample,
									})
								}
							}
						} else {
							// // Inside the px/py loops, before you iterate over shapes:
							// pixelNoise := float64((px*127+py*431)%1000) / 1000.0
							// // Every pixel gets a consistent time sample for the whole depth stack
							// tSampleForPixel := gomath.Mod(prng.NextFloat64()+pixelNoise, 1.0) * r.Shutter
							for i := 0; i < steps; i++ {
								// Use the consistent pixel time
								zThickness := aabb.Max.Z - aabb.Min.Z
								zJitter := prng.NextFloat64() * (zThickness / float64(steps))
								zSample := aabb.Min.Z + (zThickness * (float64(i) / float64(steps))) + zJitter

								worldP := r.Camera.Project(sx, sy, zSample)
								// Painterly check
								if surfaceBuffer[tileY][tileX].Hit && zSample >= surfaceBuffer[tileY][tileX].Depth {
									continue
								}

								//worldP := r.Camera.Project(sx, sy, zSample)

								// 2. TEMPORAL CHECK: Use the new time-aware Contains
								if s.Contains(worldP, tSampleForPixel) {
									// Apply thinning for moving objects
									if isMoving && prng.NextFloat64() > 0.2 {
										continue
									}

									// ASSIGN EVERYTHING
									surfaceBuffer[tileY][tileX].P = worldP
									surfaceBuffer[tileY][tileX].N = s.NormalAtPoint(worldP, tSampleForPixel)
									surfaceBuffer[tileY][tileX].S = s
									surfaceBuffer[tileY][tileX].Depth = zSample
									surfaceBuffer[tileY][tileX].Hit = true

									// CRITICAL: Every hit (even the floor) must store the time
									// so the shadow pass knows "when" to check for occluders.
									surfaceBuffer[tileY][tileX].TSample = tSampleForPixel

									break
								}
							}
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
