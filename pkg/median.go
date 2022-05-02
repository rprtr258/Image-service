package fimgs

import (
	"fmt"
	"image"
	"image/color"
	"math"
	"math/rand"
)

func Hsv2Rgb(h, s, v int) (int, int, int) {
	_h := float64(h) / 60
	_s := float64(s) / 100
	_v := float64(v) / 100
	f := _h - math.Floor(_h)
	_v *= 255
	p := _v * (1 - _s)
	q := _v * (1 - (_s * f))
	t := _v * (1 - (_s * (1 - f)))
	__v, __t, __p, __q := int(math.Floor(_v)), int(math.Floor(t)), int(math.Floor(p)), int(math.Floor(q))
	switch math.Mod(math.Floor(_h), 6) {
	case 0:
		return __v, __t, __p
	case 1:
		return __q, __v, __p
	case 2:
		return __p, __v, __t
	case 3:
		return __p, __q, __v
	case 4:
		return __t, __p, __v
	case 5:
		return __v, __p, __q
	}
	return 0, 0, 0
}

func Rgb2Hsv(rgb Color) (int, int, int) {
	var s float64
	r := float64(rgb[0]) / 0xFFFF
	g := float64(rgb[1]) / 0xFFFF
	b := float64(rgb[2]) / 0xFFFF
	min := math.Min(math.Min(r, g), b)
	v := math.Max(math.Max(r, g), b)
	diff := v - min
	diffc := func(c float64) float64 {
		return (v-c)/6.0/diff + 0.5
	}
	if diff == 0 {
		return 0, 0, int(math.Round(v * 100))
	} else {
		s = diff / v
		rdif := diffc(r)
		gdif := diffc(g)
		bdif := diffc(b)
		var h float64
		if r == v {
			h = bdif - gdif
		} else if g == v {
			h = (1.0 / 3.0) + rdif - bdif
		} else if b == v {
			h = (2.0 / 3.0) + gdif - rdif
		}
		if h < 0 {
			h += 1
		} else if h > 1 {
			h -= 1
		}
		return int(math.Round(h * 360)), int(math.Round(s * 100)), int(math.Round(v * 100))
	}
}

func randomPartition(arr []int, l, r int) int {
	n := r - l
	pivot := rand.Intn(n)
	arr[l+pivot], arr[l] = arr[l], arr[l+pivot]
	x := arr[l]
	i := l + 1
	for j := l + 2; j < r; j++ {
		if arr[j] <= x {
			arr[i], arr[j] = arr[j], arr[i]
			i++
		}
	}
	arr[i], arr[l] = arr[l], arr[i]
	return i
}

func kthSmallest(arr []int, l, r, k int) int {
	pos := randomPartition(arr, l, r)
	leftPartSize := pos - l
	switch {
	case leftPartSize == k:
		return arr[pos]
	case leftPartSize > k:
		return kthSmallest(arr, l, pos, k)
	default:
		return kthSmallest(arr, pos, r, k-leftPartSize)
	}
}

func Median(im image.Image, windowSize int) image.Image {
	halfWindowSize := windowSize / 2
	himage := image.NewRGBA(im.Bounds())
	window := make([]Color, windowSize*windowSize)
	hWindow := make([]int, windowSize*windowSize)
	sWindow := make([]int, windowSize*windowSize)
	vWindow := make([]int, windowSize*windowSize)
	for i := im.Bounds().Min.X; i < im.Bounds().Max.X; i++ {
		for j := im.Bounds().Min.Y; j < im.Bounds().Max.Y; j++ {
			k := 0
			for ki := -halfWindowSize; ki <= halfWindowSize; ki++ {
				for kj := -halfWindowSize; kj <= halfWindowSize; kj++ {
					r, g, b, _ := im.At(
						min(max(im.Bounds().Min.X, i+ki), im.Bounds().Max.X),
						min(max(im.Bounds().Min.Y, j+kj), im.Bounds().Max.Y),
					).RGBA()
					window[k] = Color{int(r), int(g), int(b)}
					hWindow[k], sWindow[k], vWindow[k] = Rgb2Hsv(window[k])
					k++
				}
			}
			// TODO: somehow sort by color
			h := kthSmallest(hWindow, 0, len(hWindow), len(hWindow)/2)
			s := kthSmallest(sWindow, 0, len(sWindow), len(sWindow)/2)
			v := kthSmallest(vWindow, 0, len(vWindow), len(vWindow)/2)
			r, g, b := Hsv2Rgb(h, s, v)
			himage.Set(i, j, color.RGBA64{
				uint16(r * 0x100),
				uint16(g * 0x100),
				uint16(b * 0x100),
				0xFFFF,
			})
		}
	}
	return himage
}

func MedianFilter(sourceImageFilename, resultImageFilename string, windowSize int) error {
	if windowSize < 0 || windowSize%2 == 0 {
		return fmt.Errorf("window size must be positive and odd, but it isn't: %d", windowSize)
	}
	im, err := LoadImageFile(sourceImageFilename)
	if err != nil {
		return fmt.Errorf("error occured during loading image:\n%q", err)
	}
	resImage := Median(im, windowSize)
	return saveImage(resImage, resultImageFilename)
}
