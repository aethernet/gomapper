package main

import "github.com/go-gl/gl/v2.1/gl"

// render shader to its framebuffer, pass an optional texture as input (will be available as t_previous)
func drawShaderToFramebuffer(program uint32, texture uint32, framebuffer uint32, width int, height int, vao uint32, GLtextureSlot uint32, previousTexture int32) {
	// here we put the resulting texture in proper slot
	
	// it's important to bind all textures we're going to use here 

	// mapping texture
	gl.ActiveTexture(gl.TEXTURE0)
	gl.BindTexture(gl.TEXTURE_2D, mappingTexture)

	// previous texture
	if previousTexture > 0 {
		gl.ActiveTexture(gl.TEXTURE1)
		gl.BindTexture(gl.TEXTURE_2D, uint32(previousTexture))
	}
	
	// texture we're going to render to
	gl.ActiveTexture(GLtextureSlot)
	gl.BindTexture(gl.TEXTURE_2D, texture)

	gl.BindFramebuffer(gl.FRAMEBUFFER, framebuffer)
	
	gl.ClearColor(0.1, 0.1, 0.1, 1.0);
	gl.Clear(gl.COLOR_BUFFER_BIT | gl.DEPTH_BUFFER_BIT)
	gl.UseProgram(program)
	
	attachDefaultUniformToShader(program)

	if previousTexture > 0 {
		// Attach TEXTURE from previous shader as t_previous
		attach1IUniformToShader(program, "t_previous", 1)
	} 

	// attach mapping texture
	attach1IUniformToShader(program, "t_mask", 0)

	gl.BindVertexArray(vao)
	gl.DrawArrays(gl.TRIANGLES, 0, int32(len(square) / 3))

	gl.BindFramebuffer(gl.FRAMEBUFFER, 0)
}