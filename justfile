@_help:
    just --list

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
