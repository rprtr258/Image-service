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
		minClusterDistance[i] = minkowskiiDist(pixelColor[:], clustersCenters[0][:])
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
			newDistance := minkowskiiDist(pixelColor[:], clustersCenters[0][:])
			if newDistance < minClusterDistance[i] {
				minClusterDistanceSum += newDistance - minClusterDistance[i]
				minClusterDistance[i] = newDistance
			}
		}
	}
	return clustersCenters
}

func kmeansIters(pixelColors, clustersCenters [][]float64, clustersCount int, tr *rtreego.Rtree) {
	for epoch := 0; epoch < 100; epoch++ {
		minCluster := make(map[int]int)
		minDist := make(map[int]float64)
		for i := 0; i < clustersCount; i++ {
			for _, x := range tr.SearchIntersect(rtreego.Point(clustersCenters[i]).ToRect(65536.)) {
				x := x.(*Somewhere)
				_, ok := minCluster[x.idx]
				ppp := pixelColors[x.idx]
				if !ok || minkowskiiDist(ppp, clustersCenters[i][:]) < minDist[x.idx] {
					minCluster[x.idx] = i
					minDist[x.idx] = minkowskiiDist(ppp, clustersCenters[i][:])
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
			r, g, b = r/count, g/count, b/count
			p := []float64{float64(r), float64(g), float64(b)}
			movement += minkowskiiDist(clustersCenters[i][:], p)
			copy(clustersCenters[i][:], p)
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
	pixelColors := make([][]float64, im.Bounds().Dx()*im.Bounds().Dy())
	for i := im.Bounds().Min.X; i < im.Bounds().Max.X; i++ {
		for j := im.Bounds().Min.Y; j < im.Bounds().Max.Y; j++ {
			r, g, b, _ := im.At(i, j).RGBA()
			pixelColors[i+j*im.Bounds().Dx()] = []float64{float64(r), float64(g), float64(b)}
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
			minDist := minkowskiiDist(pixel[:], clustersCenters[0][:])
			for k := 1; k < clustersCount; k++ {
				dist := minkowskiiDist(pixel[:], clustersCenters[k][:])
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
