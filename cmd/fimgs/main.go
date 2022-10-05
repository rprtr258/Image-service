package main

import (
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"github.com/urfave/cli/v2"

	fimgs "github.com/rprtr258/fimgs/pkg"
)

func makeResultFilename(filename string) string {
	nowString := time.Now().Format("2006-01-02-03-04-05")
	return fmt.Sprintf("%s.fimgs.%s.png", filename, nowString)
}

func main() {
	var sourceImageFilename string
	var resultImageFilename string

	var nClusters int
	clusterCmd := &cli.Command{
		Name:  "cluster",
		Usage: "Cluster colors.",
		UsageText: `Cluster colors using KMeans algorithm.
Example:
	fimgs cluster -n 4 girl.png`,
		Flags: []cli.Flag{&cli.IntFlag{
			Name:        "nclusters",
			Usage:       "number of clusters, must be greater than 1",
			Aliases:     []string{"n"},
			Required:    true,
			Destination: &nClusters,
		}},
		Action: func(*cli.Context) error {
			resultImageFilename = makeResultFilename(sourceImageFilename)
			return fimgs.ApplyKMeansFilter(sourceImageFilename, resultImageFilename, nClusters)
		},
	}

	var threshold int
	var power float64
	quadTreeCmd := &cli.Command{
		Name:  "quadtree",
		Usage: "Quad tree filter.",
		UsageText: `Apply quad tree like filter.
Example:
	fimgs quadtree girl.png`,
		Flags: []cli.Flag{
			&cli.IntFlag{
				Name:        "threshold",
				Aliases:     []string{"t"},
				Usage:       "must be from 0 to 65536 exclusive",
				Value:       32000,
				Destination: &threshold,
			},
			&cli.Float64Flag{
				Name:        "power",
				Aliases:     []string{"p"},
				Usage:       "must be greater than 0.0",
				Value:       2.0,
				Destination: &power,
			},
		},
		Action: func(*cli.Context) error {
			resultImageFilename = makeResultFilename(sourceImageFilename)
			return fimgs.QudTreeFilter(sourceImageFilename, resultImageFilename, power, threshold)
		},
	}

	var fragmentShaderFilename string
	shaderCmd := &cli.Command{
		Name:  "shader",
		Usage: "Shader filter.",
		UsageText: `Apply GLSL filter to image.
Example:
	fimgs shader -s shader_examples/rgb_coloring.glsl -i girl.png`,
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:      "shader",
				Aliases:   []string{"s"},
				Usage:     "shader file, must be valid fragment shader source, see shader_examples directory for examples",
				Required:  true,
				TakesFile: true,
			},
		},
		Action: func(*cli.Context) error {
			fragmentShaderFile, err := os.Open(fragmentShaderFilename)
			if err != nil {
				return fmt.Errorf("error opening fragment shader source file: %w", err)
			}
			fragmentShaderSourceData, err := io.ReadAll(fragmentShaderFile)
			if err != nil {
				return fmt.Errorf("error loading fragment shader source: %w", err)
			}
			resultImageFilename = makeResultFilename(sourceImageFilename)
			return fimgs.ShaderFilter(sourceImageFilename, resultImageFilename, string(fragmentShaderSourceData))
		},
	}

	hilbertCmd := &cli.Command{
		Name:      "hilbert",
		Usage:     "Hilbert curve filter.",
		UsageText: `Draws hilbert curve only through points on dark areas.`,
		Action: func(*cli.Context) error {
			resultImageFilename = makeResultFilename(sourceImageFilename)
			return fimgs.HilbertCurve(sourceImageFilename, resultImageFilename)
		},
	}

	hilbertDarkenCmd := &cli.Command{
		Name:      "hilbertdarken",
		Usage:     "Hilbert darken curve filter.",
		UsageText: `Darken(image, hilbert filter).`,
		Action: func(*cli.Context) error {
			resultImageFilename = makeResultFilename(sourceImageFilename)
			return fimgs.HilbertDarken(sourceImageFilename, resultImageFilename)
		},
	}

	zcurveCmd := &cli.Command{
		Name:      "zcurve",
		Usage:     "Z curve filter.",
		UsageText: `Draws Z curve only through points on dark areas.`,
		Action: func(*cli.Context) error {
			resultImageFilename = makeResultFilename(sourceImageFilename)
			return fimgs.ZCurve(sourceImageFilename, resultImageFilename)
		},
	}

	var windowSize int
	medianCmd := &cli.Command{
		Name:      "median",
		Usage:     "Median filter.",
		UsageText: `Replace each pixel's color with median color of neighbourhood.`,
		Flags: []cli.Flag{&cli.IntFlag{
			Name:        "window",
			Aliases:     []string{"w"},
			Usage:       "window size, must be odd and positive",
			Value:       5,
			Destination: &windowSize,
		}},
		Action: func(*cli.Context) error {
			resultImageFilename = makeResultFilename(sourceImageFilename)
			return fimgs.MedianFilter(sourceImageFilename, resultImageFilename, windowSize)
		},
	}

	convolutionCmds := []*cli.Command{}
	// TODO: allow to specify custom kernel
	for filterName, kernel := range map[string][][]int{
		"blur":            fimgs.BLUR_KERNEL,
		"weakblur":        fimgs.WEAK_BLUR_KERNEL,
		"emboss":          fimgs.EMBOSS_KERNEL,
		"sharpen":         fimgs.SHARPEN_KERNEL,
		"edgeenhance":     fimgs.EDGE_ENHANCE_KERNEL,
		"edgedetect1":     fimgs.EDGE_DETECT1_KERNEL,
		"edgedetect2":     fimgs.EDGE_DETECT2_KERNEL,
		"horizontallines": fimgs.HORIZONTAL_LINES_KERNEL,
		"verticallines":   fimgs.VERTICAL_LINES_KERNEL,
	} {
		convolutionCmds = append(convolutionCmds, &cli.Command{
			Name:  filterName,
			Usage: fmt.Sprintf("%s filter.", strings.Title(filterName)),
			UsageText: fmt.Sprintf(`Apply %s convolution filter.
Example:
	fimgs emboss -i girl.png`, filterName),
			Action: func(*cli.Context) error {
				resultImageFilename = makeResultFilename(sourceImageFilename)
				return fimgs.ApplyConvolutionFilter(sourceImageFilename, resultImageFilename, kernel)
			},
		})
	}

	app := cli.App{
		Name:      "fimgs",
		Usage:     "Applies filter to image.",
		UsageText: "Applies filter to image and saves new image.",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:        "image",
				Aliases:     []string{"i"},
				Destination: &sourceImageFilename,
				Required:    true,
				TakesFile:   true,
				Usage:       "input image filename",
				// TODO: validate available extensions ("image", "png", "jpeg", "jpg")
			},
		},
		Commands: append(
			convolutionCmds,
			clusterCmd,
			quadTreeCmd,
			shaderCmd,
			hilbertCmd,
			hilbertDarkenCmd,
			zcurveCmd,
			medianCmd,
		),
		After: func(*cli.Context) error {
			fmt.Println(resultImageFilename)
			return nil
		},
	}
	if err := app.Run(os.Args); err != nil {
		os.Exit(1)
	}
}
