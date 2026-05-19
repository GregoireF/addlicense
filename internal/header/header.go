// SPDX-License-Identifier: MIT
// Copyright 2026 Grégoire Favreau

package header

import (
	"bytes"
	"fmt"
	"strings"
	"text/template"
)

// Lang describes how to wrap a comment block for a given file type.
type Lang struct {
	LineComment  string // e.g. "//"
	BlockOpen    string // e.g. "/*"
	BlockClose   string // e.g. " */"
	BlockPrefix  string // e.g. " * "
}

var htmlComment = Lang{BlockOpen: "<!--", BlockClose: "-->", BlockPrefix: ""}
var cBlock = Lang{BlockOpen: "/*", BlockClose: " */", BlockPrefix: " * "}

var langs = map[string]Lang{
	// Go / systems
	".go":    {LineComment: "//"},
	".rs":    {LineComment: "//"},
	".c":     cBlock,
	".cpp":   cBlock,
	".h":     cBlock,
	// JVM
	".java":  cBlock,
	".kt":    {LineComment: "//"},
	".scala": {LineComment: "//"},
	// Web — scripts
	".ts":    {LineComment: "//"},
	".tsx":   {LineComment: "//"},
	".js":    {LineComment: "//"},
	".jsx":   {LineComment: "//"},
	// Web — markup / styles
	".html":   htmlComment,
	".vue":    htmlComment,
	".svelte": htmlComment,
	".css":    cBlock,
	".scss":   cBlock,
	// Mobile / cloud
	".swift": {LineComment: "//"},
	".php":   {LineComment: "//"},
	".cs":    {LineComment: "//"},
	".proto": {LineComment: "//"},
	// Infra / config
	".py":   {LineComment: "#"},
	".sh":   {LineComment: "#"},
	".bash": {LineComment: "#"},
	".yaml": {LineComment: "#"},
	".yml":  {LineComment: "#"},
	".tf":   {LineComment: "#"},
	".toml": {LineComment: "#"},
	".rb":   {LineComment: "#"},
	// Data
	".sql": {LineComment: "--"},
}

// LangFor returns the comment style for ext (e.g. ".go"), or nil if unsupported.
func LangFor(ext string) *Lang {
	l, ok := langs[strings.ToLower(ext)]
	if !ok {
		return nil
	}
	return &l
}

// Data is passed to header templates.
type Data struct {
	Year    int
	Author  string
	License string
	SPDX    string
}

// Render builds the comment block for the given template text and data.
func Render(tmplText string, data Data, lang Lang) (string, error) {
	t, err := template.New("header").Parse(tmplText)
	if err != nil {
		return "", fmt.Errorf("parsing template: %w", err)
	}

	var body bytes.Buffer
	if err := t.Execute(&body, data); err != nil {
		return "", fmt.Errorf("executing template: %w", err)
	}

	return wrapComment(body.String(), lang), nil
}

func wrapComment(text string, lang Lang) string {
	lines := strings.Split(strings.TrimRight(text, "\n"), "\n")
	var b strings.Builder

	if lang.LineComment != "" {
		for _, l := range lines {
			if l == "" {
				fmt.Fprintf(&b, "%s\n", lang.LineComment)
			} else {
				fmt.Fprintf(&b, "%s %s\n", lang.LineComment, l)
			}
		}
	} else {
		fmt.Fprintf(&b, "%s\n", lang.BlockOpen)
		for _, l := range lines {
			if l == "" {
				fmt.Fprintf(&b, "%s\n", strings.TrimRight(lang.BlockPrefix, " "))
			} else {
				fmt.Fprintf(&b, "%s%s\n", lang.BlockPrefix, l)
			}
		}
		fmt.Fprintf(&b, "%s\n", lang.BlockClose)
	}

	return b.String()
}

// SPDX returns the canonical SPDX expression for a license identifier.
func SPDX(license string) string {
	return strings.ToUpper(license)
}

// BuiltinTemplate returns the default header template for a license identifier.
func BuiltinTemplate(license string) string {
	switch strings.ToUpper(license) {
	case "MIT":
		return "Copyright {{.Year}}{{if .Author}} {{.Author}}{{end}}\nSPDX-License-Identifier: MIT"
	case "APACHE-2.0", "APACHE":
		return "Copyright {{.Year}}{{if .Author}} {{.Author}}{{end}}\nSPDX-License-Identifier: Apache-2.0"
	case "GPL-3.0", "GPL-3.0-ONLY", "GPL":
		return "Copyright {{.Year}}{{if .Author}} {{.Author}}{{end}}\nSPDX-License-Identifier: GPL-3.0-only"
	case "AGPL-3.0", "AGPL-3.0-ONLY", "AGPL":
		return "Copyright {{.Year}}{{if .Author}} {{.Author}}{{end}}\nSPDX-License-Identifier: AGPL-3.0-only"
	case "LGPL-2.1", "LGPL-2.1-ONLY":
		return "Copyright {{.Year}}{{if .Author}} {{.Author}}{{end}}\nSPDX-License-Identifier: LGPL-2.1-only"
	case "LGPL-3.0", "LGPL-3.0-ONLY", "LGPL":
		return "Copyright {{.Year}}{{if .Author}} {{.Author}}{{end}}\nSPDX-License-Identifier: LGPL-3.0-only"
	case "EUPL-1.2", "EUPL":
		return "Copyright {{.Year}}{{if .Author}} {{.Author}}{{end}}\nSPDX-License-Identifier: EUPL-1.2"
	case "MPL-2.0", "MPL":
		return "Copyright {{.Year}}{{if .Author}} {{.Author}}{{end}}\nSPDX-License-Identifier: MPL-2.0"
	case "BSD-2-CLAUSE", "BSD2":
		return "Copyright {{.Year}}{{if .Author}} {{.Author}}{{end}}\nSPDX-License-Identifier: BSD-2-Clause"
	case "BSD-3-CLAUSE", "BSD3", "BSD":
		return "Copyright {{.Year}}{{if .Author}} {{.Author}}{{end}}\nSPDX-License-Identifier: BSD-3-Clause"
	default:
		return "Copyright {{.Year}}{{if .Author}} {{.Author}}{{end}}\nSPDX-License-Identifier: {{.SPDX}}"
	}
}
