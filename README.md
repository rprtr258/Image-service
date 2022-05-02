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

||||
|-|-|-|
|![](img/static/orig.png)|![](img/static/blur.png)|![](img/static/cluster.png)|
|orig|blur|cluster|
|![](img/static/edgedetect1.png)|![](img/static/edgedetect2.png)|![](img/static/edgeenhance.png)|
|edgedetect1|edgedetect2|edgeenhance|
|![](img/static/emboss.png)|![](img/static/hilbert.png)|![](img/static/hilbertdarken.png)|
|emboss|[hilbert](https://habr.com/en/post/135344/)|hilbertdarken|
|![](img/static/horizontallines.png)|![](img/static/median.png)|![](img/static/quadtree.png)|
|horizontallines|[median](https://en.wikipedia.org/wiki/Median_filter)|[quadtree](https://habr.com/en/post/280674/)|
|![](img/static/shader_rgb.png)|![](img/static/sharpen.png)|![](img/static/verticallines.png)|
|shader/rgb|sharpen|verticallines|
|![](img/static/weakblur.png)|![](img/static/zcurve.png)||
|weakblur|zcurve||

