#!/usr/bin/awk BEGIN{for (i=2; i<ARGC; i++) { subslice = subslice ARGV[i] " " }; system("risor "ARGV[1]" -- "subslice)}

// markdown routines
func H1(text) { return '# {text}\n' }
func H2(text) { return '## {text}\n' }
func Code(lang, code) { return '```{lang}\n{code}\n```\n' }
func Table(headers /* []string */, rows /* [][]string */) {
  res := "|" + strings.join(headers, "|") + "|\n" +
    strings.repeat("|-", len(headers)) + "|\n"
	for _, row := range rows {
    res += "|" + strings.join(row, "|") + "|\n"
	}
  return res
}


// math routines
func min(x, y) { return x < y ? x : y }


// list routines
func transform(list, fn) {
  res := []
  for _, item := range list {
    res.append(fn(item))
  }
  return res
}
func filter(list, predicate) {
  res := []
  for _, item := range list {
    if predicate(item) {
      res.append(item)
    }
  }
  return res
}
func chunked(list, size) {
  res := []
  for i := 0; i < len(list); i += size {
    res.append(list[i:min(len(list), i+size)])
  }
  return res
}


from cli import command, flag

images := {
  "shader_rgb":      ["shader", "-s", "shader_examples/rgb_coloring.glsl"],
  "zcurve":          ["zcurve"],
  "verticallines":   ["verticallines"],
  "sharpen":         ["sharpen"],
  "blur":            ["blur"],
  "quadtree":        ["quadtree"],
  "weakblur":        ["weakblur"],
  "median":          ["median"],
  "hilbertdarken":   ["hilbertdarken"],
  "hilbert":         ["hilbert"],
  "emboss":          ["emboss"],
  "horizontallines": ["horizontallines"],
  "edgedetect2":     ["edgedetect2"],
  "edgeenhance":     ["edgeenhance"],
  "edgedetect1":     ["edgedetect1"],
  "cluster":         ["cluster", "-n", "7"],
}

imgs_dir := "img/static"

cli.app({
  name:  "mk",
  description: "commands runner",
  commands: [
    command({
      name:  "imgs",
      description: "update example imgs from orig.png",
      action: func(ctx) {
        origFilename := filepath.join(imgs_dir, "orig.png")

        wg := chan(len(images))
        for destination, args := range images {
          go func(destination, args) {
            print("building", destination, args)
            imageFilename := exec("go", ["run", "cmd/fimgs/main.go", "-i", origFilename] + args).stdout.trim_space()
            os.rename(imageFilename, filepath.join(imgs_dir, destination+".png"))
            wg <- destination // TODO: defer not working
          }(destination, args)
        }
        for range images {
          destination := <-wg
          print("built", destination)
        }
      },
    }),
    command({
      name:  "readme",
      description: "compile readme file",
      action: func(ctx) {
        examples := os.read_dir(imgs_dir) |
          transform(func(k) {return k.name}) |
          filter(func(name) {return strings.has_suffix(name, ".png")})

        rows := []
        for _, titles := range chunked(examples, 3) {
          pics := titles | transform(func(img) {return '![](./img/static/{img}.png)'})
          rows.append(pics) // TODO: report make append variadic
          rows.append(titles)
        }

        os.write_file("README.md", H1("fimgs - image filters tool") +
          H2("Install") +
            Code("bash", "go install github.com/rprtr258/fimgs/cmd/fimgs@latest") +
          H2("Usage") +
            Code("php", exec("go", ["run", "cmd/fimgs/main.go", "--help"]).stdout) +
          H2("Examples") +
            Table(["", "", ""], rows))
      },
    }),
  ],
}).run(["vahui"] + os.args())