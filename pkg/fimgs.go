package fimgs

import (
	"bufio"
	"compress/zlib"
	"encoding/binary"
	"hash/crc32"
	"image"
	_ "image/jpeg"
	_ "image/png"
	"io"
	"os"
)

const (
	ctTrueColor = 2
	cbTC8       = 6
)

func paeth(a, b, c uint8) uint8 {
	pc := int(c)
	pa := int(b) - pc
	pb := int(a) - pc
	pc = abs(pa + pb)
	pa = abs(pa)
	pb = abs(pb)
	if pa <= pb && pa <= pc {
		return a
	} else if pb <= pc {
		return b
	} else {
		return c
	}
}

const (
	ftNone    = 0
	ftSub     = 1
	ftUp      = 2
	ftAverage = 3
	ftPaeth   = 4
	nFilter   = 5
)

const pngHeader = "\x89PNG\r\n\x1a\n"

type encoder struct {
	w io.Writer
}

func abs8(d uint8) int {
	di := int8(d)
	return int(di ^ (di >> 7))
}

func writeChunk(w io.Writer, b []byte, name []byte) {
	n := uint32(len(b))
	header := [8]byte{}
	binary.BigEndian.PutUint32(header[:4], n)
	copy(header[4:8], name[:])
	crc := crc32.NewIEEE()
	crc.Write(header[4:8])
	crc.Write(b)
	footer := [4]byte{}
	binary.BigEndian.PutUint32(footer[:4], crc.Sum32())
	w.Write(header[:8])
	w.Write(b)
	w.Write(footer[:4])
}

func (e *encoder) Write(b []byte) (int, error) {
	writeChunk(e.w, b, []byte("IDAT"))
	return len(b), nil
}

func filter(cr *[nFilter][]byte, pr []byte) int {
	const bytesPerPixel = 3
	cdat0 := cr[0][1:]
	pdat := pr[1:]
	n := len(cdat0)
	ch := make(chan [2]int, 5)
	go func() {
		cdat2 := cr[2][1:]
		sum := 0
		for i := 0; i < n; i++ {
			cdat2[i] = cdat0[i] - pdat[i]
			sum += abs8(cdat2[i])
		}
		ch <- [2]int{sum, ftUp}
	}()
	go func() {
		cdat4 := cr[4][1:]
		sum := 0
		for i := 0; i < bytesPerPixel; i++ {
			cdat4[i] = cdat0[i] - pdat[i]
			sum += abs8(cdat4[i])
		}
		for i := bytesPerPixel; i < n; i++ {
			cdat4[i] = cdat0[i] - paeth(cdat0[i-bytesPerPixel], pdat[i], pdat[i-bytesPerPixel])
			sum += abs8(cdat4[i])
		}
		ch <- [2]int{sum, ftPaeth}
	}()
	go func() {
		sum := 0
		for i := 0; i < n; i++ {
			sum += abs8(cdat0[i])
		}
		ch <- [2]int{sum, ftNone}
	}()
	go func() {
		cdat1 := cr[1][1:]
		sum := 0
		for i := 0; i < bytesPerPixel; i++ {
			cdat1[i] = cdat0[i]
			sum += abs8(cdat1[i])
		}
		for i := bytesPerPixel; i < n; i++ {
			cdat1[i] = cdat0[i] - cdat0[i-bytesPerPixel]
			sum += abs8(cdat1[i])
		}
		ch <- [2]int{sum, ftSub}
	}()
	go func() {
		cdat3 := cr[3][1:]
		sum := 0
		for i := 0; i < bytesPerPixel; i++ {
			cdat3[i] = cdat0[i] - pdat[i]/2
			sum += abs8(cdat3[i])
		}
		for i := bytesPerPixel; i < n; i++ {
			cdat3[i] = cdat0[i] - uint8((int(cdat0[i-bytesPerPixel])+int(pdat[i]))/2)
			sum += abs8(cdat3[i])
		}
		ch <- [2]int{sum, ftAverage}
	}()
	bestAndFilter := <-ch
	for i := 0; i < 4; i++ {
		tmp := <-ch
		if tmp[0] < bestAndFilter[0] {
			bestAndFilter = tmp
		}
	}
	return bestAndFilter[1]
}

func (e *encoder) writeImage(m image.RGBA) {
	bw := bufio.NewWriterSize(e, 1<<15)
	defer bw.Flush()
	zw, _ := zlib.NewWriterLevel(bw, zlib.BestSpeed)
	defer zw.Close()
	const bitsPerPixel = 24
	b := m.Bounds()
	sz := 1 + (bitsPerPixel*b.Dx()+7)/8
	cr := [nFilter][]uint8{}
	for i := range cr {
		cr[i] = make([]uint8, sz)
		cr[i][0] = uint8(i)
	}
	pr := make([]uint8, sz)
	for y := b.Min.Y; y < b.Max.Y; y++ {
		i := 1
		cr0 := cr[0]
		j0 := (y - b.Min.Y) * m.Stride
		j1 := j0 + b.Dx()*4
		for j := j0; j < j1; j += 4 {
			copy(cr0[i:i+3], m.Pix[j:j+3])
			i += 3
		}
		f := filter(&cr, pr)
		zw.Write(cr[f])
		pr, cr[0] = cr[0], pr
	}
}

func Encode(w io.Writer, m image.RGBA) {
	io.WriteString(w, pngHeader)
	b := m.Bounds()
	tmp := [4 * 256]byte{}
	binary.BigEndian.PutUint32(tmp[0:4], uint32(b.Dx()))
	binary.BigEndian.PutUint32(tmp[4:8], uint32(b.Dy()))
	tmp[8] = 8
	tmp[9] = ctTrueColor
	tmp[10] = 0 // default compression method
	tmp[11] = 0 // default filter method
	tmp[12] = 0 // non-interlaced
	writeChunk(w, tmp[:13], []byte("IHDR"))
	(&encoder{w: w}).writeImage(m)
	writeChunk(w, nil, []byte("IEND"))
}

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
	Encode(imageFile, im)
	return
}
