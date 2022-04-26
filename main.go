package main

import (
	"fmt"
	"html/template"
	"image"
	"image/color"
	"image/draw"
	_ "image/jpeg"
	"image/png"
	"io"
	"log"
	"math"
	"math/rand"
	"net/http"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"
)

// TODO: command line application
func get_next_filename() string {
	return fmt.Sprintf("%v", time.Now())
}

// TODO: rewrite paths to paths.join
// TODO: check if it is actually image, restrict size
func load_image(url string) (res image.Image, imid string, err error) {
	// TODO: cache files by url
	imid = get_next_filename()
	image_filename := fmt.Sprintf("img/%s.orig.tmp", imid)
	f, err := os.Create(image_filename)
	if err != nil {
		return
	}
	r, err := http.Get(url)
	if err != nil {
		return
	}
	data, err := io.ReadAll(r.Body)
	if err != nil {
		return
	}
	if err = r.Body.Close(); err != nil {
		return
	}
	if _, err = f.Write(data); err != nil {
		return
	}
	ff, err := os.Open(image_filename)
	if err != nil {
		return
	}
	res, format, err := image.Decode(ff)
	if err != nil {
		return
	}
	real_image_filename := fmt.Sprintf("img/%s.orig.%s", imid, format)
	if err = os.Rename(image_filename, real_image_filename); err != nil {
		return
	}
	return
}

func save_image(im image.Image, imid string) (filtered_filename string, err error) {
	filtered_filename = fmt.Sprintf("img/%s.res.png", imid)
	f, err := os.Create(filtered_filename)
	if err != nil {
		return
	}
	if err = png.Encode(f, im); err != nil {
		return
	}
	return
}

// TODO: offload work to workers
func apply_convolution(im image.Image, imid string, kernel [][]int) (filtered_filename string, err error) {
	kernelHalfWidth, kernelHalfHeight := len(kernel)/2, len(kernel)/2
	R := make([][][]int, im.Bounds().Dx())
	for i := im.Bounds().Min.X; i < im.Bounds().Max.X; i++ {
		R[i] = make([][]int, im.Bounds().Dy())
	}
	for i := im.Bounds().Min.X; i < im.Bounds().Max.X; i++ {
		for j := im.Bounds().Min.Y; j < im.Bounds().Max.Y; j++ {
			r, g, b := 0, 0, 0
			for di := 0; di < len(kernel); di++ {
				for dj := 0; dj < len(kernel[0]); dj++ {
					i1 := i + dj - kernelHalfWidth
					if i1 < 0 {
						i1 = 0
					}
					if i1 >= im.Bounds().Dx() {
						i1 = im.Bounds().Dx() - 1
					}
					j1 := j + di - kernelHalfHeight
					if j1 < 0 {
						j1 = 0
					}
					if j1 >= im.Bounds().Dy() {
						j1 = im.Bounds().Dy() - 1
					}
					dr, dg, db, _ := im.At(i1, j1).RGBA()
					r += int(dr) * kernel[di][dj]
					g += int(dg) * kernel[di][dj]
					b += int(db) * kernel[di][dj]
				}
			}
			R[i][j] = []int{r, g, b}
		}
	}
	kernelMin, kernelMax := math.MaxInt, math.MinInt
	for i := im.Bounds().Min.X; i < im.Bounds().Max.X; i++ {
		for j := im.Bounds().Min.Y; j < im.Bounds().Max.Y; j++ {
			for k := 0; k < 3; k++ {
				if R[i][j][k] < kernelMin {
					kernelMin = R[i][j][k]
				}
				if R[i][j][k] > kernelMax {
					kernelMax = R[i][j][k]
				}
			}
		}
	}
	var (
		filtered_im *image.RGBA = image.NewRGBA(im.Bounds())
		diff        int         = kernelMax - kernelMin
	)
	for i := im.Bounds().Min.X; i < im.Bounds().Max.X; i++ {
		for j := im.Bounds().Min.Y; j < im.Bounds().Max.Y; j++ {
			filtered_im.Set(i, j, color.RGBA{
				uint8((R[i][j][0] - kernelMin) * 255 / diff),
				uint8((R[i][j][1] - kernelMin) * 255 / diff),
				uint8((R[i][j][2] - kernelMin) * 255 / diff),
				255,
			})
		}
	}
	return save_image(filtered_im, imid)
}

// TODO: don't call bash?
func transfer_style(imid string, style_name string) (res string, err error) {
	res = fmt.Sprintf("img/%s.res.png", imid)
	os.Chdir("fast-style-transfer/")
	if err = exec.Command(
		"python3", "evaluate.py",
		"--in-path", fmt.Sprintf("../%s", res),
		"--out-path", "../",
		"--checkpoint", fmt.Sprintf("../ckpts/%s.ckpt", style_name),
	).Run(); err != nil {
		err = fmt.Errorf("error running python3 evaluate.py, error: %q", err)
		return
	}
	os.Chdir("..")
	if err = exec.Command("mv", fmt.Sprintf("%s.orig.png", imid), res).Run(); err != nil {
		err = fmt.Errorf("error running mv, error: %q", err)
		return
	}
	if err = exec.Command("rm", res).Run(); err != nil {
		err = fmt.Errorf("error running rm, error: %q", err)
		return
	}
	return
}

func renderTemplateOrPanic(rootTemplate *template.Template, w io.Writer, name string, data interface{}) {
	if err := rootTemplate.ExecuteTemplate(w, name, data); err != nil {
		log.Fatalf("Error rendering template: name=%q data=%v err=%q", name, data, err)
	}
}

type (
	convolutionFilter struct {
		pages_templates *template.Template
		filterName      string
		kernel          [][]int
	}
	FilterPageData struct {
		FilterName string
		Message    string
		ImageFile  *string
	}
)

func (f *convolutionFilter) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method == "POST" {
		r.ParseForm()
		if !r.PostForm.Has("url") {
			renderTemplateOrPanic(f.pages_templates, w, "filter.html", FilterPageData{
				f.filterName,
				"'url' is not provided",
				nil,
			})
			return
		}
		imageUrl := r.PostFormValue("url")
		image, imid, err := load_image(imageUrl)

		if err != nil {
			renderTemplateOrPanic(f.pages_templates, w, "filter.html", FilterPageData{
				f.filterName,
				fmt.Sprintf("Error occured:\n%q", err),
				nil,
			})
			return
		}
		image_file, err := apply_convolution(image, imid, f.kernel)
		ff := FilterPageData{
			FilterName: f.filterName,
		}
		if err != nil {
			ff.Message = fmt.Sprintf("Error occured:\n%q", err)
		} else {
			ff.Message = fmt.Sprintf("Processed image %q", imageUrl)
			ff.ImageFile = &image_file
			// TODO: add timing
		}
		renderTemplateOrPanic(f.pages_templates, w, "filter.html", ff)
	} else {
		renderTemplateOrPanic(f.pages_templates, w, "filter.html", FilterPageData{
			f.filterName,
			"",
			nil,
		})
	}
}

// TODO: change to much faster network / remove
func style_transfer_route(pages_templates *template.Template, filterName string, styleName string) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "POST" {
			r.ParseForm()
			if !r.PostForm.Has("url") {
				renderTemplateOrPanic(pages_templates, w, "filter.html", FilterPageData{
					filterName,
					"'url' is not provided",
					nil,
				})
				return
			}
			imageUrl := r.PostFormValue("url")
			_, imid, err := load_image(imageUrl)

			if err != nil {
				renderTemplateOrPanic(pages_templates, w, "filter.html", FilterPageData{
					filterName,
					fmt.Sprintf("Error occured:\n%q", err),
					nil,
				})
				return
			}
			image_file, err := transfer_style(imid, styleName)
			ff := FilterPageData{
				FilterName: filterName,
			}
			if err != nil {
				ff.Message = fmt.Sprintf("Error occured:\n%q", err)
			} else {
				ff.Message = fmt.Sprintf("Processed image %q", imageUrl)
				ff.ImageFile = &image_file
				// TODO: add timing
			}
			renderTemplateOrPanic(pages_templates, w, "filter.html", ff)
		} else {
			renderTemplateOrPanic(pages_templates, w, "filter.html", FilterPageData{
				filterName,
				"",
				nil,
			})
		}
	}
}

func is_block_black(p image.Point, blockWidth, blockHeight int, im image.Image) bool {
	brightnessSum := 0.0
	for i := 0; i < blockWidth; i++ {
		for j := 0; j < blockHeight; j++ {
			r, g, b, _ := im.At(i+p.X*blockWidth, j+p.Y*blockHeight).RGBA()
			brightnessSum += float64(r+g+b) / 3 / 0xFFFF
		}
	}
	THRESHOLD := 0.5
	return brightnessSum < THRESHOLD*float64(blockWidth*blockHeight)
}

func d2xy(n, d int) (int, int) {
	t := d
	x, y := 0, 0
	for s := 1; s < n; s *= 2 {
		rx := 1 & (t / 2)
		ry := 1 & (t ^ rx)
		if ry == 0 {
			if rx == 1 {
				x, y = n-1-x, n-1-y
			}
			x, y = y, x
		}
		x += s * rx
		y += s * ry
		t /= 4
	}
	return x, y
}

// TODO: try also https://en.wikipedia.org/wiki/Z-order_curve
func hilbert_curve_filter(im image.Image) image.Image {
	// TODO: remove / change to absolute adjustment
	// f = ImageEnhance.Brightness(res).enhance(1.3)
	// f = ImageEnhance.Contrast(f).enhance(10)
	W := 1
	for W < im.Bounds().Dx() && W < im.Bounds().Dy() {
		W *= 2
	}
	W /= 2
	blockWidth, blockHeight := im.Bounds().Dx()/W, im.Bounds().Dy()/W // TODO: check blocks more precisely???, check article
	himage := image.NewRGBA(im.Bounds())
	draw.Draw(himage, himage.Bounds(), &image.Uniform{color.RGBA{255, 255, 255, 255}}, image.Point{}, draw.Src)

	ch := make(chan image.Point)
	go func() {
		for i := 0; i < W*W; i++ {
			x, y := d2xy(W, i)
			if is_block_black(image.Point{x, y}, blockWidth, blockHeight, im) {
				ch <- image.Point{x * blockWidth, y * blockHeight} // TODO: fix
			}
		}
		close(ch)
	}()
	last := <-ch
	// TODO: speedup/fix
	for next := range ch {
		// TODO: draw (last -> next) line with black color using https://ru.wikipedia.org/wiki/%D0%90%D0%BB%D0%B3%D0%BE%D1%80%D0%B8%D1%82%D0%BC_%D0%91%D1%80%D0%B5%D0%B7%D0%B5%D0%BD%D1%85%D1%8D%D0%BC%D0%B0
		himage.Set(last.X, last.Y, color.RGBA{0, 0, 0, 255})
		// const K = 10
		// dx := (next.X - last.X) / K
		// dy := (next.Y - last.Y) / K
		// for i := 0; i < K; i++ {
		// 	himage.Set(last.X+dx*i, last.Y+dy*i, color.RGBA{0, 0, 0, 255})
		// }
		last = next
	}
	return himage
}

func hilbert_curve(im image.Image, imid string) (string, error) {
	tmp := hilbert_curve_filter(im)
	return save_image(tmp, imid)
}

func hilbert_darken(im image.Image, imid string) (string, error) {
	tmp := hilbert_curve_filter(im)
	// TODO: uncomment
	// for i := tmp.Bounds().Min.X; i < tmp.Bounds().Max.X; i++ {
	// 	for j := tmp.Bounds().Min.Y; j < tmp.Bounds().Max.y; j++ {
	// 		for k := 0; k < 3; k++ {
	// 			if im.At(i, j)[k] < tmp.At(i, j)[k] {
	// 				tmp.At(i, j)[k] = im.At(i, j)[k]
	// 			}
	// 		}
	// 	}
	// }
	return save_image(tmp, imid)
}

func shader_filter(imid, fragment_shader_source string) (string, error) {
	return "", fmt.Errorf("Not implemented")
}

// TODO: log incoming requests in web server thoroughly, log request params, log result, timing
// TODO: store result metrics in db for monitoring
func Route(w http.ResponseWriter, r *http.Request) {
	pages_templates := template.Must(template.ParseGlob("templates/*.html")) // TODO: parse once
	log.Printf("method=%s url=%q", r.Method, r.URL)

	mux := http.NewServeMux() // TODO: create only once

	// TODO: move away to nginx
	mux.HandleFunc("/img/", func(w http.ResponseWriter, r *http.Request) {
		img_path := r.URL.Path[1:]
		file, err := os.Open(img_path)
		if err != nil {
			if os.IsNotExist(err) {
				w.WriteHeader(http.StatusNotFound)
				// TODO: write 404 extract to function
				renderTemplateOrPanic(pages_templates, w, "404.html", nil)
			} else {
				log.Printf("Error opening file %q %v", img_path, err)
			}
			return
		}
		img_data, err := io.ReadAll(file)
		if err != nil {
			log.Printf("2 %v", err)
			return
		}
		w.Write(img_data)
	})

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" {
			w.WriteHeader(http.StatusNotFound)
			renderTemplateOrPanic(pages_templates, w, "404.html", nil)
			return
		}
		renderTemplateOrPanic(pages_templates, w, "index.html", nil)
	})

	mux.HandleFunc("/lasts", func(w http.ResponseWriter, r *http.Request) {
		saved_images, err := os.ReadDir("img")
		if err != nil {
			log.Printf("Error reading images: %v", err)
			return
		}
		names := make([]string, 0)
		for _, x := range saved_images {
			filename := x.Name()
			if strings.HasSuffix(filename, "res.png") {
				names = append(names, filename[:len(filename)-8])
			}
		}
		// TODO: sort
		renderTemplateOrPanic(pages_templates, w, "lasts.html", names)
	})

	mux.Handle("/blur", &convolutionFilter{
		pages_templates,
		"Blur",
		[][]int{
			{1, 1, 1},
			{1, 1, 1},
			{1, 1, 1},
		},
	})

	mux.Handle("/weakblur", &convolutionFilter{
		pages_templates,
		"Weak blur",
		[][]int{
			{0, 1, 0},
			{1, 1, 1},
			{0, 1, 0},
		},
	})

	mux.Handle("/emboss", &convolutionFilter{
		pages_templates,
		"Emboss",
		[][]int{
			{-2, -1, 0},
			{-1, 1, 1},
			{0, 1, 2},
		},
	})

	mux.Handle("/sharpen", &convolutionFilter{
		pages_templates,
		"Sharpen",
		[][]int{
			{0, -1, 0},
			{-1, 5, -1},
			{0, -1, 0},
		},
	})

	mux.Handle("/edgeenhance", &convolutionFilter{
		pages_templates,
		"Edge enhance",
		[][]int{
			{0, 0, 0},
			{-1, 1, 0},
			{0, 0, 0},
		},
	})

	mux.Handle("/edgedetect1", &convolutionFilter{
		pages_templates,
		"Edge detect 1",
		[][]int{
			{1, 0, -1},
			{0, 0, 0},
			{-1, 0, 1},
		},
	})

	mux.Handle("/edgedetect2", &convolutionFilter{
		pages_templates,
		"Edge detect 2",
		[][]int{
			{0, -1, 0},
			{-1, 4, -1},
			{0, -1, 0},
		},
	})

	mux.Handle("/horizontallines", &convolutionFilter{
		pages_templates,
		"Horizontal lines",
		[][]int{
			{-1, -1, -1},
			{2, 2, 2},
			{-1, -1, -1},
		},
	})

	mux.Handle("/verticallines", &convolutionFilter{
		pages_templates,
		"Vertical lines",
		[][]int{
			{-1, 2, -1},
			{-1, 2, -1},
			{-1, 2, -1},
		},
	})

	// TODO: draw lokot'
	// TODO: fix overflows
	// TODO: fix double POST???
	mux.HandleFunc("/cluster", func(w http.ResponseWriter, r *http.Request) {
		filterName := "Cluster"
		if r.Method == "POST" {
			r.ParseForm()
			switch {
			case !r.PostForm.Has("url"):
				renderTemplateOrPanic(pages_templates, w, "cluster.html", FilterPageData{
					filterName,
					"'url' is not provided",
					nil,
				})
				return
			case !r.PostForm.Has("n"):
				renderTemplateOrPanic(pages_templates, w, "cluster.html", FilterPageData{
					filterName,
					"'n' (number of clusters) is not provided",
					nil,
				})
				return
			}
			n_clusters, err := strconv.Atoi(r.PostFormValue("n"))
			switch {
			case err != nil:
				renderTemplateOrPanic(pages_templates, w, "cluster.html", FilterPageData{
					filterName,
					fmt.Sprintf("Error in parameter 'n':\n%q", err),
					nil,
				})
				return
			case n_clusters < 2:
				renderTemplateOrPanic(pages_templates, w, "cluster.html", FilterPageData{
					filterName,
					fmt.Sprintf("'n' must be at least 2, you gave n=%d", n_clusters),
					nil,
				})
				return
			}
			imageUrl := r.PostFormValue("url")
			im, imid, err := load_image(imageUrl)
			if err != nil {
				renderTemplateOrPanic(pages_templates, w, "cluster.html", FilterPageData{
					filterName,
					fmt.Sprintf("Error occured:\n%q", err),
					nil,
				})
				return
			}
			X := make([][][]uint32, im.Bounds().Dx())
			for i := im.Bounds().Min.X; i < im.Bounds().Max.X; i++ {
				X[i] = make([][]uint32, im.Bounds().Dy())
				for j := im.Bounds().Min.Y; j < im.Bounds().Max.Y; j++ {
					r, g, b, _ := im.At(i, j).RGBA()
					X[i][j] = []uint32{r, g, b}
				}
			}
			kmeans := make([][]uint32, n_clusters)
			sumAndCount := make([][]uint32, n_clusters) // sum of Rs, Gs, Bs and count
			rand.Seed(0)
			for i := 0; i < n_clusters; i++ {
				kmeans[i] = []uint32{
					rand.Uint32() / 0x100,
					rand.Uint32() / 0x100,
					rand.Uint32() / 0x100,
				}
				sumAndCount[i] = make([]uint32, 4)
			}
			// TODO: optimize
			for epoch := 0; epoch < 100; epoch++ { // TODO: or diff is small enough
				for i := 0; i < n_clusters; i++ {
					sumAndCount[i][0], sumAndCount[i][1], sumAndCount[i][2], sumAndCount[i][3] = 0, 0, 0, 0
				}
				for i := im.Bounds().Min.X; i < im.Bounds().Max.X; i++ {
					for j := im.Bounds().Min.Y; j < im.Bounds().Max.Y; j++ {
						r, g, b, _ := im.At(i, j).RGBA()
						minCluster := 0
						minDist := (r-kmeans[0][0])*(r-kmeans[0][0]) + (g-kmeans[0][1])*(g-kmeans[0][1]) + (b-kmeans[0][2])*(b-kmeans[0][2])
						for k := 1; k < n_clusters; k++ {
							dist := (r-kmeans[k][0])*(r-kmeans[k][0]) + (g-kmeans[k][1])*(g-kmeans[k][1]) + (b-kmeans[k][2])*(b-kmeans[k][2])
							if dist < minDist {
								minCluster = k
								minDist = dist
							}
						}
						sumAndCount[minCluster][0] += r
						sumAndCount[minCluster][1] += g
						sumAndCount[minCluster][2] += b
						sumAndCount[minCluster][3]++
					}
				}
				for i := 0; i < n_clusters; i++ {
					count := sumAndCount[i][3]
					if count == 0 {
						continue
					}
					kmeans[i][0], kmeans[i][1], kmeans[i][2] = sumAndCount[i][0]/count, sumAndCount[i][1]/count, sumAndCount[i][2]/count
				}
			}
			filtered_im := image.NewRGBA(im.Bounds())
			for i := im.Bounds().Min.X; i < im.Bounds().Max.X; i++ {
				for j := im.Bounds().Min.Y; j < im.Bounds().Max.Y; j++ {
					r, g, b, _ := im.At(i, j).RGBA()
					minCluster := 0
					minDist := (r-kmeans[0][0])*(r-kmeans[0][0]) + (g-kmeans[0][1])*(g-kmeans[0][1]) + (b-kmeans[0][2])*(b-kmeans[0][2])
					for k := 1; k < n_clusters; k++ {
						dist := (r-kmeans[k][0])*(r-kmeans[k][0]) + (g-kmeans[k][1])*(g-kmeans[k][1]) + (b-kmeans[k][2])*(b-kmeans[k][2])
						if dist < minDist {
							minCluster = k
							minDist = dist
						}
					}
					filtered_im.Set(i, j, color.RGBA{
						uint8(kmeans[minCluster][0] / 0x100),
						uint8(kmeans[minCluster][1] / 0x100),
						uint8(kmeans[minCluster][2] / 0x100),
						255,
					})
				}
			}
			filtered_filename := fmt.Sprintf("img/%s.res.png", imid)
			f, err := os.Create(filtered_filename)
			if err != nil {
				renderTemplateOrPanic(pages_templates, w, "cluster.html", FilterPageData{
					filterName,
					fmt.Sprintf("Error occured:\n%q", err),
					nil,
				})
				return
			}
			err = png.Encode(f, filtered_im)
			ff := FilterPageData{
				FilterName: filterName,
			}
			if err != nil {
				ff.Message = fmt.Sprintf("Error occured:\n%q", err)
			} else {
				ff.Message = fmt.Sprintf("Processed image %q", imageUrl)
				ff.ImageFile = &filtered_filename
				// TODO: add timing
			}
			renderTemplateOrPanic(pages_templates, w, "cluster.html", ff)
		} else {
			renderTemplateOrPanic(pages_templates, w, "cluster.html", FilterPageData{
				filterName,
				"",
				nil,
			})
		}
	})

	mux.HandleFunc("/lamuse", style_transfer_route(pages_templates, "La muse styling", "la_muse"))
	mux.HandleFunc("/scream", style_transfer_route(pages_templates, "Scream styling", "scream"))
	mux.HandleFunc("/wave", style_transfer_route(pages_templates, "Wave styling", "wave"))
	mux.HandleFunc("/wreck", style_transfer_route(pages_templates, "Wreck styling", "wreck"))
	mux.HandleFunc("/udnie", style_transfer_route(pages_templates, "Udnie styling", "udnie"))
	mux.HandleFunc("/rain_princess", style_transfer_route(pages_templates, "Rain princess styling", "rain_princess"))

	mux.HandleFunc("/hilbert", func(w http.ResponseWriter, r *http.Request) {
		filterName := "Hilbert curve"
		if r.Method == "POST" {
			r.ParseForm()
			if !r.PostForm.Has("url") {
				renderTemplateOrPanic(pages_templates, w, "filter.html", FilterPageData{
					filterName,
					"'url' is not provided",
					nil,
				})
				return
			}
			imageUrl := r.PostFormValue("url")
			im, imid, err := load_image(imageUrl)

			if err != nil {
				renderTemplateOrPanic(pages_templates, w, "filter.html", FilterPageData{
					filterName,
					fmt.Sprintf("Error occured:\n%q", err),
					nil,
				})
				return
			}
			image_file, err := hilbert_curve(im, imid)
			ff := FilterPageData{
				FilterName: filterName,
			}
			if err != nil {
				ff.Message = fmt.Sprintf("Error occured:\n%q", err)
			} else {
				ff.Message = fmt.Sprintf("Processed image %q", imageUrl)
				ff.ImageFile = &image_file
				// TODO: add timing
			}
			renderTemplateOrPanic(pages_templates, w, "filter.html", ff)
		} else {
			renderTemplateOrPanic(pages_templates, w, "filter.html", FilterPageData{
				filterName,
				"",
				nil,
			})
		}
	})

	mux.HandleFunc("/hilbertdarken", func(w http.ResponseWriter, r *http.Request) {
		filterName := "Hilbert curve darken"
		if r.Method == "POST" {
			r.ParseForm()
			if !r.PostForm.Has("url") {
				renderTemplateOrPanic(pages_templates, w, "filter.html", FilterPageData{
					filterName,
					"'url' is not provided",
					nil,
				})
				return
			}
			imageUrl := r.PostFormValue("url")
			im, imid, err := load_image(imageUrl)

			if err != nil {
				renderTemplateOrPanic(pages_templates, w, "filter.html", FilterPageData{
					filterName,
					fmt.Sprintf("Error occured:\n%q", err),
					nil,
				})
				return
			}
			image_file, err := hilbert_darken(im, imid)
			ff := FilterPageData{
				FilterName: filterName,
			}
			if err != nil {
				ff.Message = fmt.Sprintf("Error occured:\n%q", err)
			} else {
				ff.Message = fmt.Sprintf("Processed image %q", imageUrl)
				ff.ImageFile = &image_file
				// TODO: add timing
			}
			renderTemplateOrPanic(pages_templates, w, "filter.html", ff)
		} else {
			renderTemplateOrPanic(pages_templates, w, "filter.html", FilterPageData{
				filterName,
				"",
				nil,
			})
		}
	})

	mux.HandleFunc("/shader", func(w http.ResponseWriter, r *http.Request) {
		filterName := "Shader"
		if r.Method == "POST" {
			r.ParseForm()
			if !r.PostForm.Has("url") {
				renderTemplateOrPanic(pages_templates, w, "shader.html", FilterPageData{
					filterName,
					"'url' is not provided",
					nil,
				})
				return
			}
			if !r.PostForm.Has("url") {
				renderTemplateOrPanic(pages_templates, w, "shader.html", FilterPageData{
					filterName,
					"'fragment_shader_source' is not provided",
					nil,
				})
				return
			}
			imageUrl := r.PostFormValue("url")
			fragment_shader_source := r.PostFormValue("fragment_shader_source")
			_, imid, err := load_image(imageUrl)

			if err != nil {
				renderTemplateOrPanic(pages_templates, w, "shader.html", FilterPageData{
					filterName,
					fmt.Sprintf("Error occured:\n%q", err),
					nil,
				})
				return
			}
			image_file, err := shader_filter(imid, fragment_shader_source)
			ff := FilterPageData{
				FilterName: filterName,
			}
			if err != nil {
				ff.Message = fmt.Sprintf("Error occured:\n%q", err)
			} else {
				ff.Message = fmt.Sprintf("Processed image %q", imageUrl)
				ff.ImageFile = &image_file
				// TODO: add timing
			}
			renderTemplateOrPanic(pages_templates, w, "shader.html", ff)
		} else {
			renderTemplateOrPanic(pages_templates, w, "shader.html", FilterPageData{
				filterName,
				"",
				nil,
			})
		}
	})

	mux.ServeHTTP(w, r)
}

func main() {
	s := &http.Server{
		Addr:           ":8080",
		Handler:        http.HandlerFunc(Route),
		ReadTimeout:    10 * time.Second,
		WriteTimeout:   10 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}
	log.Fatal(s.ListenAndServe())
}
