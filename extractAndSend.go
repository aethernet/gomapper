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
func extractAndSendMappedPixels () {	
	time := getUpdateTime()
	if( time >= lastTransmission + 1./fps) {
		lastTransmission = time // next frame
	} else {
		// fmt.Println(time)
		return // not yet
	}

	for key := range universeMapping {
		// read by 512 pixels

		var pixels [512]byte
		runtime.KeepAlive(pixels)
		gl.ReadPixels(0, int32(key), 170, int32(key+1), gl.RGB, gl.UNSIGNED_BYTE, unsafe.Pointer(&pixels))
		
		// select sender for universe
		sacn := channels[key]

		// // // this debug is quite expensive so turning it of while debugging other stuffs
		// fmt.Println("sending universe", key, universe)

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