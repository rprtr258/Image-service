package fimgs

import (
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"math"
)

func min(x, y int) int {
	if x < y {
		return x
	} else {
		return y
	}
}

func min4(x, y, z, w int) int {
	return min(min(x, y), min(z, w))
}

func calcMinColor4(c1, c2, c3, c4 [3]int) [3]int {
	return [3]int{
		min4(c1[0], c2[0], c3[0], c4[0]),
		min4(c1[1], c2[1], c3[1], c4[1]),
		min4(c1[2], c2[2], c3[2], c4[2]),
	}
}

func max(x, y int) int {
	if x < y {
		return y
	} else {
		return x
	}
}

func max4(x, y, z, w int) int {
	return max(max(x, y), max(z, w))
}

func calcMaxColor4(c1, c2, c3, c4 [3]int) [3]int {
	return [3]int{
		max4(c1[0], c2[0], c3[0], c4[0]),
		max4(c1[1], c2[1], c3[1], c4[1]),
		max4(c1[2], c2[2], c3[2], c4[2]),
	}
}

func parent(p []int, x int) int {
	if p[x] != x {
		p[x] = parent(p, p[x])
	}
	return p[x]
}

func merge(p, blockWidth []int, minColor, maxColor [][3]int, topLeft, topRight, bottomLeft, bottomRight int) {
	topLeft = parent(p, topLeft)
	topRight = parent(p, topRight)
	bottomLeft = parent(p, bottomLeft)
	bottomRight = parent(p, bottomRight)
	p[topRight], p[bottomLeft], p[bottomRight] = topLeft, topLeft, topLeft
	blockWidth[topLeft] *= 2
	minColor[topLeft] = calcMinColor4(minColor[topLeft], minColor[topRight], minColor[bottomLeft], minColor[bottomRight])
	maxColor[topLeft] = calcMaxColor4(maxColor[topLeft], maxColor[topRight], maxColor[bottomLeft], maxColor[bottomRight])
}

// TODO: typedef color
func colorDiff(minColor, maxColor [3]int) int {
	return max(max(maxColor[0]-minColor[0], maxColor[1]-minColor[1]), maxColor[2]-minColor[2])
}

func QuadTree(im image.Image, power float64, threshold int) *image.RGBA {
	imageWidth := im.Bounds().Dx()
	imageSize := imageWidth * im.Bounds().Dy()
	dsuParent := make([]int, imageSize)
	blockWidth := make([]int, imageSize)
	minColor := make([][3]int, imageSize)
	maxColor := make([][3]int, imageSize)
	for i := 0; i < imageSize; i++ {
		dsuParent[i] = i
		blockWidth[i] = 1
		r, g, b, _ := im.At(i%imageWidth, i/imageWidth).RGBA()
		minColor[i] = [3]int{int(r), int(g), int(b)}
		maxColor[i] = [3]int{int(r), int(g), int(b)}
	}
	for j := 1; j < imageWidth; j *= 2 {
		for x := j; x < imageWidth; x += 2 * j {
			for y := j; y < im.Bounds().Dy(); y += 2 * j { // TODO: move 2*j upper
				i := y*imageWidth + x
				cur := parent(dsuParent, i)
				if cur/imageWidth == 0 || cur%imageWidth == 0 {
					continue
				}
				left := cur - 1
				upLeft := cur - imageWidth - 1
				up := cur - imageWidth
				left = parent(dsuParent, left)
				up = parent(dsuParent, up)
				upLeft = parent(dsuParent, upLeft)
				if upLeft < 0 || up == cur || left == cur || upLeft == cur || up == left || up == upLeft || left == upLeft {
					continue
				}
				colDiff := colorDiff(
					calcMinColor4(minColor[cur], minColor[up], minColor[left], minColor[upLeft]),
					calcMaxColor4(maxColor[cur], maxColor[up], maxColor[left], maxColor[upLeft]),
				)
				if blockWidth[cur] == j && blockWidth[up] == j && blockWidth[left] == j && blockWidth[upLeft] == j && colDiff < threshold {
					merge(dsuParent, blockWidth, minColor, maxColor, upLeft, up, left, cur)
				} else {
					continue
				}
			}
		}
	}
	himage := image.NewRGBA(im.Bounds())
	draw.Draw(himage, himage.Bounds(), &image.Uniform{color.RGBA{0, 0, 0, 255}}, image.Point{}, draw.Src)
	for i := 0; i < imageSize; i++ {
		p := parent(dsuParent, i)
		halfQuadSize := blockWidth[p] / 2
		quadCenter := [2]int{p%imageWidth + halfQuadSize, p/imageWidth + halfQuadSize}
		xi, yi := i%imageWidth, i/imageWidth
		dx, dy := xi-quadCenter[0], yi-quadCenter[1]
		if math.Floor(math.Pow(math.Pow(math.Abs(float64(dx)), power)+math.Pow(math.Abs(float64(dy)), power), 1./power)) <= math.Ceil(float64(halfQuadSize)) {
			ci := minColor[p]
			ca := maxColor[p]
			himage.Set(xi, yi, color.RGBA64{uint16((ci[0] + ca[0]) / 2), uint16((ci[1] + ca[1]) / 2), uint16((ci[2] + ca[2]) / 2), 0xFFFF})
		}
	}
	return himage
}

func QudTreeFilter(sourceImageFilename, resultImageFilename string, power float64, threshold int) error {
	if power <= 0.0 {
		return fmt.Errorf("power should be greater than 0")
	}
	if threshold <= 0 || threshold > 0xFFFF {
		return fmt.Errorf("threshold should be greater than 0 and less than 65535")
	}
	im, err := LoadImageFile(sourceImageFilename)
	if err != nil {
		return err
	}
	tmp := QuadTree(im, power, threshold)
	return saveImage(tmp, resultImageFilename)
}
