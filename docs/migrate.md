# Migrating from `cobra/doc`

Change the import path to `github.com/naja-go/docgen` and the package
qualifier from `doc...` to `docgen...`. Function names are unchanged. Two things
behave differently:

- **Man pages** are written as roff directly.
  `Long`, `Short` and flag usage text are treated as plain text and are no longer run through a Markdown
  renderer. Markdown markup in these usage tests appear literally.
- **YAML** is written by a small built-in emitter. It describes the same data,
  but single-line strings are now always double-quoted and indentation is two
  spaces. This avoids having a yaml parser dependency where `docgen` only emits
  YAML.

## Where a docs generator should live

A common Cobra habit is a `docs` or `gen-man` subcommand inside the CLI
itself. Naja does not recommend this. Documentation is ultimately a build artifact,
modern CLIs and package managers do not presume users need to roll their tool's docs, and
generating docs from a shipped program has real costs:

- Doc generation dependencies are compiled into every binary you release,
  although users never really need it.
- The CLI gains a command that exists only for maintainers.
- Adding a `docs` command changes the command tree itself often leading
  to "hidden" or "maintainer only" command-set workarounds.

Instead, generate docs at development or release time from code that the
shipped binary never includes.

While there are many ways to structure your CLI code,
the layout below achieves that using Go generators.

`docgen` is still listed in your `go.mod` since
`go mod tidy` accounts for generators and test files despite it not being compiled
and linked to your main program.
It is a declared dependency of your module, not a part of your binary.

### Recommended: a separate generator program

Go cannot import a `package main`, so the generator needs the command tree to
live in an importable package. Builders often put it under `internal/` or `cmd/`: the constructor is
exported for use inside your module, and nothing outside the module can import it.

```
.
└── mycli/
    ├── main.go               package main, calls cmd.NewRootCmd().Execute()
    └── cmd/
        ├── root.go           package cmd, builds the root command tree
        ├── mycommand/...     package mycommand that provides a "mycommand"/
        │                     command and subcommand branch
        └── gendocs/main.go   package main, the docs generator
```

`cmd/root.go` builds the main root command at the base of the CLI command-tree:

```go
package cmd

import "github.com/naja-go/naja"

// NewRootCmd builds the command tree.
func NewRootCmd() *naja.Command {
	root := &naja.Command{Use: "mycli", Short: "Does things"}
	root.AddCommand(...)
	return root
}
```

The `main.go` builds the CLI and executes it. It also carries the `go:generate`
directive, so the generator runs from the module root and writes to `./man`:

```go
//go:generate go run ./cmd/gendocs

package main

import (
	"os"

	"example.com/mycli/cmd"
)

func main() {
	if err := cmd.NewRootCmd().Execute(); err != nil {
		os.Exit(1)
	}
}
```

In `cmd/gendocs/main.go`, we can import the root command tree and generate docs.
This `main()` lives and runs outside the main program:

```go
package main

import (
	"log"
	"os"

	"github.com/naja-go/docgen"

	"example.com/mycli/cmd"
)

func main() {
	if err := os.MkdirAll("man", 0o755); err != nil {
		log.Fatal(err)
	}

	if err := docgen.GenManTree(cmd.NewRootCmd(), nil, "man"); err != nil {
		log.Fatal(err)
	}
}
```

In CI or during release, simply run `go generate ./...` to execute the generate
directive and have docs created.

This prefers a constructor such as `NewRootCmd()` over a package-level root variable. 
Each call returns a fresh tree, so the generator and your tests never share state with
the running program.

### Alternative: a build-tagged file in `package main`

Alternatively, you may choose to use a build tag. This suits a small CLI that
keeps its whole command tree unexported in `package main` and has no importable
package for a separate generator to use.

Add a file that is compiled only when the `docgen` tag is set. It registers a
hidden command that writes the docs:

```go
//go:build docgen

package main

import (
	"github.com/naja-go/docgen"
	"github.com/naja-go/naja"
)

func init() {
	rootCmd.AddCommand(&naja.Command{
		Use:    "gendocs <dir>",
		Short:  "Generate man pages",
		Hidden: true,
		Args:   naja.ExactArgs(1),
		RunE: func(_ *naja.Command, args []string) error {
			return docgen.GenManTree(rootCmd, nil, args[0])
		},
	})
}
```

Generate the docs by building with the tag, either by hand or from a
`go:generate` directive in `main.go`:

```go
//go:generate go run -tags docgen . gendocs ./man
```

A normal `go build` skips the file, so release binaries contain neither the
command nor any `docgen` code.

The trade-off is that the tree is still altered while generating. For example,
giving a root command its first subcommand makes Naja add the default
`completion` commands, which then appear in the generated docs. The generator
program above documents exactly the tree you ship, so prefer it when you can.

### Moving an existing `docs` subcommand

1. Delete the subcommand from the CLI.
2. Move its body into `cmd/gendocs/main.go` as above.
3. Replace wherever your build or release scripts ran `mycli docs` with
   `go generate ./...` or `go run ./cmd/gendocs`.
