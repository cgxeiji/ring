package ring

import (
	"image/color"
	"testing"
)

func TestSerialize(t *testing.T) {
	tests := []struct {
		name  string
		color color.Color
		want  uint32
	}{
		{
			"rgb",
			color.NRGBA{0x16, 0x16, 0x16, 0xFF},
			0x161616,
		},
		{
			"alpha",
			color.NRGBA{0xFF, 0xFF, 0xFF, 0x32},
			0x323232,
		},
		{
			"16bit",
			color.NRGBA64{0x3214, 0x1234, 0x00FF, 0xFFFF},
			0x321200,
		},
		{
			"gray",
			color.Gray{0x10},
			0x101010,
		},
	}

	for _, ts := range tests {
		t.Run(ts.name, func(t *testing.T) {
			got := serialize(ts.color)
			if got != ts.want {
				t.Errorf("got: %#v, want: %#v", got, ts.want)
			}
		})
	}
}

func TestBlendOver(t *testing.T) {
	tests := []struct {
		name   string
		colors []color.Color
		want   color.RGBA
	}{
		{
			"single",
			[]color.Color{
				color.RGBA{0x15, 0x16, 0x17, 0x18},
			},
			color.RGBA{0x15, 0x16, 0x17, 0x18},
		},
		{
			"white over black",
			[]color.Color{
				color.RGBA{0x00, 0x00, 0x00, 0xFF},
				color.RGBA{0xFF, 0xFF, 0xFF, 0xFF},
			},
			color.RGBA{0xFF, 0xFF, 0xFF, 0xFF},
		},
		{
			"black over white",
			[]color.Color{
				color.RGBA{0xFF, 0xFF, 0xFF, 0xFF},
				color.RGBA{0x00, 0x00, 0x00, 0xFF},
			},
			color.RGBA{0x00, 0x00, 0x00, 0xFF},
		},
		{
			"red over green",
			[]color.Color{
				color.NRGBA{0x00, 0x80, 0x00, 0xFF},
				color.NRGBA{0x80, 0x00, 0x00, 0xA1},
			},
			color.RGBA{0x51, 0x2F, 0x00, 0xFF},
		},
	}

	for _, ts := range tests {
		t.Run(ts.name, func(t *testing.T) {
			got := *blendOver(ts.colors...)
			if got != ts.want {
				t.Errorf("got: %#v, want: %#v", got, ts.want)
			}
		})
	}
}

func TestBlendLerp(t *testing.T) {
	tests := []struct {
		name   string
		colorA color.Color
		colorB color.Color
		l      float64
		want   color.RGBA
	}{
		{
			"fullA",
			color.RGBA{128, 128, 0, 128},
			color.RGBA{0, 255, 255, 255},
			0.0,
			color.RGBA{128, 128, 0, 128},
		},
		{
			"fullB",
			color.RGBA{128, 128, 0, 128},
			color.RGBA{0, 255, 255, 255},
			1.0,
			color.RGBA{0, 255, 255, 255},
		},
		{
			"half",
			color.RGBA{128, 128, 0, 128},
			color.RGBA{0, 255, 255, 255},
			0.5,
			color.RGBA{64, 192, 127, 192},
		},
		{
			"quater",
			color.RGBA{128, 128, 0, 128},
			color.RGBA{0, 255, 255, 255},
			0.75,
			color.RGBA{32, 224, 191, 224},
		},
	}

	for _, ts := range tests {
		t.Run(ts.name, func(t *testing.T) {
			got := *blendLerp(ts.colorA, ts.colorB, ts.l)
			if got != ts.want {
				t.Errorf("got: %v, want: %v", got, ts.want)
			}
		})
	}
}
