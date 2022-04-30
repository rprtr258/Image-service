package fimgs

import (
	"fmt"
	"image"
	"image/color"
	"math/rand"
)

func absDiff(x, y uint32) uint32 {
	if x > y {
		return x - y
	} else {
		return y - x
	}
}

func distSquared(a, b [5]uint32) uint32 {
	res := uint32(0)
	for i, ai := range a {
		di := absDiff(ai, b[i])
		res += di * di
	}
	return res
}

func initClusterCenters(pixelColors [][5]uint32, clustersCount int) [][5]uint32 {
	clustersCenters := make([][5]uint32, clustersCount)
	clustersCenters[0] = pixelColors[rand.Intn(len(pixelColors))]
	minClusterDistance := make([]float64, len(pixelColors))
	minClusterDistanceSum := 0.0
	for i, pixelColor := range pixelColors {
		minClusterDistance[i] = float64(distSquared(pixelColor, clustersCenters[0]))
		minClusterDistanceSum += minClusterDistance[i]
	}
	for k := 1; k < clustersCount; k++ {
		// clustersCenters[k] = pixelColors[rand.Intn(len(pixelColors))]
		var clusterCenter [5]uint32
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
			newDistance := float64(distSquared(pixelColor, clustersCenters[0]))
			if newDistance < minClusterDistance[i] {
				minClusterDistanceSum += newDistance - minClusterDistance[i]
				minClusterDistance[i] = newDistance
			}
		}
	}
	return clustersCenters
}

func ApplyKMeans(im image.Image, clustersCount int) image.Image {
	pixelColors := make([][5]uint32, im.Bounds().Dx()*im.Bounds().Dy())
	for i := im.Bounds().Min.X; i < im.Bounds().Max.X; i++ {
		for j := im.Bounds().Min.Y; j < im.Bounds().Max.Y; j++ {
			r, g, b, _ := im.At(i, j).RGBA()
			pixelColors[i+j*im.Bounds().Dx()] = [5]uint32{r, g, b, uint32(i) * 2000, uint32(j) * 2000}
		}
	}
	rand.Seed(0)
	clustersCenters := initClusterCenters(pixelColors, clustersCount)
	sumAndCount := make([][6]uint64, clustersCount) // sum of Rs, Gs, Bs and count
	// TODO: optimize/parallelize
	for epoch := 0; epoch < 100; epoch++ {
		for i := 0; i < clustersCount; i++ {
			sumAndCount[i] = [6]uint64{}
		}
		for i, pixelColor := range pixelColors {
			minCluster := 0
			minDist := distSquared(pixelColor, clustersCenters[0])
			for k := 1; k < clustersCount; k++ {
				newDist := distSquared(pixelColor, clustersCenters[k])
				if newDist < minDist {
					minCluster = k
					minDist = newDist
				}
			}
			sumAndCount[minCluster][0] += uint64(pixelColor[0])
			sumAndCount[minCluster][1] += uint64(pixelColor[1])
			sumAndCount[minCluster][2] += uint64(pixelColor[2])
			sumAndCount[minCluster][3] += uint64(i % im.Bounds().Dx())
			sumAndCount[minCluster][4] += uint64(i / im.Bounds().Dx())
			sumAndCount[minCluster][5]++
		}
		// accumChange := uint32(0)
		for i := 0; i < clustersCount; i++ {
			count := sumAndCount[i][5]
			if count == 0 {
				continue
			}
			// newR, newG, newB := uint32(sumAndCount[i][0]/count), uint32(sumAndCount[i][1]/count), uint32(sumAndCount[i][2]/count)
			// accumChange += absDiff(newR, clustersCenters[i][0]) + absDiff(newG, clustersCenters[i][1]) + absDiff(newB, clustersCenters[i][2])
			clustersCenters[i][0] = uint32(sumAndCount[i][0] / count)
			clustersCenters[i][1] = uint32(sumAndCount[i][1] / count)
			clustersCenters[i][2] = uint32(sumAndCount[i][2] / count)
			clustersCenters[i][3] = uint32(sumAndCount[i][3] / count)
			clustersCenters[i][4] = uint32(sumAndCount[i][4] / count)
		}
		// if accumChange < 500 {
		// 	break
		// }
	}
	filtered_im := image.NewRGBA(im.Bounds())
	for i := im.Bounds().Min.X; i < im.Bounds().Max.X; i++ {
		for j := im.Bounds().Min.Y; j < im.Bounds().Max.Y; j++ {
			r, g, b, _ := im.At(i, j).RGBA()
			pixel := [5]uint32{r, g, b, uint32(i), uint32(j)}
			minCluster := 0
			minDist := distSquared(pixel, clustersCenters[0])
			for k := 1; k < clustersCount; k++ {
				dist := distSquared(pixel, clustersCenters[k])
				if dist < minDist {
					minCluster = k
					minDist = dist
				}
			}
			filtered_im.Set(i, j, color.RGBA{
				uint8(clustersCenters[minCluster][0] / 0x100),
				uint8(clustersCenters[minCluster][1] / 0x100),
				uint8(clustersCenters[minCluster][2] / 0x100),
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
