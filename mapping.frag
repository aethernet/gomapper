// Edwin Joassart <edwin@3kd.be> 2020
// MIT Licenced

#version 410

// define out
out vec4 frag_color;

// get input shader resolution
uniform vec2 u_resolution;
uniform vec2 u_maskresolution;

// get texture from previous renders
uniform sampler2D t_previous;

// get mask
uniform sampler2D t_mask;

// decode color encoded position from mask
vec2 decodePositionFromColor(vec4 encoded, float width, float height) {
    vec2 decoded;
    
    // bitshifting should be faster but would requires converting vec4 into int
    // before decoding, which would be more ops than this simple math solution
    // also bitshifting doens't works before OPENGL ES 3.0
    decoded.x = ( encoded.r * 255. * 255. + encoded.g * 255. ) / width;
    decoded.y = ( encoded.b * 255. * 255. + encoded.a * 255. ) / height;
    
    return decoded;
}
    
void main() {
    // get mask pixel coordinate
    vec2 uv = gl_FragCoord.xy / u_maskresolution;

    // get mask color at coordinate
    vec4 maskColor = texture(t_mask, uv);

    // if this is a full white pixel in the mask it means it's an unused value 
    // so no need to compute it further and we can return a black pixel
    if(maskColor == vec4(1.,1.,1.,1.))
    {
        frag_color = vec4(0.,0.,0.,0.);
        return;
    }

    // get position in input shader encoded as color in mask (require the width and height of input shader)
    vec2 position = decodePositionFromColor(maskColor, u_resolution.x, u_resolution.y);
    
    // get the color of input shader at mask coordinate
    vec4 color = texture(t_previous, position);

    // render
    frag_color = color;
}