package fimgs

import (
	"fmt"
	"image"
	"image/color"
	"image/draw"
	_ "image/jpeg"
	"image/png"
	"math"
	"math/rand"
	"os"
	"os/exec"
)

func loadImageFile(image_filename string) (res image.Image, err error) {
	ff, err := os.Open(image_filename)
	if err != nil {
		return
	}
	res, _, err = image.Decode(ff)
	if err != nil {
		return
	}
	return
}

func save_image(im image.Image, imageFilename string) (err error) {
	f, err := os.Create(imageFilename)
	if err != nil {
		return
	}
	if err = png.Encode(f, im); err != nil {
		return
	}
	return
}

func apply_convolution(im image.Image, kernel [][]int) image.Image {
	kernelHalfWidth, kernelHalfHeight := len(kernel)/2, len(kernel)/2
	R := make([][][]int, im.Bounds().Dx())
	for i := im.Bounds().Min.X; i < im.Bounds().Max.X; i++ {
		R[i] = make([][]int, im.Bounds().Dy())
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
			R[i][j] = []int{r, g, b}
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

func applyConvolutionFilter(sourceImageFilename string, resultImageFilename string, kernel [][]int) error {
	im, err := loadImageFile(sourceImageFilename)
	if err != nil {
		return fmt.Errorf("Error occured during loading image:\n%q", err)
	}
    resImage := apply_convolution(im, kernel)
	return save_image(resImage, resultImageFilename)
}

// TODO: don't call bash?
func transfer_style(imid string, style_name string) (res string, err error) {
	res = fmt.Sprintf("img/%s.res.png", imid)
	os.Chdir("fast-style-transfer/")
	if err = exec.Command(
		"python3", "evaluate.py",
		"--in-path", fmt.Sprintf("../%s", res),
		"--out-path", "../",
		"--checkpoint", fmt.Sprintf("../ckpts/%s.ckpt", style_name),
	).Run(); err != nil {
		err = fmt.Errorf("error running python3 evaluate.py, error: %q", err)
		return
	}
	os.Chdir("..")
	if err = exec.Command("mv", fmt.Sprintf("%s.orig.png", imid), res).Run(); err != nil {
		err = fmt.Errorf("error running mv, error: %q", err)
		return
	}
	if err = exec.Command("rm", res).Run(); err != nil {
		err = fmt.Errorf("error running rm, error: %q", err)
		return
	}
	return
}

func applyKMeansFilter(sourceImageFilename string, resultImageFilename string, n_clusters int) (err error) {
	if n_clusters < 2 {
		return fmt.Errorf("'n' must be at least 2, you gave n=%d", n_clusters)
	}
	im, err := loadImageFile(sourceImageFilename)
	if err != nil {
		return fmt.Errorf("Error occured while loading image:\n%q", err)
	}
	kmeans := make([][]uint32, n_clusters)
	sumAndCount := make([][]uint64, n_clusters) // sum of Rs, Gs, Bs and count
	rand.Seed(0)
	// TODO: init using https://en.wikipedia.org/wiki/K-means++
	for i := 0; i < n_clusters; i++ {
		kmeans[i] = []uint32{
			rand.Uint32() / 0x10000,
			rand.Uint32() / 0x10000,
			rand.Uint32() / 0x10000,
		}
		sumAndCount[i] = make([]uint64, 4)
	}
	// TODO: optimize
	for epoch := 0; epoch < 100; epoch++ { // TODO: or diff is small enough
		for i := 0; i < n_clusters; i++ {
			sumAndCount[i][0], sumAndCount[i][1], sumAndCount[i][2], sumAndCount[i][3] = 0, 0, 0, 0
		}
		// TODO: parallelize?
		for i := im.Bounds().Min.X; i < im.Bounds().Max.X; i++ {
			for j := im.Bounds().Min.Y; j < im.Bounds().Max.Y; j++ {
				r, g, b, _ := im.At(i, j).RGBA()
				minCluster := 0
				minDist := (r-kmeans[0][0])*(r-kmeans[0][0]) + (g-kmeans[0][1])*(g-kmeans[0][1]) + (b-kmeans[0][2])*(b-kmeans[0][2])
				for k := 1; k < n_clusters; k++ {
					dist := (r-kmeans[k][0])*(r-kmeans[k][0]) + (g-kmeans[k][1])*(g-kmeans[k][1]) + (b-kmeans[k][2])*(b-kmeans[k][2])
					if dist < minDist {
						minCluster = k
						minDist = dist
					}
				}
				sumAndCount[minCluster][0] += uint64(r)
				sumAndCount[minCluster][1] += uint64(g)
				sumAndCount[minCluster][2] += uint64(b)
				sumAndCount[minCluster][3]++
			}
		}
		for i := 0; i < n_clusters; i++ {
			count := sumAndCount[i][3]
			if count == 0 {
				continue
			}
			kmeans[i][0], kmeans[i][1], kmeans[i][2] = uint32(sumAndCount[i][0]/count), uint32(sumAndCount[i][1]/count), uint32(sumAndCount[i][2]/count)
		}
	}
	filtered_im := image.NewRGBA(im.Bounds())
	for i := im.Bounds().Min.X; i < im.Bounds().Max.X; i++ {
		for j := im.Bounds().Min.Y; j < im.Bounds().Max.Y; j++ {
			r, g, b, _ := im.At(i, j).RGBA()
			minCluster := 0
			minDist := (r-kmeans[0][0])*(r-kmeans[0][0]) + (g-kmeans[0][1])*(g-kmeans[0][1]) + (b-kmeans[0][2])*(b-kmeans[0][2])
			for k := 1; k < n_clusters; k++ {
				dist := (r-kmeans[k][0])*(r-kmeans[k][0]) + (g-kmeans[k][1])*(g-kmeans[k][1]) + (b-kmeans[k][2])*(b-kmeans[k][2])
				if dist < minDist {
					minCluster = k
					minDist = dist
				}
			}
			filtered_im.Set(i, j, color.RGBA{
				uint8(kmeans[minCluster][0] / 0x100),
				uint8(kmeans[minCluster][1] / 0x100),
				uint8(kmeans[minCluster][2] / 0x100),
				255,
			})
		}
	}
    err = save_image(filtered_im, resultImageFilename)
    if err != nil {
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

// TODO: pass 1st point and vectors 1->2, 2->3, instead of p1-4
// TODO: pass log2(size) instead of size
func hilbert(sourceImage image.Image, resultImage *image.RGBA, p1, p2, p3, p4 image.Point, size int) []image.Point {
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
	p12h := p2.Sub(p1).Div(2)
	p23h := p3.Sub(p2).Div(2)
	lt := hilbert(sourceImage, resultImage, p1, p1.Add(p23h), p1.Add(p23h).Add(p12h), p1.Add(p12h), size/2)
	lb := hilbert(sourceImage, resultImage, p2.Sub(p12h), p2, p2.Add(p23h), p2.Add(p23h).Sub(p12h), size/2)
	rb := hilbert(sourceImage, resultImage, p3.Sub(p12h).Sub(p23h), p3.Sub(p23h), p3, p3.Sub(p12h), size/2)
	rt := hilbert(sourceImage, resultImage, p4.Add(p12h), p4.Add(p12h).Sub(p23h), p4.Sub(p23h), p4, size/2)
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
func hilbert_curve_filter(im image.Image) image.Image {
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
	hilbert(im, himage, image.Point{im.Bounds().Min.X, im.Bounds().Max.Y}, im.Bounds().Min, image.Point{im.Bounds().Max.X, im.Bounds().Min.Y}, im.Bounds().Max, W)
	return himage
}

func hilbert_curve(sourceImageFilename, resultImageFilename string) error {
	im, err := loadImageFile(sourceImageFilename)
	if err != nil {
		return err
	}
	tmp := hilbert_curve_filter(im)
	return save_image(tmp, resultImageFilename)
}

func hilbert_darken(sourceImageFilename, resultImageFilename string) error {
	im, err := loadImageFile(sourceImageFilename)
	if err != nil {
		return err
	}
	tmp := hilbert_curve_filter(im)
	// TODO: uncomment
	// for i := tmp.Bounds().Min.X; i < tmp.Bounds().Max.X; i++ {
	// 	for j := tmp.Bounds().Min.Y; j < tmp.Bounds().Max.y; j++ {
	// 		for k := 0; k < 3; k++ {
	// 			if im.At(i, j)[k] < tmp.At(i, j)[k] {
	// 				tmp.At(i, j)[k] = im.At(i, j)[k]
	// 			}
	// 		}
	// 	}
	// }
	return save_image(tmp, resultImageFilename)
}

func shader_filter(imid, fragment_shader_source string) (string, error) {
	return "", fmt.Errorf("Not implemented")
}

//mux.HandleFunc("/blur", filterToHandler(&convolutionFilter{
//    BasicFilter{"Blur", "filter.html", pages_templates},
//    [][]int{
//        {1, 1, 1},
//        {1, 1, 1},
//        {1, 1, 1},
//    },
//}))
//
//mux.HandleFunc("/weakblur", filterToHandler(&convolutionFilter{
//    BasicFilter{"Weak blur", "filter.html", pages_templates},
//    [][]int{
//        {0, 1, 0},
//        {1, 1, 1},
//        {0, 1, 0},
//    },
//}))
//
//mux.HandleFunc("/emboss", filterToHandler(&convolutionFilter{
//    BasicFilter{"Emboss", "filter.html", pages_templates},
//    [][]int{
//        {-2, -1, 0},
//        {-1, 1, 1},
//        {0, 1, 2},
//    },
//}))
//
//mux.HandleFunc("/sharpen", filterToHandler(&convolutionFilter{
//    BasicFilter{"Sharpen", "filter.html", pages_templates},
//    [][]int{
//        {0, -1, 0},
//        {-1, 5, -1},
//        {0, -1, 0},
//    },
//}))
//
//mux.HandleFunc("/edgeenhance", filterToHandler(&convolutionFilter{
//    BasicFilter{"Edge enhance", "filter.html", pages_templates},
//    [][]int{
//        {0, 0, 0},
//        {-1, 1, 0},
//        {0, 0, 0},
//    },
//}))
//
//mux.HandleFunc("/edgedetect1", filterToHandler(&convolutionFilter{
//    BasicFilter{"Edge detect 1", "filter.html", pages_templates},
//    [][]int{
//        {1, 0, -1},
//        {0, 0, 0},
//        {-1, 0, 1},
//    },
//}))
//
//mux.HandleFunc("/edgedetect2", filterToHandler(&convolutionFilter{
//    BasicFilter{"Edge detect 2", "filter.html", pages_templates},
//    [][]int{
//        {0, -1, 0},
//        {-1, 4, -1},
//        {0, -1, 0},
//    },
//}))
//
//mux.HandleFunc("/horizontallines", filterToHandler(&convolutionFilter{
//    BasicFilter{"Horizontal lines", "filter.html", pages_templates},
//    [][]int{
//        {-1, -1, -1},
//        {2, 2, 2},
//        {-1, -1, -1},
//    },
//}))
//
//mux.HandleFunc("/verticallines", filterToHandler(&convolutionFilter{
//    BasicFilter{"Vertical lines", "filter.html", pages_templates},
//    [][]int{
//        {-1, 2, -1},
//        {-1, 2, -1},
//        {-1, 2, -1},
//    },
//}))

