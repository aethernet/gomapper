package main

import (
	"fmt"
	_ "image/png"
	"runtime"

	"github.com/go-gl/gl/v2.1/gl"
	"github.com/go-gl/glfw/v3.2/glfw"
)

const (
	width  = 800
	height = 600
	mappingWidth = 120
	maxUniverse = 16
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
		pixels [512]byte // will serve as an output in an unsafe way // TODO: this is only one universe, will have to be creative for multiuniverse
		currentScreenFramebufferRendered int32 = 1
)

func main() {
	runtime.LockOSThread()

	sacnChan := initsACNTransmitter([]string{"192.168.178.40"}, 1)

	window := initGlfw()
	defer glfw.Terminate()

	initOpenGL()

	// load fixtures and create mapping texture
	fixtures, err := decodeFixtureFromJson("fixtures.json")
	if err != nil {
		panic("Fixtures json not found or unreadable")
	}
	newTextureFromFixtures(fixtures)

	screenProgram := newScreenProgram()

	mappingProgram, mappingShaderTex, mappingFramebuffer := newFramebufferProgram(mappingWidth, 1, "default.vert", "mapping.frag")

	shaderOneProgram, shaderOneShaderTex, shaderOneShaderFramebuffer := newFramebufferProgram(width, height, "default.vert", "shaderOne.frag")

	// configure the vertex data for a full screen square
	vao := makeVao(square)

	// init time
	previousTime = glfw.GetTime()

	// register keyboard
	window.SetKeyCallback(keyCallback)

	for !window.ShouldClose() {	

		// render shaderOne to TEXTURE1
		drawShaderToFramebuffer(shaderOneProgram, shaderOneShaderTex, shaderOneShaderFramebuffer, width, height, vao, gl.TEXTURE1, -1)
		
		// render mappingShader to TEXTURE2
		drawShaderToFramebuffer(mappingProgram, mappingShaderTex, mappingFramebuffer, width, height, vao, gl.TEXTURE2, 1)

		// extract and print mapped pixels from latest rendered shader (Mapping Shader)
		extractAndSendMappedPixels(mappingWidth, 1, sacnChan)

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