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

// Package yaml is a small, dependency-free YAML emitter.
//
// It only writes YAML; it does not parse it, and it does not use reflection.
// Callers describe a document through the [Emitter] interface, one mapping
// entry at a time:
//
//	e := yaml.NewEmitter(os.Stdout)
//	e.String("name", "app serve")
//	e.Strings("see_also", []string{"app", "app version"})
//	e.Mappings("options", len(opts), func(i int, m yaml.Emitter) {
//		m.String("name", opts[i].Name)
//		m.String("usage", opts[i].Usage)
//	})
//	if err := e.Err(); err != nil {
//		// handle the write error
//	}
//
// # Supported YAML
//
// The data model is deliberately narrow: mappings with string keys whose
// values are strings, sequences of strings, mappings, or sequences of
// mappings. That is what naja needs to describe a command tree.
//
// # Output style
//
// The emitter never has to decide whether an unquoted scalar is safe, which is
// where most YAML emitter bugs live:
//
//   - single-line strings are always double-quoted, so values such as "no",
//     "1.0" or "a: b" stay strings in every YAML parser
//   - multi-line strings use a literal block scalar ("|") when the text can
//     be represented that way exactly, and a double-quoted scalar otherwise
//   - invalid UTF-8 is replaced with U+FFFD
//
// The output is valid YAML 1.1 and YAML 1.2.
package yaml
