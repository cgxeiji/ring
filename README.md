# Ring

[![Version](https://img.shields.io/github/v/tag/cgxeiji/ring?sort=semver)](https://github.com/cgxeiji/ring/releases)
[![PkgGoDev](https://pkg.go.dev/badge/github.com/cgxeiji/ring)](https://pkg.go.dev/github.com/cgxeiji/ring)
[![License](https://img.shields.io/github/license/cgxeiji/ring)](https://github.com/cgxeiji/ring/blob/master/LICENSE)
![Go version](https://img.shields.io/github/go-mod/go-version/cgxeiji/ring)

Ring is a wrapper of
[rpi-ws281x-go](https://github.com/rpi-ws281x/rpi-ws281x-go) specialized in
controlling ring-shaped LEDs.

The library adds the ability to use `layers` to do complex animations. Each
`layer` supports color transparency and blending is handled automatically.

## Compilation

Compiling directly on a Raspberry Pi might take too long. The recommended way
to compile this library is to cross-compile using a Docker container.

Follow the [guide from
rpi-ws281x-go](https://github.com/rpi-ws281x/rpi-ws281x-go#cross-compiling) to
learn how.

### Running

Because `rpi-ws281x` needs to access `/dev/mem` to create correct pwm timings,
you will need to run the compiled binary with root permissions.


## Example

> *Boss: "Can you make a pulsating white background with colorful rotating lights and a blinking cyan light?"*
>
> *You:*
>
> ![ring.gif](./img/ring.gif)

To create the animation above, try the following code:

```go
package main

import (
	"bufio"
	"fmt"
	"image/color"
	"log"
	"math"
	"os"
	"sync"
	"time"

	"github.com/cgxeiji/ring"
)

func main() {
	// Initialize the ring.
	r, err := ring.New(&ring.Options{
		LedCount:      12,  // adjust this to the number of LEDs you have
		MaxBrightness: 180, // value from 0 to 255
	})
	r.Offset(-math.Pi / 3) // you can set a rotation offset for the ring
	if err != nil {
		log.Fatal(err)
	}
	// Make sure to properly close the ring.
	defer r.Close()

	// Create a new layer.  This will be a static white background.
	bg, err := ring.NewLayer(&ring.LayerOptions{
		Resolution:  1,                 // set to 1 pixel because it is a uniform color background
		ContentMode: ring.ContentScale, // this scales the pixel to the whole ring
	})
	if err != nil {
		log.Fatal(err)
	}
	// Set all pixels of the layer to white.
	bg.SetAll(color.White)
	// Add the layer to the ring.
	r.AddLayer(bg)

	// Create a mask layer.  This will fade the background.
	bgMask, err := ring.NewLayer(&ring.LayerOptions{
		Resolution: 1, // set to 1 pixel because it is a uniform mask
	})
	if err != nil {
		log.Fatal(err)
	}
	r.AddLayer(bgMask)

	// Render the ring.
	if err := r.Render(); err != nil {
		log.Fatal(err)
	}

	// Wait for 1 second to see the beauty of the freshly rendered layer.
	time.Sleep(1 * time.Second)

	// Create another layer.  This will set 3 pixels to red, green and blue,
	// and a hidden purple pixel with transparency of 200, that rotate
	// counter-clockwise.
	triRotate, err := ring.NewLayer(&ring.LayerOptions{
		Resolution: 48,
	})
	if err != nil {
		log.Fatal(err)
	}
	// We can immediately add the layer to the ring.  By default, new layers
	// are initialized with transparent pixels.  The new layer is added on top
	// of the previous layers.
	r.AddLayer(triRotate)

	// Set the colors.
	triRotate.SetPixel(0, color.NRGBA{128, 0, 0, 200})    // dark red
	triRotate.SetPixel(3, color.NRGBA{0, 128, 0, 200})    // dark green
	triRotate.SetPixel(6, color.NRGBA{0, 0, 128, 200})    // dark blue
	triRotate.SetPixel(24, color.NRGBA{128, 0, 255, 200}) // purple
	// Render the ring.
	if err := r.Render(); err != nil {
		log.Fatal(err)
	}

	// Wait for 1 second to see the beauty of both layers.
	time.Sleep(1 * time.Second)

	// Create another layer. This will set a pixel that will blink every 500ms.
	blink, err := ring.NewLayer(&ring.LayerOptions{
		Resolution:  3,                // we are going to set the 3rd pixel (index: 2) to blink
		ContentMode: ring.ContentCrop, // this crops the layer to avoid repetition
	})
	if err != nil {
		log.Fatal(err)
	}
	// Add the layer to the ring. This will be on top of the previous two
	// layers.
	r.AddLayer(blink)

	// Set the color. We can use any variable that implements the color.Color
	// interface.
	blink.SetPixel(2, color.CMYK{255, 0, 0, 0})
	// Render the ring.
	if err := r.Render(); err != nil {
		log.Fatal(err)
	}

	// Wait for 1 second and enjoy the view.
	time.Sleep(1 * time.Second)

	/* ANIMATION SETUP */
	done := make(chan struct{})   // this will cancel all animations
	render := make(chan struct{}) // this will request a concurrent-safe render
	var ws sync.WaitGroup         // this makes sure we close all goroutines

	/* render goroutine */
	ws.Add(1)
	go func() {
		defer ws.Done()
		for {
			select {
			case <-done:
				return
			case <-render:
				if err := r.Render(); err != nil {
					log.Fatal(err)
				}
			}
		}
	}()

	/* fading goroutine */
	ws.Add(1)
	go func() {
		defer ws.Done()
		c := color.NRGBA{0, 0, 0, 0}
		step := uint8(5)
		for {
			for a := uint8(0); a < 255; a += step {
				c.A = a
				bgMask.SetAll(c)
				select {
				case <-done:
					return
				case render <- struct{}{}:
				}
				time.Sleep(20 * time.Millisecond)
			}
			for a := uint8(255); a > 0; a -= step {
				c.A = a
				bgMask.SetAll(c)
				select {
				case <-done:
					return
				case render <- struct{}{}:
				}
				time.Sleep(20 * time.Millisecond)
			}
		}
	}()

	/* rotation goroutine */
	ws.Add(1)
	go func() {
		defer ws.Done()
		for {
			for a := 0.0; a < math.Pi*2; a += 0.01 {
				triRotate.Rotate(a)
				select {
				case <-done:
					return
				case render <- struct{}{}:
				}
				time.Sleep(20 * time.Millisecond)
			}
		}
	}()

	/* blinking goroutine */
	ws.Add(1)
	go func() {
		defer ws.Done()
		c := color.CMYK{255, 0, 0, 0}
		timer := time.NewTicker(500 * time.Millisecond)
		on := true
		for {
			select {
			case <-done:
				return
			case <-timer.C:
				if on {
					blink.SetPixel(2, color.Transparent)
					on = false
				} else {
					blink.SetPixel(2, c)
					on = true
				}
				select {
				case <-done:
					return
				case render <- struct{}{}:
				}
			}
		}
	}()

	fmt.Println("Press [ENTER] to exit")
	stdin := bufio.NewReader(os.Stdin)
	stdin.ReadString('\n')

	// Stop all animations
	close(done)
	// Wait for goroutines to exit
	ws.Wait()

	// Remember that we called a defer `r.Close()` at the beginning of the
	// code. This will turn off the LEDs and clean up the resources used by the
	// ring before exiting. Otherwise, the ring will stay on with the latest
	// render.
}
```
