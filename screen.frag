#version 410
out vec4 frag_color;

uniform float u_time;
uniform vec2 u_resolution;

uniform int u_showmask;

uniform sampler2D t_mask;
uniform sampler2D t_tex;

vec2 decodePositionFromColor(vec4 encoded, float width, float height) {
    vec2 decoded;
    
    // bitshifting should be faster but would requires converting vec4 into int
    // before decoding, which would be more ops than this simple math solution
    // also bitshifting doens't works before OPENGL ES 3.0
    decoded.x = ( encoded.r * 256. + encoded.g ) / width;
    decoded.y = ( encoded.b * 255. + encoded.a ) / height;
    
    return decoded;
}

// pass thru
void main() {
  vec2 st = gl_FragCoord.xy / u_resolution / 2;
  
  frag_color = texture(t_tex, st);
}