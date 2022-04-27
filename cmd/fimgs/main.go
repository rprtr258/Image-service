package main

import (
	"fmt"
    "os"
	"path/filepath"
	"strings"

    "github.com/rprtr258/fimgs"
)

func convolutionFilter()

func main() {
    if len(os.Args) == 1 {
        fmt.Println("Usage: ") // TODO: usage
        os.Exit(1)
    }
    switch os.Args[1] {
        case "blur":
        case "weakblur":
        case "emboss":
        case "sharpen":
        case "edgeenhance":
        case "edgedetect1":
        case "edgedetect2":
        case "horizontallines":
        case "verticallines":
        case "cluster":
            KMeansFilter()
        // "lamuse"        "la_muse"
        // "scream"        "scream"
        // "wave"          "wave"
        // "wreck"         "wreck"
        // "udnie"         "udnie"
        // "rain_princess" "rain_princess"
        case "hilbert":
            HilbertFilter
        case "hilbertdarken":
            HilbertDarkenFilter
        case "shader":
            ShaderFilter
	}
}
