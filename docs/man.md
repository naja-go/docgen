# Generating Man Pages for a naja.Command

Generate man pages from a Naja command:

```go
package main

import (
	"log"

	"github.com/naja-go/naja"
	"github.com/naja-go/docgen"
)

func main() {
	cmd := &naja.Command{
		Use:   "test",
		Short: "my test program",
	}
	header := &docgen.GenManHeader{
		Title: "MINE",
		Section: "3",
	}
	err := docgen.GenManTree(cmd, header, "/tmp")
	if err != nil {
		log.Fatal(err)
	}
}
```

That will get you a man page `/tmp/test.3`

## How text is rendered

Man pages are written as roff directly. `Short`, `Long`, `Example` and flag
usage strings are treated as plain text:

- special characters are escaped, so the text appears exactly as written;
  Markdown or roff markup in it is not interpreted
- in `Long` and flag usage, a blank line starts a new paragraph, and other
  lines are joined and re-wrapped to the width of the terminal
- `Example` keeps its line breaks and indentation
