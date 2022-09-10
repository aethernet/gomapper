package main

import "github.com/go-gl/gl/v2.1/gl"

func attach1FUniformToShader(program uint32, name string, value float32) {
	uniform := gl.GetUniformLocation(program, gl.Str(name + "\x00"))
	gl.Uniform1f(uniform, value)
}

// func attach2FUniformToShader(program uint32, name string, value1 float32, value2 float32) {
// 	uniform := gl.GetUniformLocation(program, gl.Str(name + "\x00"))
// 	gl.Uniform2f(uniform, value1, value2)
// }

func attach1IUniformToShader(program uint32, name string, value int32) {
	uniform := gl.GetUniformLocation(program, gl.Str(name + "\x00"))
	gl.Uniform1i(uniform, value)
}

// default uniform we attach to all programs
func attachDefaultUniformToShader(program uint32) {
	// attach time
	runTime := getUpdateTime()
	attach1FUniformToShader(program, "u_time", runTime)

	// attach resolution
	resolutionUniform := gl.GetUniformLocation(program, gl.Str("u_resolution\x00"))
	gl.Uniform2f(resolutionUniform, float32(width), float32(height))

	// attach mapping texture resolution
	maskResolutionUniform := gl.GetUniformLocation(program, gl.Str("u_maskresolution\x00"))
	gl.Uniform2f(maskResolutionUniform, float32(170.), float32(len(universeMapping)))
}

