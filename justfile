@_help:
    just --list --unsorted

# install development tools
tools:
    curl -SsL -o ./mustpl https://github.com/tarampampam/mustpl/releases/latest/download/mustpl-linux-amd64
    chmod +x ./mustpl
    sudo install -g root -o root -t /usr/local/bin -v ./mustpl
    rm ./mustpl

USAGE := `go run cmd/fimgs/main.go --help`
EXAMPLES := `find . -name '*.png' | sed -e 's/\.\/img\/static\///' -e 's/\.png//' | awk 'NR%3==1{printf"%s",$1}NR%3==2{printf" %s ",$1}NR%3==0' | awk '{printf"|![](img/static/%s.png)|![](img/static/%s.png)|![](img/static/%s.png)|\n|%s|%s|%s|\n",$1,$2,$3,$1,$2,$3}'`
# compile readme file
@readme:
    mustpl -d '{"usage": "{{USAGE}}", "examples": "{{EXAMPLES}}"}' img/README.md.tpl > README.md

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
