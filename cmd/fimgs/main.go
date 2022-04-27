package main

import (
	"fmt"
    "os"
    "strconv"

    "github.com/rprtr258/fimgs"
)

func handleProcessingError(err error) {
    fmt.Fprintf(os.Stderr, "Error: %q\n", err)
    os.Exit(1)
}

func makeResultFilename(filename string) string {
    return fmt.Sprintf("%s.fimgs.png", filename)
}

func convolutionFilter(kernel [][]int) {
    if len(os.Args) != 3 {
        fmt.Fprintln(os.Stderr, "Usage: ") // TODO: usage
        os.Exit(1)
    }
    sourceImageFilename := os.Args[2]
    resultImageFilename := makeResultFilename(sourceImageFilename)
    err := fimgs.ApplyConvolutionFilter(sourceImageFilename, resultImageFilename, kernel)
    if err != nil {
        handleProcessingError(err)
    }
}

func main() {
    if len(os.Args) == 1 {
        fmt.Fprintln(os.Stderr, "Usage: ") // TODO: usage
        os.Exit(1)
    }
    switch os.Args[1] {
        case "blur":
            convolutionFilter(fimgs.BLUR_KERNEL)
        case "weakblur":
            convolutionFilter(fimgs.WEAK_BLUR_KERNEL)
        case "emboss":
            convolutionFilter(fimgs.EMBOSS_KERNEL)
        case "sharpen":
            convolutionFilter(fimgs.SHARPEN_KERNEL)
        case "edgeenhance":
            convolutionFilter(fimgs.EDGE_ENHANCE_KERNEL)
        case "edgedetect1":
            convolutionFilter(fimgs.EDGE_DETECT1_KERNEL)
        case "edgedetect2":
            convolutionFilter(fimgs.EDGE_DETECT2_KERNEL)
        case "horizontallines":
            convolutionFilter(fimgs.HORIZONTAL_LINES_KERNEL)
        case "verticallines":
            convolutionFilter(fimgs.VERTICAL_LINES_KERNEL)
        case "cluster":
            if len(os.Args) != 4 {
                fmt.Fprintln(os.Stderr, "Usage: ") // TODO: usage
                os.Exit(1)
            }
            n_clusters, err := strconv.Atoi(os.Args[2])
            if err != nil {
                fmt.Fprintf(os.Stderr, "Wrong clusters number: %q", os.Args[2])
                os.Exit(1)
            }
            sourceImageFilename := os.Args[3]
            resultImageFilename := makeResultFilename(sourceImageFilename)
            if err := fimgs.ApplyKMeansFilter(sourceImageFilename, resultImageFilename, n_clusters); err != nil {
                handleProcessingError(err)
            }
        // "lamuse"        "la_muse"
        // "scream"        "scream"
        // "wave"          "wave"
        // "wreck"         "wreck"
        // "udnie"         "udnie"
        // "rain_princess" "rain_princess"
        case "hilbert":
            if len(os.Args) != 3 {
                fmt.Fprintln(os.Stderr, "Usage: ") // TODO: usage
                os.Exit(1)
            }
            sourceImageFilename := os.Args[2]
            resultImageFilename := makeResultFilename(sourceImageFilename)
            if err := fimgs.HilbertCurve(sourceImageFilename, resultImageFilename); err != nil {
                handleProcessingError(err)
            }
        case "hilbertdarken":
            if len(os.Args) != 3 {
                fmt.Fprintln(os.Stderr, "Usage: ") // TODO: usage
                os.Exit(1)
            }
            sourceImageFilename := os.Args[2]
            resultImageFilename := makeResultFilename(sourceImageFilename)
            if err := fimgs.HilbertDarken(sourceImageFilename, resultImageFilename); err != nil {
                handleProcessingError(err)
            }
        case "shader":
            if len(os.Args) != 4 {
                fmt.Fprintln(os.Stderr, "Usage: ") // TODO: usage
                os.Exit(1)
            }
            fragmentShaderSource := os.Args[2]
            sourceImageFilename := os.Args[3]
            resultImageFilename := makeResultFilename(sourceImageFilename)
            if err := fimgs. ShaderFilter(sourceImageFilename, resultImageFilename, fragmentShaderSource); err != nil {
                handleProcessingError(err)
            }
	}
}
