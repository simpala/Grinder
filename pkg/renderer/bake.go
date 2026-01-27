package renderer

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"grinder/pkg/camera"
	"grinder/pkg/geometry"
	"grinder/pkg/math"
	"grinder/pkg/shading"
	"io"
	gomath "math"
	"os"
	"sort"
	"unsafe"

	"golang.org/x/exp/mmap"
)

// BakedAtom represents a single voxel in the baked scene.
type BakedAtom struct {
	Pos        [3]float32
	HalfExtent float32
	Normal     uint32
	Albedo     [3]uint8
	MaterialID uint8
	LightDir   uint32
	LightColor [3]uint8
	Padding    uint8
}

// TLASNode represents a node in the Top-Level Acceleration Structure.
type TLASNode struct {
	Min, Max    [3]float32
	BLASOffset  int64 // Absolute file offset to the root BLASNode of a shape
	IsLeaf      int32 // 1 = Leaf (points to BLAS), 0 = Node
	Left, Right int32 // Relative indices to other TLASNodes
	Padding     int32
}

// BLASNode represents a node in the Bottom-Level Acceleration Structure.
type BLASNode struct {
	Min, Max    [3]float32
	AtomOffset  int64 // Absolute file offset to the first atom in a leaf
	AtomCount   int32 // Number of atoms in leaf, 0 for internal nodes
	Left, Right int32 // Relative indices to other BLASNodes in the same shape's block
	Padding     int32
}

// CameraData stores camera parameters for the bake.
type CameraData struct {
	Eye    [3]float32
	Target [3]float32
	Up     [3]float32
	Fov    float32
	Aspect float32
	Near   float32
	Far    float32
}

// Header is the file header for the baked scene.
type Header struct {
	Magic      [4]byte
	Version    uint32
	AtomCount  int64
	TLASRoot   int64 // Absolute file offset to the root TLASNode
	BakeCamera CameraData
	VoxelSize  float32
	Epsilon    float32
}

type blasResult struct {
	shapeID    uint8
	rootOffset int64
	aabb       math.AABB3D
}

func OctEncode(n math.Point3D) uint32 {
	l1 := gomath.Abs(n.X) + gomath.Abs(n.Y) + gomath.Abs(n.Z)
	if l1 == 0 {
		return 0
	}
	x, y := n.X/l1, n.Y/l1
	if n.Z < 0 {
		oldX := x
		x = (1.0 - gomath.Abs(y)) * sign(oldX)
		y = (1.0 - gomath.Abs(oldX)) * sign(y)
	}
	ux := uint32((x*0.5 + 0.5) * 65535.0)
	uy := uint32((y*0.5 + 0.5) * 65535.0)
	return (ux << 16) | uy
}

func OctDecode(packed uint32) math.Point3D {
	ux := packed >> 16
	uy := packed & 0xFFFF
	x := (float64(ux)/65535.0)*2.0 - 1.0
	y := (float64(uy)/65535.0)*2.0 - 1.0
	n := math.Point3D{X: x, Y: y, Z: 1.0 - gomath.Abs(x) - gomath.Abs(y)}
	if n.Z < 0 {
		oldX := n.X
		n.X = (1.0 - gomath.Abs(n.Y)) * sign(oldX)
		n.Y = (1.0 - gomath.Abs(oldX)) * sign(n.Y)
	}
	return n.Normalize()
}

func sign(f float64) float64 {
	if f >= 0 {
		return 1
	}
	return -1
}

func (a *BakedAtom) Write(w io.Writer) error { return binary.Write(w, binary.LittleEndian, a) }
func (a *BakedAtom) Read(r io.Reader) error  { return binary.Read(r, binary.LittleEndian, a) }

type BakeEngine struct {
	Camera   camera.Camera
	Shapes   []geometry.Shape
	Light    shading.Light
	Width    int
	Height   int
	MinSize  float64
	Near     float64
	Far      float64
	Shutter  float64
	shapeIDs map[geometry.Shape]uint8

	CamTarget math.Point3D
	CamUp     math.Point3D
	CamFov    float64
}

func NewBakeEngine(cam camera.Camera, shapes []geometry.Shape, light shading.Light, width, height int, minSize, near, far, shutter float64, target, up math.Point3D, fov float64) *BakeEngine {
	shapeIDs := make(map[geometry.Shape]uint8)
	for i, s := range shapes {
		shapeIDs[s] = uint8(i)
	}
	return &BakeEngine{
		Camera: cam, Shapes: shapes, Light: light, Width: width, Height: height,
		MinSize: minSize, Near: near, Far: far, Shutter: shutter,
		shapeIDs: shapeIDs, CamTarget: target, CamUp: up, CamFov: fov,
	}
}

func (e *BakeEngine) Bake(tempFile string, finalFile string) error {
	fmt.Printf("Starting Pass A (The Raw Bake)... writing to %s\n", tempFile)
	f, err := os.Create(tempFile)
	if err != nil {
		return err
	}
	defer f.Close()
	initialAABB := math.AABB3D{Min: math.Point3D{X: 0, Y: 0, Z: e.Near}, Max: math.Point3D{X: 1, Y: 1, Z: e.Far}}
	bvh := geometry.NewBVH(e.Shapes)
	atomCount := int64(0)
	e.subdivideBake(initialAABB, f, bvh, &atomCount)
	fmt.Printf("Pass A complete. Baked %d atoms.\n", atomCount)
	return e.Indexer(tempFile, finalFile, atomCount)
}

func (e *BakeEngine) Indexer(tempFile string, finalFile string, totalAtoms int64) error {
	fmt.Printf("Starting Pass B (The Indexer)... writing to %s\n", finalFile)
	f, err := os.Open(tempFile)
	if err != nil {
		return err
	}
	defer f.Close()
	atomsByShape := make(map[uint8][]BakedAtom)
	for {
		var atom BakedAtom
		if err := atom.Read(f); err != nil {
			if err == io.EOF {
				break
			}
			return err
		}
		atomsByShape[atom.MaterialID] = append(atomsByShape[atom.MaterialID], atom)
	}
	out, err := os.Create(finalFile)
	if err != nil {
		return err
	}
	defer out.Close()
	header := Header{
		Magic: [4]byte{'S', 'D', 'S', 'B'}, Version: 1, AtomCount: totalAtoms,
		VoxelSize: float32(e.MinSize),
		Epsilon:   float32(e.MinSize * 1.5),
	}
	eye := e.Camera.GetEye()
	header.BakeCamera = CameraData{
		Eye:    [3]float32{float32(eye.X), float32(eye.Y), float32(eye.Z)},
		Target: [3]float32{float32(e.CamTarget.X), float32(e.CamTarget.Y), float32(e.CamTarget.Z)},
		Up:     [3]float32{float32(e.CamUp.X), float32(e.CamUp.Y), float32(e.CamUp.Z)},
		Fov:    float32(e.CamFov), Aspect: float32(float64(e.Width) / float64(e.Height)),
		Near: float32(e.Near), Far: float32(e.Far),
	}
	binary.Write(out, binary.LittleEndian, header)

	var blasResults []blasResult
	for shapeID, atoms := range atomsByShape {
		fmt.Printf("Building BLAS for Shape %d (%d atoms)...\n", shapeID, len(atoms))
		nodes, sortedAtoms := e.buildBLAS(atoms)
		atomStartOffset, _ := out.Seek(0, io.SeekCurrent)
		for _, a := range sortedAtoms {
			a.Write(out)
		}
		blasStartOffset, _ := out.Seek(0, io.SeekCurrent)
		for i := range nodes {
			if nodes[i].AtomCount > 0 {
				nodes[i].AtomOffset = atomStartOffset + int64(nodes[i].AtomOffset)
			}
		}
		for _, n := range nodes {
			binary.Write(out, binary.LittleEndian, n)
		}
		shapeAABB := math.AABB3D{
			Min: math.Point3D{X: float64(nodes[0].Min[0]), Y: float64(nodes[0].Min[1]), Z: float64(nodes[0].Min[2])},
			Max: math.Point3D{X: float64(nodes[0].Max[0]), Y: float64(nodes[0].Max[1]), Z: float64(nodes[0].Max[2])},
		}
		blasResults = append(blasResults, blasResult{shapeID: shapeID, rootOffset: blasStartOffset, aabb: shapeAABB})
	}
	tlasNodes := e.buildTLAS(blasResults)
	tlasStartOffset, _ := out.Seek(0, io.SeekCurrent)
	for _, n := range tlasNodes {
		binary.Write(out, binary.LittleEndian, n)
	}
	header.TLASRoot = tlasStartOffset
	out.Seek(0, io.SeekStart)
	binary.Write(out, binary.LittleEndian, header)
	fmt.Printf("Pass B complete. Final scene written to %s\n", finalFile)
	return nil
}

func (e *BakeEngine) buildBLAS(atoms []BakedAtom) ([]BLASNode, []BakedAtom) {
	if len(atoms) == 0 {
		return nil, nil
	}
	minP, maxP := atoms[0].Pos, atoms[0].Pos
	for _, a := range atoms {
		for i := 0; i < 3; i++ {
			if a.Pos[i] < minP[i] {
				minP[i] = a.Pos[i]
			}
			if a.Pos[i] > maxP[i] {
				maxP[i] = a.Pos[i]
			}
		}
	}
	sceneAABB := math.AABB3D{Min: math.Point3D{X: float64(minP[0]), Y: float64(minP[1]), Z: float64(minP[2])}, Max: math.Point3D{X: float64(maxP[0]), Y: float64(maxP[1]), Z: float64(maxP[2])}}
	diag := sceneAABB.Max.Sub(sceneAABB.Min)
	type atomWithCode struct {
		atom BakedAtom
		code uint32
	}
	codedAtoms := make([]atomWithCode, len(atoms))
	for i, a := range atoms {
		nx, ny, nz := 0.5, 0.5, 0.5
		if diag.X > 0 {
			nx = (float64(a.Pos[0]) - sceneAABB.Min.X) / diag.X
		}
		if diag.Y > 0 {
			ny = (float64(a.Pos[1]) - sceneAABB.Min.Y) / diag.Y
		}
		if diag.Z > 0 {
			nz = (float64(a.Pos[2]) - sceneAABB.Min.Z) / diag.Z
		}
		codedAtoms[i] = atomWithCode{atom: a, code: math.Morton3D(nx, ny, nz)}
	}
	sort.Slice(codedAtoms, func(i, j int) bool { return codedAtoms[i].code < codedAtoms[j].code })
	sortedAtoms := make([]BakedAtom, len(atoms))
	for i, ca := range codedAtoms {
		sortedAtoms[i] = ca.atom
	}
	var nodes []BLASNode
	var build func(start, end int) int32
	build = func(start, end int) int32 {
		nodeIdx := int32(len(nodes))
		nodes = append(nodes, BLASNode{Left: -1, Right: -1})
		curMin, curMax := sortedAtoms[start].Pos, sortedAtoms[start].Pos
		for i := start + 1; i < end; i++ {
			for j := 0; j < 3; j++ {
				if sortedAtoms[i].Pos[j] < curMin[j] {
					curMin[j] = sortedAtoms[i].Pos[j]
				}
				if sortedAtoms[i].Pos[j] > curMax[j] {
					curMax[j] = sortedAtoms[i].Pos[j]
				}
			}
		}
		nodes[nodeIdx].Min, nodes[nodeIdx].Max = curMin, curMax
		count := end - start
		if count <= 64 {
			nodes[nodeIdx].AtomOffset = int64(start) * 32 // This is relative to atomStartOffset
			nodes[nodeIdx].AtomCount = int32(count)
			return nodeIdx
		}
		mid := start + count/2
		l, r := build(start, mid), build(mid, end)
		nodes[nodeIdx].Left, nodes[nodeIdx].Right = l, r
		return nodeIdx
	}
	build(0, len(sortedAtoms))
	return nodes, sortedAtoms
}

func (e *BakeEngine) buildTLAS(blasInfos []blasResult) []TLASNode {
	if len(blasInfos) == 0 {
		return nil
	}
	var nodes []TLASNode
	var build func(infos []blasResult) int32
	build = func(infos []blasResult) int32 {
		nodeIdx := int32(len(nodes))
		nodes = append(nodes, TLASNode{Left: -1, Right: -1})
		overallAABB := infos[0].aabb
		for i := 1; i < len(infos); i++ {
			overallAABB = overallAABB.Expand(infos[i].aabb.Min).Expand(infos[i].aabb.Max)
		}
		nodes[nodeIdx].Min = [3]float32{float32(overallAABB.Min.X), float32(overallAABB.Min.Y), float32(overallAABB.Min.Z)}
		nodes[nodeIdx].Max = [3]float32{float32(overallAABB.Max.X), float32(overallAABB.Max.Y), float32(overallAABB.Max.Z)}
		if len(infos) == 1 {
			nodes[nodeIdx].IsLeaf = 1
			nodes[nodeIdx].BLASOffset = infos[0].rootOffset
			return nodeIdx
		}
		mid := len(infos) / 2
		l, r := build(infos[:mid]), build(infos[mid:])
		nodes[nodeIdx].Left, nodes[nodeIdx].Right = l, r
		return nodeIdx
	}
	build(blasInfos)
	return nodes
}

func (e *BakeEngine) computeAABBWorld(aabb math.AABB3D) math.AABB3D {
	corners := aabb.GetCorners()
	first := true
	var res math.AABB3D
	for _, c := range corners {
		p := e.Camera.Project(c.X, c.Y, c.Z)
		if first {
			res = math.AABB3D{Min: p, Max: p}
			first = false
		} else {
			res = res.Expand(p)
		}
	}
	return res
}

func (e *BakeEngine) subdivideBake(aabb math.AABB3D, w io.Writer, bvh *geometry.BVH, atomCount *int64) {
	worldAABB := e.computeAABBWorld(aabb)
	shapes := bvh.IntersectsShapes(worldAABB)
	if len(shapes) == 0 {
		return
	}
	if (aabb.Max.X - aabb.Min.X) < e.MinSize {
		// Surface Pruning: discard if entirely inside a single solid shape.
		// !IsVolumetric() identifies solid geometry (vs participating media),
		// allowing us to hollow out the interior and keep only the shell.
		if len(shapes) == 1 && !shapes[0].IsVolumetric() {
			allInside := true
			for _, c := range aabb.GetCorners() {
				worldC := e.Camera.Project(c.X, c.Y, c.Z)
				if !shapes[0].Contains(worldC, 0) {
					allInside = false
					break
				}
			}
			if allInside {
				return
			}
		}

		center := aabb.Center()
		worldP := e.Camera.Project(center.X, center.Y, center.Z)
		for _, s := range shapes {
			if s.Contains(worldP, 0) {
				id, ok := e.shapeIDs[s]
				if !ok {
					continue
				}
				albedo, normal := s.GetColor(), s.NormalAtPoint(worldP, 0)
				lightDir := e.Light.Position.Sub(worldP).Normalize()
				//checkP := worldP.Add(normal.ToVector().Mul(1e-4))
				//attenuation := shading.CalculateShadowAttenuation(checkP, e.Light.Position, e.Shapes, e.Light.Radius, 0)
				//lIntensity := e.Light.Intensity * attenuation
				lIntensity := e.Light.Intensity // we dont ever want to bake approximated shadows.
				pCorner := e.Camera.Project(aabb.Max.X, aabb.Max.Y, aabb.Max.Z)
				halfExtent := pCorner.Sub(worldP).Length()
				atom := BakedAtom{
					Pos:        [3]float32{float32(worldP.X), float32(worldP.Y), float32(worldP.Z)},
					HalfExtent: float32(halfExtent),
					Normal:     OctEncode(normal.ToVector()),
					Albedo:     [3]uint8{albedo.R, albedo.G, albedo.B}, MaterialID: id,
					LightDir:   OctEncode(lightDir),
					LightColor: [3]uint8{uint8(gomath.Min(255, 255*lIntensity)), uint8(gomath.Min(255, 255*lIntensity)), uint8(gomath.Min(255, 255*lIntensity))},
				}
				atom.Write(w)
				*atomCount++
			}
		}
		return
	}
	mx, my, mz := (aabb.Min.X+aabb.Max.X)/2, (aabb.Min.Y+aabb.Max.Y)/2, (aabb.Min.Z+aabb.Max.Z)/2
	xs, ys, zs := [3]float64{aabb.Min.X, mx, aabb.Max.X}, [3]float64{aabb.Min.Y, my, aabb.Max.Y}, [3]float64{aabb.Min.Z, mz, aabb.Max.Z}
	for zi := 0; zi < 2; zi++ {
		for xi := 0; xi < 2; xi++ {
			for yi := 0; yi < 2; yi++ {
				e.subdivideBake(math.AABB3D{Min: math.Point3D{X: xs[xi], Y: ys[yi], Z: zs[zi]}, Max: math.Point3D{X: xs[xi+1], Y: ys[yi+1], Z: zs[zi+1]}}, w, bvh, atomCount)
			}
		}
	}
}

func (e *BakeEngine) Verify(bakedFile string) error {
	scene, err := LoadBakedScene(bakedFile)
	if err != nil {
		return err
	}
	defer scene.Close()
	fmt.Printf("Verifying baked scene %s...\nAtoms: %d, TLASRoot offset: %d\n", bakedFile, scene.Header.AtomCount, scene.Header.TLASRoot)
	for y := 0.4; y <= 0.6; y += 0.05 {
		for x := 0.4; x <= 0.6; x += 0.05 {
			pNear, pFar := e.Camera.Project(x, y, e.Near), e.Camera.Project(x, y, e.Far)
			ray := math.Ray{Origin: pNear, Direction: pFar.Sub(pNear).Normalize()}
			hit, atom := scene.Intersect(ray)
			if hit {
				fmt.Printf("Ray at (%.2f, %.2f): HIT shape %d at (%.2f, %.2f, %.2f)\n", x, y, atom.MaterialID, atom.Pos[0], atom.Pos[1], atom.Pos[2])
			} else {
				fmt.Printf("Ray at (%.2f, %.2f): MISS\n", x, y)
			}
		}
	}
	return nil
}

type BakedScene struct {
	Header Header
	Data   []byte
	closer io.Closer
}

func (s *BakedScene) Close() error {
	if s.closer != nil {
		return s.closer.Close()
	}
	return nil
}

func LoadBakedScene(filename string, memLimit ...int64) (*BakedScene, error) {
	limit := int64(2 * 1024 * 1024 * 1024) // Default 2GB
	if len(memLimit) > 0 {
		limit = memLimit[0]
	}

	f, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	info, err := f.Stat()
	if err != nil {
		return nil, err
	}
	size := info.Size()

	var data []byte
	var closer io.Closer

	if size < limit {
		data, err = os.ReadFile(filename)
		if err != nil {
			return nil, err
		}
	} else {
		r, err := mmap.Open(filename)
		if err != nil {
			return nil, err
		}
		closer = r
		// Use unsafe to access the unexported data []byte field of mmap.ReaderAt.
		// This provides the requested consistent []byte access portably.
		data = *(*[]byte)(unsafe.Pointer(r))
	}

	if len(data) < 84 {
		if closer != nil {
			closer.Close()
		}
		return nil, fmt.Errorf("file too small")
	}

	var header Header
	if err := binary.Read(bytes.NewReader(data), binary.LittleEndian, &header); err != nil {
		if closer != nil {
			closer.Close()
		}
		return nil, err
	}
	return &BakedScene{Header: header, Data: data, closer: closer}, nil
}

func (s *BakedScene) Intersect(ray math.Ray) (bool, BakedAtom) {
	return s.intersectTLAS(s.Header.TLASRoot, ray)
}

func (s *BakedScene) getTLASNode(offset int64) TLASNode {
	data := s.Data[offset:]
	return TLASNode{
		Min: [3]float32{
			gomath.Float32frombits(binary.LittleEndian.Uint32(data[0:4])),
			gomath.Float32frombits(binary.LittleEndian.Uint32(data[4:8])),
			gomath.Float32frombits(binary.LittleEndian.Uint32(data[8:12])),
		},
		Max: [3]float32{
			gomath.Float32frombits(binary.LittleEndian.Uint32(data[12:16])),
			gomath.Float32frombits(binary.LittleEndian.Uint32(data[16:20])),
			gomath.Float32frombits(binary.LittleEndian.Uint32(data[20:24])),
		},
		BLASOffset: int64(binary.LittleEndian.Uint64(data[24:32])),
		IsLeaf:     int32(binary.LittleEndian.Uint32(data[32:36])),
		Left:       int32(binary.LittleEndian.Uint32(data[36:40])),
		Right:      int32(binary.LittleEndian.Uint32(data[40:44])),
		Padding:    int32(binary.LittleEndian.Uint32(data[44:48])),
	}
}

func (s *BakedScene) getBLASNode(offset int64) BLASNode {
	data := s.Data[offset:]
	return BLASNode{
		Min: [3]float32{
			gomath.Float32frombits(binary.LittleEndian.Uint32(data[0:4])),
			gomath.Float32frombits(binary.LittleEndian.Uint32(data[4:8])),
			gomath.Float32frombits(binary.LittleEndian.Uint32(data[8:12])),
		},
		Max: [3]float32{
			gomath.Float32frombits(binary.LittleEndian.Uint32(data[12:16])),
			gomath.Float32frombits(binary.LittleEndian.Uint32(data[16:20])),
			gomath.Float32frombits(binary.LittleEndian.Uint32(data[20:24])),
		},
		AtomOffset: int64(binary.LittleEndian.Uint64(data[24:32])),
		AtomCount:  int32(binary.LittleEndian.Uint32(data[32:36])),
		Left:       int32(binary.LittleEndian.Uint32(data[36:40])),
		Right:      int32(binary.LittleEndian.Uint32(data[40:44])),
		Padding:    int32(binary.LittleEndian.Uint32(data[44:48])),
	}
}

func (s *BakedScene) getBakedAtom(offset int64) BakedAtom {
	data := s.Data[offset:]
	return BakedAtom{
		Pos: [3]float32{
			gomath.Float32frombits(binary.LittleEndian.Uint32(data[0:4])),
			gomath.Float32frombits(binary.LittleEndian.Uint32(data[4:8])),
			gomath.Float32frombits(binary.LittleEndian.Uint32(data[8:12])),
		},
		HalfExtent: gomath.Float32frombits(binary.LittleEndian.Uint32(data[12:16])),
		Normal:     binary.LittleEndian.Uint32(data[16:20]),
		Albedo:     [3]uint8{data[20], data[21], data[22]},
		MaterialID: data[23],
		LightDir:   binary.LittleEndian.Uint32(data[24:28]),
		LightColor: [3]uint8{data[28], data[29], data[30]},
		Padding:    data[31],
	}
}

func (s *BakedScene) intersectTLAS(offset int64, ray math.Ray) (bool, BakedAtom) {
	if offset < 0 || offset+48 > int64(len(s.Data)) {
		return false, BakedAtom{}
	}
	node := s.getTLASNode(offset)
	aabb := math.AABB3D{Min: math.Point3D{X: float64(node.Min[0]), Y: float64(node.Min[1]), Z: float64(node.Min[2])}, Max: math.Point3D{X: float64(node.Max[0]), Y: float64(node.Max[1]), Z: float64(node.Max[2])}}
	if _, _, ok := aabb.IntersectRay(ray); !ok {
		return false, BakedAtom{}
	}
	if node.IsLeaf == 1 {
		return s.intersectBLAS(node.BLASOffset, node.BLASOffset, ray)
	}
	var hitL, hitR bool
	var atomL, atomR BakedAtom
	if node.Left != -1 {
		hitL, atomL = s.intersectTLAS(s.Header.TLASRoot+int64(node.Left)*48, ray)
	}
	if node.Right != -1 {
		hitR, atomR = s.intersectTLAS(s.Header.TLASRoot+int64(node.Right)*48, ray)
	}
	if hitL && hitR {
		dL := math.Point3D{X: float64(atomL.Pos[0]), Y: float64(atomL.Pos[1]), Z: float64(atomL.Pos[2])}.Sub(ray.Origin).LengthSquared()
		dR := math.Point3D{X: float64(atomR.Pos[0]), Y: float64(atomR.Pos[1]), Z: float64(atomR.Pos[2])}.Sub(ray.Origin).LengthSquared()
		if dL < dR {
			return true, atomL
		} else {
			return true, atomR
		}
	}
	if hitL {
		return true, atomL
	}
	return hitR, atomR
}

func (s *BakedScene) intersectBLAS(baseOffset int64, offset int64, ray math.Ray) (bool, BakedAtom) {
	if offset < 0 || offset+48 > int64(len(s.Data)) {
		return false, BakedAtom{}
	}
	node := s.getBLASNode(offset)
	aabb := math.AABB3D{Min: math.Point3D{X: float64(node.Min[0]), Y: float64(node.Min[1]), Z: float64(node.Min[2])}, Max: math.Point3D{X: float64(node.Max[0]), Y: float64(node.Max[1]), Z: float64(node.Max[2])}}
	if _, _, ok := aabb.IntersectRay(ray); !ok {
		return false, BakedAtom{}
	}
	if node.AtomCount > 0 {
		if node.AtomOffset < 0 || node.AtomOffset+int64(node.AtomCount)*32 > int64(len(s.Data)) {
			return false, BakedAtom{}
		}
		var nearest BakedAtom
		found, minDist := false, 1e18
		for i := 0; i < int(node.AtomCount); i++ {
			atomOffset := node.AtomOffset + int64(i)*32
			// Lazy Decoding: extract only Pos and HalfExtent (first 16 bytes) for the AABB check.
			atomData := s.Data[atomOffset:]
			posX := gomath.Float32frombits(binary.LittleEndian.Uint32(atomData[0:4]))
			posY := gomath.Float32frombits(binary.LittleEndian.Uint32(atomData[4:8]))
			posZ := gomath.Float32frombits(binary.LittleEndian.Uint32(atomData[8:12]))
			halfExtent := gomath.Float32frombits(binary.LittleEndian.Uint32(atomData[12:16]))

			atomAABB := math.AABB3D{
				Min: math.Point3D{X: float64(posX - halfExtent), Y: float64(posY - halfExtent), Z: float64(posZ - halfExtent)},
				Max: math.Point3D{X: float64(posX + halfExtent), Y: float64(posY + halfExtent), Z: float64(posZ + halfExtent)},
			}
			if tmin, _, ok := atomAABB.IntersectRay(ray); ok {
				if tmin < minDist {
					minDist = tmin
					nearest = s.getBakedAtom(atomOffset)
					found = true
				}
			}
		}
		return found, nearest
	}
	var hitL, hitR bool
	var atomL, atomR BakedAtom
	if node.Left != -1 {
		hitL, atomL = s.intersectBLAS(baseOffset, baseOffset+int64(node.Left)*48, ray)
	}
	if node.Right != -1 {
		hitR, atomR = s.intersectBLAS(baseOffset, baseOffset+int64(node.Right)*48, ray)
	}
	if hitL && hitR {
		dL := math.Point3D{X: float64(atomL.Pos[0]), Y: float64(atomL.Pos[1]), Z: float64(atomL.Pos[2])}.Sub(ray.Origin).LengthSquared()
		dR := math.Point3D{X: float64(atomR.Pos[0]), Y: float64(atomR.Pos[1]), Z: float64(atomR.Pos[2])}.Sub(ray.Origin).LengthSquared()
		if dL < dR {
			return true, atomL
		} else {
			return true, atomR
		}
	}
	if hitL {
		return true, atomL
	}
	return hitR, atomR
}
