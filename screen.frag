#version 410
out vec4 frag_color;

uniform float u_time;
uniform vec2 u_resolution;

uniform sampler2D t_tex;

// pass thru
void main() {
  vec2 st = gl_FragCoord.xy/ u_resolution / 2;
  frag_color = texture(t_tex, st);
}