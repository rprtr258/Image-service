# fimgs - image filters tool

## Install:
```bash
go install github.com/rprtr258/fimgs/cmd/fimgs@latest
```

## Usage:
```bash
Usage:
fimgs <filter_name> <filter_params> <source_image_file>
Applies filter to image and saves new image. Available filters:
Convolution filters:
	blur, weakblur, emboss, sharpen, edgeenhance, edgedetect1, edgedetect2, horizontallines, verticallines
Hilbert filters:
	hilbert, hilbertdarken
Other:
	cluster - clusters colors, required parameters:
		number of clusters, must be integer and greater than 1
	shader - apply GLSL filter to image, required parameters:
		shader file, must be valid fragment shader source, see shader_examples for examples
Example usage:
	fimgs emboss girl.png
	fimgs cluster 4 rain.jpeg`
```