# Go experiments around a shader based pixel mapper

PoC implementation in p5.ps : https://github.com/Drunk-Developper/absolute-crazy-led-mapper

This is half baked : 

- [x] setup go, glsl, ...
- [x] load shader from file
- [x] render shader
- [x] pass float uniform to shader (time)
- [x] pass texture to shader as uniform
- [x] extract texture data from shader
- [x] chain shaders (send output texture from a shader as input texture for the next)
- [x] refactor code for readability
- [x] render each shader in its own texture
- [x] use proper mapping shaders
- [x] extract pixels from mapping texture
- [x] generate artnet or sACN udp packages
- [x] fix broken mapping
- [x] add mask generator from a json
- [-] add multi-universe support
- [x] add a fps throttling for sacn
- [ ] basic debug UI [draw mapping lines on top of screen]
- [ ] make it run on a jetson nano
- [-] make basic control UI
- [ ] add isf support
- [ ] add automatic OSC -> ISF controls

Controls
- [Space] - toggle view between Rendering and Mapping Shader
- [Esc] - Quit


Current limitations (todo for later):
- Only one shader as input
- Limited to 16 universe (can be easily changed in main.go)

Notes : 
- Encoding the mapping mask as a texture is good idea, but using .png as storage was stupid (it's a lossy format), let's just do a raw bytes format. It's unidimensional anyway.

### Multi-Universe : 
First prototype was using a p x 1 mapping texture, n beeing the quantity of pixel to map.

Updated approach is to use a 512 x u mapping texture, u beeing the quantity of universe to send.

If a fixture is overflowing a universe, it will continue on the next one.

Each universe is encoded as one line in the mapping texture and all unused pixels will be "blank" (actually white -> vec4(255 255 255 255) as black -> vec4(0 0 0 0) translate is a valid position vec2(0, 0)) We'll filter out blank pixel in the mapping shader (no computation) and won't send non mapped universe out on the network.

### Thanks to : 
- https://github.com/KyleBanks/conways-gol
- https://github.com/go-gl/example/blob/master/gl41core-cube/cube.go
- https://github.com/mjw6i/shadr/blob/master/app.go
- https://thebookofshaders.com
- https://stackoverflow.com/questions/59433403/how-to-save-fragment-shader-image-changes-to-image-output-file
- https://www.haroldserrano.com/blog/how-to-pass-data-from-shader-to-shader-in-opengl
- https://learnopengl.com/book/book_pdf.pdf
- https://github.com/Hundemeier/go-sacn

Maight thank you later :
- https://github.com/hypebeast/go-osc
