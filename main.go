package main

import (
	"fmt"
	"image"
	"image/color"
	"image/png"
	"math"
	"os"
//	"sync"
)

// --- Basic Types ---

type Point3D struct{ x, y, z float64 }
type AABB3D struct{ min, max Point3D }
type Normal3D struct{ x, y, z float64 }

func normalize(p Point3D) Point3D {
	d := math.Sqrt(p.x*p.x + p.y*p.y + p.z*p.z)
	if d == 0 { return p }
	return Point3D{p.x / d, p.y / d, p.z / d}
}

func cross(a, b Point3D) Point3D {
	return Point3D{a.y*b.z - a.z*b.y, a.z*b.x - a.x*b.z, a.x*b.y - a.y*b.x}
}

// --- Camera System ---

type Camera interface {
	Project(sx, sy, z float64) Point3D
}

type PerspectiveCamera struct {
	position, forward, right, up Point3D
	fovScale, aspect             float64
}

func NewLookAtCamera(pos, target, up Point3D, fov, aspect float64) *PerspectiveCamera {
	f := normalize(Point3D{target.x - pos.x, target.y - pos.y, target.z - pos.z})
	r := normalize(cross(f, up))
	u := cross(r, f)
	return &PerspectiveCamera{
		position: pos, forward: f, right: r, up: u,
		fovScale: math.Tan(fov * 0.5 * math.Pi / 180.0),
		aspect:   aspect,
	}
}

func (c *PerspectiveCamera) Project(sx, sy, z float64) Point3D {
	nx := (2.0*sx - 1.0) * c.aspect * c.fovScale * z
	ny := (1.0 - 2.0*sy) * c.fovScale * z
	return Point3D{
		x: c.position.x + c.forward.x*z + c.right.x*nx + c.up.x*ny,
		y: c.position.y + c.forward.y*z + c.right.y*nx + c.up.y*ny,
		z: c.position.z + c.forward.z*z + c.right.z*nx + c.up.z*ny,
	}
}

// --- Shapes ---

type Shape interface {
	Contains(p Point3D) bool
	Intersects(aabb AABB3D) bool
	NormalAtPoint(p Point3D) Normal3D
	GetColor() color.RGBA
}

type Plane3D struct {
	point  Point3D
	normal Normal3D
	color  color.RGBA
}

func (pl Plane3D) Contains(p Point3D) bool {
	v := Point3D{p.x - pl.point.x, p.y - pl.point.y, p.z - pl.point.z}
	// Add a tiny epsilon (0.0001) to reduce sampling noise
	return (v.x*pl.normal.x + v.y*pl.normal.y + v.z*pl.normal.z) <= 0.0001
}

func (pl Plane3D) Intersects(aabb AABB3D) bool {
	// Check if any of the 8 corners are on opposite sides of the plane
	// This is much more robust for volume rendering
	points := [8]Point3D{
		{aabb.min.x, aabb.min.y, aabb.min.z}, {aabb.max.x, aabb.min.y, aabb.min.z},
		{aabb.min.x, aabb.max.y, aabb.min.z}, {aabb.max.x, aabb.max.y, aabb.min.z},
		{aabb.min.x, aabb.min.y, aabb.max.z}, {aabb.max.x, aabb.min.y, aabb.max.z},
		{aabb.min.x, aabb.max.y, aabb.max.z}, {aabb.max.x, aabb.max.y, aabb.max.z},
	}
	
	hasIn, hasOut := false, false
	for _, p := range points {
		v := Point3D{p.x - pl.point.x, p.y - pl.point.y, p.z - pl.point.z}
		dot := v.x*pl.normal.x + v.y*pl.normal.y + v.z*pl.normal.z
		if dot <= 0 { hasIn = true } else { hasOut = true }
		if hasIn && hasOut { return true }
	}
	return hasIn // Returns true if the volume is partially or fully "under" the plane
}

func (pl Plane3D) NormalAtPoint(p Point3D) Normal3D { return pl.normal }
func (pl Plane3D) GetColor() color.RGBA { return pl.color }

type Sphere3D struct {
	center Point3D
	radius float64
	color  color.RGBA
}

func (s Sphere3D) Contains(p Point3D) bool {
	dx, dy, dz := p.x-s.center.x, p.y-s.center.y, p.z-s.center.z
	return dx*dx+dy*dy+dz*dz <= s.radius*s.radius
}

func (s Sphere3D) Intersects(aabb AABB3D) bool {
	closestX := math.Max(aabb.min.x, math.Min(s.center.x, aabb.max.x))
	closestY := math.Max(aabb.min.y, math.Min(s.center.y, aabb.max.y))
	closestZ := math.Max(aabb.min.z, math.Min(s.center.z, aabb.max.z))
	dx, dy, dz := closestX-s.center.x, closestY-s.center.y, closestZ-s.center.z
	return dx*dx+dy*dy+dz*dz <= s.radius*s.radius
}

func (s Sphere3D) NormalAtPoint(p Point3D) Normal3D {
	dx, dy, dz := p.x-s.center.x, p.y-s.center.y, p.z-s.center.z
	len := math.Sqrt(dx*dx + dy*dy + dz*dz)
	return Normal3D{dx / len, dy / len, dz / len}
}

func (s Sphere3D) GetColor() color.RGBA { return s.color }

// --- Shading ---

type Light struct {
	Position  Point3D
	Intensity float64
}

// 1. Updated ShadedColor with shadow logic
func ShadedColor(p Point3D, n Normal3D, l Light, base color.RGBA, shapes []Shape) color.RGBA {
	lightVec := Point3D{l.Position.x - p.x, l.Position.y - p.y, l.Position.z - p.z}
	lightDir := normalize(lightVec)
	
	// Shadow Check: Offset the starting point slightly to avoid self-intersection
	shadowBias := 0.01
	checkP := Point3D{p.x + n.x*shadowBias, p.y + n.y*shadowBias, p.z + n.z*shadowBias}
	
	inShadow := false
	// Trace towards the light (cheap volumetric shadow check)
	for t := 0.1; t < 5.0; t += 0.2 {
		sampleP := Point3D{checkP.x + lightDir.x*t, checkP.y + lightDir.y*t, checkP.z + lightDir.z*t}
		for _, s := range shapes {
			// Ignore planes for this specific simple shadow check to prevent infinite floor shadows
			if _, ok := s.(Plane3D); ok { continue } 
			if s.Contains(sampleP) {
				inShadow = true
				break
			}
		}
		if inShadow { break }
	}

	dot := n.x*lightDir.x + n.y*lightDir.y + n.z*lightDir.z
	factor := math.Max(0.15, dot) * l.Intensity
	
	if inShadow {
		factor = 0.15 // Ambient only
	}

	return color.RGBA{
		uint8(math.Min(255, float64(base.R)*factor)),
		uint8(math.Min(255, float64(base.G)*factor)),
		uint8(math.Min(255, float64(base.B)*factor)),
		255,
	}
}

// 2. Updated call inside the render function
// Ensure you pass 'shapes' as the last argument
/*
if hitShape.Contains(worldP) {
    norm := hitShape.NormalAtPoint(worldP)
    // ADD 'shapes' HERE:
    img.Set(px, py, ShadedColor(worldP, norm, light, hitShape.GetColor(), shapes))
}
*/
func getWorldAABB(aabb AABB3D, cam *PerspectiveCamera) AABB3D {
	corners := []Point3D{
		cam.Project(aabb.min.x, aabb.min.y, aabb.min.z),
		cam.Project(aabb.max.x, aabb.min.y, aabb.min.z),
		cam.Project(aabb.min.x, aabb.max.y, aabb.min.z),
		cam.Project(aabb.max.x, aabb.max.y, aabb.min.z),
		cam.Project(aabb.min.x, aabb.min.y, aabb.max.z),
		cam.Project(aabb.max.x, aabb.min.y, aabb.max.z),
		cam.Project(aabb.min.x, aabb.max.y, aabb.max.z),
		cam.Project(aabb.max.x, aabb.max.y, aabb.max.z),
	}
	minP := Point3D{math.Inf(1), math.Inf(1), math.Inf(1)}
	maxP := Point3D{math.Inf(-1), math.Inf(-1), math.Inf(-1)}
	for _, p := range corners {
		minP.x, minP.y, minP.z = math.Min(minP.x, p.x), math.Min(minP.y, p.y), math.Min(minP.z, p.z)
		maxP.x, maxP.y, maxP.z = math.Max(maxP.x, p.x), math.Max(maxP.y, p.y), math.Max(maxP.z, p.z)
	}
	return AABB3D{minP, maxP}
}

// --- Renderer ---

func render(shapes []Shape, cam *PerspectiveCamera, light Light, width, height int, minSize float64) *image.RGBA {
	img := image.NewRGBA(image.Rect(0, 0, width, height))
	bg := color.RGBA{30, 30, 35, 255}
	for i := 0; i < len(img.Pix); i += 4 {
		img.Pix[i], img.Pix[i+1], img.Pix[i+2], img.Pix[i+3] = bg.R, bg.G, bg.B, bg.A
	}

	var drawAABB func(aabb AABB3D)
	drawAABB = func(aabb AABB3D) {
		worldAABB := getWorldAABB(aabb, cam)
		
		var hitShape Shape
		anyHit := false
		for _, s := range shapes {
			if s.Intersects(worldAABB) {
				anyHit = true
				hitShape = s
				break
			}
		}
		if !anyHit { return }

		if (aabb.max.x - aabb.min.x) < minSize {
			minX, minY := int(aabb.min.x*float64(width)), int(aabb.min.y*float64(height))
			maxX, maxY := int(aabb.max.x*float64(width)), int(aabb.max.y*float64(height))

			for py := minY; py <= maxY; py++ {
				for px := minX; px <= maxX; px++ {
					if px >= 0 && px < width && py >= 0 && py < height {
						// Check if pixel is still background
						curr := img.RGBAAt(px, py)
						if curr == bg {
							sx, sy := float64(px)/float64(width), float64(py)/float64(height)
							zMid := (aabb.min.z + aabb.max.z) / 2
							worldP := cam.Project(sx, sy, zMid)
							
							if hitShape.Contains(worldP) {
								norm := hitShape.NormalAtPoint(worldP)
								img.Set(px, py, ShadedColor(worldP, norm, light, hitShape.GetColor(), shapes))
							}
						}
					}
				}
			}
			return
		}

		mx, my, mz := (aabb.min.x+aabb.max.x)/2, (aabb.min.y+aabb.max.y)/2, (aabb.min.z+aabb.max.z)/2
		xs := [3]float64{aabb.min.x, mx, aabb.max.x}
		ys := [3]float64{aabb.min.y, my, aabb.max.y}
		zs := [3]float64{aabb.min.z, mz, aabb.max.z}

		// STRATEGY: Loop Z first (inner bit) to ensure Front-to-Back order
		for zi := 0; zi < 2; zi++ {
			for xi := 0; xi < 2; xi++ {
				for yi := 0; yi < 2; yi++ {
					drawAABB(AABB3D{
						min: Point3D{xs[xi], ys[yi], zs[zi]},
						max: Point3D{xs[xi+1], ys[yi+1], zs[zi+1]},
					})
				}
			}
		}
	}

	drawAABB(AABB3D{Point3D{0, 0, 0.1}, Point3D{1, 1, 15.0}})
	return img
}


// func renderParallel(shapes []Shape, cam *PerspectiveCamera, light Light, width, height int, minSize float64) *image.RGBA {
//     img := image.NewRGBA(image.Rect(0, 0, width, height))
//     var wg sync.WaitGroup
//     numWorkers := 8 // Or use runtime.NumCPU()

//     for i := 0; i < numWorkers; i++ {
//         wg.Add(1)
//         go func(workerID int) {
//             defer wg.Done()
//             // Define a sub-region of the initial Screen-AABB
//             startY := float64(workerID) / float64(numWorkers)
//             endY := float64(workerID+1) / float64(numWorkers)
            
//             // Call the drawAABB logic for just this vertical strip
//             drawAABB(AABB3D{
//                 min: Point3D{0, startY, 0.1}, 
//                 max: Point3D{1, endY, 15.0},
//             })
//         }(i)
//     }
//     wg.Wait()
//     return img
// }

func main() {
	scene := []Shape{
		Sphere3D{Point3D{0, 0, 0}, 1.0, color.RGBA{255, 80, 80, 255}},
		Sphere3D{Point3D{1.2, 0.5, -0.5}, 0.5, color.RGBA{80, 255, 80, 255}},
		Plane3D{
			point:  Point3D{0, -1.0, 0}, 
			normal: Normal3D{0, 1, 0}, 
			color:  color.RGBA{100, 100, 100, 255},
		},
	}

	cam := NewLookAtCamera(
		Point3D{4, 3, 6},   // Eye
		Point3D{0, 0, 0},   // Target
		Point3D{0, 1, 0},   // Up
		45.0, 1.0,
	)

	light := Light{Position: Point3D{10, 10, 10}, Intensity: 1.3}

	fmt.Println("Rendering...")
	img := render(scene, cam, light, 512, 512, 0.004)
	
	f, _ := os.Create("render.png")
	png.Encode(f, img)
	fmt.Println("Saved to render.png")
}
