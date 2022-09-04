package main

import (
	"fmt"
	"image"
	"image/draw"
	_ "image/png"
	"io/ioutil"
	"log"
	"os"
	"runtime"
	"strings"

	"github.com/go-gl/gl/v2.1/gl"
	"github.com/go-gl/glfw/v3.2/glfw"
)

const (
	width  = 800
	height = 600
	mappingWidth = 20
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
		pixels [width*height*4]byte // will serve as an output in an unsafe way
)

func main() {
	runtime.LockOSThread()

	window := initGlfw()
	defer glfw.Terminate()

	initOpenGL()

	screenProgram := newScreenProgram()

	mappingProgram, mappingShaderTex, mappingFramebuffer := newFramebufferProgram(mappingWidth, 1, "default.vert", "mapping.frag")
	passTextureFromFileTo(mappingProgram, "mask.png", gl.TEXTURE0)

	shaderOneProgram, shaderOneShaderTex, shaderOneShaderFramebuffer := newFramebufferProgram(width, height, "default.vert", "shaderOne.frag")

	// configure the vertex data
	vao := makeVao(square)

	// init time
	previousTime = glfw.GetTime()

	for !window.ShouldClose() {	

		// drawShaderToFramebuffer(mappingProgram, mappingShaderTex, mappingFramebuffer, mappingWidth, 1)

		drawShaderToFramebuffer(shaderOneProgram, shaderOneShaderTex, shaderOneShaderFramebuffer, width, height, vao, gl.TEXTURE1)
		
		// forward texture from shader1 to shader2
		passTextureUniform := gl.GetUniformLocation(mappingProgram, gl.Str("t_tex\x00"))
		gl.Uniform1i(passTextureUniform, 1)
		
		drawShaderToFramebuffer(mappingProgram, mappingShaderTex, mappingFramebuffer, mappingWidth, 1, vao, gl.TEXTURE2)

		//Bind the default framebuffer so that we can draw to screen
		gl.BindFramebufferEXT(gl.FRAMEBUFFER, 0);
		
		gl.Clear(gl.COLOR_BUFFER_BIT | gl.DEPTH_BUFFER_BIT)
		// gl.ActiveTexture(gl.TEXTURE0) // get back to texture 0

		// // extract mapping shader pixels to buffer
		// // printing is quite costly, so we skip that atm
		// gl.ReadPixels(0, 0, width, height, gl.RGBA, gl.UNSIGNED_BYTE, unsafe.Pointer(&pixels))
		// fmt.Println(pixels) 

		/** go to screen **/
		renderFramebufferToScreen(screenProgram, vao, 1)

		// Maintenance
		window.SwapBuffers()
		glfw.PollEvents()
	}
}
	
// pass TEXTURE1 to our screenProgram as we want to display on screen "Rendering shader" instead of MappingShader.
func renderFramebufferToScreen(screenProgram uint32, vao uint32, framebufferIndex int) {
	gl.UseProgram(screenProgram)

	// attach texture from latest shader in chain to screen
	textureUniform := gl.GetUniformLocation(screenProgram, gl.Str("t_tex\x00"));
	gl.Uniform1i(textureUniform, 1);

	resolutionUniform := gl.GetUniformLocation(screenProgram, gl.Str("u_resolution\x00"))
	gl.Uniform2f(resolutionUniform, float32(width), float32(height))
	
	gl.BindVertexArray(vao)
	gl.DrawArrays(gl.TRIANGLES, 0, int32(len(square) / 3))
}

func passTextureFromFileTo(program uint32, file string, GLTextureSlot uint32) uint32 {
	// Load mapping texture to slot 0
	texture, err := newTextureFromFile(file, GLTextureSlot)
	if err != nil {
		log.Fatalln(err)
	}

	// pass mask as a texture uniform on slot 0
	maskUniform := gl.GetUniformLocation(program, gl.Str("t_mask\x00"))
	gl.Uniform1i(maskUniform, 0)

	return texture
}

func getUpdateTime() float32 {
	time := glfw.GetTime()
	elapsed := time - previousTime
	previousTime = time

	runTime += elapsed

	return float32(runTime)
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
		
		// // activate texture
		// gl.ActiveTexture(gl.TEXTURE0)
		// gl.BindTexture(gl.TEXTURE_2D, mappingTexture)

		gl.BindVertexArray(vao)
		gl.DrawArrays(gl.TRIANGLES, 0, int32(len(square) / 3))

		// here we put the resulting texture in slot 1
		gl.ActiveTexture(GLtextureSlot)
		gl.BindTexture(gl.TEXTURE_2D, texture)

		// unbind framebuffer (back to default -> render to screen)
		gl.BindFramebuffer(gl.FRAMEBUFFER, 0)
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

// load texture file in texture slot0
func newTextureFromFile(file string, bindToTextureSlot uint32) (uint32, error) {
	imgFile, err := os.Open(file)
	if err != nil {
		return 0, fmt.Errorf("texture %q not found on disk: %v", file, err)
	}
	img, _, err := image.Decode(imgFile)
	if err != nil {
		return 0, err
	}

	rgba := image.NewRGBA(img.Bounds())
	if rgba.Stride != rgba.Rect.Size().X*4 {
		return 0, fmt.Errorf("unsupported stride")
	}
	draw.Draw(rgba, rgba.Bounds(), img, image.Point{0, 0}, draw.Src)

	var texture uint32
	gl.GenTextures(1, &texture)
	gl.ActiveTexture(bindToTextureSlot)
	gl.BindTexture(gl.TEXTURE_2D, texture)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MIN_FILTER, gl.LINEAR)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MAG_FILTER, gl.LINEAR)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_WRAP_S, gl.CLAMP_TO_EDGE)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_WRAP_T, gl.CLAMP_TO_EDGE)
	gl.TexImage2D(
		gl.TEXTURE_2D,
		0,
		gl.RGBA,
		int32(rgba.Rect.Size().X),
		int32(rgba.Rect.Size().Y),
		0,
		gl.RGBA,
		gl.UNSIGNED_BYTE,
		gl.Ptr(rgba.Pix))

	return texture, nil
}