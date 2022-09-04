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
- [ ] generate artnet or sACN udp packages
- [ ] add mask generator
- [ ] add isf support
- [ ] add automatic OSC -> ISF controls
- [ ] make it run on a jetson nano

Notes : 
- frag shader 1 coordinate doens't requires a /2, while we do it on frag shader2 before displaying (the logic behind this escapes me atm, but it works)

inspired by : 
- https://github.com/KyleBanks/conways-gol
- https://github.com/go-gl/example/blob/master/gl41core-cube/cube.go
- https://github.com/mjw6i/shadr/blob/master/app.go
- https://thebookofshaders.com
- https://stackoverflow.com/questions/59433403/how-to-save-fragment-shader-image-changes-to-image-output-file
- https://www.haroldserrano.com/blog/how-to-pass-data-from-shader-to-shader-in-opengl
- https://learnopengl.com/book/book_pdf.pdf