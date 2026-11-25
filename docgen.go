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

// Package docgen generates reference documentation for a [naja.Command]
// tree: man pages, Markdown, reStructuredText and YAML.
//
// It is meant to be used at development or release time, from a small
// program or a go:generate directive that imports your root command, rather
// than from the actual CLI you ship:
//
//	//go:generate go run ./internal/gendocs
//
//	func main() {
//		if err := docgen.GenManTree(cmd.Root(), nil, "./man"); err != nil {
//			log.Fatal(err)
//		}
//	}
//
// Each format has a function for a single command (GenMan, GenMarkdown,
// GenReST, GenYaml) and one for a command and all of its descendants
// (GenManTree, GenMarkdownTree, GenReSTTree, GenYamlTree).
//
// # Text handling
//
// Text taken from commands and flags is treated as plain text. In man pages
// it is escaped and never interpreted as Markdown or roff: blank lines
// separate paragraphs, and examples keep their line breaks and indentation.
//
// docgen depends only on naja, gpflag and the standard library.
package docgen
