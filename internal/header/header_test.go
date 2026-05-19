// SPDX-License-Identifier: MIT
// Copyright 2026 Grégoire Favreau

package header_test

import (
	"fmt"
	"strings"
	"testing"

	"github.com/GregoireF/addlicense/internal/header"
)

func TestLangFor(t *testing.T) {
	cases := []struct {
		ext  string
		want bool
	}{
		{".go", true},
		{".ts", true},
		{".py", true},
		{".sh", true},
		{".yaml", true},
		{".tf", true},
		{".rs", true},
		// v0.2.0 additions
		{".proto", true},
		{".vue", true},
		{".svelte", true},
		{".html", true},
		{".css", true},
		{".scss", true},
		{".sql", true},
		// unsupported
		{".md", false},
		{".json", false},
	}
	for _, tc := range cases {
		t.Run(tc.ext, func(t *testing.T) {
			l := header.LangFor(tc.ext)
			if (l != nil) != tc.want {
				t.Errorf("LangFor(%q) = %v, want supported=%v", tc.ext, l, tc.want)
			}
		})
	}
}

func testData(year int, author string) header.Data {
	line := fmt.Sprintf("Copyright %d", year)
	if author != "" {
		line += " " + author
	}
	return header.Data{Year: year, Author: author, License: "MIT", SPDX: "MIT", CopyrightLine: line}
}

func TestRenderMITLineComment(t *testing.T) {
	lang := header.LangFor(".go")
	data := testData(2026, "Grégoire")

	got, err := header.Render(header.BuiltinTemplate("MIT"), data, *lang)
	if err != nil {
		t.Fatalf("Render: %v", err)
	}

	if !strings.Contains(got, "// Copyright 2026 Grégoire") {
		t.Errorf("missing copyright line, got:\n%s", got)
	}
	if !strings.Contains(got, "// SPDX-License-Identifier: MIT") {
		t.Errorf("missing SPDX line, got:\n%s", got)
	}
}

func TestRenderReuseLineComment(t *testing.T) {
	lang := header.LangFor(".go")
	data := header.Data{
		Year:          2026,
		Author:        "Grégoire",
		License:       "MIT",
		SPDX:          "MIT",
		CopyrightLine: "SPDX-FileCopyrightText: 2026 Grégoire",
	}

	got, err := header.Render(header.BuiltinTemplate("MIT"), data, *lang)
	if err != nil {
		t.Fatalf("Render: %v", err)
	}
	if !strings.Contains(got, "// SPDX-FileCopyrightText: 2026 Grégoire") {
		t.Errorf("missing REUSE copyright line, got:\n%s", got)
	}
	if !strings.Contains(got, "// SPDX-License-Identifier: MIT") {
		t.Errorf("missing SPDX line, got:\n%s", got)
	}
}

func TestRenderBlockComment(t *testing.T) {
	lang := header.LangFor(".java")
	data := testData(2026, "Grégoire")

	got, err := header.Render(header.BuiltinTemplate("MIT"), data, *lang)
	if err != nil {
		t.Fatalf("Render: %v", err)
	}

	if !strings.Contains(got, "/*") || !strings.Contains(got, "*/") {
		t.Errorf("missing block comment markers, got:\n%s", got)
	}
}

func TestRenderHTMLComment(t *testing.T) {
	lang := header.LangFor(".html")
	data := testData(2026, "Grégoire")

	got, err := header.Render(header.BuiltinTemplate("MIT"), data, *lang)
	if err != nil {
		t.Fatalf("Render: %v", err)
	}
	if !strings.Contains(got, "<!--") || !strings.Contains(got, "-->") {
		t.Errorf("missing HTML comment markers, got:\n%s", got)
	}
}

func TestRenderSQLComment(t *testing.T) {
	lang := header.LangFor(".sql")
	data := testData(2026, "Grégoire")

	got, err := header.Render(header.BuiltinTemplate("MIT"), data, *lang)
	if err != nil {
		t.Fatalf("Render: %v", err)
	}
	if !strings.Contains(got, "-- Copyright") {
		t.Errorf("missing SQL comment prefix, got:\n%s", got)
	}
}

func TestBuiltinTemplateLicenses(t *testing.T) {
	licenses := []string{"MIT", "Apache-2.0", "GPL-3.0", "AGPL-3.0", "LGPL-2.1", "LGPL-3.0", "EUPL-1.2", "MPL-2.0", "BSD-2-Clause", "BSD-3-Clause"}
	for _, l := range licenses {
		tmpl := header.BuiltinTemplate(l)
		if tmpl == "" {
			t.Errorf("BuiltinTemplate(%q) returned empty string", l)
		}
	}
}
