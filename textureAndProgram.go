package main

import (
	"fmt"
	"strings"

	"github.com/go-gl/gl/v2.1/gl"
)

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

func newFramebufferProgram(width int32, height int32, vertexFile string, fragFile string,) (uint32, uint32, uint32) {
	var vertexShaderSource = readShaderFile(vertexFile)    // the vertex shader
	var fragmentShaderSource = readShaderFile(fragFile)		 // the fragment shader

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

	// attach texture to framebuffer
	gl.FramebufferTexture2D(gl.FRAMEBUFFER, gl.COLOR_ATTACHMENT0, gl.TEXTURE_2D, shaderTex, 0);

	if gl.CheckFramebufferStatus(gl.FRAMEBUFFER) != gl.FRAMEBUFFER_COMPLETE {
		panic("something's wrong with our framebuffer")
	}

	// unbind the framebuffer to prevent accidentaly writing to it
	gl.BindFramebuffer(gl.FRAMEBUFFER, 0);

	return program, shaderTex, framebuffer
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