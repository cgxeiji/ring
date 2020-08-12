package ring

import (
	"fmt"
	"image/color"
	"math"
)

// Layer represents a drawable layer of the LED ring.
type Layer struct {
	pixels           []color.Color
	rotation, pixArc float64 // in radians
	opt              *LayerOptions
	buffer           []color.Color
}

// LayerOptions is the list of options of a layer.
type LayerOptions struct {
	// Resolution sets the number of pixels a layer has. Usually, this is set
	// to the same number of LEDs the ring has.
	Resolution int
	// ContentMode sets how the layer will be rendered (default: Tile).
	ContentMode ContentMode
}

// ContentMode defines how the layer will be rendered.
type ContentMode uint8

const (
	// ContentTile sets the layer to crop its content if it is larger that the ring
	// and to repeat the content.
	ContentTile ContentMode = iota
	// ContentCrop sets the layer to crop its content if it is larger than the ring
	// and does not repeat the content.
	ContentCrop
	// ContentScale sets the layer to scale up or down its content to fit the ring.
	ContentScale
)

// NewLayer creates a new drawable layer.
func NewLayer(options *LayerOptions) (*Layer, error) {
	if options.Resolution == 0 {
		return nil, fmt.Errorf("ring: resolution of new layer is 0")
	}

	l := &Layer{
		pixels: make([]color.Color, options.Resolution),
		buffer: make([]color.Color, options.Resolution),
		pixArc: 2 * math.Pi / float64(options.Resolution),
		opt:    options,
	}
	l.SetAll(color.Transparent)
	l.update()

	return l, nil
}

// SetAll sets all the pixels of a layer to an uniform color.
func (l *Layer) SetAll(c color.Color) {
	for i := range l.pixels {
		l.pixels[i] = c
	}
	l.update()
}

// SetPixel sets the color of a single pixel in the layer.
func (l *Layer) SetPixel(i int, c color.Color) {
	l.pixels[i] = c
	l.update()
}

// Rotate sets the rotation of the layer. A positive angle makes a counter-clockwise rotation.
func (l *Layer) Rotate(angle float64) {
	l.rotation = angle
	l.update()
}

// pixelRotated returns the color of the pixel at position i adjusted for the
// rotation of the layer.
func (l *Layer) pixelRotated(i int) (c color.Color) {
	rotFloat := l.rotation / l.pixArc
	rotInt := math.Floor(rotFloat)
	rotFloat -= rotInt

	i += int(rotInt)

	c = blendLerp(l.pixel(i), l.pixel(i+1), rotFloat)

	return c
}

func (l *Layer) led(i int) (c color.Color) {
	return l.buffer[mod(i, l.opt.Resolution)]
}

func (l *Layer) update() {
	for i := range l.pixels {
		l.buffer[i] = l.pixelRotated(i)
	}
}

// pixel returns the color of the pixel at position i.
func (l *Layer) pixel(i int) (c color.Color) {
	return l.pixels[mod(i, l.opt.Resolution)]
}

func mod(p, n int) (r int) {
	r = p % n
	if r < 0 {
		r += n
	}

	return r
}
