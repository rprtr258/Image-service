# fimgs - image filters tool

## Install:
```bash
go install github.com/rprtr258/fimgs/cmd/fimgs@latest
```

## Usage:
```
Usage:
fimgs <filter_name> <filter_params> <source_image_file>
Applies filter to image and saves new image. Available filters:
Convolution filters:
	blur, weakblur, emboss, sharpen, edgeenhance, edgedetect1, edgedetect2, horizontallines, verticallines
Curve filters:
	hilbert, hilbertdarken, zcurve
Other:
	cluster - clusters colors, required parameters:
		number of clusters, must be integer and greater than 1
	shader - apply GLSL filter to image, required parameters:
		shader file, must be valid fragment shader source, see shader_examples for examples
	quadtree - apply quad tree like filter, required parameters:
		threshold must be integer from 0 to 65536 exclusive
		power must be float greater than 0.0
Example usage:
	fimgs emboss girl.png
	fimgs cluster 4 rain.jpeg
	fimgs quadtree 40000 3.14 girl.png
```

TODO: add image examples (they are in img/static, I just need to add them here)
