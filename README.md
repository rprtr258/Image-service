# fimgs - image filters tool
## Install
```bash
go install github.com/rprtr258/fimgs/cmd/fimgs@latest
```
## Usage
```php
NAME:
   fimgs - Applies filter to image

USAGE:
   Applies filter to image and saves new image

COMMANDS:
   weakblur         Weakblur filter
   edgeenhance      Edgeenhance filter
   sharpen          Sharpen filter
   edgedetect1      Edgedetect1 filter
   edgedetect2      Edgedetect2 filter
   horizontallines  Horizontallines filter
   verticallines    Verticallines filter
   blur             Blur filter
   emboss           Emboss filter
   cluster          Cluster colors
   quadtree         Quad tree filter
   shader           Shader filter
   hilbert          Hilbert curve filter
   hilbertdarken    Hilbert darken curve filter
   zcurve           Z curve filter
   median           Median filter
   help, h          Shows a list of commands or help for one command

GLOBAL OPTIONS:
   --image value, -i value  input image filename
   --help, -h               show help

```
## Examples
||||
|-|-|-|
|![](./img/static/blur.png.png)|![](./img/static/cluster.png.png)|![](./img/static/edgedetect1.png.png)|
|blur.png|cluster.png|edgedetect1.png|
|![](./img/static/edgedetect2.png.png)|![](./img/static/edgeenhance.png.png)|![](./img/static/emboss.png.png)|
|edgedetect2.png|edgeenhance.png|emboss.png|
|![](./img/static/hilbert.png.png)|![](./img/static/hilbertdarken.png.png)|![](./img/static/horizontallines.png.png)|
|hilbert.png|hilbertdarken.png|horizontallines.png|
|![](./img/static/median.png.png)|![](./img/static/orig.png.png)|![](./img/static/quadtree.png.png)|
|median.png|orig.png|quadtree.png|
|![](./img/static/shader_rgb.png.png)|![](./img/static/sharpen.png.png)|![](./img/static/verticallines.png.png)|
|shader_rgb.png|sharpen.png|verticallines.png|
|![](./img/static/weakblur.png.png)|![](./img/static/zcurve.png.png)|
|weakblur.png|zcurve.png|
