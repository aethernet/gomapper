# Go experiments around a shader based pixel mapper

PoC implementation in p5.ps : https://github.com/Drunk-Developper/absolute-crazy-led-mapper

This is half baked : 

- [x] setup go, glsl, ...
- [x] load shader from file
- [x] render shader
- [x] pass float uniform to shader (time)
- [x] pass texture to shader as uniform
- [ ] extract texture data from shader
- [ ] chain shaders (send output texture from a shader as input texture for the next)
- [ ] generate artnet or sACN upd packages
- [ ] add mask generator
- [ ] load isf



inspired by : 
- https://github.com/KyleBanks/conways-gol
- https://github.com/go-gl/example/blob/master/gl41core-cube/cube.go
- https://github.com/mjw6i/shadr/blob/master/app.go
- https://thebookofshaders.com
- https://stackoverflow.com/questions/59433403/how-to-save-fragment-shader-image-changes-to-image-output-file
- https://www.haroldserrano.com/blog/how-to-pass-data-from-shader-to-shader-in-opengl