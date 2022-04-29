package fimgs

import (
	"fmt"
	"image"
	"image/color"
	"image/draw"
	_ "image/jpeg"
	"image/png"
	"math"
	"os"
	"os/exec"
)

var (
	BLUR_KERNEL = [][]int{
		{1, 1, 1},
		{1, 1, 1},
		{1, 1, 1},
	}
	WEAK_BLUR_KERNEL = [][]int{
		{0, 1, 0},
		{1, 1, 1},
		{0, 1, 0},
	}
	EMBOSS_KERNEL = [][]int{
		{-2, -1, 0},
		{-1, 1, 1},
		{0, 1, 2},
	}
	SHARPEN_KERNEL = [][]int{
		{0, -1, 0},
		{-1, 5, -1},
		{0, -1, 0},
	}
	EDGE_ENHANCE_KERNEL = [][]int{
		{0, 0, 0},
		{-1, 1, 0},
		{0, 0, 0},
	}
	EDGE_DETECT1_KERNEL = [][]int{
		{1, 0, -1},
		{0, 0, 0},
		{-1, 0, 1},
	}
	EDGE_DETECT2_KERNEL = [][]int{
		{0, -1, 0},
		{-1, 4, -1},
		{0, -1, 0},
	}
	HORIZONTAL_LINES_KERNEL = [][]int{
		{-1, -1, -1},
		{2, 2, 2},
		{-1, -1, -1},
	}
	VERTICAL_LINES_KERNEL = [][]int{
		{-1, 2, -1},
		{-1, 2, -1},
		{-1, 2, -1},
	}
)

func LoadImageFile(image_filename string) (im image.Image, err error) {
	imageFile, err := os.Open(image_filename)
	if err != nil {
		return
	}
	defer imageFile.Close()
	im, _, err = image.Decode(imageFile)
	if err != nil {
		return
	}
	return
}

func saveImage(im image.Image, imageFilename string) (err error) {
	imageFile, err := os.Create(imageFilename)
	if err != nil {
		return
	}
	defer imageFile.Close()
	if err = png.Encode(imageFile, im); err != nil {
		return
	}
	return
}

func ApplyConvolution(im image.Image, kernel [][]int) image.Image {
	kernelHalfWidth, kernelHalfHeight := len(kernel)/2, len(kernel)/2
	R := make([][][3]int, im.Bounds().Dx())
	for i := im.Bounds().Min.X; i < im.Bounds().Max.X; i++ {
		R[i] = make([][3]int, im.Bounds().Dy())
	}
	for i := im.Bounds().Min.X; i < im.Bounds().Max.X; i++ {
		for j := im.Bounds().Min.Y; j < im.Bounds().Max.Y; j++ {
			r, g, b := 0, 0, 0
			for di := 0; di < len(kernel); di++ {
				for dj := 0; dj < len(kernel[0]); dj++ {
					i1 := i + dj - kernelHalfWidth
					if i1 < 0 {
						i1 = 0
					}
					if i1 >= im.Bounds().Dx() {
						i1 = im.Bounds().Dx() - 1
					}
					j1 := j + di - kernelHalfHeight
					if j1 < 0 {
						j1 = 0
					}
					if j1 >= im.Bounds().Dy() {
						j1 = im.Bounds().Dy() - 1
					}
					dr, dg, db, _ := im.At(i1, j1).RGBA()
					r += int(dr) * kernel[di][dj]
					g += int(dg) * kernel[di][dj]
					b += int(db) * kernel[di][dj]
				}
			}
			R[i][j] = [3]int{r, g, b}
		}
	}
	kernelMin, kernelMax := math.MaxInt, math.MinInt
	for i := im.Bounds().Min.X; i < im.Bounds().Max.X; i++ {
		for j := im.Bounds().Min.Y; j < im.Bounds().Max.Y; j++ {
			for k := 0; k < 3; k++ {
				if R[i][j][k] < kernelMin {
					kernelMin = R[i][j][k]
				}
				if R[i][j][k] > kernelMax {
					kernelMax = R[i][j][k]
				}
			}
		}
	}
	var (
		filtered_im *image.RGBA = image.NewRGBA(im.Bounds())
		diff        int         = kernelMax - kernelMin
	)
	for i := im.Bounds().Min.X; i < im.Bounds().Max.X; i++ {
		for j := im.Bounds().Min.Y; j < im.Bounds().Max.Y; j++ {
			filtered_im.Set(i, j, color.RGBA{
				uint8((R[i][j][0] - kernelMin) * 255 / diff),
				uint8((R[i][j][1] - kernelMin) * 255 / diff),
				uint8((R[i][j][2] - kernelMin) * 255 / diff),
				255,
			})
		}
	}
	return filtered_im
}

func ApplyConvolutionFilter(sourceImageFilename string, resultImageFilename string, kernel [][]int) error {
	im, err := LoadImageFile(sourceImageFilename)
	if err != nil {
		return fmt.Errorf("rror occured during loading image:\n%q", err)
	}
	resImage := ApplyConvolution(im, kernel)
	return saveImage(resImage, resultImageFilename)
}

// TODO: don't call bash?
func TransferStyle(sourceImageFilename, resultImageFilename, style_name string) (err error) {
	os.Chdir("fast-style-transfer/")
	if err = exec.Command(
		"python3", "evaluate.py",
		"--in-path", sourceImageFilename,
		"--out-path", resultImageFilename,
		"--checkpoint", fmt.Sprintf("../ckpts/%s.ckpt", style_name),
	).Run(); err != nil {
		err = fmt.Errorf("error running python3 evaluate.py, error: %q", err)
		return
	}
	return
}

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

// TODO: pass log2(size) instead of size
func hilbert(sourceImage image.Image, resultImage *image.RGBA, p1, p12, p23 image.Point, size int) []image.Point {
	p2 := p1.Add(p12)
	p3 := p2.Add(p23)
	p4 := p1.Add(p23)
	if size <= 2 {
		if is_block_black(rectFrom4Points(p1, p2, p3, p4), sourceImage) {
			mid := p1.Add(p2).Add(p3).Add(p4).Div(4)
			return []image.Point{mid, mid}
		} else {
			return nil
		}
	}
	// . 1       4
	// | |1-2 3-4|
	// | | a| |d |
	// | |4-3 2-1|
	// v ||     ||
	// p |1 4-1 4|
	// 1 ||b| |c||
	// 2 |2-3 2-3|
	// h 2-------3
	//   .--->p23h
	p12h := p12.Div(2)
	p23h := p23.Div(2)
	lt := hilbert(sourceImage, resultImage, p1, p23h, p12h, size/2)
	lb := hilbert(sourceImage, resultImage, p2.Sub(p12h), p12h, p23h, size/2)
	rb := hilbert(sourceImage, resultImage, p3.Sub(p12h).Sub(p23h), p12h, p23h, size/2)
	rt := hilbert(sourceImage, resultImage, p4.Add(p12h), p23h.Mul(-1), p12h.Mul(-1), size/2)
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

// TODO: try also https://en.wikipedia.org/wiki/Z-order_curve
func HilbertCurveFilter(im image.Image) *image.RGBA {
	// TODO: remove / change to absolute adjustment
	// f = ImageEnhance.Brightness(res).enhance(1.3)
	// f = ImageEnhance.Contrast(f).enhance(10)
	W := 1
	for W < im.Bounds().Dx() && W < im.Bounds().Dy() {
		W *= 2
	}
	W /= 2
	himage := image.NewRGBA(im.Bounds())
	draw.Draw(himage, himage.Bounds(), &image.Uniform{color.RGBA{255, 255, 255, 255}}, image.Point{}, draw.Src)
	hilbert(im, himage, im.Bounds().Min, image.Point{im.Bounds().Dx(), 0}, image.Point{0, im.Bounds().Dy()}, W)
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
