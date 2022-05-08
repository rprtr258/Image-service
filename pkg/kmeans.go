package fimgs

import (
	"fmt"
	"image"
	"image/color"
	"math"
	"math/rand"

	"github.com/dhconnelly/rtreego"
)

func minkowskiiDist(a, b []float64) float64 {
	return math.Abs(a[0]-b[0]) + math.Abs(a[1]-b[1]) + math.Abs(a[2]-b[2])
}

func initClusterCenters(pixelColors [][]float64, clustersCount int) [][]float64 {
	clustersCenters := make([][]float64, clustersCount)
	clustersCenters[0] = pixelColors[rand.Intn(len(pixelColors))]
	minClusterDistance := make([]float64, len(pixelColors))
	minClusterDistanceSum := float64(0)
	for i, pixelColor := range pixelColors {
		minClusterDistance[i] = minkowskiiDist(pixelColor, clustersCenters[0])
		minClusterDistanceSum += minClusterDistance[i]
	}
	for k := 1; k < clustersCount; k++ {
		var clusterCenter []float64
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
	for epoch := 0; epoch < 300; epoch++ {
		sumAndCount := make([]int64, clustersCount*4) // sum of Rs, Gs, Bs and count
		for _, pixelColor := range pixelColors {
			minCluster := 0
			minDist := minkowskiiDist(pixelColor, clustersCenters[0])
			for k := 1; k < clustersCount; k++ {
				newDist := minkowskiiDist(pixelColor, clustersCenters[k])
				if newDist < minDist {
					minCluster = k
					minDist = newDist
				}
			}
		}
		rgbSum := make([]float64, clustersCount*4)
		for k, v := range minCluster {
			rgbSum[v*4+0]++
			rgbSum[v*4+1] += pixelColors[k][0]
			rgbSum[v*4+2] += pixelColors[k][1]
			rgbSum[v*4+3] += pixelColors[k][2]
		}
		movement := 0.0
		for i := 0; i < clustersCount; i++ {
			count := rgbSum[i*4]
			r, g, b := rgbSum[i*4+1], rgbSum[i*4+2], rgbSum[i*4+3]
			if count == 0 {
				continue
			}
			sumAndCount[i*4+0] /= count
			sumAndCount[i*4+1] /= count
			sumAndCount[i*4+2] /= count
			movement += minkowskiiDist(clustersCenters[i], sumAndCount[i*4:i*4+4])
			clustersCenters[i] = sumAndCount[i*4 : i*4+4]
		}
		if movement < 100 {
			break
		}
	}
}

type Somewhere struct {
	pixelColors [][]float64
	idx         int
}

func (s Somewhere) Bounds() *rtreego.Rect {
	return rtreego.Point(s.pixelColors[s.idx]).ToRect(0.0)
}

func ApplyKMeans(im image.Image, clustersCount int) image.Image {
	pixelColors := make([][]int64, im.Bounds().Dx()*im.Bounds().Dy())
	mean := [3]int64{}
	for i := im.Bounds().Min.X; i < im.Bounds().Max.X; i++ {
		for j := im.Bounds().Min.Y; j < im.Bounds().Max.Y; j++ {
			r, g, b, _ := im.At(i, j).RGBA()
			pixelColors[i+j*im.Bounds().Dx()] = []int64{int64(r), int64(g), int64(b)}
			mean[0] += int64(r)
			mean[1] += int64(g)
			mean[2] += int64(b)
		}
	}
	tr := rtreego.NewTree(3, 5000, 10000)
	for i := range pixelColors {
		if i%100 == 0 {
			tr.Insert(&Somewhere{pixelColors, i})
		}
	}
	rand.Seed(0)
	clustersCenters := initClusterCenters(pixelColors, clustersCount)
	// TODO: try to sample mini-batches (random subdatasets)
	kmeansIters(pixelColors, clustersCenters, clustersCount, tr)
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
				uint8((clustersCenters[minCluster][0]) / 0x100),
				uint8((clustersCenters[minCluster][1]) / 0x100),
				uint8((clustersCenters[minCluster][2]) / 0x100),
				255,
			})
		}
	}
	return filtered_im
}

// TODO: filter init is also validation?
func ApplyKMeansFilter(sourceImageFilename string, resultImageFilename string, clustersCount int) (err error) {
	// f, _ := os.Create("cpu.pb")
	// defer f.Close() // error handling omitted for example
	// pprof.StartCPUProfile(f)
	// defer pprof.StopCPUProfile()

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
