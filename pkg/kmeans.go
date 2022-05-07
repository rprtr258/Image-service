package fimgs

import (
	"fmt"
	"image"
	"image/color"
	"math/rand"
)

func abs64(x int64) int64 {
	if x < 0 {
		return -x
	} else {
		return x
	}
}

func minkowskiiDist(a, b [5]int64) int64 {
	res := int64(0)
	for i, ai := range a {
		di := abs64(ai - b[i])
		res += di
	}
	return res
}

func initClusterCenters(pixelColors [][5]int64, clustersCount int) [][5]int64 {
	clustersCenters := make([][5]int64, clustersCount)
	clustersCenters[0] = pixelColors[rand.Intn(len(pixelColors))]
	minClusterDistance := make([]float64, len(pixelColors))
	minClusterDistanceSum := 0.0
	for i, pixelColor := range pixelColors {
		minClusterDistance[i] = float64(minkowskiiDist(pixelColor, clustersCenters[0]))
		minClusterDistanceSum += minClusterDistance[i]
	}
	for k := 1; k < clustersCount; k++ {
		var clusterCenter [5]int64
		x := rand.Float64() * minClusterDistanceSum
		for i, pixelColor := range pixelColors {
			x -= minClusterDistance[i]
			if x < 0 {
				clusterCenter = pixelColor
				break
			}
		}
		clustersCenters[k] = clusterCenter
		if k == clustersCount-1 {
			break
		}
		for i, pixelColor := range pixelColors {
			newDistance := float64(minkowskiiDist(pixelColor, clustersCenters[0]))
			if newDistance < minClusterDistance[i] {
				minClusterDistanceSum += newDistance - minClusterDistance[i]
				minClusterDistance[i] = newDistance
			}
		}
	}
	return clustersCenters
}

func ApplyKMeans(im image.Image, clustersCount int) image.Image {
	pixelColors := make([][5]int64, im.Bounds().Dx()*im.Bounds().Dy())
	mean := [3]int64{}
	for i := im.Bounds().Min.X; i < im.Bounds().Max.X; i++ {
		for j := im.Bounds().Min.Y; j < im.Bounds().Max.Y; j++ {
			r, g, b, _ := im.At(i, j).RGBA()
			pixelColors[i+j*im.Bounds().Dx()] = [5]int64{int64(r), int64(g), int64(b), int64(i * 2000), int64(j * 2000)}
			mean[0] += int64(r)
			mean[1] += int64(g)
			mean[2] += int64(b)
		}
	}
	mean[0] /= int64(len(pixelColors))
	mean[1] /= int64(len(pixelColors))
	mean[2] /= int64(len(pixelColors))
	for i := range pixelColors {
		pixelColors[i][0] -= mean[0]
		pixelColors[i][1] -= mean[1]
		pixelColors[i][2] -= mean[2]
	}
	rand.Seed(0)
	clustersCenters := initClusterCenters(pixelColors, clustersCount)
	sumAndCount := make([][6]int64, clustersCount) // sum of Rs, Gs, Bs and count
	// TODO: optimize/parallelize
	for epoch := 0; epoch < 100; epoch++ {
		for i := 0; i < clustersCount; i++ {
			sumAndCount[i] = [6]int64{}
		}
		for i, pixelColor := range pixelColors {
			minCluster := 0
			minDist := minkowskiiDist(pixelColor, clustersCenters[0])
			for k := 1; k < clustersCount; k++ {
				newDist := minkowskiiDist(pixelColor, clustersCenters[k])
				if newDist < minDist {
					minCluster = k
					minDist = newDist
				}
			}
			sumAndCount[minCluster][0] += pixelColor[0]
			sumAndCount[minCluster][1] += pixelColor[1]
			sumAndCount[minCluster][2] += pixelColor[2]
			sumAndCount[minCluster][3] += int64(i % im.Bounds().Dx())
			sumAndCount[minCluster][4] += int64(i / im.Bounds().Dx())
			sumAndCount[minCluster][5]++
		}
		for i := 0; i < clustersCount; i++ {
			count := sumAndCount[i][5]
			if count == 0 {
				continue
			}
			clustersCenters[i][0] = sumAndCount[i][0] / count
			clustersCenters[i][1] = sumAndCount[i][1] / count
			clustersCenters[i][2] = sumAndCount[i][2] / count
			clustersCenters[i][3] = sumAndCount[i][3] / count
			clustersCenters[i][4] = sumAndCount[i][4] / count
		}
	}
	filtered_im := image.NewRGBA(im.Bounds())
	for i := im.Bounds().Min.X; i < im.Bounds().Max.X; i++ {
		for j := im.Bounds().Min.Y; j < im.Bounds().Max.Y; j++ {
			pixel := pixelColors[i+j*im.Bounds().Dx()]
			minCluster := 0
			minDist := minkowskiiDist(pixel, clustersCenters[0])
			for k := 1; k < clustersCount; k++ {
				dist := minkowskiiDist(pixel, clustersCenters[k])
				if dist < minDist {
					minCluster = k
					minDist = dist
				}
			}
			filtered_im.Set(i, j, color.RGBA{
				uint8((clustersCenters[minCluster][0] + mean[0]) / 0x100),
				uint8((clustersCenters[minCluster][1] + mean[1]) / 0x100),
				uint8((clustersCenters[minCluster][2] + mean[2]) / 0x100),
				255,
			})
		}
	}
	return filtered_im
}

// TODO: filter init is also validation?
func ApplyKMeansFilter(sourceImageFilename string, resultImageFilename string, clustersCount int) (err error) {
	if clustersCount < 2 {
		return fmt.Errorf("'n' must be at least 2, you gave n=%d", clustersCount)
	}
	im, err := LoadImageFile(sourceImageFilename)
	if err != nil {
		return fmt.Errorf("error occured while loading image:\n%q", err)
	}
	filtered_im := ApplyKMeans(im, clustersCount)
	err = saveImage(filtered_im, resultImageFilename)
	if err != nil {
		return
	}
	return
}
