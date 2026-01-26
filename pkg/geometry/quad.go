package geometry

import (
	"grinder/pkg/math"
	"image/color"
	gomath "math"
)

type BilinearQuad struct {
	P00, P10, P11, P01 math.Point3D
	N00, N10, N11, N01 math.Normal3D
	AABB               math.AABB3D
	Color              color.RGBA
	Thickness          float64 // Since meshes are thin, we need a small volume
	Shininess          float64
	SpecularIntensity  float64
	SpecularColor      color.RGBA
}

// PositionAt calculates the point on the quad at parameters u, v
func (q *BilinearQuad) PositionAt(u, v float64) math.Point3D {
	// Standard bilinear interpolation formula:
	// P(u,v) = (1-u)(1-v)P00 + u(1-v)P10 + uvP11 + (1-u)vP01
	res := q.P00.Mul((1 - u) * (1 - v)).
		Add(q.P10.Mul(u * (1 - v))).
		Add(q.P11.Mul(u * v)).
		Add(q.P01.Mul((1 - u) * v))
	return res
}

func (q *BilinearQuad) DistanceToPoint(p math.Point3D) float64 {
	// 1. Get the local u,v coordinates for the point
	u, v := q.findUVForPoint(p)

	// 2. If the point is outside the 0-1 range, it's NOT on this quad.
	// Return a huge distance so the marcher ignores this plane.
	if u < 0 || u > 1 || v < 0 || v > 1 {
		return 1e18
	}

	// 3. Otherwise, return the distance to the surface
	center := q.P00.Add(q.P11).Mul(0.5)
	n := q.P10.Sub(q.P00).Cross(q.P01.Sub(q.P00)).Normalize()
	return gomath.Abs(p.Sub(center).Dot(n))
}
func (q *BilinearQuad) NormalAtPoint(p math.Point3D, t float64) math.Normal3D {
	// Use smooth normals interpolated from vertex normals if they are defined (non-zero)
	zeroNormal := math.Normal3D{}
	if q.N00 != zeroNormal || q.N10 != zeroNormal || q.N11 != zeroNormal || q.N01 != zeroNormal {
		// Use smooth normals interpolated from vertex normals
		u, v := q.findUVForPoint(p)

		// Bilinear interpolation of the vertex normals
		n := q.N00.Mul((1 - u) * (1 - v)).
			Add(q.N10.Mul(u * (1 - v))).
			Add(q.N11.Mul(u * v)).
			Add(q.N01.Mul((1 - u) * v))

		return n.Normalize()
	} else {
		// Fallback to geometric normal calculation
		u, v := 0.5, 0.5
		dpdu := q.P00.Mul(-(1 - v)).Add(q.P10.Mul(1 - v)).Add(q.P11.Mul(v)).Add(q.P01.Mul(-v))
		dpdv := q.P00.Mul(-(1 - u)).Add(q.P10.Mul(-u)).Add(q.P11.Mul(u)).Add(q.P01.Mul(1 - u))

		n := dpdu.Cross(dpdv).Normalize()
		return math.Normal3D{X: n.X, Y: n.Y, Z: n.Z}
	}
}

func (q *BilinearQuad) Contains(p math.Point3D, t float64) bool {
	// 1. AABB check is mandatory and must be FAST
	aabb := q.GetAABB()
	if p.X < aabb.Min.X || p.X > aabb.Max.X ||
		p.Y < aabb.Min.Y || p.Y > aabb.Max.Y ||
		p.Z < aabb.Min.Z || p.Z > aabb.Max.Z {
		return false
	}

	// 2. Newton-Raphson
	u, v := q.findUVForPoint(p)
	// Only do the heavy math if we are inside the box
	surfacePoint := q.PositionAt(u, v)

	// Check distance to the calculated surface point
	return p.Sub(surfacePoint).Length() <= q.Thickness
}

func (q *BilinearQuad) findUVForPoint(target math.Point3D) (float64, float64) {
	u, v := 0.5, 0.5 // Start at center

	for iter := 0; iter < 8; iter++ { // Limit to 8 iterations
		currentPoint := q.PositionAt(u, v)
		residual := target.Sub(currentPoint)

		if residual.Length() < 1e-4 {
			break
		}

		du := q.partialDerivativeU(u, v)
		dv := q.partialDerivativeV(u, v)

		// Least Squares / Normal Equations
		jTj00 := du.Dot(du)
		jTj01 := du.Dot(dv)
		jTj11 := dv.Dot(dv)
		jTr0 := du.Dot(residual)
		jTr1 := dv.Dot(residual)

		det := jTj00*jTj11 - jTj01*jTj01

		// Safety: If det is near zero, the point is likely degenerate or far away
		if gomath.Abs(det) < 1e-9 {
			break
		}

		u += (jTj11*jTr0 - jTj01*jTr1) / det
		v += (jTj00*jTr1 - jTj01*jTr0) / det

		// Keep it on the patch
		u = gomath.Max(0, gomath.Min(1, u))
		v = gomath.Max(0, gomath.Min(1, v))
	}
	return u, v
}

// partialDerivativeU computes the partial derivative of the bilinear surface with respect to u
func (q *BilinearQuad) partialDerivativeU(u, v float64) math.Point3D {
	// dP/du = (1-v)(P10 - P00) + v(P11 - P01)
	term1 := q.P10.Sub(q.P00).Mul(1 - v)
	term2 := q.P11.Sub(q.P01).Mul(v)
	return term1.Add(term2)
}

// partialDerivativeV computes the partial derivative of the bilinear surface with respect to v
func (q *BilinearQuad) partialDerivativeV(u, v float64) math.Point3D {
	// dP/dv = (1-u)(P01 - P00) + u(P11 - P10)
	term1 := q.P01.Sub(q.P00).Mul(1 - u)
	term2 := q.P11.Sub(q.P10).Mul(u)
	return term1.Add(term2)
}

func (q *BilinearQuad) GetAABB() math.AABB3D {
	aabb := math.AABB3D{Min: q.P00, Max: q.P00}
	aabb = aabb.Expand(q.P10)
	aabb = aabb.Expand(q.P11)
	aabb = aabb.Expand(q.P01)

	// Add a tiny bit of "thickness" padding to the Z-axis
	// to ensure the dicing pass doesn't miss it due to float precision.
	aabb.Min.Z -= q.Thickness
	aabb.Max.Z += q.Thickness
	return aabb
}

func (q *BilinearQuad) Intersects(aabb math.AABB3D) bool {
	return q.GetAABB().Intersects(aabb)
}

func (q *BilinearQuad) GetColor() color.RGBA {
	return q.Color
}

func (q *BilinearQuad) GetShininess() float64 {
	return 32.0 // Default or add to struct
}

func (q *BilinearQuad) GetSpecularIntensity() float64 {
	return 0.5
}

func (q *BilinearQuad) GetSpecularColor() color.RGBA {
	return color.RGBA{255, 255, 255, 255}
}

func (q *BilinearQuad) GetCenter() math.Point3D {
	return q.PositionAt(0.5, 0.5)
}

func (q *BilinearQuad) IsVolumetric() bool {
	return false
}
