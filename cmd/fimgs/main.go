package main

import (
	"fmt"
    "os"

    "github.com/rprtr258/fimgs"
)

func convolutionFilter(kernel [][]int) {
    if len(os.Args) != 3 {
        fmt.Println("Usage: ")
        os.Exit(1)
    }
    sourceImageFilename := os.Args[2]
    resultImageFilename := fmt.Sprintf("%s.fimgs.png", sourceImageFilename)
    err := fimgs.ApplyConvolutionFilter(sourceImageFilename, resultImageFilename, kernel)
    if err != nil {
        fmt.Printf("Error: %q\n", err)
        os.Exit(1)
    }
}

func main() {
    if len(os.Args) == 1 {
        fmt.Println("Usage: ") // TODO: usage
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
            //KMeansFilter()
        // "lamuse"        "la_muse"
        // "scream"        "scream"
        // "wave"          "wave"
        // "wreck"         "wreck"
        // "udnie"         "udnie"
        // "rain_princess" "rain_princess"
        case "hilbert":
            //HilbertFilter
        case "hilbertdarken":
            //HilbertDarkenFilter
        case "shader":
            //ShaderFilter
	}
}
