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
	"errors"
	"strings"
	"testing"
)

func TestEmitterDocument(t *testing.T) {
	var b strings.Builder
	e := NewEmitter(&b)
	e.String("name", "app serve")
	e.String("description", "Serve the app.\n\nRuns until interrupted.\n")
	e.Mappings("options", 2, func(i int, m Emitter) {
		if i == 0 {
			m.String("name", "port")
			m.String("default_value", "8080")
			return
		}
		m.String("name", "verbose")
		m.String("usage", "first line\nsecond line")
	})
	e.Mapping("meta", func(m Emitter) {
		m.String("source", "naja")
		m.Strings("tags", []string{"a", "b"})
		m.Mapping("nested", func(n Emitter) { n.String("deep", "yes") })
	})
	e.Strings("see_also", []string{"app - The app", "multi\nline"})
	if err := e.Err(); err != nil {
		t.Fatal(err)
	}

	want := `name: "app serve"
description: |
  Serve the app.

  Runs until interrupted.
options:
  - name: "port"
    default_value: "8080"
  - name: "verbose"
    usage: |-
      first line
      second line
meta:
  source: "naja"
  tags:
    - "a"
    - "b"
  nested:
    deep: "yes"
see_also:
  - "app - The app"
  - |-
    multi
    line
`
	if got := b.String(); got != want {
		t.Errorf("unexpected output\n got:\n%s\nwant:\n%s", got, want)
	}
}

func TestEmitterEmptyCollections(t *testing.T) {
	var b strings.Builder
	e := NewEmitter(&b)
	e.String("s", "")
	e.Strings("list", nil)
	e.Mapping("map", func(Emitter) {})
	e.Mappings("maps", 0, func(int, Emitter) { t.Error("fn called for n == 0") })
	e.Mappings("holes", 2, func(i int, m Emitter) {
		if i == 1 {
			m.String("k", "v")
		}
	})
	if err := e.Err(); err != nil {
		t.Fatal(err)
	}

	want := `s: ""
list: []
map: {}
maps: []
holes:
  - {}
  - k: "v"
`
	if got := b.String(); got != want {
		t.Errorf("unexpected output\n got:\n%s\nwant:\n%s", got, want)
	}
}

func TestFormatScalar(t *testing.T) {
	tests := []struct {
		name string
		in   string
		want string
	}{
		{"empty", "", `""`},
		{"plain", "hello", `"hello"`},
		{"bool-like", "no", `"no"`},
		{"number-like", "1.0", `"1.0"`},
		{"indicator characters", "a: b # c", `"a: b # c"`},
		{"quote and backslash", `say "hi" \o/`, `"say \"hi\" \\o/"`},
		{"tab", "a\tb", `"a\tb"`},
		{"carriage return", "a\rb", `"a\rb"`},
		{"control character", "a\x01b", `"a\x01b"`},
		{"delete", "a\x7fb", `"a\x7Fb"`},
		{"next line", "a\u0085b", `"a\x85b"`},
		{"line separator", "a\u2028b", `"a\u2028b"`},
		{"byte order mark", "\ufeffa", `"\uFEFFa"`},
		{"unicode", "héllo ✓ 🐍", `"héllo ✓ 🐍"`},
		{"invalid utf-8", "a\xffb", "\"a\ufffdb\""},

		{"literal strip", "a\nb", "|-\n  a\n  b"},
		{"literal clip", "a\nb\n", "|\n  a\n  b"},
		{"literal keep", "a\nb\n\n", "|+\n  a\n  b\n"},
		{"literal single line clip", "a\n", "|\n  a"},
		{"literal blank line", "a\n\nb", "|-\n  a\n\n  b"},
		{"literal inner indentation", "a\n  b", "|-\n  a\n    b"},
		{"literal inner tab", "a\n\tb\tc", "|-\n  a\n  \tb\tc"},
		{"literal with indicators", "---\n# x: y", "|-\n  ---\n  # x: y"},

		{"leading newline", "\na", `"\na"`},
		{"leading space", " a\nb", `" a\nb"`},
		{"leading tab", "\ta\nb", `"\ta\nb"`},
		{"only newlines", "\n\n", `"\n\n"`},
		{"trailing space", "a \nb", `"a \nb"`},
		{"white space only line", "a\n  \nb", `"a\n  \nb"`},
		{"multi-line carriage return", "a\r\nb", `"a\r\nb"`},
		{"multi-line control character", "a\n\x1bb", `"a\n\x1Bb"`},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := formatScalar(tt.in, "  "); got != tt.want {
				t.Errorf("formatScalar(%q)\n got: %q\nwant: %q", tt.in, got, tt.want)
			}
		})
	}
}

func TestFormatKey(t *testing.T) {
	tests := map[string]string{
		"name":          `name`,
		"default_value": `default_value`,
		"see-also":      `see-also`,
		"_x9":           `_x9`,
		"":              `""`,
		"9lives":        `"9lives"`,
		"-dash":         `"-dash"`,
		"two words":     `"two words"`,
		"a:b":           `"a:b"`,
		"null":          `"null"`,
		"True":          `"True"`,
		"no":            `"no"`,
		"Y":             `"Y"`,
		"héllo":         `"héllo"`,
	}
	for in, want := range tests {
		if got := formatKey(in); got != want {
			t.Errorf("formatKey(%q) = %s, want %s", in, got, want)
		}
	}
}

type failingWriter struct {
	writes int
	err    error
}

func (w *failingWriter) Write(p []byte) (int, error) {
	w.writes++
	return 0, w.err
}

func TestEmitterStickyError(t *testing.T) {
	w := &failingWriter{err: errors.New("disk full")}
	e := NewEmitter(w)
	e.String("a", "1")
	e.Mapping("m", func(m Emitter) {
		m.String("b", "2")
		if !errors.Is(m.Err(), w.err) {
			t.Errorf("nested Err() = %v, want %v", m.Err(), w.err)
		}
	})
	e.Strings("c", []string{"3"})
	if !errors.Is(e.Err(), w.err) {
		t.Errorf("Err() = %v, want %v", e.Err(), w.err)
	}
	if w.writes != 1 {
		t.Errorf("writer called %d times after failing, want 1", w.writes)
	}
}
