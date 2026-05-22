// SPDX-License-Identifier: MIT
// Copyright 2026 Grégoire Favreau

// Package addlicense provides a Go API for programmatic license header
// management. It exposes the same functionality as the addlicense CLI
// but as an importable library with a stable, versioned interface.
//
// Basic usage:
//
//	import addlicense "github.com/GregoireF/addlicense/pkg/addlicense"
//
//	// Add MIT headers to all supported files under the current directory.
//	err := addlicense.Run(addlicense.Options{
//	    License: "MIT",
//	    Author:  "Acme Corp",
//	    Paths:   []string{"."},
//	})
//
//	// Check that all files carry a license header (CI-friendly).
//	err = addlicense.Run(addlicense.Options{
//	    CheckOnly: true,
//	    Paths:     []string{"."},
//	})
package addlicense

import (
	"github.com/GregoireF/addlicense/internal/config"
	"github.com/GregoireF/addlicense/internal/runner"
)

// Output format constants for Options.Format.
const (
	FormatText = "text" // human-readable output (default)
	FormatJSON = "json" // JSON Lines: {"file":"…","status":"…","header":"…","error":"…"}
)

// DefaultIgnore is the default set of path patterns excluded when Options.Ignore is nil.
// Callers may extend it without modifying the original slice:
//
//	opts.Ignore = append(append([]string{}, addlicense.DefaultIgnore...), "testdata")
var DefaultIgnore = config.DefaultIgnore

// Options configures a single addlicense run.
// All fields have sensible zero values — Run applies the same defaults as the CLI
// (License defaults to "MIT", Year defaults to the current year, etc.).
type Options struct {
	// License is the SPDX identifier (e.g. "MIT", "Apache-2.0").
	// Defaults to "MIT" when empty.
	License string

	// Author is the copyright holder.
	// Use a comma-separated string for multiple holders:
	//   "Alice Dupont, Bob Martin"
	// emits one copyright line per author.
	Author string

	// Template is the path to a custom header template file (Go text/template syntax).
	// When set, it overrides the built-in template for License.
	Template string

	// Ignore is the list of glob patterns and path substrings to skip.
	// Defaults to DefaultIgnore when nil.
	Ignore []string

	// Year is the copyright year. Defaults to the current year when zero.
	Year int

	// Paths is the list of directories or files to process. At least one is required.
	Paths []string

	// CheckOnly reports which files are missing a license header without modifying
	// anything. Run returns a non-nil error if any file is missing a header.
	CheckOnly bool

	// Reuse emits "SPDX-FileCopyrightText:" instead of "Copyright" for
	// REUSE/FSFE compliance (https://reuse.software/).
	Reuse bool

	// Remove strips existing license headers from files.
	Remove bool

	// Update replaces existing headers with the current options in one pass
	// (equivalent to Remove followed by inject). The previous header is fully
	// replaced; to keep existing authors include them in Author.
	Update bool

	// DryRun previews changes without writing to disk.
	// Prints "[dry-run] <action>: <file>" for each affected file.
	DryRun bool

	// Diff emits JSON Lines with the rendered header for each file that would
	// change, without writing to disk. Run returns a non-nil error if any file
	// would be modified (useful for CI gating). Mutually exclusive with CheckOnly.
	Diff bool

	// YearRange preserves the original copyright year during Update, emitting a
	// "YYYY-YYYY" range (e.g. "Copyright 2023-2026 Author"). Ignored when the
	// original year equals Year or cannot be extracted.
	YearRange bool

	// Dep5 generates a REUSE-compliant .reuse/dep5 file for files that cannot
	// carry inline headers (images, fonts, binaries, and other unsupported types).
	Dep5 bool

	// Verbose prints every processed file, including already-licensed ones.
	// Mutually exclusive with Quiet.
	Verbose bool

	// Quiet suppresses all stdout. Errors still go to stderr.
	// Mutually exclusive with Verbose.
	Quiet bool

	// Format selects the output format: FormatText (default) or FormatJSON.
	Format string

	// Workers sets the number of parallel goroutines for file processing.
	// Zero (default) uses runtime.NumCPU().
	Workers int
}

// Run executes the license header operation described by opts.
//
// It applies the same validation, config-file loading (.addlicenserc.yaml /
// .addlicenserc.json), and defaults as the CLI binary.
//
// Run returns a non-nil error when:
//   - flag validation fails (mutually exclusive options)
//   - a config file exists but cannot be parsed
//   - any file produces an I/O error during processing
//   - CheckOnly: one or more files are missing a license header
//   - Diff: one or more files would be modified by the operation
func Run(opts Options) error {
	return runner.Run(toInternal(opts))
}

func toInternal(o Options) config.Options {
	return config.Options{
		License:   o.License,
		Author:    o.Author,
		Template:  o.Template,
		Ignore:    o.Ignore,
		Year:      o.Year,
		Paths:     o.Paths,
		CheckOnly: o.CheckOnly,
		Reuse:     o.Reuse,
		Remove:    o.Remove,
		Update:    o.Update,
		DryRun:    o.DryRun,
		Diff:      o.Diff,
		YearRange: o.YearRange,
		Dep5:      o.Dep5,
		Verbose:   o.Verbose,
		Quiet:     o.Quiet,
		Format:    o.Format,
		Workers:   o.Workers,
	}
}
