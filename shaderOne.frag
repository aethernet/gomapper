#version 410
out vec4 frag_color;

uniform float u_time;
uniform vec2 u_resolution;

uniform sampler2D t_mask;

void main() {
    vec2 st = gl_FragCoord.xy/u_resolution ;

    vec4 color = texture(t_mask, st);

    frag_color = vec4(color);
}