package fimgs

import (
	"fmt"
	"image"
	"image/color"
	"math"
)

var (
	// TODO: flat data layout
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

func ApplyConvolution(im image.Image, kernel [][]int) image.Image {
	kernelHalfWidth, kernelHalfHeight := len(kernel)/2, len(kernel)/2
	// TODO: flat data layout
	R := make([][]Color, im.Bounds().Dx())
	for i := im.Bounds().Min.X; i < im.Bounds().Max.X; i++ {
		R[i] = make([]Color, im.Bounds().Dy())
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
			R[i][j] = Color{r, g, b}
		}
	}
	kernelMin, kernelMax := math.MaxInt, math.MinInt
	for i := im.Bounds().Min.X; i < im.Bounds().Max.X; i++ {
		for j := im.Bounds().Min.Y; j < im.Bounds().Max.Y; j++ {
			kernelMin = min4(kernelMin, R[i][j][0], R[i][j][1], R[i][j][2])
			kernelMax = max4(kernelMax, R[i][j][0], R[i][j][1], R[i][j][2])
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

func ApplyConvolutionFilter(sourceImageFilename, resultImageFilename string, kernel [][]int) error {
	im, err := LoadImageFile(sourceImageFilename)
	if err != nil {
		return fmt.Errorf("error occured during loading image:\n%q", err)
	}
	resImage := ApplyConvolution(im, kernel)
	return saveImage(resImage, resultImageFilename)
}
