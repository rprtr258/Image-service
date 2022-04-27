package main

import (
	"fmt"
    "os"
	"path/filepath"
	"strings"

    "github.com/rprtr258/fimgs"
)

func main() {
    if len(os.Args) == 1 {
        fmt.Println("Usage: ") // TODO: usage
        os.Exit(1)
    }
    switch os.Args[1] {
        case "blur":
            [][]int{
                {1, 1, 1},
                {1, 1, 1},
                {1, 1, 1},
            },
        case "weakblur":
            [][]int{
                {0, 1, 0},
                {1, 1, 1},
                {0, 1, 0},
            },
        case "emboss":
            [][]int{
                {-2, -1, 0},
                {-1, 1, 1},
                {0, 1, 2},
            },
        case "sharpen":
            [][]int{
                {0, -1, 0},
                {-1, 5, -1},
                {0, -1, 0},
            },
        case "edgeenhance":
            [][]int{
                {0, 0, 0},
                {-1, 1, 0},
                {0, 0, 0},
            },
        case "edgedetect1":
            [][]int{
                {1, 0, -1},
                {0, 0, 0},
                {-1, 0, 1},
            },
        case "edgedetect2":
            [][]int{
                {0, -1, 0},
                {-1, 4, -1},
                {0, -1, 0},
            },
        case "horizontallines":
            [][]int{
                {-1, -1, -1},
                {2, 2, 2},
                {-1, -1, -1},
            },
        case "verticallines":
            [][]int{
                {-1, 2, -1},
                {-1, 2, -1},
                {-1, 2, -1},
            },
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
