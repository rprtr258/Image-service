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

func distSquared(a, b [3]uint32) uint32 {
	dx, dy, dz := absDiff(a[0], b[0]), absDiff(a[1], b[1]), absDiff(a[2], b[2])
	return dx*dx + dy*dy + dz*dz
}

func initClusterCenters(pixelColors [][3]uint32, clustersCount int) [][3]uint32 {
	clustersCenters := make([][3]uint32, clustersCount)
	clustersCenters[0] = pixelColors[rand.Intn(len(pixelColors))]
	minClusterDistance := make([]float64, len(pixelColors))
	minClusterDistanceSum := 0.0
	for i, pixelColor := range pixelColors {
		minClusterDistance[i] = float64(distSquared(pixelColor, clustersCenters[0]))
		minClusterDistanceSum += minClusterDistance[i]
	}
	for k := 1; k < clustersCount; k++ {
		var clusterCenter [3]uint32
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
	pixelColors := make([][3]uint32, im.Bounds().Dx()*im.Bounds().Dy())
	for i := im.Bounds().Min.X; i < im.Bounds().Max.X; i++ {
		for j := im.Bounds().Min.Y; j < im.Bounds().Max.Y; j++ {
			r, g, b, _ := im.At(i, j).RGBA()
			pixelColors[i+j*im.Bounds().Dx()] = [3]uint32{r, g, b}
		}
	}
	rand.Seed(0)
	clustersCenters := initClusterCenters(pixelColors, clustersCount)
	sumAndCount := make([][4]uint64, clustersCount) // sum of Rs, Gs, Bs and count
	// TODO: optimize/parallelize
	for epoch := 0; epoch < 100; epoch++ {
		for i := 0; i < clustersCount; i++ {
			sumAndCount[i][0], sumAndCount[i][1], sumAndCount[i][2], sumAndCount[i][3] = 0, 0, 0, 0
		}
		for _, pixelColor := range pixelColors {
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
			sumAndCount[minCluster][3]++
		}
		accumChange := uint32(0)
		for i := 0; i < clustersCount; i++ {
			count := sumAndCount[i][3]
			if count == 0 {
				continue
			}
			newR, newG, newB := uint32(sumAndCount[i][0]/count), uint32(sumAndCount[i][1]/count), uint32(sumAndCount[i][2]/count)
			accumChange += absDiff(newR, clustersCenters[i][0]) + absDiff(newG, clustersCenters[i][1]) + absDiff(newB, clustersCenters[i][2])
			clustersCenters[i][0], clustersCenters[i][1], clustersCenters[i][2] = newR, newG, newB
		}
		if accumChange < 500 {
			break
		}
	}
	filtered_im := image.NewRGBA(im.Bounds())
	for i := im.Bounds().Min.X; i < im.Bounds().Max.X; i++ {
		for j := im.Bounds().Min.Y; j < im.Bounds().Max.Y; j++ {
			r, g, b, _ := im.At(i, j).RGBA()
			minCluster := 0
			minDist := (r-clustersCenters[0][0])*(r-clustersCenters[0][0]) + (g-clustersCenters[0][1])*(g-clustersCenters[0][1]) + (b-clustersCenters[0][2])*(b-clustersCenters[0][2])
			for k := 1; k < clustersCount; k++ {
				dist := (r-clustersCenters[k][0])*(r-clustersCenters[k][0]) + (g-clustersCenters[k][1])*(g-clustersCenters[k][1]) + (b-clustersCenters[k][2])*(b-clustersCenters[k][2])
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
