package main

import "github.com/go-gl/gl/v2.1/gl"

// render shader to its framebuffer, pass an optional texture as input (will be available as t_previous)
func drawShaderToFramebuffer(program uint32, texture uint32, framebuffer uint32, width int, height int, vao uint32, GLtextureSlot uint32, passTexture int32) {
		gl.BindFramebuffer(gl.FRAMEBUFFER, framebuffer)
		gl.Clear(gl.COLOR_BUFFER_BIT | gl.DEPTH_BUFFER_BIT)
		gl.UseProgram(program)
		
		attachDefaultUniformToShader(program)

		if passTexture >= 0 {
			// Attach TEXTURE from previous shader as t_previous
			attach1IUniformToShader(program, "t_previous", passTexture)
		}

		gl.BindVertexArray(vao)
		gl.DrawArrays(gl.TRIANGLES, 0, int32(len(square) / 3))

		// here we put the resulting texture in proper slot
		gl.ActiveTexture(GLtextureSlot)
		gl.BindTexture(gl.TEXTURE_2D, texture)
}