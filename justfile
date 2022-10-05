_help:
    just --list

USAGE := `go run cmd/fimgs/main.go --help`
# compile readme file
readme:
    # TODO: generate table with examples also
    mustpl -d '{"usage": "{{USAGE}}"}' img/README.md.tpl > README.md
