package main

import (
	"log"
	"unsafe"

	"github.com/Hundemeier/go-sacn/sacn"
	"github.com/go-gl/gl/v2.1/gl"
)

// extract mapped pixels from current framebuffer and send them using sacn

// TODO : we have to map to the correct universe here

func extractAndSendMappedPixels (width int32, height int32, sacn chan<-[512]byte) {
	gl.ReadPixels(0, 0, width, height, gl.RGB, gl.UNSIGNED_BYTE, unsafe.Pointer(&pixels))
	// // this is quite expensive so turning it of while debugging other stuffs
	// fmt.Println(pixels)
	sacn<-pixels
}

func initsACNTransmitter(ips []string, universe uint16) chan<-[512]byte {
	transmitter, err := sacn.NewTransmitter("", [16]byte{1, 2, 3}, "test")
	if err != nil {
		log.Fatal(err)
	}

	//activates the first universe
	channel, err := transmitter.Activate(1)
	if err != nil {
		log.Fatal(err)
	}
	//deactivate the channel on exit
	// defer close(channel)

	//set a unicast destination, and/or use multicast
	// trans.SetMulticast(1, true)//this specific setup will not multicast on windows, 
	//because no bind address was provided
	
	//set some example ip-addresses
	transmitter.SetDestinations(universe, ips)

	return channel
}