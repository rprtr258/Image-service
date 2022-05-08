package fimgs

import (
	"image"
	_ "image/jpeg"
	"image/png"
	"os"
)

type Color = [3]int

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

func saveImage(im image.RGBA, imageFilename string) (err error) {
	imageFile, err := os.Create(imageFilename)
	if err != nil {
		return
	}
	defer imageFile.Close()
	if err = png.Encode(imageFile, &im); err != nil {
		return err
	}
	return
}
