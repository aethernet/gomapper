package main

import (
	"image"
	"image/draw"
	_ "image/png"
	"log"
	"os"
	"runtime"
	"unsafe"

	// OR: github.com/go-gl/gl/v2.1/gl
	"fmt"
	"io/ioutil"
	"strings"

	"github.com/go-gl/gl/v2.1/gl"
	"github.com/go-gl/glfw/v3.2/glfw"
)

const (
	width  = 50
	height = 50
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

	program, err := newProgram(vertexShaderSource, fragmentShaderSource)

	if err != nil {
			panic(err)
	}

	// pass mask as a texture uniform
	textureUniform := gl.GetUniformLocation(program, gl.Str("tex_mask\x00"))
	gl.Uniform1i(textureUniform, 0)

	// Load the texture
	texture, err := newTexture("square.png")
	if err != nil {
		log.Fatalln(err)
	}

	// configure the vertex data
	vao := makeVao(square)
	
	// init time
	previousTime = glfw.GetTime()

	for !window.ShouldClose() {
		gl.Clear(gl.COLOR_BUFFER_BIT | gl.DEPTH_BUFFER_BIT)

		// Update
		time := glfw.GetTime()
		elapsed := time - previousTime
		previousTime = time

		runTime += elapsed

		// Render
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

		// extract buffer result
		gl.ReadPixels(0, 0, width, height, gl.RGBA, gl.UNSIGNED_BYTE, unsafe.Pointer(&pixels))
		fmt.Println(pixels)
    
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

func newTexture(file string) (uint32, error) {
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