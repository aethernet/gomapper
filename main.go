package main

import (
	"fmt"
	"runtime"

	"github.com/Hundemeier/go-sacn/sacn"
	"github.com/go-gl/gl/v2.1/gl"
	"github.com/go-gl/glfw/v3.2/glfw"
)

const (
	width  = 800
	height = 600
	fps = 60
)

var (
    square = []float32{
			-1., -1., 0., // top left point
			1., 1., 0., // top right point
			-1, 1, 0., // bottom left point
			-1, -1, 0., // top left point
			1., 1., 0., // top right point
			1., -1., 0., // bottom right point
    }
    previousTime float64
    runTime float64 = 0.
		currentScreenFramebufferRendered int32 = 1
		ips = []string{"192.168.178.40"}
		transmitter sacn.Transmitter
		universeMapping []int // will keep the relation line in the mapping texture -> universe to output
		mappingTexture uint32
)

func main() {
	runtime.LockOSThread()

	window := initGlfw()
	defer glfw.Terminate()

	initOpenGL()

	// load fixtures and create mapping texture
	fixtures, err := decodeFixtureFromJson("fixtures.json")
	if err != nil {
		panic("Fixtures json not found or unreadable")
	}
	
	mappingTexture = newTextureFromFixtures(fixtures)

	transmitter = initSACNTransmitter()

	//beware ! setupChannels needs to be called AFTER fixtures has been loaded
	setupChannels(ips, universeMapping, transmitter)

	screenProgram := newScreenProgram()

	
	mappingProgram, mappingShaderTex, mappingFramebuffer := newFramebufferProgram(170, int32(len(universeMapping)), "default.vert", "mapping.frag")
	shaderOneProgram, shaderOneShaderTex, shaderOneShaderFramebuffer := newFramebufferProgram(width, height, "default.vert", "shaderOne.frag")
	
	// configure the vertex data for a full screen square
	vao := makeVao(square)

	// init time
	previousTime = glfw.GetTime()

	// register keyboard
	window.SetKeyCallback(keyCallback)

	for !window.ShouldClose() {	

		// throttle to x fps (screen and sacn output alike)
		time := getUpdateTime()
		if( time >= lastTransmission + 1./fps) {
			lastTransmission = time // next frame
		} else {
			// fmt.Println(time)
			continue // not yet
		}

		// render shaderOne to TEXTURE1
		drawShaderToFramebuffer(shaderOneProgram, shaderOneShaderTex, shaderOneShaderFramebuffer, width, height, vao, gl.TEXTURE1, int32(-1))
		
		// render mappingShader to TEXTURE2
		drawShaderToFramebuffer(mappingProgram, mappingShaderTex, mappingFramebuffer, width, height, vao, gl.TEXTURE2, int32(shaderOneShaderTex))

		// extract and print mapped pixels from latest rendered shader (Mapping Shader)
		// FIXME: getting the pixels from the "output" framebuffer works, while taking them from the mapping framebuffer doesn't
		extractAndSendMappedPixelsFrom(0)

		/** go to screen **/
		renderFramebufferToScreen(screenProgram, vao, currentScreenFramebufferRendered)

		// Maintenance
		window.SwapBuffers()
		glfw.PollEvents()
	}
}

// toggle screen output between ShaderOne and Output Shader
func toggleView(){
	if currentScreenFramebufferRendered == 2 {
		currentScreenFramebufferRendered = 1
		fmt.Printf("Change view to : Shader 1\n")
	} else {
		currentScreenFramebufferRendered = 2
		fmt.Printf("Change view to : Mapping Output\n")
	}
}

// initGlfw initializes glfw and returns a Window to use.
// cannot move this function to another file without breaking the app
func initGlfw() *glfw.Window {
	if err := glfw.Init(); err != nil {
		panic(err)
	}
	
	glfw.WindowHint(glfw.Resizable, glfw.False)
	glfw.WindowHint(glfw.ContextVersionMajor, 4) // OR 2
	glfw.WindowHint(glfw.ContextVersionMinor, 1)
	glfw.WindowHint(glfw.OpenGLProfile, glfw.OpenGLCoreProfile)
	glfw.WindowHint(glfw.OpenGLForwardCompatible, glfw.True)

	window, err := glfw.CreateWindow(width, height, "Go shader go go go", nil, nil)
	if err != nil {
		panic(err)
	}
	window.MakeContextCurrent()

	return window
}