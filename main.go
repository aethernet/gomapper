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
	"unsafe"

	"github.com/go-gl/gl/v2.1/gl"
	"github.com/go-gl/glfw/v3.2/glfw"
)

const (
	width  = 800
	height = 600
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

	var vertexShaderSource = readShaderFile("shaderOne.vert")    // the vertex shader
	var fragmentShaderSource = readShaderFile("shaderOne.frag")		// the fragment shader
	var mappingShaderSource = readShaderFile("shaderTwo.frag")		// the fragment shader

	program, err := newProgram(vertexShaderSource, fragmentShaderSource)
	if err != nil {
			panic(err)
	}

	program2, err := newProgram(vertexShaderSource, mappingShaderSource)
	if err != nil {
			panic(err)
	}

	// pass mask as a texture uniform on slot 0
	textureUniform := gl.GetUniformLocation(program, gl.Str("tex_mask\x00"))
	gl.Uniform1i(textureUniform, 0)

	// Load texture to slot 0
	texture, err := newTextureFromFile("square.png")
	if err != nil {
		log.Fatalln(err)
	}

	// configure the vertex data
	vao := makeVao(square)

	// init time
	previousTime = glfw.GetTime()
	
	// prepare texture for framebuffer for shader 1
	var shaderOneTex uint32
	gl.GenTextures(1, &shaderOneTex)
	gl.BindTexture(gl.TEXTURE_2D, shaderOneTex);
	gl.TexImage2D(gl.TEXTURE_2D, 0, gl.RGB, width, height, 0, gl.RGB, gl.UNSIGNED_BYTE, nil);

	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MIN_FILTER, gl.LINEAR);
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MAG_FILTER, gl.LINEAR);

	// prepare framebuffer for shader1
	var frameBuffer uint32
	gl.GenFramebuffers(1, &frameBuffer)
	gl.BindFramebuffer(gl.FRAMEBUFFER, frameBuffer)

	gl.FramebufferTexture2D(gl.FRAMEBUFFER, gl.COLOR_ATTACHMENT0, gl.TEXTURE_2D, shaderOneTex, 0);

	if gl.CheckFramebufferStatus(gl.FRAMEBUFFER) != gl.FRAMEBUFFER_COMPLETE {
		panic("something's wrong with our framebuffer")
	}

	for !window.ShouldClose() {
		gl.Clear(gl.COLOR_BUFFER_BIT | gl.DEPTH_BUFFER_BIT)

		// Update
		time := glfw.GetTime()
		elapsed := time - previousTime
		previousTime = time

		runTime += elapsed

		/** Shader 1 go to a framebuffer **/

		// Shader 1 is rendered here to the framebuffer
		gl.UseProgram(program)
		
		// pass time and resolution as uniforms
		timeUniform := gl.GetUniformLocation(program, gl.Str("u_time\x00"))
		gl.Uniform1f(timeUniform, float32(runTime))
		
		resolutionUniform := gl.GetUniformLocation(program, gl.Str("u_resolution\x00"))
		gl.Uniform2f(resolutionUniform, float32(width), float32(height))
		
		// activate texture
		gl.ActiveTexture(gl.TEXTURE0)
		gl.BindTexture(gl.TEXTURE_2D, texture)

		gl.BindVertexArray(vao)
		gl.DrawArrays(gl.TRIANGLES, 0, int32(len(square) / 3))

		//Unbind the framebuffer so that you can draw normally again
		gl.BindFramebufferEXT(gl.FRAMEBUFFER, 0);

		// here we put the resulting texture in slot 1
		gl.ActiveTexture(gl.TEXTURE1)
		gl.BindTexture(gl.TEXTURE_2D, shaderOneTex)
		
		gl.ActiveTexture(gl.TEXTURE0) // get back to texture 0

		/** Shader 2 go to screen **/
		gl.UseProgram(program2)

		// attach texture from shader 1 to uniform on shader2
		shader1TUniform := gl.GetUniformLocation(program2, gl.Str("t_shaderOne\x00"));
		gl.Uniform1i(shader1TUniform, 1); //We use 1 here because we used GL_TEXTURE1 to bind our texture
		
		// pass time and resolution as uniforms
		timeUniform2 := gl.GetUniformLocation(program2, gl.Str("u_time\x00"))
		gl.Uniform1f(timeUniform2, float32(runTime))
		
		resolutionUniform2 := gl.GetUniformLocation(program2, gl.Str("u_resolution\x00"))
		gl.Uniform2f(resolutionUniform2, float32(width), float32(height))

		gl.BindVertexArray(vao)
		gl.DrawArrays(gl.TRIANGLES, 0, int32(len(square) / 3))
		
		// // extract buffer result
		// // printing is quite costly, so we skip that atm
		gl.ReadPixels(0, 0, width, height, gl.RGBA, gl.UNSIGNED_BYTE, unsafe.Pointer(&pixels))
		// fmt.Println(pixels) 

		// Maintenance
		window.SwapBuffers()
		glfw.PollEvents()
	}
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
func newTextureFromFile(file string) (uint32, error) {
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
		int32(rgba.Rect.Size().X),
		int32(rgba.Rect.Size().Y),
		0,
		gl.RGBA,
		gl.UNSIGNED_BYTE,
		gl.Ptr(rgba.Pix))

	return texture, nil
}