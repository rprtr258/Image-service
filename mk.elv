#!/usr/bin/env elvish

use "str"

fn chunked {|size|
  var chunk = []
  each {|item|
    if (== (count $chunk) $size) {
      put $chunk
      set chunk = []
    }
    set chunk = (conj $chunk $item)
  }
  if (!= (count $chunk) 0) {
    put $chunk
  }
}

var images = [
  &shader_rgb=      ["shader" "-s" "shader_examples/rgb_coloring.glsl"]
  &zcurve=          ["zcurve"]
  &verticallines=   ["verticallines"]
  &sharpen=         ["sharpen"]
  &blur=            ["blur"]
  &quadtree=        ["quadtree"]
  &weakblur=        ["weakblur"]
  &median=          ["median"]
  &hilbertdarken=   ["hilbertdarken"]
  &hilbert=         ["hilbert"]
  &emboss=          ["emboss"]
  &horizontallines= ["horizontallines"]
  &edgedetect2=     ["edgedetect2"]
  &edgeenhance=     ["edgeenhance"]
  &edgedetect1=     ["edgedetect1"]
  &cluster=         ["cluster" "-n" "7"]
]

var imgs_dir = "img/static"

var command = $args[0]
# name: "mk",
# description: "commands runner",
if (==s $command "imgs") {
  # description: "update example imgs from orig.png",
  var origFilename = $imgs_dir"/orig.png"
  keys $images | each {|destination|
    var args = $images[$destination]
    echo "building" $destination $args
    var imageFilename = (str:trim-space (go run cmd/fimgs/main.go -i $origFilename (all $args)))
    mv $imageFilename $imgs_dir"/"$destination".png"
    echo "built" $destination
  }
} elif (==s $command "readme") {
  # description: "compile readme file",

  # markdown routines
  fn H1 {|text| put "# "$text"\n" }
  fn H2 {|text| put "## "$text"\n" }
  fn Code {|lang code| put "```"$lang"\n"$code"\n```\n" }
  fn Table {|headers|
    put "|"(str:join "|" $headers)"|\n"(repeat (count $headers) "|-" | str:join "")"|"
    each {|row|
      put "|"(str:join "|" $row)"|"
    }
  }

  var content = ({
    H1 "fimgs - image filters tool"
    H2 "Install"
    Code "bash" "go install github.com/rprtr258/fimgs/cmd/fimgs@latest"
    H2 "Usage"
    Code "php" (go run cmd/fimgs/main.go --help | slurp)
    H2 "Examples"
    ls $imgs_dir |
      each {|name|
        if (str:has-suffix $name ".png") {
          put $name
        }
      } |
      chunked 3 |
      each {|titles|
        all $titles | each {|img| put "![](./img/static/"$img")"} | put [(all)]
        put $titles
      } | Table ["" "" ""]
  } | str:join "\n")
  echo $content >"README.md"
}