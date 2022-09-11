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
vec2 decodePositionFromColor(vec4 encoded, vec2 resolution) {
    vec2 decoded;
    
    // considered bitshifting but it would requires converting vec4 into in before decoding, 
    // which would be more ops than this simple math solution
    // also bitshifting doens't works before OPENGL ES 3.0
    // encoded is normalized ([0 -> 1]) so we need to first multiply by 256
    decoded.x = ( encoded.r * 256. * 256. + encoded.g * 256. ) / resolution.x;
    decoded.y = ( encoded.b * 256. * 256. + encoded.a * 256. ) / resolution.y;
    
    return decoded;
}
    
void main() {
    // get mask pixel coordinate
    vec2 uv = vec2(gl_FragCoord.xy) / u_maskresolution;

    // get mask color at coordinate
    vec4 maskColor = texture(t_mask, uv);

    // if this is a full white pixel in the mask it means it's an unused value 
    // so no need to compute it further and we can return a black pixel
    // if(maskColor == vec4(1.,1.,1.,1.))
    // {
    //     frag_color = vec4(0.,0.,0.,0.);
    //     return;
    // }

    // // get position in input shader encoded as color in mask (require the width and height of input shader)
    vec2 position = decodePositionFromColor(maskColor, u_resolution);

    // get the color of input shader at mask coordinate
    vec4 color = texture(t_previous, position * 10.); // this is so weird why do we need to multiply by 10 here ? it makes no sense

    // render
    frag_color = color;
}