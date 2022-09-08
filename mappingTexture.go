package main

import (
	"fmt"
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
		// so our width will always be 128 (512 / 4) which is the limit of pixel we can address on a single universe
		// and our height depends on the number of universe
		// a white pixel means we don't need data for that pixel and it will be skipped
		// note : it prevents us to send one pixel to two universe but we can live with that i suppose
		// to decode we'll simply split our output by 512 and send to each universe

		for pixelKey, pixel := range pixels {
			pixelOffset := int(offset) + pixelKey * 4
			universeOffset := int(universe) - 1 + int(math.Floor(float64(pixelOffset) / 511.))

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

// load texture from bytearray
func newTextureFromBytes(rgba [512*maxUniverse]byte) uint32 {

	// hardcoded 120 pixel mapping mask [0:0] -> [800:600]
	// TODO: generate instead of hardcode
	// rgba := [mappingWidth*4]byte{0,0,0,0,0,6,0,5,0,13,0,10,0,20,0,15,0,26,0,20,0,33,0,25,0,40,0,30,0,47,0,35,0,53,0,40,0,60,0,45,0,67,0,50,0,73,0,55,0,80,0,60,0,87,0,65,0,94,0,70,0,100,0,75,0,107,0,80,0,114,0,85,0,121,0,90,0,127,0,95,0,134,0,100,0,141,0,105,0,147,0,110,0,154,0,115,0,161,0,121,0,168,0,126,0,174,0,131,0,181,0,136,0,188,0,141,0,194,0,146,0,201,0,151,0,208,0,156,0,215,0,161,0,221,0,166,0,228,0,171,0,235,0,176,0,242,0,181,0,248,0,186,0,255,0,191,1,6,0,196,1,12,0,201,1,19,0,206,1,26,0,211,1,33,0,216,1,39,0,221,1,46,0,226,1,53,0,231,1,59,0,236,1,66,0,242,1,73,0,247,1,80,0,252,1,86,1,1,1,93,1,6,1,100,1,11,1,107,1,16,1,113,1,21,1,120,1,26,1,127,1,31,1,133,1,36,1,140,1,41,1,147,1,46,1,154,1,51,1,160,1,56,1,167,1,61,1,174,1,66,1,180,1,71,1,187,1,76,1,194,1,81,1,201,1,86,1,207,1,91,1,214,1,96,1,221,1,101,1,228,1,107,1,234,1,112,1,241,1,117,1,248,1,122,1,254,1,127,2,5,1,132,2,12,1,137,2,19,1,142,2,25,1,147,2,32,1,152,2,39,1,157,2,45,1,162,2,52,1,167,2,59,1,172,2,66,1,177,2,72,1,182,2,79,1,187,2,86,1,192,2,93,1,197,2,99,1,202,2,106,1,207,2,113,1,212,2,119,1,217,2,126,1,222,2,133,1,228,2,140,1,233,2,146,1,238,2,153,1,243,2,160,1,248,2,166,1,253,2,173,2,2,2,180,2,7,2,187,2,12,2,193,2,17,2,200,2,22,2,207,2,27,2,214,2,32,2,220,2,37,2,227,2,42,2,234,2,47,2,240,2,52,2,247,2,57,2,254,2,62,3,5,2,67,3,11,2,72,3,18,2,77,3,25,2,82,3,32,2,88}

	fmt.Println(len(rgba))

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
		int32(128),
		int32(maxUniverse),
		0,
		gl.RGBA,
		gl.UNSIGNED_BYTE,
		unsafe.Pointer(&rgba),
	)

	fmt.Println("texture done")

	return texture
}
