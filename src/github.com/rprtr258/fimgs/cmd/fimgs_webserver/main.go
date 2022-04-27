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
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"

    "github.com/rprtr258/fimgs"
)

// TODO: command line application
func get_next_filename() string {
	return fmt.Sprintf("%v", time.Now())
}

// TODO: rewrite paths to paths.join
// TODO: check if it is actually image, restrict size
func load_image(url string) (res string, imid string, err error) {
	// TODO: cache files by url
	imid = get_next_filename()
	r, err := http.Get(url)
	if err != nil {
		return
	}
	// TODO: check for content type to deduce image format
	var format string
	switch contentType := r.Header.Get("Content-Type"); contentType {
	case "image/jpeg":
		format = "jpeg"
	case "image/png":
		format = "png"
	default:
		err = fmt.Errorf("Image format %q is not supported", contentType)
		return
	}
	res = fmt.Sprintf("img/%s.orig.%s", imid, format)
	f, err := os.Create(res)
	defer func() {
		f.Close()
	}()
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
	return
}

// TODO: offload work to workers

type FilterPageData struct {
	FilterName string
	Message    string
	ImageFile  *string
}

func renderTemplateOrPanic(rootTemplate *template.Template, w io.Writer, name string, data interface{}) {
	if err := rootTemplate.ExecuteTemplate(w, name, data); err != nil {
		log.Fatalf("Error rendering template: name=%q data=%v err=%q", name, data, err)
	}
}

func renderFilterPage(pages_templates *template.Template, w http.ResponseWriter, templateName, filterName, message string) {
	renderTemplateOrPanic(pages_templates, w, templateName, FilterPageData{
		filterName,
		message,
		nil,
	})
}

type Filter interface {
	filterName() string
	templateName() string
	pages_templates() *template.Template
	validate(url.Values) error
	process(imageFilename string, imageId string, form url.Values) (string, error)
}

func filterToHandler(f Filter) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		filterName := f.filterName()
		if r.Method == "POST" {
			r.ParseForm()
			if !r.PostForm.Has("url") {
				renderFilterPage(f.pages_templates(), w, f.templateName(), filterName, "'url' is not provided")
				return
			}
			imageUrl := r.PostFormValue("url")
			if imageUrl == "" {
				renderFilterPage(f.pages_templates(), w, f.templateName(), filterName, "'url' is not provided")
				return
			}
			if err := f.validate(r.PostForm); err != nil {
				renderFilterPage(f.pages_templates(), w, f.templateName(), filterName, fmt.Sprintf("Error in request params:\n%q", err))
				return
			}
			sourceImageFilename, imid, err := load_image(imageUrl)
			if err != nil {
				renderFilterPage(f.pages_templates(), w, f.templateName(), filterName, fmt.Sprintf("Error occured during loading image:\n%q", err))
				return
			}
			image_file, err := f.process(sourceImageFilename, imid, r.PostForm)
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
			renderTemplateOrPanic(f.pages_templates(), w, f.templateName(), ff)
		} else {
			renderFilterPage(f.pages_templates(), w, f.templateName(), filterName, "")
		}
	}
}

type BasicFilter struct {
	_filterName      string
	_templateName    string
	_pages_templates *template.Template
}

func (f *BasicFilter) filterName() string {
	return f._filterName
}

func (f *BasicFilter) templateName() string {
	return f._templateName
}

func (f *BasicFilter) pages_templates() *template.Template {
	return f._pages_templates
}

func (f *BasicFilter) validate(url.Values) error {
	return nil
}

type convolutionFilter struct {
	BasicFilter
	kernel [][]int
}

func (f *convolutionFilter) process(imageFilename string, imageId string, _ url.Values) (string, error) {
	im, err := loadImageFile(imageFilename)
	if err != nil {
		return "", fmt.Errorf("Error occured during loading image:\n%q", err)
	}
	return apply_convolution(im, imageId, f.kernel)
}

type StyleTransferFilter struct {
	BasicFilter
	styleName string
}

// TODO: change to much faster network / remove
func (f *StyleTransferFilter) process(imageFilename string, imageId string, _ url.Values) (string, error) {
	return transfer_style(imageId, f.styleName)
}

type KMeansFilter struct {
	BasicFilter
}

func (f KMeansFilter) validate(form url.Values) error {
	if !form.Has("n") {
		return fmt.Errorf("'n' (number of clusters) is not provided")
	}
	n_clusters, err := strconv.Atoi(form.Get("n"))
	switch {
	case err != nil:
		return fmt.Errorf("Error parsing parameter 'n':\n%q", err)
	case n_clusters < 2:
		return fmt.Errorf("'n' must be at least 2, you gave n=%d", n_clusters)
	}
	return nil
}

func (f KMeansFilter) process(imageFilename string, imageId string, form url.Values) (filtered_filename string, err error) {
	n_clusters, _ := strconv.Atoi(form.Get("n"))
}

type HilbertFilter struct {
	BasicFilter
}

func (f *HilbertFilter) process(imageFilename string, imageId string, form url.Values) (string, error) {
	im, err := loadImageFile(imageFilename)
	if err != nil {
		return "", err
	}
	return hilbert_curve(im, imageId)
}

type HilbertDarkenFilter struct {
	BasicFilter
}

func (f *HilbertDarkenFilter) process(imageFilename string, imageId string, form url.Values) (string, error) {
	im, err := loadImageFile(imageFilename)
	if err != nil {
		return "", err
	}
	return hilbert_darken(im, imageId)
}

type ShaderFilter struct {
	BasicFilter
}

func (f *ShaderFilter) validate(form url.Values) error {
	if !form.Has("fragment_shader_source") {
		return fmt.Errorf("'fragment_shader_source' is not provided")
	}
	//fragment_shader_source := r.PostFormValue("fragment_shader_source")
	// TODO: compile shader and return any errors
	return nil
}

func (f *ShaderFilter) process(imageFilename string, imageId string, form url.Values) (string, error) {
	fragment_shader_source := form.Get("fragment_shader_source")
	return shader_filter(imageId, fragment_shader_source)
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
		sourceImages := make(map[string]template.URL)
		resultImages := make(map[string]template.URL)
		for _, x := range saved_images {
			filename := x.Name()
			dotBeforeExtension := strings.LastIndex(filename, ".")
			if dotBeforeExtension == -1 {
				continue
			}
			dotBeforeOrigOrRes := strings.LastIndex(filename[:dotBeforeExtension], ".")
			if dotBeforeOrigOrRes == -1 {
				continue
			}
			imageId := filename[:dotBeforeOrigOrRes]
			origOrRes := filename[dotBeforeOrigOrRes+1 : dotBeforeExtension]
			fullFilepath := filepath.Join("img", filename)
			switch origOrRes {
			case "orig":
				sourceImages[imageId] = template.URL(fullFilepath)
			case "res":
				resultImages[imageId] = template.URL(fullFilepath)
			}
		}
		// TODO: sort
		renderTemplateOrPanic(pages_templates, w, "lasts.html", struct {
			SourceImages map[string]template.URL
			ResultImages map[string]template.URL
		}{sourceImages, resultImages})
	})

	mux.HandleFunc("/blur", filterToHandler(&convolutionFilter{
		BasicFilter{"Blur", "filter.html", pages_templates},
		[][]int{
			{1, 1, 1},
			{1, 1, 1},
			{1, 1, 1},
		},
	}))

	mux.HandleFunc("/weakblur", filterToHandler(&convolutionFilter{
		BasicFilter{"Weak blur", "filter.html", pages_templates},
		[][]int{
			{0, 1, 0},
			{1, 1, 1},
			{0, 1, 0},
		},
	}))

	mux.HandleFunc("/emboss", filterToHandler(&convolutionFilter{
		BasicFilter{"Emboss", "filter.html", pages_templates},
		[][]int{
			{-2, -1, 0},
			{-1, 1, 1},
			{0, 1, 2},
		},
	}))

	mux.HandleFunc("/sharpen", filterToHandler(&convolutionFilter{
		BasicFilter{"Sharpen", "filter.html", pages_templates},
		[][]int{
			{0, -1, 0},
			{-1, 5, -1},
			{0, -1, 0},
		},
	}))

	mux.HandleFunc("/edgeenhance", filterToHandler(&convolutionFilter{
		BasicFilter{"Edge enhance", "filter.html", pages_templates},
		[][]int{
			{0, 0, 0},
			{-1, 1, 0},
			{0, 0, 0},
		},
	}))

	mux.HandleFunc("/edgedetect1", filterToHandler(&convolutionFilter{
		BasicFilter{"Edge detect 1", "filter.html", pages_templates},
		[][]int{
			{1, 0, -1},
			{0, 0, 0},
			{-1, 0, 1},
		},
	}))

	mux.HandleFunc("/edgedetect2", filterToHandler(&convolutionFilter{
		BasicFilter{"Edge detect 2", "filter.html", pages_templates},
		[][]int{
			{0, -1, 0},
			{-1, 4, -1},
			{0, -1, 0},
		},
	}))

	mux.HandleFunc("/horizontallines", filterToHandler(&convolutionFilter{
		BasicFilter{"Horizontal lines", "filter.html", pages_templates},
		[][]int{
			{-1, -1, -1},
			{2, 2, 2},
			{-1, -1, -1},
		},
	}))

	mux.HandleFunc("/verticallines", filterToHandler(&convolutionFilter{
		BasicFilter{"Vertical lines", "filter.html", pages_templates},
		[][]int{
			{-1, 2, -1},
			{-1, 2, -1},
			{-1, 2, -1},
		},
	}))

	// TODO: draw lokot'
	// TODO: fix double POST???
	mux.HandleFunc("/cluster", filterToHandler(&KMeansFilter{BasicFilter{"Cluster", "cluster.html", pages_templates}}))

	mux.HandleFunc("/lamuse", filterToHandler(&StyleTransferFilter{BasicFilter{"La muse styling", "filter.html", pages_templates}, "la_muse"}))
	mux.HandleFunc("/scream", filterToHandler(&StyleTransferFilter{BasicFilter{"Scream styling", "filter.html", pages_templates}, "scream"}))
	mux.HandleFunc("/wave", filterToHandler(&StyleTransferFilter{BasicFilter{"Wave styling", "filter.html", pages_templates}, "wave"}))
	mux.HandleFunc("/wreck", filterToHandler(&StyleTransferFilter{BasicFilter{"Wreck styling", "filter.html", pages_templates}, "wreck"}))
	mux.HandleFunc("/udnie", filterToHandler(&StyleTransferFilter{BasicFilter{"Udnie styling", "filter.html", pages_templates}, "udnie"}))
	mux.HandleFunc("/rain_princess", filterToHandler(&StyleTransferFilter{BasicFilter{"Rain princess styling", "filter.html", pages_templates}, "rain_princess"}))

	mux.HandleFunc("/hilbert", filterToHandler(&HilbertFilter{BasicFilter{"Hilbert curve", "filter.html", pages_templates}}))

	mux.HandleFunc("/hilbertdarken", filterToHandler(&HilbertFilter{BasicFilter{"Hilbert curve darken", "filter.html", pages_templates}}))

	mux.HandleFunc("/shader", filterToHandler(&ShaderFilter{BasicFilter{"Shader", "shader.html", pages_templates}}))

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
