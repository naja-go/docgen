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

package yaml

import (
	"io"
	"strings"
	"unicode/utf8"
)

// Emitter writes the entries of one YAML mapping, in the order its methods
// are called. The caller is responsible for not repeating a key.
//
// Write errors are sticky: after the first one, every method is a no-op and
// [Emitter.Err] reports it. Nested emitters handed to callbacks share the
// error of the emitter that created them and are only valid during the
// callback.
type Emitter interface {
	// String emits key with a string value.
	String(key, value string)

	// Strings emits key with a sequence of strings.
	Strings(key string, values []string)

	// Mapping emits key with a nested mapping. fn fills in its entries.
	Mapping(key string, fn func(m Emitter))

	// Mappings emits key with a sequence of n mappings. fn is called once
	// per element, in order, to fill in the entries of element i.
	Mappings(key string, n int, fn func(i int, m Emitter))

	// Err returns the first error encountered while writing, if any.
	Err() error
}

// NewEmitter returns an [Emitter] that writes a YAML document to w. The
// document's root is a mapping; the returned emitter writes its entries.
//
// Output is written to w as it is produced and is not buffered.
func NewEmitter(w io.Writer) Emitter {
	return &emitter{out: &output{w: w}}
}

const indentStep = "  "

// output is the destination shared by an emitter and all of its children.
type output struct {
	w   io.Writer
	err error
}

func (o *output) write(parts ...string) {
	if o.err != nil {
		return
	}
	_, o.err = io.WriteString(o.w, strings.Join(parts, ""))
}

type emitter struct {
	out    *output
	indent string // indentation of this mapping's keys
	first  string // written instead of indent before the first entry, if set
	count  int    // number of entries written
}

// key starts a new entry: indentation, the key and the colon.
func (e *emitter) key(k string) {
	prefix := e.indent
	if e.count == 0 && e.first != "" {
		prefix = e.first
	}
	e.count++
	e.out.write(prefix, formatKey(k), ":")
}

func (e *emitter) String(key, value string) {
	e.key(key)
	e.out.write(" ", formatScalar(value, e.indent+indentStep), "\n")
}

func (e *emitter) Strings(key string, values []string) {
	e.key(key)
	if len(values) == 0 {
		e.out.write(" []\n")
		return
	}
	e.out.write("\n")
	item := e.indent + indentStep
	for _, v := range values {
		e.out.write(item, "- ", formatScalar(v, item+indentStep), "\n")
	}
}

func (e *emitter) Mapping(key string, fn func(m Emitter)) {
	e.key(key)
	inner := e.indent + indentStep
	child := &emitter{out: e.out, indent: inner, first: "\n" + inner}
	fn(child)
	if child.count == 0 {
		e.out.write(" {}\n")
	}
}

func (e *emitter) Mappings(key string, n int, fn func(i int, m Emitter)) {
	e.key(key)
	if n <= 0 {
		e.out.write(" []\n")
		return
	}
	e.out.write("\n")
	item := e.indent + indentStep
	for i := range n {
		child := &emitter{out: e.out, indent: item + indentStep, first: item + "- "}
		fn(i, child)
		if child.count == 0 {
			e.out.write(item, "- {}\n")
		}
	}
}

func (e *emitter) Err() error {
	return e.out.err
}

// formatKey returns k as a plain scalar when that is unambiguous, and as a
// double-quoted scalar otherwise.
func formatKey(k string) string {
	if isPlainKey(k) {
		return k
	}
	return quote(k)
}

// isPlainKey reports whether k is an identifier-like key that no YAML parser
// resolves to a non-string type.
func isPlainKey(k string) bool {
	if k == "" {
		return false
	}
	for i := 0; i < len(k); i++ {
		c := k[i]
		switch {
		case c >= 'a' && c <= 'z', c >= 'A' && c <= 'Z', c == '_':
		case i > 0 && (c >= '0' && c <= '9' || c == '-'):
		default:
			return false
		}
	}
	switch strings.ToLower(k) {
	case "null", "true", "false", "yes", "no", "on", "off", "y", "n":
		return false
	}
	return true
}

// formatScalar renders s as a YAML scalar. indent is the indentation to use
// for the content lines of a block scalar.
func formatScalar(s, indent string) string {
	s = strings.ToValidUTF8(s, string(utf8.RuneError))
	if !isLiteralSafe(s) {
		return quote(s)
	}

	var b strings.Builder
	body := s
	switch trailing := len(s) - len(strings.TrimRight(s, "\n")); trailing {
	case 0:
		b.WriteString("|-") // strip: no final line break
	case 1:
		b.WriteString("|") // clip: exactly one final line break
		body = s[:len(s)-1]
	default:
		b.WriteString("|+") // keep: every final line break
		body = s[:len(s)-1]
	}
	for line := range strings.SplitSeq(body, "\n") {
		b.WriteString("\n")
		if line != "" {
			b.WriteString(indent)
			b.WriteString(line)
		}
	}
	return b.String()
}

// isLiteralSafe reports whether s can be written as a literal block scalar
// and read back unchanged without an explicit indentation indicator.
func isLiteralSafe(s string) bool {
	if !strings.Contains(s, "\n") {
		return false
	}
	// A first line that is empty or starts with white space would need an
	// indentation indicator.
	switch s[0] {
	case '\n', ' ', '\t':
		return false
	}
	for line := range strings.SplitSeq(s, "\n") {
		// Lines holding only white space are easily confused with empty
		// lines, and trailing white space is easily lost in transit.
		if line != "" && strings.TrimRight(line, " \t") != line {
			return false
		}
		for _, r := range line {
			if r != '\t' && !isPrintable(r) {
				return false
			}
		}
	}
	return true
}

// isPrintable reports whether r may appear unescaped in YAML output. It is
// stricter than the YAML specification: the Unicode line and paragraph
// separators, NEL and the byte order mark are escaped as well.
func isPrintable(r rune) bool {
	switch {
	case r >= 0x20 && r <= 0x7E:
		return true
	case r >= 0xA0 && r <= 0xD7FF:
		return r != 0x2028 && r != 0x2029
	case r >= 0xE000 && r <= 0xFFFD:
		return r != 0xFEFF
	case r >= 0x10000 && r <= 0x10FFFF:
		return true
	}
	return false
}

// quote renders s as a double-quoted YAML scalar on a single line.
func quote(s string) string {
	const hex = "0123456789ABCDEF"

	var b strings.Builder
	b.Grow(len(s) + 2)
	b.WriteByte('"')
	for _, r := range s {
		switch {
		case r == '"':
			b.WriteString(`\"`)
		case r == '\\':
			b.WriteString(`\\`)
		case r == '\n':
			b.WriteString(`\n`)
		case r == '\t':
			b.WriteString(`\t`)
		case r == '\r':
			b.WriteString(`\r`)
		case isPrintable(r):
			b.WriteRune(r)
		case r <= 0xFF:
			b.WriteString(`\x`)
			b.WriteByte(hex[r>>4])
			b.WriteByte(hex[r&0xF])
		default:
			// Everything else that is not printable lies in the BMP.
			b.WriteString(`\u`)
			for shift := 12; shift >= 0; shift -= 4 {
				b.WriteByte(hex[(r>>shift)&0xF])
			}
		}
	}
	b.WriteByte('"')
	return b.String()
}
