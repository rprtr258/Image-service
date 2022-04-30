package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"strconv"
	"time"

	"github.com/rprtr258/fimgs"
)

// TODO: reduce text
const (
	ProgramUsage = `Usage:
%[1]s <filter_name> <filter_params> <source_image_file>
Applies filter to image and saves new image. Available filters:
Convolution filters:
	blur, weakblur, emboss, sharpen, edgeenhance, edgedetect1, edgedetect2, horizontallines, verticallines
Hilbert filters:
	hilbert, hilbertdarken
Other:
	cluster - clusters colors, required parameters:
		number of clusters, must be integer and greater than 1
	shader - apply GLSL filter to image, required parameters:
		shader file, must be valid fragment shader source, see shader_examples for examples
	quadtree - apply quad tree like filter, required parameters:
		threshold must be integer from 0 to 65536 exclusive
		power must be float greater than 0.0
Example usage:
	%[1]s emboss girl.png
	%[1]s cluster 4 rain.jpeg
	%[1]s quadtree 40000 3.14 girl.png`
	ClusterFilterUsage = `Usage:
%[1]s %[2]s <number_of_clusters> <source_image_file>
Applies cluster filter to image and saves new image.
Example usage:
	%[1]s %[2]s 5 girl.png`
	ShaderFilterUsage = `Usage:
%[1]s %[2]s <fragment_shader_source_file> <source_image_file>
Applies GLSL shader to image and saves new image. Shader file must be valid fragment shader source, see shader_examples for examples.
Example usage:
	%[1]s %[2]s shader.glsl girl.png`
	QuadTreeUsage = `Usage:
%[1]s %[2]s <threshold> <power> <source_image_file>
Applies quad tree like filter to image and saves new image.
	threshold must be integer from 0 to 65536 exclusive
	power must be float greater than 0.0
Example usage:
	%[1]s %[2]s 40000 3.14 girl.png`
	SimpleFilterUsage = `Usage:
%[1]s %[2]s <source_image_file>
Applies filter to image and saves new image.
Example usage:
	%[1]s %[2]s girl.png`
)

func makeResultFilename(filename string) string {
	nowString := time.Now().Format("2006-01-02-03-04-05")
	return fmt.Sprintf("%s.fimgs.%s.png", filename, nowString)
}

// TODO: interface not to return string?
func convolutionFilter(kernel [][]int) (string, string) {
	if len(os.Args) != 3 {
		return "", fmt.Sprintf(SimpleFilterUsage, os.Args[0], os.Args[1])
	}
	sourceImageFilename := os.Args[2]
	resultImageFilename := makeResultFilename(sourceImageFilename)
	err := fimgs.ApplyConvolutionFilter(sourceImageFilename, resultImageFilename, kernel)
	if err != nil {
		return "", fmt.Sprintf("Error applying filter:\n%s", err)
	}
	return resultImageFilename, ""
}

func mainRoutine() (string, string) {
	if len(os.Args) == 1 {
		return "", fmt.Sprintf(ProgramUsage, os.Args[0])
	}
	var resultImageFilename string
	// TODO: init filter using os.Args, only then apply
	switch os.Args[1] {
	case "--help", "-h":
		return "", fmt.Sprintf(ProgramUsage, os.Args[0])
	case "blur":
		return convolutionFilter(fimgs.BLUR_KERNEL)
	case "weakblur":
		return convolutionFilter(fimgs.WEAK_BLUR_KERNEL)
	case "emboss":
		return convolutionFilter(fimgs.EMBOSS_KERNEL)
	case "sharpen":
		return convolutionFilter(fimgs.SHARPEN_KERNEL)
	case "edgeenhance":
		return convolutionFilter(fimgs.EDGE_ENHANCE_KERNEL)
	case "edgedetect1":
		return convolutionFilter(fimgs.EDGE_DETECT1_KERNEL)
	case "edgedetect2":
		return convolutionFilter(fimgs.EDGE_DETECT2_KERNEL)
	case "horizontallines":
		return convolutionFilter(fimgs.HORIZONTAL_LINES_KERNEL)
	case "verticallines":
		return convolutionFilter(fimgs.VERTICAL_LINES_KERNEL)
	case "cluster":
		if len(os.Args) != 4 {
			return "", fmt.Sprintf(ClusterFilterUsage, os.Args[0], os.Args[1])
		}
		n_clusters, err := strconv.Atoi(os.Args[2])
		if err != nil {
			return "", fmt.Sprintf("Clusters number should be number, not %q", os.Args[2])
		}
		sourceImageFilename := os.Args[3]
		resultImageFilename = makeResultFilename(sourceImageFilename)
		if err := fimgs.ApplyKMeansFilter(sourceImageFilename, resultImageFilename, n_clusters); err != nil {
			return "", fmt.Sprintf("Error applying filter:\n%s", err)
		}
		return resultImageFilename, ""
	// "lamuse"        "la_muse"
	// "scream"        "scream"
	// "wave"          "wave"
	// "wreck"         "wreck"
	// "udnie"         "udnie"
	// "rain_princess" "rain_princess"
	case "zcurve":
		if len(os.Args) != 3 {
			return "", fmt.Sprintf(SimpleFilterUsage, os.Args[0], os.Args[1])
		}
		sourceImageFilename := os.Args[2]
		resultImageFilename = makeResultFilename(sourceImageFilename)
		if err := fimgs.ZCurve(sourceImageFilename, resultImageFilename); err != nil {
			return "", fmt.Sprintf("Error applying filter:\n%s", err)
		}
		return resultImageFilename, ""
	case "hilbert":
		if len(os.Args) != 3 {
			return "", fmt.Sprintf(SimpleFilterUsage, os.Args[0], os.Args[1])
		}
		sourceImageFilename := os.Args[2]
		resultImageFilename = makeResultFilename(sourceImageFilename)
		if err := fimgs.HilbertCurve(sourceImageFilename, resultImageFilename); err != nil {
			return "", fmt.Sprintf("Error applying filter:\n%s", err)
		}
		return resultImageFilename, ""
	case "hilbertdarken":
		if len(os.Args) != 3 {
			return "", fmt.Sprintf(SimpleFilterUsage, os.Args[0], os.Args[1])
		}
		sourceImageFilename := os.Args[2]
		resultImageFilename = makeResultFilename(sourceImageFilename)
		if err := fimgs.HilbertDarken(sourceImageFilename, resultImageFilename); err != nil {
			return "", fmt.Sprintf("Error applying filter:\n%s", err)
		}
		return resultImageFilename, ""
	case "quadtree":
		if len(os.Args) != 5 {
			return "", fmt.Sprintf(QuadTreeUsage, os.Args[0], os.Args[1])
		}
		threshold, err := strconv.Atoi(os.Args[2])
		if err != nil {
			return "", fmt.Sprintf("Error parsing threshold: %s", err)
		}
		power, err := strconv.ParseFloat(os.Args[3], 64)
		if err != nil {
			return "", fmt.Sprintf("Error parsing power: %s", err)
		}
		sourceImageFilename := os.Args[4]
		resultImageFilename = makeResultFilename(sourceImageFilename)
		if err := fimgs.QudTreeFilter(sourceImageFilename, resultImageFilename, power, threshold); err != nil {
			return "", fmt.Sprintf("Error applying filter:\n%s", err)
		}
		return resultImageFilename, ""
	case "shader":
		if len(os.Args) != 4 {
			return "", fmt.Sprintf(ShaderFilterUsage, os.Args[0], os.Args[1])
		}
		fragmentShaderFilename := os.Args[2]
		sourceImageFilename := os.Args[3]
		fragmentShaderFile, err := os.Open(fragmentShaderFilename)
		if err != nil {
			return "", fmt.Sprintf("error opening fragment shader source: %q", err)
		}
		fragmentShaderSourceData, err := ioutil.ReadAll(fragmentShaderFile)
		if err != nil {
			return "", fmt.Sprintf("error loading fragment shader source: %q", err)
		}
		resultImageFilename = makeResultFilename(sourceImageFilename)
		if err := fimgs.ShaderFilter(sourceImageFilename, resultImageFilename, string(fragmentShaderSourceData)); err != nil {
			return "", fmt.Sprintf("Error applying filter:\n%s", err)
		}
		return resultImageFilename, ""
	default:
		return "", fmt.Sprintf("unknown command: %q", os.Args[1])
	}
}

func main() {
	resultImageFilename, message := mainRoutine()
	if message != "" {
		fmt.Fprintf(os.Stderr, "%s\n", message)
		os.Exit(1)
	}
	fmt.Print(resultImageFilename)
}
