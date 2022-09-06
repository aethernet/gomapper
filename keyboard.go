package main

import "github.com/go-gl/glfw/v3.2/glfw"

// keyboard callbacks
func keyCallback(w *glfw.Window, key glfw.Key, scancode int, action glfw.Action, mods glfw.ModifierKey) {
	//if u only want the on press, do = && && action == glfw.Press
	if key == glfw.KeySpace && action == glfw.Press {
		toggleView()
	}
	
	if key == glfw.KeyEscape && action == glfw.Press {
		w.SetShouldClose(true)
	}
}