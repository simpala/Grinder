package geometry

import (
	"grinder/pkg/math"
	"image/color"
	gomath "math"
	"sort"
)

type BVHNode struct {
	AABB   math.AABB3D
	Left   *BVHNode
	Right  *BVHNode
	Shapes []Shape // Only for leaf nodes
}

type BVH struct {
	Root           *BVHNode
	InfiniteShapes []Shape
}

func NewBVH(shapes []Shape) *BVH {
	if len(shapes) == 0 {
		return &BVH{}
	}

	var finite []Shape
	var infinite []Shape
	for _, s := range shapes {
		aabb := s.GetAABB()
		if gomath.IsInf(aabb.Min.X, -1) || gomath.IsInf(aabb.Max.X, 1) {
			infinite = append(infinite, s)
		} else {
			finite = append(finite, s)
		}
	}

	bvh := &BVH{InfiniteShapes: infinite}

	if len(finite) > 0 {
		// 1. Compute overall scene AABB for finite shapes
		sceneAABB := finite[0].GetAABB()
		for i := 1; i < len(finite); i++ {
			aabb := finite[i].GetAABB()
			sceneAABB.Min.X = gomath.Min(sceneAABB.Min.X, aabb.Min.X)
			sceneAABB.Min.Y = gomath.Min(sceneAABB.Min.Y, aabb.Min.Y)
			sceneAABB.Min.Z = gomath.Min(sceneAABB.Min.Z, aabb.Min.Z)
			sceneAABB.Max.X = gomath.Max(sceneAABB.Max.X, aabb.Max.X)
			sceneAABB.Max.Y = gomath.Max(sceneAABB.Max.Y, aabb.Max.Y)
			sceneAABB.Max.Z = gomath.Max(sceneAABB.Max.Z, aabb.Max.Z)
		}

		// 2. Compute Morton codes for each finite shape
		type shapeWithCode struct {
			shape Shape
			code  uint32
		}
		codedShapes := make([]shapeWithCode, len(finite))
		diag := sceneAABB.Max.Sub(sceneAABB.Min)
		for i, s := range finite {
			center := s.GetCenter()
			// Normalize center to [0, 1]
			nx, ny, nz := 0.5, 0.5, 0.5
			if diag.X > 0 {
				nx = (center.X - sceneAABB.Min.X) / diag.X
			}
			if diag.Y > 0 {
				ny = (center.Y - sceneAABB.Min.Y) / diag.Y
			}
			if diag.Z > 0 {
				nz = (center.Z - sceneAABB.Min.Z) / diag.Z
			}
			codedShapes[i] = shapeWithCode{
				shape: s,
				code:  math.Morton3D(nx, ny, nz),
			}
		}

		// 3. Sort by Morton code
		sort.Slice(codedShapes, func(i, j int) bool {
			return codedShapes[i].code < codedShapes[j].code
		})

		// 4. Build tree recursively
		sortedShapes := make([]Shape, len(finite))
		for i, cs := range codedShapes {
			sortedShapes[i] = cs.shape
		}
		bvh.Root = buildBVH(sortedShapes)
	}

	return bvh
}

func buildBVH(shapes []Shape) *BVHNode {
	if len(shapes) == 0 {
		return nil
	}

	node := &BVHNode{}
	// Compute AABB for all shapes in this node
	node.AABB = shapes[0].GetAABB()
	for i := 1; i < len(shapes); i++ {
		aabb := shapes[i].GetAABB()
		node.AABB.Min.X = gomath.Min(node.AABB.Min.X, aabb.Min.X)
		node.AABB.Min.Y = gomath.Min(node.AABB.Min.Y, aabb.Min.Y)
		node.AABB.Min.Z = gomath.Min(node.AABB.Min.Z, aabb.Min.Z)
		node.AABB.Max.X = gomath.Max(node.AABB.Max.X, aabb.Max.X)
		node.AABB.Max.Y = gomath.Max(node.AABB.Max.Y, aabb.Max.Y)
		node.AABB.Max.Z = gomath.Max(node.AABB.Max.Z, aabb.Max.Z)
	}

	if len(shapes) <= 4 {
		node.Shapes = shapes
		return node
	}

	mid := len(shapes) / 2
	node.Left = buildBVH(shapes[:mid])
	node.Right = buildBVH(shapes[mid:])

	return node
}

// IntersectsShapes returns all shapes in the BVH that might intersect the given AABB.
func (b *BVH) IntersectsShapes(aabb math.AABB3D) []Shape {
	result := append([]Shape{}, b.InfiniteShapes...)
	if b.Root != nil {
		b.Root.intersectsShapes(aabb, &result)
	}
	return result
}

func (node *BVHNode) intersectsShapes(aabb math.AABB3D, result *[]Shape) {
	if !node.AABB.Intersects(aabb) {
		return
	}

	if node.Shapes != nil {
		for _, s := range node.Shapes {
			if s.Intersects(aabb) {
				*result = append(*result, s)
			}
		}
		return
	}

	if node.Left != nil {
		node.Left.intersectsShapes(aabb, result)
	}
	if node.Right != nil {
		node.Right.intersectsShapes(aabb, result)
	}
}

// --- Shape Interface Implementation ---

func (b *BVH) Contains(p math.Point3D, t float64) bool {
	// BVH is a non-renderable acceleration structure.
	return false
}

func (b *BVH) Intersects(aabb math.AABB3D) bool {
	// For infinite shapes, we say it intersects.
	if len(b.InfiniteShapes) > 0 {
		return true
	}
	return b.Root != nil && b.Root.AABB.Intersects(aabb)
}

func (b *BVH) NormalAtPoint(p math.Point3D, t float64) math.Normal3D {
	return math.Normal3D{}
}

func (b *BVH) GetColor() color.RGBA {
	return color.RGBA{}
}

func (b *BVH) GetShininess() float64 { return 0 }

func (b *BVH) GetSpecularIntensity() float64 { return 0 }

func (b *BVH) GetSpecularColor() color.RGBA { return color.RGBA{} }

func (b *BVH) GetAABB() math.AABB3D {
	if b.Root == nil {
		if len(b.InfiniteShapes) > 0 {
			return b.InfiniteShapes[0].GetAABB() // Returns infinite AABB
		}
		return math.AABB3D{}
	}
	// Technically should merge with infinite AABB but that's already infinite.
	return b.Root.AABB
}

func (b *BVH) GetCenter() math.Point3D {
	return b.GetAABB().Center()
}

func (b *BVH) IsVolumetric() bool {
	return false
}
