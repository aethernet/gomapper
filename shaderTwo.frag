#version 410
out vec4 frag_color;

uniform float u_time;
uniform vec2 u_resolution;

uniform sampler2D t_shaderOne;

vec4 mix(vec4 color1, vec4 color2) { 
  float r = ((color1.r * color1.a) + (color2.r * color2.a))/ 2;
  float g = ((color1.g * color1.a) + (color2.g * color2.a))/ 2;
  float b = ((color1.b * color1.a) + (color2.b * color2.a))/ 2;
  float a = (color1.a + color2.a) / 2;
  return vec4(r,g,b,a);
}

void main() {
  vec2 st = gl_FragCoord.xy/ u_resolution / 2;

  vec4 color = texture(t_shaderOne, st);

  vec4 gradiant = vec4(st.x, abs(sin(u_time / 2)) / 2, 0.2, 1);

  frag_color = vec4(mix(color, gradiant));
}