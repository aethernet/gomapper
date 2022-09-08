package main

import (
	"math"
	"unsafe"

	"github.com/go-gl/gl/v2.1/gl"
	"golang.org/x/exp/slices"
)

//keep the list of universe in the order they appears in the mask
var universeMapping []int

func distributeEncodePixelsForLine(start []int64, end []int64, quantity int64) [][4]byte {
	var pixels [][4]byte

	for key := 0; key < int(quantity); key++ {
		incX := int64(end[0] - start[0]) / quantity - 1.
		incY := int64(end[1] - start[1]) / quantity - 1.
		
		x := start[0] + int64(key) * incX
		y := start[1] + int64(key) * incY

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
	var mask [512*maxUniverse]byte
	for i := range mask {
    mask[i] = 255
	}

	for _, fixture := range fixtures {
		pixels := distributeEncodePixelsForLine(fixture.start, fixture.end, fixture.pixelCount)
		offset := fixture.pixelAddressStartsAt
		universe := fixture.universe

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
			pixelOffset := int(offset) + pixelKey * 3
			universeOffset := int(universe) - 1 + int(math.Floor(float64(pixelOffset) / 510.))

			// we keep track of which universe we've used for which line
			// when pushing to sacn, for each pack of 512 value, we'll get corresponding universe from that table
			universeExist := slices.IndexFunc(universeMapping, func(u int) bool { return u == universeOffset })
			if universeExist == -1 {
				universeMapping = append(universeMapping, universeOffset)
			}

			maskIndex := int(universe - 1) * 512 + pixelOffset - 1

			mask[maskIndex] = pixel[0]
			mask[maskIndex + 1] = pixel[1]
			mask[maskIndex + 2] = pixel[2]
			mask[maskIndex + 3] = pixel[3]
		}
	}

	texture := newTextureFromBytes(mask)

	return texture
}

// load a bytearray and return a 170 x maxUniverse rgba pixels texture
func newTextureFromBytes(rgba [512*maxUniverse]byte) uint32 {

	var texture uint32
	gl.GenTextures(1, &texture)
	gl.ActiveTexture(gl.TEXTURE0)
	gl.BindTexture(gl.TEXTURE_2D, texture)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MIN_FILTER, gl.LINEAR)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MAG_FILTER, gl.LINEAR)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_WRAP_S, gl.CLAMP_TO_EDGE)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_WRAP_T, gl.CLAMP_TO_EDGE)
	gl.TexImage2D(
		gl.TEXTURE_2D,
		0,
		gl.RGBA,
		int32(170),
		int32(maxUniverse),
		0,
		gl.RGBA,
		gl.UNSIGNED_BYTE,
		unsafe.Pointer(&rgba),
	)

	return texture
}
