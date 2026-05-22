// SPDX-License-Identifier: MIT
// Copyright 2026 Grégoire Favreau

// Package bench measures end-to-end pipeline performance via runner.Run.
// Run with: go test -bench=. -benchtime=3s ./tests/bench/
package bench_test

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/GregoireF/addlicense/internal/config"
	"github.com/GregoireF/addlicense/internal/runner"
)

// makeFiles creates n unlicensed Go files in a temp directory.
func makeFiles(b *testing.B, n int) string {
	b.Helper()
	root := b.TempDir()
	for i := range n {
		path := filepath.Join(root, fmt.Sprintf("file_%04d.go", i))
		if err := os.WriteFile(path, []byte("package bench\n"), 0o644); err != nil {
			b.Fatal(err)
		}
	}
	return root
}

// makeLicensedFiles creates n already-licensed Go files (idempotence scenario).
func makeLicensedFiles(b *testing.B, n int) string {
	b.Helper()
	root := b.TempDir()
	header := "// SPDX-License-Identifier: MIT\n// Copyright 2026 Acme\n\n"
	for i := range n {
		path := filepath.Join(root, fmt.Sprintf("file_%04d.go", i))
		if err := os.WriteFile(path, []byte(header+"package bench\n"), 0o644); err != nil {
			b.Fatal(err)
		}
	}
	return root
}

func baseOpts(root string) config.Options {
	return config.Options{
		License: "MIT",
		Author:  "Acme Corp",
		Year:    2026,
		Quiet:   true,
		Ignore:  config.DefaultIgnore,
		Paths:   []string{root},
	}
}

// BenchmarkRun_Add_N measures adding headers to N fresh files.
func BenchmarkRun_Add_10(b *testing.B)   { benchAdd(b, 10) }
func BenchmarkRun_Add_100(b *testing.B)  { benchAdd(b, 100) }
func BenchmarkRun_Add_500(b *testing.B)  { benchAdd(b, 500) }
func BenchmarkRun_Add_1000(b *testing.B) { benchAdd(b, 1000) }

func benchAdd(b *testing.B, n int) {
	b.Helper()
	b.ResetTimer()
	for range b.N {
		b.StopTimer()
		root := makeFiles(b, n)
		opts := baseOpts(root)
		b.StartTimer()
		if err := runner.Run(opts); err != nil {
			b.Fatal(err)
		}
	}
	b.ReportMetric(float64(n), "files/op")
}

// BenchmarkRun_Check_N measures check-only mode on N already-licensed files.
func BenchmarkRun_Check_10(b *testing.B)   { benchCheck(b, 10) }
func BenchmarkRun_Check_100(b *testing.B)  { benchCheck(b, 100) }
func BenchmarkRun_Check_1000(b *testing.B) { benchCheck(b, 1000) }

func benchCheck(b *testing.B, n int) {
	b.Helper()
	root := makeLicensedFiles(b, n)
	opts := baseOpts(root)
	opts.CheckOnly = true
	b.ResetTimer()
	for range b.N {
		if err := runner.Run(opts); err != nil {
			b.Fatal(err)
		}
	}
	b.ReportMetric(float64(n), "files/op")
}

// BenchmarkRun_Idempotent_N measures add mode when all files already have headers.
func BenchmarkRun_Idempotent_100(b *testing.B)  { benchIdempotent(b, 100) }
func BenchmarkRun_Idempotent_1000(b *testing.B) { benchIdempotent(b, 1000) }

func benchIdempotent(b *testing.B, n int) {
	b.Helper()
	root := makeLicensedFiles(b, n)
	opts := baseOpts(root)
	b.ResetTimer()
	for range b.N {
		if err := runner.Run(opts); err != nil {
			b.Fatal(err)
		}
	}
	b.ReportMetric(float64(n), "files/op")
}

// BenchmarkRun_Workers_* shows scaling with different worker counts.
func BenchmarkRun_Workers_1(b *testing.B)      { benchWorkers(b, 1) }
func BenchmarkRun_Workers_4(b *testing.B)      { benchWorkers(b, 4) }
func BenchmarkRun_Workers_NumCPU(b *testing.B) { benchWorkers(b, 0) }

func benchWorkers(b *testing.B, workers int) {
	b.Helper()
	b.ResetTimer()
	for range b.N {
		b.StopTimer()
		root := makeFiles(b, 200)
		opts := baseOpts(root)
		opts.Workers = workers
		b.StartTimer()
		if err := runner.Run(opts); err != nil {
			b.Fatal(err)
		}
	}
	b.ReportMetric(200, "files/op")
}
