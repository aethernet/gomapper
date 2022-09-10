package main

import (
	"fmt"
	"math"
	"reflect"
	"unsafe"

	"github.com/go-gl/gl/v2.1/gl"
	"golang.org/x/exp/slices"
)

func distributeEncodePixelsForLine(start []int64, end []int64, quantity int64) [][4]byte {
	var pixels [][4]byte

	for key := 0; key < int(quantity); key++ {
		incX := float64(end[0] - start[0]) / float64(quantity - 1)
		incY := float64(end[1] - start[1]) / float64(quantity - 1)
		
		x := float64(start[0]) + float64(key) * incX
		y := float64(start[1]) + float64(key) * incY

		pixel := [2]int64{int64(x), int64(y)}

		pixels = append(pixels, encodePositionAsColor(pixel))
	}

	return pixels
}

func encodePositionAsColor (pixel [2]int64) [4]byte {
	intPixel := int((pixel[0] << 16) | (pixel[1] >> 0))

	r := byte((intPixel & 0xFF000000) >> 24)
	g := byte((intPixel & 0x00FF0000) >> 16)
	b := byte((intPixel & 0x0000FF00) >> 8)
	a := byte((intPixel & 0x000000FF))
	
	return [4]byte{r, g, b, a}
} 

func newTextureFromFixtures(fixtures []Fixture) uint32  {
	// initialise our mask with all white pixels (we'll skip those values when computing our mask in opengl)
	// even if our mask has no pre-defined size, we only fill it up to 680 (170 pixels rgba) * our mapped universe quantity
	var mask []byte

	for _, fixture := range fixtures {
		pixels := distributeEncodePixelsForLine(fixture.start, fixture.end, fixture.pixelCount)
		offset := fixture.pixelAddressStartsAt
		startUniverse := fixture.universe

		// we need to order the pixels in the order of the output
		// we'll do a new mapping : 
		// each line of the map represent a universe
		// each column a pixel in that universe
		// so our width will always be 170 (512 / 3) which is the limit of pixel we can address on a single universe
		// and our height depends on the number of universe
		// a white pixel means we don't need data for that pixel and it will be skipped
		// note : it prevents us to send one pixel to two universe but we can live with that i suppose
		// to decode we'll simply split our output by 512 and send to each universe

		for pixelKey, pixel := range pixels {
			universeOffset := int(math.Floor(float64(pixelKey) / 170.))

			// pixel first byte (r)
			// = offset (where to start in the universe)
			// + key (posisition of the pixel in the chain) * 4
			pixelOffset := int(offset) + pixelKey * 4

			universe := startUniverse + int64(universeOffset)
			
			// we keep track of which universe we've used for which line, a universe in mapping texture is 680 bytes (170 pixels rgba)
			// when pushing to sacn, for each pack of 512 bytes (170 pixels in rgb), we'll get corresponding universe from that table
			universeExist := slices.IndexFunc(universeMapping, func(u int) bool { return u == int(universe) })
			
			if universeExist == -1 {
				// if our universe is not mapped yet, let's map it and prepare a black line in the mapping texture
				for i := 0; i < 680; i++ {
					mask = append(mask, 255)
				}
				universeMapping = append(universeMapping, int(universe))
			}
			
			// 1 universe in RGBA is 680 bytes (4*170) while in RGB it will be only 510 bytes (3x170)
			// our mask is in rgba so we multiply by 680
			maskIndex := int(universe - int64(universeOffset) - 1) * 680 + pixelOffset - 1

			mask[maskIndex] = pixel[0]
			mask[maskIndex + 1] = pixel[1]
			mask[maskIndex + 2] = pixel[2]
			mask[maskIndex + 3] = pixel[3]
		}
	}

	fmt.Println(mask)

	texture := newTextureFromBytes(mask)

	return texture
}

// load a bytearray and return a 170 x maxUniverse rgba pixels texture
func newTextureFromBytes(rgba []byte) uint32 {

	// lot of unsafe in here : https://stackoverflow.com/questions/36706843/how-to-get-the-underlying-array-of-a-slice-in-go
	// but it seems like it's fine as we only do this once to copy all those bytes into the gpu memory
	hdr := (*reflect.SliceHeader)(unsafe.Pointer(&rgba))
	data := *(*[4]int)(unsafe.Pointer(hdr.Data))

	var texture uint32
	gl.GenTextures(1, &texture)
	gl.ActiveTexture(gl.TEXTURE0)
	gl.BindTexture(gl.TEXTURE_2D, texture)
	// gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MIN_FILTER, gl.LINEAR)
	// gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MAG_FILTER, gl.LINEAR)
	// gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_WRAP_S, gl.CLAMP_TO_EDGE)
	// gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_WRAP_T, gl.CLAMP_TO_EDGE)
	gl.TexImage2D(
		gl.TEXTURE_2D,
		0,
		gl.RGBA,
		int32(170),
		int32(len(universeMapping)),
		0, 
		gl.RGBA,
		gl.UNSIGNED_BYTE,
		unsafe.Pointer(&data),
	)

	return texture
}
