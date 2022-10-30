// Taken from caddy source code (https://github.com/mholt/caddy/)
// Copyright 2015 Matthew Holt and The Caddy Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//	http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
//
//nolint:all
package gitea

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strings"
	"unicode"

	"github.com/BurntSushi/toml"
	"github.com/alecthomas/chroma/formatters/html"
	"github.com/yuin/goldmark"
	highlighting "github.com/yuin/goldmark-highlighting"
	"github.com/yuin/goldmark/extension"
	"github.com/yuin/goldmark/parser"
	gmhtml "github.com/yuin/goldmark/renderer/html"
	"gopkg.in/yaml.v3"
)

func extractFrontMatter(input string) (map[string]any, string, error) {
	// get the bounds of the first non-empty line
	var firstLineStart, firstLineEnd int
	lineEmpty := true
	for i, b := range input {
		if b == '\n' {
			firstLineStart = firstLineEnd
			if firstLineStart > 0 {
				firstLineStart++ // skip newline character
			}
			firstLineEnd = i
			if !lineEmpty {
				break
			}
			continue
		}
		lineEmpty = lineEmpty && unicode.IsSpace(b)
	}
	firstLine := input[firstLineStart:firstLineEnd]

	// ensure residue windows carriage return byte is removed
	firstLine = strings.TrimSpace(firstLine)

	// see what kind of front matter there is, if any
	var closingFence []string
	var fmParser func([]byte) (map[string]any, error)
	for _, fmType := range supportedFrontMatterTypes {
		if firstLine == fmType.FenceOpen {
			closingFence = fmType.FenceClose
			fmParser = fmType.ParseFunc
		}
	}

	if fmParser == nil {
		// no recognized front matter; whole document is body
		return nil, input, nil
	}

	// find end of front matter
	var fmEndFence string
	fmEndFenceStart := -1
	for _, fence := range closingFence {
		index := strings.Index(input[firstLineEnd:], "\n"+fence)
		if index >= 0 {
			fmEndFenceStart = index
			fmEndFence = fence
			break
		}
	}
	if fmEndFenceStart < 0 {
		return nil, "", fmt.Errorf("unterminated front matter")
	}
	fmEndFenceStart += firstLineEnd + 1 // add 1 to account for newline

	// extract and parse front matter
	frontMatter := input[firstLineEnd:fmEndFenceStart]
	fm, err := fmParser([]byte(frontMatter))
	if err != nil {
		return nil, "", err
	}

	// the rest is the body
	body := input[fmEndFenceStart+len(fmEndFence):]

	return fm, body, nil
}

func yamlFrontMatter(input []byte) (map[string]any, error) {
	m := make(map[string]any)
	err := yaml.Unmarshal(input, &m)
	return m, err
}

func tomlFrontMatter(input []byte) (map[string]any, error) {
	m := make(map[string]any)
	err := toml.Unmarshal(input, &m)
	return m, err
}

func jsonFrontMatter(input []byte) (map[string]any, error) {
	input = append([]byte{'{'}, input...)
	input = append(input, '}')
	m := make(map[string]any)
	err := json.Unmarshal(input, &m)
	return m, err
}

type parsedMarkdownDoc struct {
	Meta map[string]any `json:"meta,omitempty"`
	Body string         `json:"body,omitempty"`
}

type frontMatterType struct {
	FenceOpen  string
	FenceClose []string
	ParseFunc  func(input []byte) (map[string]any, error)
}

var supportedFrontMatterTypes = []frontMatterType{
	{
		FenceOpen:  "---",
		FenceClose: []string{"---", "..."},
		ParseFunc:  yamlFrontMatter,
	},
	{
		FenceOpen:  "+++",
		FenceClose: []string{"+++"},
		ParseFunc:  tomlFrontMatter,
	},
	{
		FenceOpen:  "{",
		FenceClose: []string{"}"},
		ParseFunc:  jsonFrontMatter,
	},
}

func markdown(input []byte) ([]byte, error) {
	md := goldmark.New(
		goldmark.WithExtensions(
			extension.GFM,
			extension.Footnote,
			highlighting.NewHighlighting(
				highlighting.WithFormatOptions(
					html.WithClasses(true),
				),
			),
		),
		goldmark.WithParserOptions(
			parser.WithAutoHeadingID(),
		),
		goldmark.WithRendererOptions(
			gmhtml.WithUnsafe(), // TODO: this is not awesome, maybe should be configurable?
		),
	)

	buf := bufPool.Get().(*bytes.Buffer)
	buf.Reset()

	defer bufPool.Put(buf)

	if err := md.Convert(input, buf); err != nil {
		return input, err
	}

	return buf.Bytes(), nil
}
