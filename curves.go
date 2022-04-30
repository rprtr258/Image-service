package fimgs

import (
	"image"
	"image/color"
	"image/draw"
	"math"
)

func is_block_black(r image.Rectangle, im image.Image) bool {
	brightnessSum := 0.0
	for i := r.Min.X; i < r.Max.X; i++ {
		for j := r.Min.Y; j < r.Max.Y; j++ {
			r, g, b, _ := im.At(i, j).RGBA()
			brightnessSum += float64(r+g+b) / 3 / 0xFFFF
		}
	}
	THRESHOLD := 0.6
	return brightnessSum < THRESHOLD*float64(r.Dx()*r.Dy())
}

func rectFrom4Points(p1, p2, p3, p4 image.Point) image.Rectangle {
	xmin, ymin, xmax, ymax := p1.X, p1.Y, p2.X, p2.Y
	if xmin > p3.X {
		xmin = p3.X
	}
	if xmax < p3.X {
		xmax = p3.X
	}
	if ymin > p3.Y {
		ymin = p3.Y
	}
	if ymax < p3.Y {
		ymax = p3.Y
	}
	if xmin > p4.X {
		xmin = p4.X
	}
	if xmax < p4.X {
		xmax = p4.X
	}
	if ymin > p4.Y {
		ymin = p4.Y
	}
	if ymax < p4.Y {
		ymax = p4.Y
	}
	return image.Rect(xmin, ymin, xmax, ymax)
}

func sgn(x int) int {
	switch {
	case x < 0:
		return -1
	case x == 0:
		return 0
	default:
		return 1
	}
}

func plotLineLow(im *image.RGBA, p, q image.Point) {
	dx := q.X - p.X
	dy := q.Y - p.Y
	yi := sgn(dy)
	dy *= yi
	D := 2*dy - dx
	y := p.Y
	for x := p.X; x <= q.X; x++ {
		im.Set(x, y, color.RGBA{0, 0, 0, 255})
		if D > 0 {
			y += yi
			D += 2 * (dy - dx)
		} else {
			D += 2 * dy
		}
	}
}

func plotLineHigh(im *image.RGBA, p, q image.Point) {
	dx := q.X - p.X
	dy := q.Y - p.Y
	xi := sgn(dx)
	dx *= xi
	D := 2*dx - dy
	x := p.X
	for y := p.Y; y <= q.Y; y++ {
		im.Set(x, y, color.RGBA{0, 0, 0, 255})
		if D > 0 {
			x += xi
			D += 2 * (dx - dy)
		} else {
			D += 2 * dx
		}
	}
}

func abs(x int) int {
	if x < 0 {
		return -x
	} else {
		return x
	}
}

func drawLine(im *image.RGBA, p, q image.Point) {
	if abs(q.Y-p.Y) < abs(q.X-p.X) {
		if p.X > q.X {
			p, q = q, p
		}
		plotLineLow(im, p, q)
	} else {
		if p.Y > q.Y {
			p, q = q, p
		}
		plotLineHigh(im, p, q)
	}
}

// TODO: unite hilbert and zcurve
func hilbert(sourceImage image.Image, resultImage *image.RGBA, p1, p12, p23 image.Point, size int) []image.Point {
	p2 := p1.Add(p12)
	p3 := p2.Add(p23)
	p4 := p1.Add(p23)
	if size <= 2 {
		if is_block_black(rectFrom4Points(p1, p2, p3, p4), sourceImage) {
			mid := p1.Add(p3).Div(2)
			return []image.Point{mid, mid}
		} else {
			return nil
		}
	}
	// . 1       4
	// | |1-2 3-4|
	// | |  | |  |
	// | |4-3 2-1|
	// v ||     ||
	// p |1 4-1 4|
	// 1 || | | ||
	// 2 |2-3 2-3|
	// h 2-------3
	//   .--->p23h
	p12h := p12.Div(2)
	p23h := p23.Div(2)
	lt := hilbert(sourceImage, resultImage, p1, p23h, p12h, size-1)
	lb := hilbert(sourceImage, resultImage, p2.Sub(p12h), p12h, p23h, size-1)
	rb := hilbert(sourceImage, resultImage, p3.Sub(p12h).Sub(p23h), p12h, p23h, size-1)
	rt := hilbert(sourceImage, resultImage, p4.Add(p12h), p23h.Mul(-1), p12h.Mul(-1), size-1)
	if lt == nil && lb == nil && rb == nil && rt == nil && !is_block_black(rectFrom4Points(p1, p2, p3, p4), sourceImage) {
		return nil
	}
	if lt == nil {
		p := p1.Add(p12h.Add(p23h).Div(2))
		lt = []image.Point{p, p}
	}
	if lb == nil {
		p := p2.Add(p23h.Sub(p12h).Div(2))
		lb = []image.Point{p, p}
	}
	if rb == nil {
		p := p3.Sub(p12h.Add(p23h).Div(2))
		rb = []image.Point{p, p}
	}
	if rt == nil {
		p := p4.Add(p12h.Sub(p23h).Div(2))
		rt = []image.Point{p, p}
	}
	drawLine(resultImage, lt[1], lb[0])
	drawLine(resultImage, lb[1], rb[0])
	drawLine(resultImage, rb[1], rt[0])
	return []image.Point{lt[0], rt[1]}
}

func zcurve(sourceImage image.Image, resultImage *image.RGBA, p1, p12, p13 image.Point, size int) []image.Point {
	p2 := p1.Add(p12)
	p3 := p1.Add(p13)
	p4 := p2.Add(p13)
	if size <= 2 {
		if is_block_black(rectFrom4Points(p1, p2, p3, p4), sourceImage) {
			mid := p1.Add(p3).Div(2)
			return []image.Point{mid, mid}
		} else {
			return nil
		}
	}
	// 1-------2
	// |1-2 1-2
	// | / / /
	// |3-4 3-4
	// |     /
	// | /--/
	// |/
	// |1-2 1-2
	// | / / /
	// |3-4 3-4
	// 3       4
	p12h := p12.Div(2)
	p13h := p13.Div(2)
	part0 := zcurve(sourceImage, resultImage, p1, p12h, p13h, size-1)
	part1 := zcurve(sourceImage, resultImage, p1.Add(p12h), p12h, p13h, size-1)
	part2 := zcurve(sourceImage, resultImage, p1.Add(p13h), p12h, p13h, size-1)
	part3 := zcurve(sourceImage, resultImage, p1.Add(p12h).Add(p13h), p12h, p13h, size-1)
	if part0 == nil && part1 == nil && part2 == nil && part3 == nil && !is_block_black(rectFrom4Points(p1, p2, p3, p4), sourceImage) {
		return nil
	}
	if part0 == nil {
		p := p1.Add(p12h.Add(p13h).Div(2))
		part0 = []image.Point{p, p}
	}
	if part1 == nil {
		p := p2.Add(p13h.Sub(p12h).Div(2))
		part1 = []image.Point{p, p}
	}
	if part2 == nil {
		p := p3.Sub(p13h.Sub(p12h).Div(2))
		part2 = []image.Point{p, p}
	}
	if part3 == nil {
		p := p4.Sub(p12h.Add(p13h).Div(2))
		part3 = []image.Point{p, p}
	}
	drawLine(resultImage, part0[1], part1[0])
	drawLine(resultImage, part1[1], part2[0])
	drawLine(resultImage, part2[1], part3[0])
	return []image.Point{part0[0], part3[1]}
}

func HilbertCurveFilter(im image.Image) *image.RGBA {
	// TODO: remove / change to absolute adjustment
	// f = ImageEnhance.Brightness(res).enhance(1.3)
	// f = ImageEnhance.Contrast(f).enhance(10)
	W := int(math.Min(math.Log2(float64(im.Bounds().Dx())), math.Log2(float64(im.Bounds().Dx()))))
	himage := image.NewRGBA(im.Bounds())
	draw.Draw(himage, himage.Bounds(), &image.Uniform{color.RGBA{255, 255, 255, 255}}, image.Point{}, draw.Src)
	hilbert(im, himage, im.Bounds().Min, image.Point{0, im.Bounds().Dy()}, image.Point{im.Bounds().Dx(), 0}, W)
	return himage
}

func HilbertCurve(sourceImageFilename, resultImageFilename string) error {
	im, err := LoadImageFile(sourceImageFilename)
	if err != nil {
		return err
	}
	tmp := HilbertCurveFilter(im)
	return saveImage(tmp, resultImageFilename)
}

// TODO: extract and make blendings
func HilbertDarken(sourceImageFilename, resultImageFilename string) error {
	im, err := LoadImageFile(sourceImageFilename)
	if err != nil {
		return err
	}
	tmp := HilbertCurveFilter(im)
	for i := tmp.Bounds().Min.X; i < tmp.Bounds().Max.X; i++ {
		for j := tmp.Bounds().Min.Y; j < tmp.Bounds().Max.Y; j++ {
			r, g, b, _ := tmp.At(i, j).RGBA()
			if r > 0 || g > 0 || b > 0 {
				tmp.Set(i, j, im.At(i, j))
			}
		}
	}
	return saveImage(tmp, resultImageFilename)
}

func ZCurveFilter(im image.Image) *image.RGBA {
	// TODO: remove / change to absolute adjustment
	// f = ImageEnhance.Brightness(res).enhance(1.3)
	// f = ImageEnhance.Contrast(f).enhance(10)
	W := int(math.Min(math.Log2(float64(im.Bounds().Dx())), math.Log2(float64(im.Bounds().Dx()))))
	himage := image.NewRGBA(im.Bounds())
	draw.Draw(himage, himage.Bounds(), &image.Uniform{color.RGBA{255, 255, 255, 255}}, image.Point{}, draw.Src)
	zcurve(im, himage, im.Bounds().Min, image.Point{im.Bounds().Dx(), 0}, image.Point{0, im.Bounds().Dy()}, W)
	return himage
}

// TODO: extract repeating loading, saving image
func ZCurve(sourceImageFilename, resultImageFilename string) error {
	im, err := LoadImageFile(sourceImageFilename)
	if err != nil {
		return err
	}
	tmp := ZCurveFilter(im)
	return saveImage(tmp, resultImageFilename)
}
