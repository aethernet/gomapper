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
- [-] add mask generator from a json
- [ ] basic debug UI [draw mapping lines on top of screen]
- [ ] make it run on a jetson nano
- [-] make basic control UI
- [ ] add isf support
- [ ] add automatic OSC -> ISF controls

Controls
- [Space] - toggle view between Rendering and Mapping Shader
- [Esc] - Quit


Current limitations (todo for later):
- Only send 1 universe (512 pixels)
- Only one shader as input
- Lots of hardcoded values

Notes : 
- Encoding the mapping mask as a texture is good idea, but using .png as storage was stupid (it's a lossy format), let's just do a raw bytes format. It's unidimensional anyway.

Thanks to : 
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
