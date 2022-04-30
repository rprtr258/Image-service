package fimgs

import (
	"fmt"
	"image"
	"image/color"
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
