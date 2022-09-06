package main

import "github.com/go-gl/glfw/v3.2/glfw"

func getUpdateTime() float32 {
	time := glfw.GetTime()
	elapsed := time - previousTime
	previousTime = time

	runTime += elapsed

	return float32(runTime)
}