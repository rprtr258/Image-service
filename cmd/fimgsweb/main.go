package main

import (
	"fmt"
	"html/template"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	fimgs "github.com/rprtr258/fimgs/pkg"
)

func generateNewImageId() string {
	return time.Now().Format("2006-01-02-03-04-05")
}

// TODO: check if it is actually image, restrict size
func downloadImage(url string) (_imageFilename string, _imageId string, _err error) {
	// TODO: cache files by url
	imageId := generateNewImageId()
	resp, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return
	}

	r, err := http.DefaultClient.Do(resp)
	if err != nil {
		return
	}
	defer r.Body.Close()

	var format string
	switch contentType := r.Header.Get("Content-Type"); contentType {
	case "image/jpeg":
		format = "jpeg"
	case "image/png":
		format = "png"
	default:
		return "", "", fmt.Errorf("image format %q is not supported", contentType)
	}

	imageFilename := filepath.Join("img", fmt.Sprintf("%s.orig.%s", imageId, format))
	f, err := os.Create(imageFilename)
	if err != nil {
		return "", "", err
	}
	defer f.Close()

	if _, err := io.Copy(f, r.Body); err != nil {
		return "", "", err
	}

	return imageFilename, imageId, nil
}

// TODO: offload work to workers

var rootTemplate = template.Must(template.ParseGlob("templates/*.html")) // TODO: parse once // TODO: embed

func renderTemplateOrPanic(w io.Writer, name string, data interface{}) {
	if err := rootTemplate.ExecuteTemplate(w, name, data); err != nil {
		// TODO: return and handle error
		log.Fatalf("Error rendering template: name=%q data=%v err=%q", name, data, err)
	}
}

type FilterPageData struct {
	FilterName string
	Message    string
	ImageFile  *string
}

func renderFilterPage(w http.ResponseWriter, templateName, filterName, message string) {
	renderTemplateOrPanic(w, templateName, FilterPageData{
		filterName,
		message,
		nil,
	})
}

func noValidate(form url.Values) (struct{}, error) { return struct{}{}, nil }

type Filter[P any] struct {
	filterName   string
	templateName string

	validate func(url.Values) (P, error)
	process  func(sourceImageFilename, resultImageFilename string, params P) error
}

func filterHandler[P any](f Filter[P]) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			renderFilterPage(w, f.templateName, f.filterName, "")
			return
		}

		r.ParseForm()
		if !r.PostForm.Has("url") {
			renderFilterPage(w, f.templateName, f.filterName, "'url' is not provided")
			return
		}
		imageUrl := r.PostFormValue("url")
		if imageUrl == "" {
			renderFilterPage(w, f.templateName, f.filterName, "'url' is not provided")
			return
		}

		params, err := f.validate(r.PostForm)
		if err != nil {
			renderFilterPage(w, f.templateName, f.filterName, fmt.Sprintf("Error in request params:\n%q", err))
			return
		}

		sourceImageFilename, imageId, err := downloadImage(imageUrl)
		if err != nil {
			renderFilterPage(w, f.templateName, f.filterName, fmt.Sprintf("Error occured during loading image:\n%q", err))
			return
		}

		resultImageFile := filepath.Join("img", fmt.Sprintf("img/%s.res.png", imageId))

		ff := FilterPageData{
			FilterName: f.filterName,
		}
		if err := f.process(sourceImageFilename, resultImageFile, params); err != nil {
			ff.Message = fmt.Sprintf("Error occured:\n%q", err)
		} else {
			ff.Message = fmt.Sprintf("Processed image %q", imageUrl)
			ff.ImageFile = &resultImageFile
			// TODO: add timing
		}
		renderTemplateOrPanic(w, f.templateName, ff)
	}
}

// TODO: load assets https://github.com/go-gl/example/blob/d71b0d9f823d97c3b5ac2a79fdcdb56ca1677eba/gl41core-cube/cube.go#L322
// or include at compile time
func main() {
	mux := http.NewServeMux()
	// TODO: move away to nginx
	mux.HandleFunc("/img/", func(w http.ResponseWriter, r *http.Request) {
		img_path := r.URL.Path[1:]
		file, err := os.Open(img_path)
		if err != nil {
			if os.IsNotExist(err) {
				w.WriteHeader(http.StatusNotFound)
				// TODO: write 404 extract to function
				renderTemplateOrPanic(w, "404.html", nil)
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
			renderTemplateOrPanic(w, "404.html", nil)
			return
		}
		renderTemplateOrPanic(w, "index.html", nil)
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
			fullFilepathURL := template.URL(filepath.Join("img", filename))
			switch filename[dotBeforeOrigOrRes+1 : dotBeforeExtension] {
			case "orig":
				sourceImages[imageId] = fullFilepathURL
			case "res":
				resultImages[imageId] = fullFilepathURL
			}
		}
		// TODO: sort
		renderTemplateOrPanic(w, "lasts.html", struct {
			SourceImages map[string]template.URL
			ResultImages map[string]template.URL
		}{sourceImages, resultImages})
	})
	for route, hndlr := range map[string]struct {
		name   string
		kernel [][]int
	}{
		"blur":            {"Blur", fimgs.BLUR_KERNEL},
		"weakblur":        {"Weak blur", fimgs.WEAK_BLUR_KERNEL},
		"emboss":          {"Emboss", fimgs.EMBOSS_KERNEL},
		"sharpen":         {"Sharpen", fimgs.SHARPEN_KERNEL},
		"edgeenhance":     {"Edge enhance", fimgs.EDGE_ENHANCE_KERNEL},
		"edgedetect1":     {"Edge detect 1", fimgs.EDGE_DETECT1_KERNEL},
		"edgedetect2":     {"Edge detect 2", fimgs.EDGE_DETECT2_KERNEL},
		"horizontallines": {"Horizontal lines", fimgs.HORIZONTAL_LINES_KERNEL},
		"verticallines":   {"Vertical lines", fimgs.VERTICAL_LINES_KERNEL},
	} {
		mux.HandleFunc("/"+route, filterHandler(Filter[struct{}]{
			hndlr.name, "filter.html", noValidate,
			func(sourceImageFilename, resultImageFilename string, _ struct{}) error {
				return fimgs.ApplyConvolutionFilter(sourceImageFilename, resultImageFilename, hndlr.kernel)
			},
		}))
	}
	// TODO: draw lokot'
	// TODO: fix double POST???
	mux.HandleFunc("/cluster", filterHandler(Filter[int]{
		"Cluster", "cluster.html",
		func(form url.Values) (int, error) {
			if !form.Has("n") {
				return 0, fmt.Errorf("'n' (number of clusters) is not provided")
			}
			n_clusters, err := strconv.Atoi(form.Get("n"))
			switch {
			case err != nil:
				return 0, fmt.Errorf("error parsing parameter 'n':\n%q", err)
			case n_clusters < 2:
				return 0, fmt.Errorf("'n' must be at least 2, you gave n=%d", n_clusters)
			}
			return n_clusters, nil
		},
		func(sourceImageFilename, resultImageFilename string, n_clusters int) error {
			return fimgs.ApplyKMeansFilter(sourceImageFilename, resultImageFilename, n_clusters)
		},
	}))
	mux.HandleFunc("/hilbert", filterHandler(Filter[struct{}]{
		"Hilbert curve", "filter.html", noValidate,
		func(sourceImageFilename, resultImageFilename string, _ struct{}) error {
			return fimgs.HilbertCurve(sourceImageFilename, resultImageFilename)
		},
	}))
	mux.HandleFunc("/hilbertdarken", filterHandler(Filter[struct{}]{
		"Hilbert curve darken", "filter.html", noValidate,
		func(sourceImageFilename, resultImageFilename string, _ struct{}) error {
			return fimgs.HilbertDarken(sourceImageFilename, resultImageFilename)
		},
	}))
	mux.HandleFunc("/shader", filterHandler(Filter[string]{
		"Shader", "shader.html",
		func(form url.Values) (string, error) {
			if !form.Has("fragment_shader_source") {
				return "", fmt.Errorf("'fragment_shader_source' is not provided")
			}
			fragment_shader_source := form.Get("fragment_shader_source")
			// TODO: compile shader and return any errors
			return fragment_shader_source, nil
		},
		func(sourceImageFilename, resultImageFilename string, fragment_shader_source string) error {
			return fimgs.ShaderFilter(sourceImageFilename, resultImageFilename, fragment_shader_source)
		},
	}))

	s := &http.Server{
		Addr: ":8080",
		Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// TODO: log incoming requests in web server thoroughly, log request params, log result, timing
			// TODO: store result metrics in db for monitoring
			log.Printf("method=%s url=%q", r.Method, r.URL)
			mux.ServeHTTP(w, r)
		}),
		ReadTimeout:    10 * time.Second,
		WriteTimeout:   10 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}
	log.Fatal(s.ListenAndServe())
}
