# docgen

Documentation generators for [Naja](https://github.com/naja-go/naja) commands:
man pages, Markdown, reStructuredText and YAML.

## Install

Requires Go 1.27+:

```sh
go get github.com/naja-go/docgen
```

## Usage

Generate docs at **development or release time** from a small program that imports
your root command, for example `internal/gendocs/main.go`:

```go
package main

import (
	"log"

	"github.com/naja-go/docgen"

	"example.com/mycli/cmd"
)

func main() {
	if err := docgen.GenManTree(cmd.Root(), nil, "./man"); err != nil {
		log.Fatal(err)
	}
}
```

Run it by hand or with `//go:generate go run ./internal/gendocs`.

| Format           | Single command | Whole command tree | Guide                        |
|------------------|----------------|--------------------|------------------------------|
| Man pages        | `GenMan`       | `GenManTree`       | [docs/man.md](docs/man.md)   |
| Markdown         | `GenMarkdown`  | `GenMarkdownTree`  | [docs/md.md](docs/md.md)     |
| reStructuredText | `GenReST`      | `GenReSTTree`      | [docs/rest.md](docs/rest.md) |
| YAML             | `GenYaml`      | `GenYamlTree`      | [docs/yaml.md](docs/yaml.md) |

See [docs/README.md](docs/README.md) for options shared by all formats.


## License

Apache 2.0, see [LICENSE.txt](LICENSE.txt). Derived from
[spf13/cobra](https://github.com/spf13/cobra)'s `doc` package.
