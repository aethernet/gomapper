package main

import (
	"log"
	"runtime"
	"unsafe"

	"github.com/Hundemeier/go-sacn/sacn"
	"github.com/go-gl/gl/v2.1/gl"
)

// extract mapped pixels from current framebuffer and send them using sacn

// Beware !
// Due to the unsafe pointer if we try to read and write at the same time, we'll cause a segfault
// To prevent that (and save our network) we need to throttle the update
// We do this by capping transmission at a fixed fps rate defined in a main const

var lastTransmission float32 = 0.
func extractAndSendMappedPixelsFrom (framebuffer uint32) {	
	
	// set source framebuffer (reset when leaving)
	gl.BindFramebuffer(gl.FRAMEBUFFER, framebuffer)
	defer gl.BindFramebuffer(gl.FRAMEBUFFER, 0)

	for key := range universeMapping {
		// read by 512 pixels

		var pixels [512]byte
		runtime.KeepAlive(pixels) // absolutely needed to prevent pixel to be garbage collected before they're read by C, 
		
		// here we take 170 pixels at a time (1 universe) as it's in RGB it will be 510 bytes, which we store in our 512 bytes array
		// Note that the texture we extract from is RGBA so each lines are 680 bytes, opengl should extract just what we need
		// gl.ActiveTexture(gl.TEXTURE0)
		// gl.BindTexture(gl.TEXTURE_2D, mappingTexture)
		gl.ReadPixels(0, 1, 170, 1, gl.RGB, gl.UNSIGNED_BYTE, unsafe.Pointer(&pixels))
		
		// just to be sure, we'll black out the last two bytes before sending
		pixels[510] = 0;
		pixels[511] = 0;

		// select sender for universe
		sacn := channels[key]

		// // this debug is quite expensive so turning it of while debugging other stuffs
		// fmt.Println("sending universe", key, universe)
		// fmt.Println("sending pixels", key, pixels)

		// sending the pixels 
		sacn<-pixels

	}
}

func initSACNTransmitter() sacn.Transmitter {
	transmitter, err := sacn.NewTransmitter("", [16]byte{1, 2, 3}, "test")
	if err != nil {
		log.Fatal(err)
	}
	
	return transmitter
}

func activateChannel(ips []string, universe uint16, transmitter sacn.Transmitter) chan<-[512]byte {
	// activate universe
	channel, err := transmitter.Activate(universe)
	if err != nil {
		log.Fatal(err)
	}

	// set destination (unicast but possibly to multiple dest)
	transmitter.SetDestinations(universe, ips)

	return channel
}

var channels []chan<-[512]byte
func setupChannels(ips []string, universes []int, transmitter sacn.Transmitter){
	for _, universe := range universes {
		channels = append(channels, activateChannel(ips, uint16(universe), transmitter))
	}
}