package fimgs

import (
	"fmt"
	"image"
	"image/color"
	"math/rand"
	"os"
	"runtime/pprof"
)

func abs64(x int64) int64 {
	if x < 0 {
		return -x
	} else {
		return x
	}
}

func minkowskiiDist(a, b []int64) int64 {
	return abs64(a[0]-b[0]) + abs64(a[1]-b[1]) + abs64(a[2]-b[2]) + abs64(a[3]-b[3]) + abs64(a[4]-b[4])
}

func initClusterCenters(pixelColors [][]int64, clustersCount int) [][]int64 {
	clustersCenters := make([][]int64, clustersCount)
	clustersCenters[0] = pixelColors[rand.Intn(len(pixelColors))]
	minClusterDistance := make([]float64, len(pixelColors))
	minClusterDistanceSum := 0.0
	for i, pixelColor := range pixelColors {
		minClusterDistance[i] = float64(minkowskiiDist(pixelColor, clustersCenters[0]))
		minClusterDistanceSum += minClusterDistance[i]
	}
	for k := 1; k < clustersCount; k++ {
		var clusterCenter []int64
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

func kmeansIters(clustersCenters, pixelColors [][]int64, clustersCount, imageWidth int) {
	// TODO: optimize/parallelize
	for epoch := 0; epoch < 300; epoch++ {
		sumAndCount := make([]int64, clustersCount*6) // sum of Rs, Gs, Bs and count
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
			sumAndCount[minCluster*6+0] += pixelColor[0]
			sumAndCount[minCluster*6+1] += pixelColor[1]
			sumAndCount[minCluster*6+2] += pixelColor[2]
			sumAndCount[minCluster*6+3] += int64(i % imageWidth)
			sumAndCount[minCluster*6+4] += int64(i / imageWidth)
			sumAndCount[minCluster*6+5]++
		}
		movement := int64(0)
		for i := 0; i < clustersCount; i++ {
			count := sumAndCount[i*6+5]
			if count == 0 {
				continue
			}
			sumAndCount[i*6+0] /= count
			sumAndCount[i*6+1] /= count
			sumAndCount[i*6+2] /= count
			sumAndCount[i*6+3] /= count
			sumAndCount[i*6+4] /= count
			movement += minkowskiiDist(clustersCenters[i], sumAndCount[i*6:i*6+5])
			clustersCenters[i] = sumAndCount[i*6 : i*6+5]
		}
		if movement < 100 {
			break
		}
	}
}

func ApplyKMeans(im image.Image, clustersCount int) image.Image {
	pixelColors := make([][]int64, im.Bounds().Dx()*im.Bounds().Dy())
	mean := [3]int64{}
	for i := im.Bounds().Min.X; i < im.Bounds().Max.X; i++ {
		for j := im.Bounds().Min.Y; j < im.Bounds().Max.Y; j++ {
			r, g, b, _ := im.At(i, j).RGBA()
			pixelColors[i+j*im.Bounds().Dx()] = []int64{int64(r), int64(g), int64(b), int64(i), int64(j)}
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
	kmeansIters(clustersCenters, pixelColors, clustersCount, im.Bounds().Dx())
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
	f, _ := os.Create("cpu.pb")
	defer f.Close() // error handling omitted for example
	pprof.StartCPUProfile(f)
	defer pprof.StopCPUProfile()

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
