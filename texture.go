package main

import "github.com/go-gl/gl/v2.1/gl"

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