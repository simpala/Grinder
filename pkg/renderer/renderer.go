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
	P     math.Point3D
	N     math.Normal3D
	S     geometry.Shape
	Depth float64
	Hit   bool
}

// Renderer is a configurable rendering engine.
type Renderer struct {
	Camera     camera.Camera
	Shapes     []geometry.Shape
	Light      shading.Light
	Width      int
	Height     int
	MinSize    float64
	bgColor    color.RGBA
	Near       float64
	Far        float64
	Atmosphere shading.AtmosphereConfig
}

// NewRenderer creates a new renderer with the given configuration.
func NewRenderer(cam camera.Camera, shapes []geometry.Shape, light shading.Light, width, height int, minSize, near, far float64, atmos shading.AtmosphereConfig) *Renderer {
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

// Render performs the recursive subdivision rendering for a specific screen area.
// It returns a new image containing only the rendered tile.
// func (r *Renderer) Render(bounds ScreenBounds) *image.RGBA {
// 	tileWidth := bounds.MaxX - bounds.MinX
// 	tileHeight := bounds.MaxY - bounds.MinY
// 	img := image.NewRGBA(image.Rect(0, 0, tileWidth, tileHeight))
// 	surfaceBuffer := make([][]SurfaceData, tileHeight)
// 	for i := range surfaceBuffer {
// 		surfaceBuffer[i] = make([]SurfaceData, tileWidth)
// 	}

// 	// Initial AABB for the specified screen bounds
// 	initialAABB := math.AABB3D{
// 		Min: math.Point3D{X: float64(bounds.MinX) / float64(r.Width), Y: float64(bounds.MinY) / float64(r.Height), Z: r.Near},
// 		Max: math.Point3D{X: float64(bounds.MaxX) / float64(r.Width), Y: float64(bounds.MaxY) / float64(r.Height), Z: r.Far},
// 	}

// 	// Do not cull. AABB culling is incorrect for rotated shapes.
// 	// The painter's algorithm with depth testing is the ground truth.
// 	primaryShapes := make([]geometry.Shape, len(r.Shapes))
// 	copy(primaryShapes, r.Shapes)

// 	// Sort shapes from Furthest to Closest relative to the camera
// 	sort.Slice(primaryShapes, func(i, j int) bool {
// 		distI := primaryShapes[i].GetCenter().Sub(r.Camera.GetEye()).Length()
// 		distJ := primaryShapes[j].GetCenter().Sub(r.Camera.GetEye()).Length()
// 		return distI > distJ
// 	})

// 	r.subdivide(initialAABB, bounds, surfaceBuffer, primaryShapes, r.Shapes)

// 	// Pass 2: The Final Overdraw (Shading & SSAA)
// 	prng := math.NewXorShift32(uint32(bounds.MinX*r.Width + bounds.MinY))
// 	for y := 0; y < tileHeight; y++ {
// 		for x := 0; x < tileWidth; x++ {
// 			surface := surfaceBuffer[y][x]
// 			if surface.Hit {
// 				var rTotal, gTotal, bTotal float64
// 				samples := 9

// 				for i := 0; i < samples; i++ {
// 					// Jittered sample point for SSAA
// 					sx := (float64(bounds.MinX+x) + prng.NextFloat64()) / float64(r.Width)
// 					sy := (float64(bounds.MinY+y) + prng.NextFloat64()) / float64(r.Height)

// 					// We use the original hit point's depth (Z) for the sample ray
// 					worldP := r.Camera.Project(sx, sy, surface.Depth)

// 					// Jitter the light position for soft shadows
// 					var jitteredLight shading.Light
// 					if r.Light.Radius > 0 {
// 						var offsetX, offsetY, offsetZ float64
// 						// Rejection sampling for a uniform spherical distribution
// 						for {
// 							offsetX = (prng.NextFloat64()*2 - 1) * r.Light.Radius
// 							offsetY = (prng.NextFloat64()*2 - 1) * r.Light.Radius
// 							offsetZ = (prng.NextFloat64()*2 - 1) * r.Light.Radius
// 							if offsetX*offsetX+offsetY*offsetY+offsetZ*offsetZ <= r.Light.Radius*r.Light.Radius {
// 								break
// 							}
// 						}
// 						jitteredLight = shading.Light{
// 							Position:  r.Light.Position.Add(math.Point3D{X: offsetX, Y: offsetY, Z: offsetZ}),
// 							Intensity: r.Light.Intensity,
// 							Radius:    r.Light.Radius,
// 						}
// 					} else {
// 						jitteredLight = r.Light
// 					}

// 					// Use the pre-calculated normal from the Dicer pass
// 					finalColor := shading.ShadedColor(worldP, surface.N, r.Camera.GetEye(), jitteredLight, surface.S, r.Shapes)
// 					rTotal += float64(finalColor.R)
// 					gTotal += float64(finalColor.G)
// 					bTotal += float64(finalColor.B)
// 				}

// 				img.Set(x, y, color.RGBA{
// 					R: uint8(rTotal / float64(samples)),
// 					G: uint8(gTotal / float64(samples)),
// 					B: uint8(bTotal / float64(samples)),
// 					A: 255,
// 				})
// 			} else {
// 				img.Set(x, y, r.bgColor)
// 			}
// 		}
// 	}

//		return img
//	}
func (r *Renderer) Render(bounds ScreenBounds) *image.RGBA {
	r.FitDepthPlanes()
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

	// Pass 2: Shading with Stratified Light Sampling
	prng := math.NewXorShift32(uint32(bounds.MinX*r.Width + bounds.MinY))

	// Restored Pixel Loops
	for y := 0; y < tileHeight; y++ {
		for x := 0; x < tileWidth; x++ {
			surface := surfaceBuffer[y][x]

			if !surface.Hit {
				img.Set(x, y, r.bgColor)
				continue
			}

			var rTotal, gTotal, bTotal float64
			gridSize := int(gomath.Sqrt(float64(r.Light.Samples)))
			if gridSize < 1 {
				gridSize = 1
			}
			totalSamples := float64(gridSize * gridSize)
			// Calculate light orientation ONCE per pixel to save CPU
			pixelCenterP := r.Camera.Project(float64(bounds.MinX+x)/float64(r.Width), float64(bounds.MinY+y)/float64(r.Height), surface.Depth)
			lightVec := r.Light.Position.Sub(pixelCenterP)
			lightDir := lightVec.Normalize()

			// Standard 'LookAt' math to find vectors perpendicular to the light direction
			var up math.Point3D
			if gomath.Abs(lightDir.Y) < 0.9 {
				up = math.Point3D{X: 0, Y: 1, Z: 0}
			} else {
				up = math.Point3D{X: 1, Y: 0, Z: 0}
			}
			right := lightDir.Cross(up).Normalize()
			vUp := lightDir.Cross(right).Normalize()
			// Stratified Loops (Inside each pixel)
			for gy := 0; gy < gridSize; gy++ {
				for gx := 0; gx < gridSize; gx++ {
					// SSAA Jitter (Pixel)
					sx := (float64(bounds.MinX+x) + prng.NextFloat64()) / float64(r.Width)
					sy := (float64(bounds.MinY+y) + prng.NextFloat64()) / float64(r.Height)
					worldP := r.Camera.Project(sx, sy, surface.Depth)

					var jitteredLight shading.Light
					if r.Light.Radius > 0 {
						u := (float64(gx) + prng.NextFloat64()) / float64(gridSize)
						v := (float64(gy) + prng.NextFloat64()) / float64(gridSize)

						// Offset within the light's local "facing" plane
						offU := (u*2 - 1) * r.Light.Radius
						offV := (v*2 - 1) * r.Light.Radius

						// New Position: Original + (Right * u) + (Up * v)
						jitteredPos := r.Light.Position.Add(right.Mul(offU)).Add(vUp.Mul(offV))

						jitteredLight = shading.Light{
							Position:  jitteredPos,
							Intensity: r.Light.Intensity,
							Radius:    r.Light.Radius,
						}
					} else {
						jitteredLight = r.Light
					}

					//finalColor := shading.ShadedColor(worldP, surface.N, r.Camera.GetEye(), jitteredLight, surface.S, r.Shapes)
					finalColor := shading.ShadedColor(worldP, surface.N, r.Camera.GetEye(), jitteredLight, surface.S, r.Shapes)

					rTotal += float64(finalColor.R)
					gTotal += float64(finalColor.G)
					bTotal += float64(finalColor.B)
				}
			}

			img.Set(x, y, color.RGBA{
				R: uint8(rTotal / totalSamples),
				G: uint8(gTotal / totalSamples),
				B: uint8(bTotal / totalSamples),
				A: 255,
			})
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

		for py := minY; py <= maxY; py++ {
			for px := minX; px <= maxX; px++ {
				if px >= bounds.MinX && px < bounds.MaxX && py >= bounds.MinY && py < bounds.MaxY {
					tileX, tileY := px-bounds.MinX, py-bounds.MinY
					sx, sy := float64(px)/float64(r.Width), float64(py)/float64(r.Height)

					// Fine-grind search: find the actual surface within this depth slice
					for _, s := range primaryShapes {
						for i := 0; i < 8; i++ {
							zSample := aabb.Min.Z + (aabb.Max.Z-aabb.Min.Z)*(float64(i)/7.0)

							// The Painterly Check:
							// If we already hit something closer, don't even bother sampling this shape.
							if surfaceBuffer[tileY][tileX].Hit && zSample >= surfaceBuffer[tileY][tileX].Depth {
								continue
							}

							worldP := r.Camera.Project(sx, sy, zSample)
							if s.Contains(worldP) {
								surfaceBuffer[tileY][tileX] = SurfaceData{
									P:     worldP,
									N:     s.NormalAtPoint(worldP),
									S:     s,
									Depth: zSample,
									Hit:   true,
								}
								break // Found the surface for this specific shape; move to next shape
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
