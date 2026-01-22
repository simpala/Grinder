package math

// XorShift32 is a simple 32-bit pseudo-random number generator.
type XorShift32 struct {
	state uint32
}

// NewXorShift32 creates a new XorShift32 PRNG with a given seed.
func NewXorShift32(seed uint32) *XorShift32 {
	if seed == 0 {
		seed = 1 // 0 is a bad seed
	}
	return &XorShift32{state: seed}
}

// Next returns the next pseudo-random number.
func (r *XorShift32) Next() uint32 {
	x := r.state
	x ^= x << 13
	x ^= x >> 17
	x ^= x << 5
	r.state = x
	return x
}

// Float64 returns a pseudo-random float64 in [0, 1).
func (r *XorShift32) Float64() float64 {
	return float64(r.Next()) / 4294967296.0
}
