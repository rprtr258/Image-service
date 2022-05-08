package fimgs

import (
	"fmt"
	"image"
	"image/color"
	"math"
	"math/rand"
	"os"
	"runtime/pprof"
)

func makeColorArray(len int) [][]int64 {
	data := make([]int64, len*3)
	res := make([][]int64, len)
	for i := 0; i < len; i++ {
		res[i] = data[i*3 : i*3+3]
	}
	return res
}

func abs64(x int64) int64 {
	mask := x >> 63
	return (x + mask) ^ mask
}

func minkowskiiDist(a, b []int64) int64 {
	return abs64(a[0]-b[0]) + abs64(a[1]-b[1]) + abs64(a[2]-b[2])
}

func initClusterCenters(pixelColors [][]int64, clustersCount int) [][]int64 {
	clustersCenters := makeColorArray(clustersCount)
	copy(clustersCenters[0], pixelColors[rand.Intn(len(pixelColors))])
	minClusterDistance := make([]int64, len(pixelColors))
	minClusterDistanceSum := int64(0)
	for i, pixelColor := range pixelColors {
		minClusterDistance[i] = minkowskiiDist(pixelColor, clustersCenters[0])
		minClusterDistanceSum += minClusterDistance[i]
	}
	for k := 1; k < clustersCount; k++ {
		x := rand.Int63n(minClusterDistanceSum)
		for i, pixelColor := range pixelColors {
			x -= minClusterDistance[i]
			if x < 0 {
				copy(clustersCenters[k], pixelColor)
				break
			}
		}
		if k == clustersCount-1 {
			break
		}
		for i, pixelColor := range pixelColors {
			newDistance := minkowskiiDist(pixelColor, clustersCenters[0])
			if newDistance < minClusterDistance[i] {
				minClusterDistanceSum += newDistance - minClusterDistance[i]
				minClusterDistance[i] = newDistance
			}
		}
	}
	return clustersCenters
}

func kmeansIters(clustersCenters, pixelColors [][]int64, clustersCount int) {
	batchMaxSize := int(math.Sqrt(float64(len(pixelColors))))
	sumAndCount := make([]int64, clustersCount*4) // sum of Rs, Gs, Bs and count
	for epoch := 0; epoch < 300; epoch++ {
		k := rand.Intn(batchMaxSize) + 1
		sumAndCount[0] = 0
		for i := 1; i < len(sumAndCount); i *= 2 {
			copy(sumAndCount[i:], sumAndCount[:i])
		}
		for i := k; i < len(pixelColors); i += k {
			pixelColor := pixelColors[i]
			minCluster := 0
			minDist := minkowskiiDist(pixelColor, clustersCenters[0])
			for k := 1; k < clustersCount; k++ {
				newDist := minkowskiiDist(pixelColor, clustersCenters[k])
				if newDist < minDist {
					minCluster = k
					minDist = newDist
				}
			}
			sumAndCount[minCluster*4+0] += pixelColor[0]
			sumAndCount[minCluster*4+1] += pixelColor[1]
			sumAndCount[minCluster*4+2] += pixelColor[2]
			sumAndCount[minCluster*4+3]++
		}
		movement := int64(0)
		for i := 0; i < clustersCount; i++ {
			count := sumAndCount[i*4+3]
			if count == 0 {
				continue
			}
			sumAndCount[i*4+0] /= count
			sumAndCount[i*4+1] /= count
			sumAndCount[i*4+2] /= count
			movement += minkowskiiDist(clustersCenters[i], sumAndCount[i*4:i*4+4])
			copy(clustersCenters[i], sumAndCount[i*4:i*4+4])
		}
		if movement < 100 {
			break
		}
	}
}

func ApplyKMeans(im image.Image, clustersCount int) image.RGBA {
	imageWidth := im.Bounds().Dx()
	pixelColors := makeColorArray(imageWidth * im.Bounds().Dy())
	for j := 0; j < im.Bounds().Dy(); j++ {
		for i := 0; i < im.Bounds().Dx(); i++ {
			k := i + j*imageWidth
			r, g, b, _ := im.At(i, j).RGBA()
			pixelColors[k][0] = int64(r)
			pixelColors[k][1] = int64(g)
			pixelColors[k][2] = int64(b)
		}
	}
	rand.Seed(0)
	clustersCenters := initClusterCenters(pixelColors, clustersCount)
	// TODO: try to sample mini-batches (random subdatasets)
	kmeansIters(clustersCenters, pixelColors, clustersCount)
	filtered_im := image.NewRGBA(im.Bounds())
	for j := 0; j < im.Bounds().Dy(); j++ {
		for i := 0; i < im.Bounds().Dx(); i++ {
			pixel := pixelColors[i+j*imageWidth]
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
				uint8((clustersCenters[minCluster][0] & 0xFF00) >> 8),
				uint8((clustersCenters[minCluster][1] & 0xFF00) >> 8),
				uint8((clustersCenters[minCluster][2] & 0xFF00) >> 8),
				255,
			})
		}
	}
	return *filtered_im
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
