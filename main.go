package main

import (
	"fmt"
	_ "image/png"
	"io/ioutil"
	"log"
	"runtime"
	"strings"
	"unsafe"

	"github.com/Hundemeier/go-sacn/sacn"
	"github.com/go-gl/gl/v2.1/gl"
	"github.com/go-gl/glfw/v3.2/glfw"
)

const (
	width  = 800
	height = 600
	mappingWidth = 120
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

	sacnChan := initsACNTranmitter([]string{"192.168.178.40"}, 1)

	window := initGlfw()
	defer glfw.Terminate()

	initOpenGL()

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
		
		drawShaderToFramebuffer(shaderOneProgram, shaderOneShaderTex, shaderOneShaderFramebuffer, width, height, vao, gl.TEXTURE1)
		
		// render mappingShader to TEXTURE2
		drawMappingShaderToFramebuffer(mappingProgram, mappingShaderTex, mappingFramebuffer, width, height, vao, gl.TEXTURE2)

		// extract and print mapped pixels from latest rendered shader (Mapping Shader)
		extractAndSendMappedPixels(mappingWidth, 1, sacnChan)

		/** go to screen **/
		renderFramebufferToScreen(screenProgram, vao, currentScreenFramebufferRendered)

		// Maintenance
		window.SwapBuffers()
		glfw.PollEvents()
	}
}

// keyboard callbacks
func keyCallback(w *glfw.Window, key glfw.Key, scancode int, action glfw.Action, mods glfw.ModifierKey) {
	//if u only want the on press, do = && && action == glfw.Press
	if key == glfw.KeySpace && action == glfw.Press {
		toggleView()
	}
	
	if key == glfw.KeyEscape && action == glfw.Press {
		w.SetShouldClose(true)
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

// extract mapped pixels from current framebuffer and send them using sacn
func extractAndSendMappedPixels (width int32, height int32, sacn chan<-[512]byte) {
	gl.ReadPixels(0, 0, width, height, gl.RGB, gl.UNSIGNED_BYTE, unsafe.Pointer(&pixels))
	// // this is quite expensive so turning it of while debugging other stuffs
	// fmt.Println(pixels)
	sacn<-pixels
}

// pass TEXTURE1 to our screenProgram as we want to display on screen "Rendering shader" instead of MappingShader.
func renderFramebufferToScreen(screenProgram uint32, vao uint32, framebufferIndex int32) {
	gl.BindFramebufferEXT(gl.FRAMEBUFFER, 0);
	gl.Clear(gl.COLOR_BUFFER_BIT | gl.DEPTH_BUFFER_BIT)
	
	gl.UseProgram(screenProgram)
	// passTextureFromFileTo(screenProgram, "mask.png", gl.TEXTURE0)
	newTextureFromBytes()

	// attach texture from shaderOne to screen
	textureUniform := gl.GetUniformLocation(screenProgram, gl.Str("t_tex\x00"));
	gl.Uniform1i(textureUniform, framebufferIndex);

	// pass mask texture to screen
	maskUniform := gl.GetUniformLocation(screenProgram, gl.Str("t_mask\x00"))
	gl.Uniform1i(maskUniform, 0)

	// boolean to show mask uniform or not
	showMaskUniform := gl.GetUniformLocation(screenProgram, gl.Str("u_showmask\x00"))
	gl.Uniform1i(showMaskUniform, 1)

	resolutionUniform := gl.GetUniformLocation(screenProgram, gl.Str("u_resolution\x00"))
	gl.Uniform2f(resolutionUniform, float32(width), float32(height))

	maskResolutionUniform := gl.GetUniformLocation(screenProgram, gl.Str("u_maskresolution\x00"))
	gl.Uniform2f(maskResolutionUniform, float32(mappingWidth), float32(1.))
	
	gl.BindVertexArray(vao)
	gl.DrawArrays(gl.TRIANGLES, 0, int32(len(square) / 3))
}

func getUpdateTime() float32 {
	time := glfw.GetTime()
	elapsed := time - previousTime
	previousTime = time

	runTime += elapsed

	return float32(runTime)
}

// render shaderOne to its framebuffer
func drawMappingShaderToFramebuffer(program uint32, texture uint32, framebuffer uint32, width int, height int, vao uint32, GLtextureSlot uint32) {
		// get time
		runTime := getUpdateTime()

		gl.BindFramebuffer(gl.FRAMEBUFFER, framebuffer)
		gl.Clear(gl.COLOR_BUFFER_BIT | gl.DEPTH_BUFFER_BIT)
		gl.UseProgram(program)
		
		// pass time and resolution as uniforms
		timeUniform := gl.GetUniformLocation(program, gl.Str("u_time\x00"))
		gl.Uniform1f(timeUniform, float32(runTime))
		
		resolutionUniform := gl.GetUniformLocation(program, gl.Str("u_resolution\x00"))
		gl.Uniform2f(resolutionUniform, float32(width), float32(height))

		maskResolutionUniform := gl.GetUniformLocation(program, gl.Str("u_maskresolution\x00"))
		gl.Uniform2f(maskResolutionUniform, float32(mappingWidth), float32(1.))

		// Attach TEXTURE1 (shaderOne) to _Mapping Shader_
		passTextureUniform := gl.GetUniformLocation(program, gl.Str("t_shaderOne\x00"))
		gl.Uniform1i(passTextureUniform, 1)

		gl.BindVertexArray(vao)
		gl.DrawArrays(gl.TRIANGLES, 0, int32(len(square) / 3))

		// here we put the resulting texture in slot 1
		gl.ActiveTexture(GLtextureSlot)
		gl.BindTexture(gl.TEXTURE_2D, texture)
}

// render shaderOne to its framebuffer
func drawShaderToFramebuffer(program uint32, texture uint32, framebuffer uint32, width int, height int, vao uint32, GLtextureSlot uint32) {
		// get time
		runTime := getUpdateTime()

		gl.BindFramebuffer(gl.FRAMEBUFFER, framebuffer)
		gl.Clear(gl.COLOR_BUFFER_BIT | gl.DEPTH_BUFFER_BIT)
		gl.UseProgram(program)
		
		// pass time and resolution as uniforms
		timeUniform := gl.GetUniformLocation(program, gl.Str("u_time\x00"))
		gl.Uniform1f(timeUniform, float32(runTime))
		
		resolutionUniform := gl.GetUniformLocation(program, gl.Str("u_resolution\x00"))
		gl.Uniform2f(resolutionUniform, float32(width), float32(height))

		maskResolutionUniform := gl.GetUniformLocation(program, gl.Str("u_maskresolution\x00"))
		gl.Uniform2f(maskResolutionUniform, float32(mappingWidth), float32(1.))

		gl.BindVertexArray(vao)
		gl.DrawArrays(gl.TRIANGLES, 0, int32(len(square) / 3))

		// here we put the resulting texture in slot 1
		gl.ActiveTexture(GLtextureSlot)
		gl.BindTexture(gl.TEXTURE_2D, texture)
}

func newScreenProgram() uint32 {
	var vertexShaderSource = readShaderFile("default.vert")    // the vertex shader
	var fragmentShaderSource = readShaderFile("screen.frag")		// the fragment shader
	
	program, err := newProgram(vertexShaderSource, fragmentShaderSource)
	if err != nil {
		panic(err)
	}

	return program
}

func newFramebufferProgram(width int32, height int32, vertexFile string, fragFile string,) (uint32, uint32, uint32) {
	var vertexShaderSource = readShaderFile(vertexFile)    // the vertex shader
	var fragmentShaderSource = readShaderFile(fragFile)		// the fragment shader

	program, err := newProgram(vertexShaderSource, fragmentShaderSource)
	if err != nil {
			panic(err)
	}

	// prepare texture for framebuffer
	var shaderTex uint32
	gl.GenTextures(1, &shaderTex)
	gl.BindTexture(gl.TEXTURE_2D, shaderTex);
	gl.TexImage2D(gl.TEXTURE_2D, 0, gl.RGB, width, height, 0, gl.RGB, gl.UNSIGNED_BYTE, nil);

	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MIN_FILTER, gl.LINEAR);
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MAG_FILTER, gl.LINEAR);

	// prepare framebuffer for shader
	var framebuffer uint32
	gl.GenFramebuffers(1, &framebuffer)
	gl.BindFramebuffer(gl.FRAMEBUFFER, framebuffer)

	gl.FramebufferTexture2D(gl.FRAMEBUFFER, gl.COLOR_ATTACHMENT0, gl.TEXTURE_2D, shaderTex, 0);

	if gl.CheckFramebufferStatus(gl.FRAMEBUFFER) != gl.FRAMEBUFFER_COMPLETE {
		panic("something's wrong with our framebuffer")
	}

	return program, shaderTex, framebuffer
}

// initGlfw initializes glfw and returns a Window to use.
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

// read file and return the contents with an ending \x00
func readShaderFile(fileName string) string {
	content, err := ioutil.ReadFile(fileName)
	if err != nil {fmt.Println("Err")}		
	return string(content) + "\x00"
}

// create a program from shader
func newProgram(vertexShaderSource, fragmentShaderSource string) (uint32, error) {
	vertexShader, err := compileShader(vertexShaderSource, gl.VERTEX_SHADER)
	if err != nil {
		return 0, err
	}

	fragmentShader, err := compileShader(fragmentShaderSource, gl.FRAGMENT_SHADER)
	if err != nil {
		return 0, err
	}

	program := gl.CreateProgram()

	gl.AttachShader(program, vertexShader)
	gl.AttachShader(program, fragmentShader)
	gl.LinkProgram(program)

	var status int32
	gl.GetProgramiv(program, gl.LINK_STATUS, &status)
	if status == gl.FALSE {
		var logLength int32
		gl.GetProgramiv(program, gl.INFO_LOG_LENGTH, &logLength)

		log := strings.Repeat("\x00", int(logLength+1))
		gl.GetProgramInfoLog(program, logLength, nil, gl.Str(log))

		return 0, fmt.Errorf("failed to link program: %v", log)
	}

	gl.DeleteShader(vertexShader)
	gl.DeleteShader(fragmentShader)

	return program, nil
}


// initOpenGL initializes OpenGL
func initOpenGL() {
    if err := gl.Init(); err != nil {
			panic(err)
    }

    version := gl.GoStr(gl.GetString(gl.VERSION))
    log.Println("OpenGL version", version)
}

// makeVao initializes and returns a vertex array from the points provided.
func makeVao(points []float32) uint32 {
    var vbo uint32
    gl.GenBuffers(1, &vbo)
    gl.BindBuffer(gl.ARRAY_BUFFER, vbo)
    gl.BufferData(gl.ARRAY_BUFFER, 4*len(points), gl.Ptr(points), gl.STATIC_DRAW)
    
    var vao uint32
    gl.GenVertexArrays(1, &vao)
    gl.BindVertexArray(vao)
    gl.EnableVertexAttribArray(0)
    gl.BindBuffer(gl.ARRAY_BUFFER, vbo)
    gl.VertexAttribPointer(0, 3, gl.FLOAT, false, 0, nil)
    
    return vao
}

func compileShader(source string, shaderType uint32) (uint32, error) {
    shader := gl.CreateShader(shaderType)
    
    csources, free := gl.Strs(source)
    gl.ShaderSource(shader, 1, csources, nil)
    free()
    gl.CompileShader(shader)
    
    var status int32
    gl.GetShaderiv(shader, gl.COMPILE_STATUS, &status)
    if status == gl.FALSE {
        var logLength int32
        gl.GetShaderiv(shader, gl.INFO_LOG_LENGTH, &logLength)
        
        log := strings.Repeat("\x00", int(logLength+1))
        gl.GetShaderInfoLog(shader, logLength, nil, gl.Str(log))
        
        return 0, fmt.Errorf("failed to compile %v: %v", source, log)
    }
    
    return shader, nil
}

// load texture from bytearray 
func newTextureFromBytes() uint32 {

	// hardcoded 120 pixel mapping mask [0:0] -> [800:600]
	// TODO: generate instead of hardcode
	rgba := [120*4]byte{0,0,0,0,0,6,0,5,0,13,0,10,0,20,0,15,0,26,0,20,0,33,0,25,0,40,0,30,0,47,0,35,0,53,0,40,0,60,0,45,0,67,0,50,0,73,0,55,0,80,0,60,0,87,0,65,0,94,0,70,0,100,0,75,0,107,0,80,0,114,0,85,0,121,0,90,0,127,0,95,0,134,0,100,0,141,0,105,0,147,0,110,0,154,0,115,0,161,0,121,0,168,0,126,0,174,0,131,0,181,0,136,0,188,0,141,0,194,0,146,0,201,0,151,0,208,0,156,0,215,0,161,0,221,0,166,0,228,0,171,0,235,0,176,0,242,0,181,0,248,0,186,0,255,0,191,1,6,0,196,1,12,0,201,1,19,0,206,1,26,0,211,1,33,0,216,1,39,0,221,1,46,0,226,1,53,0,231,1,59,0,236,1,66,0,242,1,73,0,247,1,80,0,252,1,86,1,1,1,93,1,6,1,100,1,11,1,107,1,16,1,113,1,21,1,120,1,26,1,127,1,31,1,133,1,36,1,140,1,41,1,147,1,46,1,154,1,51,1,160,1,56,1,167,1,61,1,174,1,66,1,180,1,71,1,187,1,76,1,194,1,81,1,201,1,86,1,207,1,91,1,214,1,96,1,221,1,101,1,228,1,107,1,234,1,112,1,241,1,117,1,248,1,122,1,254,1,127,2,5,1,132,2,12,1,137,2,19,1,142,2,25,1,147,2,32,1,152,2,39,1,157,2,45,1,162,2,52,1,167,2,59,1,172,2,66,1,177,2,72,1,182,2,79,1,187,2,86,1,192,2,93,1,197,2,99,1,202,2,106,1,207,2,113,1,212,2,119,1,217,2,126,1,222,2,133,1,228,2,140,1,233,2,146,1,238,2,153,1,243,2,160,1,248,2,166,1,253,2,173,2,2,2,180,2,7,2,187,2,12,2,193,2,17,2,200,2,22,2,207,2,27,2,214,2,32,2,220,2,37,2,227,2,42,2,234,2,47,2,240,2,52,2,247,2,57,2,254,2,62,3,5,2,67,3,11,2,72,3,18,2,77,3,25,2,82,3,32,2,88}

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
		int32(120),
		int32(1),
		0,
		gl.RGBA,
		gl.UNSIGNED_BYTE,
		unsafe.Pointer(&rgba),
	)

	return texture
}

// // load texture file in texture slot0
// func newTextureFromFile(file string, bindToTextureSlot uint32) (uint32, error) {
// 	imgFile, err := os.Open(file)
// 	if err != nil {
// 		return 0, fmt.Errorf("texture %q not found on disk: %v", file, err)
// 	}
// 	img, _, err := image.Decode(imgFile)
// 	if err != nil {
// 		return 0, err
// 	}

// 	rgba := image.NewRGBA(img.Bounds())
// 	if rgba.Stride != rgba.Rect.Size().X*4 {
// 		return 0, fmt.Errorf("unsupported stride")
// 	}
// 	draw.Draw(rgba, rgba.Bounds(), img, image.Point{0, 0}, draw.Src)

// 	var texture uint32
// 	gl.GenTextures(1, &texture)
// 	gl.ActiveTexture(bindToTextureSlot)
// 	gl.BindTexture(gl.TEXTURE_2D, texture)
// 	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MIN_FILTER, gl.LINEAR)
// 	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MAG_FILTER, gl.LINEAR)
// 	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_WRAP_S, gl.CLAMP_TO_EDGE)
// 	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_WRAP_T, gl.CLAMP_TO_EDGE)
// 	gl.TexImage2D(
// 		gl.TEXTURE_2D,
// 		0,
// 		gl.RGBA,
// 		int32(rgba.Rect.Size().X),
// 		int32(rgba.Rect.Size().Y),
// 		0,
// 		gl.RGBA,
// 		gl.UNSIGNED_BYTE,
// 		gl.Ptr(rgba.Pix))

// 	return texture, nil
// }

func initsACNTranmitter(ips []string, universe uint16) chan<-[512]byte {
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