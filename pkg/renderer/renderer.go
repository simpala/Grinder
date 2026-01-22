package renderer

import (
	"grinder/pkg/math"
	"grinder/pkg/camera"
	"grinder/pkg/geometry"
	"grinder/pkg/shading"
	"image"
	"image/color"
	gomath "math"
)

// ScreenBounds defines the rectangular region of the screen to be rendered.
type ScreenBounds struct {
	MinX, MinY, MaxX, MaxY int
}

// Renderer is a configurable rendering engine.
type Renderer struct {
	Camera  camera.Camera
	Shapes  []geometry.Shape
	Light   shading.Light
	Width   int
	Height  int
	MinSize float64
	bgColor color.RGBA
}

// NewRenderer creates a new renderer with the given configuration.
func NewRenderer(cam camera.Camera, shapes []geometry.Shape, light shading.Light, width, height int, minSize float64) *Renderer {
	return &Renderer{
		Camera:  cam,
		Shapes:  shapes,
		Light:   light,
		Width:   width,
		Height:  height,
		MinSize: minSize,
		bgColor: color.RGBA{30, 30, 35, 255},
	}
}

// Render performs the recursive subdivision rendering for a specific screen area.
// It returns a new image containing only the rendered tile.
func (r *Renderer) Render(bounds ScreenBounds) *image.RGBA {
	tileWidth := bounds.MaxX - bounds.MinX
	tileHeight := bounds.MaxY - bounds.MinY
	img := image.NewRGBA(image.Rect(0, 0, tileWidth, tileHeight))

	// Fill background
	for i := 0; i < len(img.Pix); i += 4 {
		img.Pix[i], img.Pix[i+1], img.Pix[i+2], img.Pix[i+3] = r.bgColor.R, r.bgColor.G, r.bgColor.B, r.bgColor.A
	}

	// Initial AABB for the specified screen bounds
	initialAABB := math.AABB3D{
		Min: math.Point3D{X: float64(bounds.MinX) / float64(r.Width), Y: float64(bounds.MinY) / float64(r.Height), Z: 0.1},
		Max: math.Point3D{X: float64(bounds.MaxX) / float64(r.Width), Y: float64(bounds.MaxY) / float64(r.Height), Z: 15.0},
	}

	r.subdivide(initialAABB, bounds, img)

	return img
}

// subdivide is the core recursive rendering function.
func (r *Renderer) subdivide(aabb math.AABB3D, bounds ScreenBounds, img *image.RGBA) {
	worldAABB := r.getWorldAABB(aabb)

	var hitShape geometry.Shape
	anyHit := false
	for _, s := range r.Shapes {
		if s.Intersects(worldAABB) {
			anyHit = true
			hitShape = s // Simplification: Just consider the first hit
			break
		}
	}
	if !anyHit {
		return
	}

	// Base case: If the AABB is small enough, fill the pixels.
// Base case: If the AABB is small enough, fill the pixels.
if (aabb.Max.X - aabb.Min.X) < r.MinSize {
    minX, minY := int(aabb.Min.X*float64(r.Width)), int(aabb.Min.Y*float64(r.Height))
    maxX, maxY := int(aabb.Max.X*float64(r.Width)), int(aabb.Max.Y*float64(r.Height))

    for py := minY; py <= maxY; py++ {
        for px := minX; px <= maxX; px++ {
            if px >= bounds.MinX && px < bounds.MaxX && py >= bounds.MinY && py < bounds.MaxY {
                tileX, tileY := px-bounds.MinX, py-bounds.MinY

                if img.RGBAAt(tileX, tileY) == r.bgColor {
                    sx, sy := float64(px)/float64(r.Width), float64(py)/float64(r.Height)
                    
                    // Fine-grind search: find the actual surface within this depth slice
                    for zi := 0; zi < 4; zi++ { 
                        zSample := aabb.Min.Z + (aabb.Max.Z-aabb.Min.Z)*(float64(zi)/3.0)
                        worldP := r.Camera.Project(sx, sy, zSample)

                        if hitShape.Contains(worldP) {
                            norm := hitShape.NormalAtPoint(worldP)
                            img.Set(tileX, tileY, shading.ShadedColor(worldP, norm, r.Camera.GetEye(), r.Light, hitShape, r.Shapes))
                            break // Found the surface, move to next pixel
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
				}, bounds, img)
			}
		}
	}
}

func (r *Renderer) getWorldAABB(aabb math.AABB3D) math.AABB3D {
	corners := []math.Point3D{
		r.Camera.Project(aabb.Min.X, aabb.Min.Y, aabb.Min.Z),
		r.Camera.Project(aabb.Max.X, aabb.Min.Y, aabb.Min.Z),
		r.Camera.Project(aabb.Min.X, aabb.Max.Y, aabb.Min.Z),
		r.Camera.Project(aabb.Max.X, aabb.Max.Y, aabb.Min.Z),
		r.Camera.Project(aabb.Min.X, aabb.Min.Y, aabb.Max.Z),
		r.Camera.Project(aabb.Max.X, aabb.Min.Y, aabb.Max.Z),
		r.Camera.Project(aabb.Min.X, aabb.Max.Y, aabb.Max.Z),
		r.Camera.Project(aabb.Max.X, aabb.Max.Y, aabb.Max.Z),
	}
	minP := math.Point3D{X: gomath.Inf(1), Y: gomath.Inf(1), Z: gomath.Inf(1)}
	maxP := math.Point3D{X: gomath.Inf(-1), Y: gomath.Inf(-1), Z: gomath.Inf(-1)}
	for _, p := range corners {
		minP.X, minP.Y, minP.Z = gomath.Min(minP.X, p.X), gomath.Min(minP.Y, p.Y), gomath.Min(minP.Z, p.Z)
		maxP.X, maxP.Y, maxP.Z = gomath.Max(maxP.X, p.X), gomath.Max(maxP.Y, p.Y), gomath.Max(maxP.Z, p.Z)
	}
	return math.AABB3D{Min: minP, Max: maxP}
}
