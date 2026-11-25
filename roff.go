// Copyright 2026 The Naja Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package docgen

import (
	"io"
	"strings"

	"github.com/naja-go/naja"
)

// Man pages are written as roff directly. Text taken from a command (its
// descriptions, flag usage and examples) is treated as plain text: it is
// escaped so that roff prints it literally, and is never interpreted as
// markup of any kind.

// roffText escapes s so that roff prints it literally:
//
//   - a backslash, roff's escape character, becomes \e
//   - a hyphen becomes \-, so that command and flag names can be copied
//     from the rendered page
//   - a line starting with "." or "'" is prefixed with the zero-width \&,
//     so that it is not taken for a request
//
// Line breaks are preserved.
func roffText(s string) string {
	s = strings.ReplaceAll(s, "\r\n", "\n")

	var b strings.Builder
	b.Grow(len(s) + len(s)/8)
	lineStart := true
	for i := 0; i < len(s); i++ {
		c := s[i]
		if lineStart && (c == '.' || c == '\'') {
			b.WriteString(`\&`)
		}
		lineStart = c == '\n'
		switch c {
		case '\\':
			b.WriteString(`\e`)
		case '-':
			b.WriteString(`\-`)
		default:
			b.WriteByte(c)
		}
	}
	return b.String()
}

// roffArg renders s as a single quoted macro argument.
func roffArg(s string) string {
	s = strings.Join(strings.Fields(s), " ")
	return `"` + strings.ReplaceAll(roffText(s), `"`, `\(dq`) + `"`
}

// roffBold renders s in bold.
func roffBold(s string) string {
	return `\fB` + roffText(s) + `\fP`
}

// roffParagraphs writes s as filled text. Paragraphs are separated by blank
// lines in s and by the request sep (".PP" or ".IP") in the output. Nothing
// is written if s holds no text.
func roffParagraphs(buf io.StringWriter, s, sep string) {
	first := true
	inParagraph := false
	for line := range strings.SplitSeq(strings.ReplaceAll(s, "\r\n", "\n"), "\n") {
		line = strings.TrimRight(line, " \t\r")
		if line == "" {
			inParagraph = false
			continue
		}
		if !inParagraph && !first {
			naja.WriteStringAndCheck(buf, sep+"\n")
		}
		first, inParagraph = false, true
		naja.WriteStringAndCheck(buf, roffText(line)+"\n")
	}
}

// roffVerbatim writes s as an unfilled example block, keeping its line
// breaks and indentation.
func roffVerbatim(buf io.StringWriter, s string) {
	s = strings.Trim(strings.ReplaceAll(s, "\r\n", "\n"), "\n")
	naja.WriteStringAndCheck(buf, ".EX\n"+roffText(s)+"\n.EE\n")
}
