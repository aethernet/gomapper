package main

import (
	"log"

	"github.com/go-gl/gl/v2.1/gl"
)

// initOpenGL initializes OpenGL
func initOpenGL() {
    if err := gl.Init(); err != nil {
			panic(err)
    }

    version := gl.GoStr(gl.GetString(gl.VERSION))
    log.Println("OpenGL version", version)
}


