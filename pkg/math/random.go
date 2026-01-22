package math

// XorShift32 is a simple pseudo-random number generator.
type XorShift32 struct {
	state uint32
}

// NewXorShift32 creates a new XorShift32 with a given seed.
func NewXorShift32(seed uint32) *XorShift32 {
	if seed == 0 {
		seed = 1 // Avoid seed 0
	}
	return &XorShift32{state: seed}
}

// Next returns a pseudo-random uint32.
func (r *XorShift32) Next() uint32 {
	x := r.state
	x ^= x << 13
	x ^= x >> 17
	x ^= x << 5
	r.state = x
	return x
}

// NextFloat64 returns a pseudo-random float64 in [0, 1).
func (r *XorShift32) NextFloat64() float64 {
	return float64(r.Next()) / 4294967296.0
}
