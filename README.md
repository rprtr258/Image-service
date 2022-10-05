# fimgs - image filters tool

## Install
```bash
go install github.com/rprtr258/fimgs/cmd/fimgs@latest
```

## Usage
```
NAME:
   fimgs - Applies filter to image

USAGE:
   Applies filter to image and saves new image

COMMANDS:
   sharpen          Sharpen filter
   edgeenhance      Edgeenhance filter
   edgedetect2      Edgedetect2 filter
   weakblur         Weakblur filter
   emboss           Emboss filter
   horizontallines  Horizontallines filter
   verticallines    Verticallines filter
   blur             Blur filter
   edgedetect1      Edgedetect1 filter
   cluster          Cluster colors
   quadtree         Quad tree filter
   shader           Shader filter
   hilbert          Hilbert curve filter
   hilbertdarken    Hilbert darken curve filter
   zcurve           Z curve filter
   median           Median filter
   help, h          Shows a list of commands or help for one command

GLOBAL OPTIONS:
   --help, -h               show help (default: false)
   --image value, -i value  input image filename
   

```

## Examples
||||
|-|-|-|
|![](img/static/shader_rgb.png)|![](img/static/zcurve.png)|![](img/static/verticallines.png)|
|shader_rgb|zcurve|verticallines|
|![](img/static/sharpen.png)|![](img/static/blur.png)|![](img/static/quadtree.png)|
|sharpen|blur|quadtree|
|![](img/static/weakblur.png)|![](img/static/median.png)|![](img/static/hilbertdarken.png)|
|weakblur|median|hilbertdarken|
|![](img/static/emboss.png)|![](img/static/horizontallines.png)|![](img/static/edgedetect2.png)|
|emboss|horizontallines|edgedetect2|
|![](img/static/edgeenhance.png)|![](img/static/edgedetect1.png)|![](img/static/orig.png)|
|edgeenhance|edgedetect1|orig|
|![](img/static/hilbert.png)|![](img/static/cluster.png)|![](img/static/.png)|
|hilbert|cluster||
