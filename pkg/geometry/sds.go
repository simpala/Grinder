package geometry

import (
	"grinder/pkg/math"
	"image/color"
)

// Mesh defines the topology for subdivision
type Mesh struct {
	Vertices []math.Point3D
	Faces    [][]int
}

// SDSObject satisfies the Shape interface and wraps the subdivided quads
type SDSObject struct {
	Quads             []*BilinearQuad
	AABB              math.AABB3D
	Color             color.RGBA
	Shininess         float64
	SpecularIntensity float64
	SpecularColor     color.RGBA
}

// --- Shape Interface Implementation ---

func (s *SDSObject) Contains(p math.Point3D, t float64) bool {
	if !s.AABB.Contains(p) {
		return false
	}
	for _, q := range s.Quads {
		if q.Contains(p, t) {
			return true
		}
	}
	return false
}

func (s *SDSObject) NormalAtPoint(p math.Point3D, t float64) math.Normal3D {
	// Find the quad closest to the point to get the surface normal
	var bestQuad *BilinearQuad
	minDist := 1e18

	for _, q := range s.Quads {
		if q.Contains(p, t) {
			// Calculate distance from point to quad center to find the closest quad
			quadCenter := q.PositionAt(0.5, 0.5)
			dist := p.Sub(quadCenter).Length()
			if dist < minDist {
				minDist = dist
				bestQuad = q
			}
		}
	}

	if bestQuad != nil {
		return bestQuad.NormalAtPoint(p, t)
	}

	// Fallback if no quad is found
	return math.Normal3D{X: 0, Y: 1, Z: 0}
}
func (s *SDSObject) GetColor() color.RGBA { return s.Color }

func (s *SDSObject) GetAABB() math.AABB3D    { return s.AABB }
func (s *SDSObject) GetCenter() math.Point3D { return s.AABB.Center() }
func (s *SDSObject) GetShininess() float64   { return s.Shininess }
func (s *SDSObject) GetSpecular() (float64, color.RGBA) {
	return s.SpecularIntensity, s.SpecularColor
}
func (s *SDSObject) GetSpecularIntensity() float64 {
	return s.SpecularIntensity
}
func (s *SDSObject) GetSpecularColor() color.RGBA {
	return s.SpecularColor
}
func (s *SDSObject) IsVolumetric() bool {
	return false
}
func (s *SDSObject) Intersects(aabb math.AABB3D) bool {
	return s.AABB.Intersects(aabb)
}

// --- Catmull-Clark Subdivision Logic ---

func (m *Mesh) Subdivide() *Mesh {
	// 1. Face Points
	facePoints := make([]math.Point3D, len(m.Faces))
	for i, face := range m.Faces {
		var sum math.Point3D
		for _, vIdx := range face {
			sum = sum.Add(m.Vertices[vIdx])
		}
		facePoints[i] = sum.Mul(1.0 / float64(len(face)))
	}

	// 2. Topology Mapping
	type EdgeKey struct{ V1, V2 int }
	getEdgeKey := func(v1, v2 int) EdgeKey {
		if v1 < v2 {
			return EdgeKey{v1, v2}
		}
		return EdgeKey{v2, v1}
	}

	edgeToFacePoints := make(map[EdgeKey][]int)
	vertexToFaces := make([][]int, len(m.Vertices))
	vertexToEdges := make(map[int]map[EdgeKey]bool)

	for fIdx, face := range m.Faces {
		for i := 0; i < len(face); i++ {
			v1, v2 := face[i], face[(i+1)%len(face)]
			key := getEdgeKey(v1, v2)
			edgeToFacePoints[key] = append(edgeToFacePoints[key], fIdx)
			vertexToFaces[v1] = append(vertexToFaces[v1], fIdx)
			if vertexToEdges[v1] == nil {
				vertexToEdges[v1] = make(map[EdgeKey]bool)
			}
			vertexToEdges[v1][key] = true
		}
	}

	// 3. Edge Points
	edgePointIndices := make(map[EdgeKey]int)
	newVertices := append([]math.Point3D{}, facePoints...)
	for key, fIndices := range edgeToFacePoints {
		edgePointIndices[key] = len(newVertices)
		sum := m.Vertices[key.V1].Add(m.Vertices[key.V2])
		for _, fIdx := range fIndices {
			sum = sum.Add(facePoints[fIdx])
		}
		newVertices = append(newVertices, sum.Mul(1.0/float64(2+len(fIndices))))
	}

	// 4. Reposition Original Vertices
	oldVertexOffset := len(newVertices)
	for i, V := range m.Vertices {
		n := float64(len(vertexToFaces[i]))
		var F, E math.Point3D
		for _, fIdx := range vertexToFaces[i] {
			F = F.Add(facePoints[fIdx])
		}
		F = F.Mul(1.0 / n)
		for key := range vertexToEdges[i] {
			mid := m.Vertices[key.V1].Add(m.Vertices[key.V2]).Mul(0.5)
			E = E.Add(mid)
		}
		E = E.Mul(1.0 / n)
		newVertices = append(newVertices, F.Add(E.Mul(2.0)).Add(V.Mul(n-3.0)).Mul(1.0/n))
	}

	// 5. Build New Faces
	var newFaces [][]int
	for fIdx, face := range m.Faces {
		for i, vIdx := range face {
			vPrev := face[(i+len(face)-1)%len(face)]
			vNext := face[(i+1)%len(face)]
			e1 := edgePointIndices[getEdgeKey(vIdx, vPrev)]
			e2 := edgePointIndices[getEdgeKey(vIdx, vNext)]
			newFaces = append(newFaces, []int{fIdx, e1, oldVertexOffset + vIdx, e2})
		}
	}
	return &Mesh{Vertices: newVertices, Faces: newFaces}
}
