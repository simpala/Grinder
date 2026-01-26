package math

// ExpandBits spreads 10 bits to 30 bits by inserting 2 zeros between each bit.
func ExpandBits(v uint32) uint32 {
	v &= 0x000003FF
	v = (v | (v << 16)) & 0x030000FF
	v = (v | (v << 8)) & 0x0300F00F
	v = (v | (v << 4)) & 0x030C30C3
	v = (v | (v << 2)) & 0x09249249
	return v
}

// Morton3D computes a 30-bit Morton code for a 3D point in [0, 1] range.
func Morton3D(x, y, z float64) uint32 {
	ux := uint32(x * 1023.0)
	uy := uint32(y * 1023.0)
	uz := uint32(z * 1023.0)
	return (ExpandBits(ux) << 2) | (ExpandBits(uy) << 1) | ExpandBits(uz)
}
