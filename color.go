package ring

import (
	"image/color"
)

// serialize transforms color information to uint32 with the shape 0x00RRGGBB
func serialize(c color.Color) uint32 {
	r, g, b, _ := c.RGBA()

	return ((r >> 8) << 16) |
		((g >> 8) << 8) |
		(b >> 8)
}

// blendOver blends multiple colors using the over operator and returns an
// alpha pre-multiplied color. The first color is considered to be at the
// bottom and the last color is considered to be at the top.
func blendOver(cs ...color.Color) (blend *color.RGBA) {
	over := func(a, b, delta uint32) uint8 {
		return uint8((a + b*delta/0xFFFF) >> 8)
	}
	blend = &color.RGBA{0, 0, 0, 0}
	for _, c := range cs {
		r, g, b, a := c.RGBA()
		bR, bG, bB, bA := blend.RGBA()
		delta := (0xFFFF - a)

		blend.R = over(r, bR, delta)
		blend.G = over(g, bG, delta)
		blend.B = over(b, bB, delta)
		blend.A = over(a, bA, delta)
	}

	return blend
}

// blendLerp blends two colors by linearly interpolating between them given the
// amount l: (0.0 to 1.0) -> (a to b).
func blendLerp(a, b color.Color, l float64) (blend *color.RGBA) {
	lerp := func(a, b, l uint32) uint8 {
		return uint8((a - (a-b)*l/0xFFFF) >> 8)
	}

	aR, aG, aB, aA := a.RGBA()
	bR, bG, bB, bA := b.RGBA()

	l16 := uint32(l * 0xFFFF)

	blend = &color.RGBA{
		R: lerp(aR, bR, l16),
		G: lerp(aG, bG, l16),
		B: lerp(aB, bB, l16),
		A: lerp(aA, bA, l16),
	}

	return blend
}
