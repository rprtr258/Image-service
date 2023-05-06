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
   emboss           Emboss filter
   edgeenhance      Edgeenhance filter
   horizontallines  Horizontallines filter
   blur             Blur filter
   sharpen          Sharpen filter
   edgedetect1      Edgedetect1 filter
   edgedetect2      Edgedetect2 filter
   verticallines    Verticallines filter
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
|![](./img/static/blur.png)|![](./img/static/cluster.png)|![](./img/static/edgedetect1.png)|
|blur|cluster|edgedetect1|
|![](./img/static/edgedetect2.png)|![](./img/static/edgeenhance.png)|![](./img/static/emboss.png)|
|edgedetect2|edgeenhance|emboss|
|![](./img/static/hilbert.png)|![](./img/static/hilbertdarken.png)|![](./img/static/horizontallines.png)|
|hilbert|hilbertdarken|horizontallines|
|![](./img/static/median.png)|![](./img/static/orig.png)|![](./img/static/quadtree.png)|
|median|orig|quadtree|
|![](./img/static/shader_rgb.png)|![](./img/static/sharpen.png)|![](./img/static/verticallines.png)|
|shader_rgb|sharpen|verticallines|
|![](./img/static/weakblur.png)|![](./img/static/zcurve.png)|
|weakblur|zcurve|
