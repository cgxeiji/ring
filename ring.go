package ring

import (
	"fmt"
	"image/color"
	"math"
	"os"

	ws2811 "github.com/rpi-ws281x/rpi-ws281x-go"
)

// Ring represents the WS2811 LED device.
type Ring struct {
	device    *ws2811.WS2811
	layers    []Pixeler
	ledArc    float64
	ledOffset int
	offset    float64
	opt       *Options
}

// Pixeler is an interface that returns the color of a pixel at a specific
// location, with a set resolution.
type Pixeler interface {
	Pixel(int) color.Color
	Options() *LayerOptions
}

// Options is the list of ring options.
type Options struct {
	// LedCount is the number of LEDs in the ring.
	LedCount int
	// MinBrightness is the minimum output of the LED> Goes from 0 to 255
	// (default: 0).
	// MaxBrightness is the maximum output of the LED. Goes from 0 to 255
	// (default: 64).
	//
	// The color will be scaled to these values. For example, color.RGBA{255,
	// 255, 255, 255} will output led(R: 128, G: 128, B: 128) if MaxBrightness
	// is set to 128, and color.RGBA(0, 0, 0, 0) will output led(R: 10, G: 10,
	// B: 10) if MinBrightness is set to 10.
	MinBrightness, MaxBrightness int
	// GpioPin is the GPIO pin on the Raspberry Pi with PWM output (default:
	// GPIO 18). *Do not confuse with the physical pin number*
	GpioPin int
}

// New creates a new LED ring with given options.
func New(options *Options) (*Ring, error) {
	if os.Getuid() != 0 {
		return nil, fmt.Errorf("ring: rpi-ws281x needs root permissions (try running as sudo)")
	}

	opt := ws2811.DefaultOptions
	if options.LedCount != 0 {
		opt.Channels[0].LedCount = options.LedCount
	}
	if options.MaxBrightness != 0 {
		opt.Channels[0].Brightness = options.MaxBrightness
	}
	if options.GpioPin != 0 {
		opt.Channels[0].GpioPin = options.GpioPin
	}

	dev, err := ws2811.MakeWS2811(&opt)
	if err != nil {
		return nil, fmt.Errorf("ring: could not create ws2811 device: %w", err)
	}

	r := &Ring{
		device: dev,
		ledArc: 2 * math.Pi / float64(options.LedCount),
		opt:    options,
	}

	if err := r.device.Init(); err != nil {
		return nil, fmt.Errorf("ring: could not start ws2811 device: %w", err)
	}

	return r, nil
}

// Render updates the LED ring.
func (r *Ring) Render() error {
	pixels := make([]color.Color, r.Size())
	pixel := make([]color.Color, len(r.layers))

	for i := range r.device.Leds(0) {
		for j, l := range r.layers {
			switch l.Options().ContentMode {
			case ContentTile:
				pixel[j] = l.Pixel(i)
			case ContentCrop:
				if i < l.Options().Resolution {
					pixel[j] = l.Pixel(i)
				} else {
					pixel[j] = color.Transparent
				}
			case ContentScale:
				pixel[j] = l.Pixel(scale(i, r.Size(), l.Options().Resolution))
			}
		}
		pixels[i] = blendOver(pixel...)
	}
	rotInt := math.Floor(r.offset)
	rotFloat := r.offset - rotInt
	for i := range r.device.Leds(0) {
		r.device.Leds(0)[i] = serialize(lerp(int(rotInt)+i, pixels, rotFloat))
	}

	if err := r.device.Render(); err != nil {
		return err
	}

	return nil
}

func lerp(i int, pixels []color.Color, alpha float64) color.Color {
	return blendLerp(pixels[mod(i, len(pixels))], pixels[mod(i+1, len(pixels))], alpha)
}

// AddLayer adds a drawable layer to the ring.
func (r *Ring) AddLayer(l Pixeler) {
	r.layers = append(r.layers, l)
}

// Close turns off the LED ring and closes the device.
func (r *Ring) Close() {
	r.TurnOff()
	r.device.Fini()
}

// TurnOff tuns off the LED ring without closing the device.
func (r *Ring) TurnOff() {
	for i := range r.device.Leds(0) {
		r.device.Leds(0)[i] = 0
	}
	r.device.Render()
}

// Size returns the total number of LEDs of the ring.
func (r *Ring) Size() int {
	return r.opt.LedCount
}

// Offset sets an angular offset (in radians) to render the layers.
// A positive angle rotates counter-clockwise.
func (r *Ring) Offset(rotation float64) {
	if rotation < 0 {
		r.ledOffset = int(math.Ceil(rotation / r.ledArc))
	} else {
		r.ledOffset = int(math.Floor(rotation / r.ledArc))
	}
	r.offset = rotation / r.ledArc
}

func scale(v, fmax, tmax int) int {
	return v * tmax / fmax
}
