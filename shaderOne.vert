#version 410
in vec3 vp;

uniform vec2 u_resolution;

void main() {
  gl_Position = vec4(vp, 1.0);
}