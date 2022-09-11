#version 410
out vec4 frag_color;

uniform float u_time;
uniform vec2 u_resolution;

uniform int u_showmask;

uniform sampler2D t_mask;
uniform sampler2D t_tex;

vec2 decodePositionFromColor(vec4 encoded, float width, float height) {
    
    // bitshifting should be faster but would requires converting vec4 into int
    // before decoding, which would be more ops than this simple math solution
    // also bitshifting doens't works before OPENGL ES 3.0
    float x = ( encoded.r * 256. * 256. + encoded.g * 256. ) / width;
    float y = ( encoded.b * 255. * 256. + encoded.a * 256. ) / height;
    
    return vec2(x, y);
}

// pass thru
void main() {
  vec2 st = gl_FragCoord.xy / u_resolution / 2;

  // // get mask pixel coordinate
  // vec2 uv = gl_FragCoord.xy / u_resolution / 4;
  // // get mask color at coordinate
  // vec4 maskColor = texture(t_mask, vec2(uv.x, 1));
  // // get position in input shader encoded as color in mask (require the width and height of input shader)
  // vec2 position = decodePositionFromColor(maskColor, u_resolution.x, u_resolution.y);
  
  vec4 color = texture(t_tex, st);

  frag_color = color;
}