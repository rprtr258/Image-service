package main

import (
	"bytes"
	"context"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/rprtr258/log"
	"github.com/rprtr258/mk"
	"github.com/rprtr258/mk/cache"
	md "github.com/rprtr258/mk/contrib/markdown"
	"github.com/sourcegraph/conc"
	"github.com/urfave/cli/v2"
)

func main() {
	if err := (&cli.App{
		Name:  "mk",
		Usage: "commands runner",
		Commands: []*cli.Command{
			{
				Name:  "imgs",
				Usage: "update example imgs from orig.png",
				Action: func(ctx *cli.Context) error {
					imgsDir := "img/static"
					origFilename := filepath.Join(imgsDir, "orig.png")
					fimgsCmd := mk.ExecAliasContext("go", "run", "cmd/fimgs/main.go", "-i", origFilename)

					cch := cache.LoadFromFile[string, string](".cache.json")

					origHash, err := cache.HashFile(origFilename)
					if err != nil {
						log.Warnf("can't get orig.png file hash", log.F{"err": err.Error()})
					}

					wg := conc.NewWaitGroup()
					for destination, args := range map[string][]string{
						"shader_rgb":      {"shader", "-s", "shader_examples/rgb_coloring.glsl"},
						"zcurve":          {"zcurve"},
						"verticallines":   {"verticallines"},
						"sharpen":         {"sharpen"},
						"blur":            {"blur"},
						"quadtree":        {"quadtree"},
						"weakblur":        {"weakblur"},
						"median":          {"median"},
						"hilbertdarken":   {"hilbertdarken"},
						"hilbert":         {"hilbert"},
						"emboss":          {"emboss"},
						"horizontallines": {"horizontallines"},
						"edgedetect2":     {"edgedetect2"},
						"edgeenhance":     {"edgeenhance"},
						"edgedetect1":     {"edgedetect1"},
						"cluster":         {"cluster", "-n", "7"},
					} {
						if !cache.CompareAndSwap(cch, args[0], origHash) {
							continue
						}

						destination := destination
						args := args
						wg.Go(func() {
							imageFilename, _ := mk.Must2(fimgsCmd(ctx.Context, args...))
							mk.Must0(os.Rename(strings.TrimSpace(imageFilename), filepath.Join(imgsDir, destination+".png")))
						})
					}
					wg.Wait()

					cache.SaveToFile(".cache.json", cch)

					return nil
				},
			},
			{
				Name:  "readme",
				Usage: "compile readme file",
				Action: func(*cli.Context) error {
					b := &bytes.Buffer{}
					md.H1(b, "fimgs - image filters tool")

					md.H2(b, "Install")
					md.Code(b, "bash", "go install github.com/rprtr258/fimgs/cmd/fimgs@latest")

					md.H2(b, "Usage")
					usage, _ := mk.Must2(mk.ExecContext(context.Background(), "go", "run", "cmd/fimgs/main.go", "--help"))
					md.Code(b, "php", usage)

					examples, err := fs.Glob(os.DirFS("img/static"), "*.png")
					if err != nil {
						return err
					}

					rows := make([][]string, 0, 2*(len(examples)/3+1))
					for i := 0; i < len(examples); i += 3 {
						pics := []string{}
						titles := []string{}
						for j := i; j < len(examples) && j < i+3; j++ {
							pics = append(pics, fmt.Sprintf("![](./img/static/%s)", examples[j]))
							titles = append(titles, strings.TrimSuffix(examples[j], ".png"))
						}
						rows = append(rows, pics, titles)
					}

					md.H2(b, "Examples")
					md.Table(b, []string{"", "", ""}, rows)

					mk.Must0(os.WriteFile("README.md", b.Bytes(), 0o644))

					return nil
				},
			},
		},
	}).Run(os.Args); err != nil {
		log.Fatal(err.Error())
	}
}
