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

var langs = map[string]Lang{
	".go":    {LineComment: "//"},
	".ts":    {LineComment: "//"},
	".tsx":   {LineComment: "//"},
	".js":    {LineComment: "//"},
	".jsx":   {LineComment: "//"},
	".java":  {BlockOpen: "/*", BlockClose: " */", BlockPrefix: " * "},
	".c":     {BlockOpen: "/*", BlockClose: " */", BlockPrefix: " * "},
	".cpp":   {BlockOpen: "/*", BlockClose: " */", BlockPrefix: " * "},
	".h":     {BlockOpen: "/*", BlockClose: " */", BlockPrefix: " * "},
	".rs":    {LineComment: "//"},
	".py":    {LineComment: "#"},
	".sh":    {LineComment: "#"},
	".bash":  {LineComment: "#"},
	".yaml":  {LineComment: "#"},
	".yml":   {LineComment: "#"},
	".tf":    {LineComment: "#"},
	".toml":  {LineComment: "#"},
	".rb":    {LineComment: "#"},
	".swift": {LineComment: "//"},
	".kt":    {LineComment: "//"},
	".scala": {LineComment: "//"},
	".php":   {LineComment: "//"},
	".cs":    {LineComment: "//"},
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
