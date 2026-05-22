// SPDX-License-Identifier: MIT
// Copyright 2026 Grégoire Favreau

package injector_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/GregoireF/addlicense/internal/injector"
)

const benchHeader = "// SPDX-License-Identifier: MIT\n// Copyright 2026 Acme Corp\n"

func BenchmarkHasHeader_WithHeader(b *testing.B) {
	path := benchFile(b, benchHeader+"package main\n\nfunc main() {}\n")
	b.ResetTimer()
	for range b.N {
		_, _ = injector.HasHeader(path)
	}
}

func BenchmarkHasHeader_WithoutHeader(b *testing.B) {
	path := benchFile(b, "package main\n\nfunc main() {}\n")
	b.ResetTimer()
	for range b.N {
		_, _ = injector.HasHeader(path)
	}
}

func BenchmarkInject(b *testing.B) {
	b.ResetTimer()
	for range b.N {
		path := benchFile(b, "package main\n\nfunc main() {}\n")
		_ = injector.Inject(path, benchHeader)
	}
}

func BenchmarkRemove_LineComment(b *testing.B) {
	b.ResetTimer()
	for range b.N {
		path := benchFile(b, benchHeader+"package main\n\nfunc main() {}\n")
		_, _ = injector.Remove(path, "//", "", "")
	}
}

func BenchmarkExtractYear(b *testing.B) {
	path := benchFile(b, benchHeader+"package main\n")
	b.ResetTimer()
	for range b.N {
		_ = injector.ExtractYear(path)
	}
}

func benchFile(b *testing.B, content string) string {
	b.Helper()
	f, err := os.CreateTemp(b.TempDir(), "bench*.go")
	if err != nil {
		b.Fatal(err)
	}
	if _, err := f.WriteString(content); err != nil {
		b.Fatal(err)
	}
	if err := f.Close(); err != nil {
		b.Fatal(err)
	}
	return f.Name()
}

func BenchmarkHasHeader_LargeFile(b *testing.B) {
	// Simulate a file where the copyright is absent and the scanner reads all 20 lines.
	var body string
	for i := range 25 {
		body += "// line number placeholder\n"
		_ = i
	}
	body += "package main\n"
	path := benchFile(b, body)
	b.ResetTimer()
	for range b.N {
		_, _ = injector.HasHeader(path)
	}
}

func BenchmarkInject_LargeFile(b *testing.B) {
	// 500-line file — measures WriteFile overhead on realistic input.
	var body string
	for range 500 {
		body += "// some code line here\n"
	}
	body += "package main\n"
	b.ResetTimer()
	for range b.N {
		path := filepath.Join(b.TempDir(), "large.go")
		_ = os.WriteFile(path, []byte(body), 0o644)
		_ = injector.Inject(path, benchHeader)
	}
}
