@_help:
    just --list --unsorted

# compile readme file
@readme $USAGE = `go run cmd/fimgs/main.go --help` $EXAMPLES = `find . -name '*.png'`:
    go install github.com/hairyhenderson/gomplate/v3/cmd/gomplate@latest
    cat ./img/README.md.tpl | gomplate > README.md

IMGS := "img/static"
FIMGS := "go run cmd/fimgs/main.go -i "+IMGS/"orig.png"
# update example imgs from orig.png
imgs:
    mv $({{FIMGS}} shader -s shader_examples/rgb_coloring.glsl) {{IMGS}}/shader_rgb.png
    mv $({{FIMGS}} zcurve) {{IMGS}}/zcurve.png
    mv $({{FIMGS}} verticallines) {{IMGS}}/verticallines.png
    mv $({{FIMGS}} sharpen) {{IMGS}}/sharpen.png
    mv $({{FIMGS}} blur) {{IMGS}}/blur.png
    mv $({{FIMGS}} verticallines) {{IMGS}}/verticallines.png
    mv $({{FIMGS}} sharpen) {{IMGS}}/sharpen.png
    mv $({{FIMGS}} quadtree) {{IMGS}}/quadtree.png
    mv $({{FIMGS}} weakblur) {{IMGS}}/weakblur.png
    mv $({{FIMGS}} median) {{IMGS}}/median.png
    mv $({{FIMGS}} hilbertdarken) {{IMGS}}/hilbertdarken.png
    mv $({{FIMGS}} hilbert) {{IMGS}}/hilbert.png
    mv $({{FIMGS}} emboss) {{IMGS}}/emboss.png
    mv $({{FIMGS}} horizontallines) {{IMGS}}/horizontallines.png
    mv $({{FIMGS}} edgedetect2) {{IMGS}}/edgedetect2.png
    mv $({{FIMGS}} edgeenhance) {{IMGS}}/edgeenhance.png
    mv $({{FIMGS}} edgedetect1) {{IMGS}}/edgedetect1.png
    mv $({{FIMGS}} verticallines) {{IMGS}}/verticallines.png
    mv $({{FIMGS}} cluster -n 7) {{IMGS}}/cluster.png
