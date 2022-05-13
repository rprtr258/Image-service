package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"strings"
	"time"

	fimgs "github.com/rprtr258/fimgs/pkg"
	"github.com/spf13/cobra"
)

func makeResultFilename(filename string) string {
	nowString := time.Now().Format("2006-01-02-03-04-05")
	return fmt.Sprintf("%s.fimgs.%s.png", filename, nowString)
}

func main() {
	var sourceImageFilename string
	var resultImageFilename string
	rootCmd := cobra.Command{
		Use:   "fimgs",
		Short: "Applies filter to image.",
		Long:  `Applies filter to image and saves new image.`,
		PersistentPostRun: func(cmd *cobra.Command, args []string) {
			fmt.Println(resultImageFilename)
		},
	}
	rootCmd.PersistentFlags().StringVarP(&sourceImageFilename, "image", "i", "", "input image filename")
	rootCmd.MarkPersistentFlagFilename("image", "png", "jpeg", "jpg")
	rootCmd.MarkPersistentFlagRequired("image")

	var nClusters int
	clusterCmd := cobra.Command{
		Use:   "cluster",
		Short: "Cluster colors.",
		Long:  `Cluster colors using KMeans algorithm.`,
		Args:  cobra.MaximumNArgs(0),
		RunE: func(cmd *cobra.Command, args []string) error {
			resultImageFilename = makeResultFilename(sourceImageFilename)
			return fimgs.ApplyKMeansFilter(sourceImageFilename, resultImageFilename, nClusters)
		},
		Example: "fimgs cluster -n 4 girl.png",
	}
	clusterCmd.Flags().IntVarP(&nClusters, "nclusters", "n", 0, "number of clusters, must be greater than 1")
	clusterCmd.MarkFlagRequired("nclusters")
	rootCmd.AddCommand(&clusterCmd)

	var threshold int
	var power float64
	quadTreeCmd := cobra.Command{
		Use:   "quadtree",
		Short: "Quad tree filter.",
		Long:  `Apply quad tree like filter.`,
		Args:  cobra.MaximumNArgs(0),
		RunE: func(cmd *cobra.Command, args []string) error {
			resultImageFilename = makeResultFilename(sourceImageFilename)
			return fimgs.QudTreeFilter(sourceImageFilename, resultImageFilename, power, threshold)
		},
		Example: "fimgs quadtree girl.png",
	}
	quadTreeCmd.Flags().IntVarP(&threshold, "threshold", "t", 32000, "must be from 0 to 65536 exclusive")
	quadTreeCmd.Flags().Float64VarP(&power, "power", "p", 2.0, "must be greater than 0.0")
	rootCmd.AddCommand(&quadTreeCmd)

	var fragmentShaderFilename string
	shaderCmd := cobra.Command{
		Use:   "shader",
		Short: "Shader filter.",
		Long:  `Apply GLSL filter to image.`,
		Args:  cobra.MaximumNArgs(0),
		RunE: func(cmd *cobra.Command, args []string) error {
			fragmentShaderFile, err := os.Open(fragmentShaderFilename)
			if err != nil {
				return fmt.Errorf("Error opening fragment shader source: %q", err)
			}
			fragmentShaderSourceData, err := ioutil.ReadAll(fragmentShaderFile)
			if err != nil {
				return fmt.Errorf("Error loading fragment shader source: %q", err)
			}
			resultImageFilename = makeResultFilename(sourceImageFilename)
			return fimgs.ShaderFilter(sourceImageFilename, resultImageFilename, string(fragmentShaderSourceData))
		},
		Example: "fimgs shader -s shader_examples/rgb_coloring.glsl -i girl.png",
	}
	shaderCmd.Flags().StringVarP(&fragmentShaderFilename, "shader", "s", "", "shader file, must be valid fragment shader source, see shader_examples directory for examples")
	shaderCmd.MarkFlagFilename("shader", "glsl")
	shaderCmd.MarkFlagRequired("shader")
	rootCmd.AddCommand(&shaderCmd)

	hilbertCmd := cobra.Command{
		Use:   "hilbert",
		Short: "Hilbert curve filter.",
		Long:  `Draws hilbert curve only through points on dark areas.`,
		Args:  cobra.MaximumNArgs(0),
		RunE: func(cmd *cobra.Command, args []string) error {
			resultImageFilename = makeResultFilename(sourceImageFilename)
			return fimgs.HilbertCurve(sourceImageFilename, resultImageFilename)
		},
	}
	rootCmd.AddCommand(&hilbertCmd)

	hilbertDarkenCmd := cobra.Command{
		Use:   "hilbertdarken",
		Short: "Hilbert darken curve filter.",
		Long:  `Darken(image, hilbert filter).`,
		Args:  cobra.MaximumNArgs(0),
		RunE: func(cmd *cobra.Command, args []string) error {
			resultImageFilename = makeResultFilename(sourceImageFilename)
			return fimgs.HilbertDarken(sourceImageFilename, resultImageFilename)
		},
	}
	rootCmd.AddCommand(&hilbertDarkenCmd)

	zcurveCmd := cobra.Command{
		Use:   "zcurve",
		Short: "Z curve filter.",
		Long:  `Draws Z curve only through points on dark areas.`,
		Args:  cobra.MaximumNArgs(0),
		RunE: func(cmd *cobra.Command, args []string) error {
			resultImageFilename = makeResultFilename(sourceImageFilename)
			return fimgs.ZCurve(sourceImageFilename, resultImageFilename)
		},
	}
	rootCmd.AddCommand(&zcurveCmd)

	var windowSize int
	medianCmd := cobra.Command{
		Use:   "median",
		Short: "Median filter.",
		Long:  `Replace each pixel's color with median color of neighbourhood.`,
		Args:  cobra.MaximumNArgs(0),
		RunE: func(cmd *cobra.Command, args []string) error {
			resultImageFilename = makeResultFilename(sourceImageFilename)
			return fimgs.MedianFilter(sourceImageFilename, resultImageFilename, windowSize)
		},
	}
	medianCmd.Flags().IntVarP(&windowSize, "window", "w", 5, "window size, must be odd and positive")
	rootCmd.AddCommand(&medianCmd)

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
		rootCmd.AddCommand(&cobra.Command{
			Use:   filterName,
			Short: fmt.Sprintf("%s filter.", strings.Title(filterName)),
			Long:  fmt.Sprintf("Apply %s convolution filter.", filterName),
			Args:  cobra.MaximumNArgs(0),
			RunE: func(cmd *cobra.Command, args []string) error {
				resultImageFilename = makeResultFilename(sourceImageFilename)
				return fimgs.ApplyConvolutionFilter(sourceImageFilename, resultImageFilename, kernel)
			},
			Example: "fimgs emboss -i girl.png",
		})
	}

	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
