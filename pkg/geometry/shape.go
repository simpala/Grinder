package geometry

import (
	"grinder/pkg/math"
	"image/color"
)

// Shape defines the interface for all geometric objects in the scene.
type Shape interface {
	AtTime(t float64) Shape
	Contains(p math.Point3D, time float64) bool
	Intersects(aabb math.AABB3D) bool
	NormalAtPoint(p math.Point3D, t float64) math.Normal3D
	GetColor() color.RGBA
	GetShininess() float64
	GetSpecularIntensity() float64
	GetSpecularColor() color.RGBA
	GetAABB() math.AABB3D
	GetCenter() math.Point3D
	IsVolumetric() bool
}

// VolumetricShape defines the interface for all volumetric objects in the scene.
type VolumetricShape interface {
	Shape
	GetDensity() float64
}
