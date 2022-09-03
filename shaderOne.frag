#version 410
out vec4 frag_colour;

uniform float u_runtime;
// 

void main() {

  vec4 color = vec4(u_runtime, 1, 1, 1);

  frag_colour = vec4(color.r, 0.1, 0.1, 1);
}