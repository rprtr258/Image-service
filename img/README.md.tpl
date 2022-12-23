# fimgs - image filters tool

## Install
```bash
go install github.com/rprtr258/fimgs/cmd/fimgs@latest
```

## Usage
```php
{{.Env.USAGE}}
```

## Examples
||||
|-|-|-|
{{with $v := .Env.EXAMPLE | strings.Split "\n" | append "" | append ""}}{{range $i := seq 0 (sub (len .) 3) 3}}|{{range $r := coll.JSONPath (printf `$[%d:%d]` $i (add $i 3)) $v}}![]({{$r}})|{{end}}
|{{range $r := coll.JSONPath (printf `$[%d:%d]` $i (add $i 3)) $v}}{{$r | strings.TrimPrefix "./img/static/" | strings.TrimSuffix ".png"}}|{{end}}
{{end}}{{end}}
