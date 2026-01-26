package geometry

import "grinder/pkg/math"

// Mesh defines the raw topology
// type Mesh struct {
// 	Vertices []math.Point3D
// 	Faces    [][]int
// }

func CreateCubeMesh(center math.Point3D, radius float64) *Mesh {
	r := radius
	cx, cy, cz := center.X, center.Y, center.Z

	// Exactly 8 vertices - no more, no less
	vertices := []math.Point3D{
		{X: cx - r, Y: cy - r, Z: cz - r}, // 0: Left-Bottom-Back
		{X: cx + r, Y: cy - r, Z: cz - r}, // 1: Right-Bottom-Back
		{X: cx + r, Y: cy + r, Z: cz - r}, // 2: Right-Top-Back
		{X: cx - r, Y: cy + r, Z: cz - r}, // 3: Left-Top-Back
		{X: cx - r, Y: cy - r, Z: cz + r}, // 4: Left-Bottom-Front
		{X: cx + r, Y: cy - r, Z: cz + r}, // 5: Right-Bottom-Front
		{X: cx + r, Y: cy + r, Z: cz + r}, // 6: Right-Top-Front
		{X: cx - r, Y: cy + r, Z: cz + r}, // 7: Left-Top-Front
	}

	// Define 6 faces using the indices above (Counter-Clockwise)
	faces := [][]int{
		{0, 3, 2, 1}, // Back
		{4, 5, 6, 7}, // Front
		{0, 1, 5, 4}, // Bottom
		{3, 7, 6, 2}, // Top
		{0, 4, 7, 3}, // Left
		{1, 2, 6, 5}, // Right
	}

	return &Mesh{Vertices: vertices, Faces: faces}
}
