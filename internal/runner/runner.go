// SPDX-License-Identifier: MIT
// Copyright 2026 Grégoire Favreau

package runner

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"sync"

	"github.com/GregoireF/addlicense/internal/config"
	"github.com/GregoireF/addlicense/internal/header"
	"github.com/GregoireF/addlicense/internal/injector"
	"github.com/GregoireF/addlicense/internal/scanner"
)

// fileResult holds the outcome of processing one file.
type fileResult struct {
	Path   string
	Action string // added | removed | updated | skipped | missing | ok | error
	// dry-run variants: would-add | would-remove | would-update
	Err error
}

// jsonRecord is the JSON Lines shape for --format json.
type jsonRecord struct {
	File   string `json:"file"`
	Status string `json:"status"`
	Error  string `json:"error,omitempty"`
}

// Run executes the add/check/remove/update operation based on opts.
func Run(opts config.Options) error {
	if err := validateOpts(&opts); err != nil {
		return err
	}
	if err := config.Load(&opts); err != nil {
		return err
	}
	opts.Normalize()

	tmplText := header.BuiltinTemplate(opts.License)
	if opts.Template != "" {
		data, err := os.ReadFile(opts.Template)
		if err != nil {
			return fmt.Errorf("reading template: %w", err)
		}
		tmplText = string(data)
	}

	files, err := scanner.Walk(opts.Paths, opts.Ignore)
	if err != nil {
		return err
	}

	hData := header.Data{
		Year:          opts.Year,
		Author:        opts.Author,
		License:       opts.License,
		SPDX:          header.SPDX(opts.License),
		CopyrightLine: buildCopyrightLine(opts.Year, opts.Author, opts.Reuse),
	}

	results := runParallel(files, opts, hData, tmplText)
	return handleResults(results, opts)
}

func validateOpts(opts *config.Options) error {
	if opts.Verbose && opts.Quiet {
		return fmt.Errorf("--verbose and --quiet are mutually exclusive")
	}
	if opts.CheckOnly && (opts.Remove || opts.Update) {
		return fmt.Errorf("--check cannot be combined with --remove or --update")
	}
	if opts.Remove && opts.Update {
		return fmt.Errorf("--remove and --update are mutually exclusive; --update implies removal")
	}
	if opts.Format != "" && opts.Format != "text" && opts.Format != "json" {
		return fmt.Errorf("--format must be \"text\" or \"json\", got %q", opts.Format)
	}
	return nil
}

func runParallel(files []scanner.File, opts config.Options, hData header.Data, tmplText string) <-chan fileResult {
	n := opts.Workers
	if n <= 0 {
		n = runtime.NumCPU()
	}
	if n > len(files) {
		n = len(files)
	}
	if n == 0 {
		n = 1
	}

	jobs := make(chan scanner.File, len(files))
	out := make(chan fileResult, len(files))

	var wg sync.WaitGroup
	for range n {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for f := range jobs {
				out <- processFile(f, opts, hData, tmplText)
			}
		}()
	}

	for _, f := range files {
		jobs <- f
	}
	close(jobs)

	go func() {
		wg.Wait()
		close(out)
	}()

	return out
}

func processFile(f scanner.File, opts config.Options, hData header.Data, tmplText string) fileResult {
	rel, _ := filepath.Rel(".", f.Path)
	lang := header.LangFor(f.Ext)
	if lang == nil {
		return fileResult{Path: rel, Action: "skipped"}
	}
	switch {
	case opts.CheckOnly:
		return checkFile(rel, f.Path)
	case opts.Update:
		return updateFile(rel, f.Path, lang, hData, tmplText, opts.DryRun)
	case opts.Remove:
		return removeFile(rel, f.Path, lang, opts.DryRun)
	default:
		return addFile(rel, f.Path, lang, hData, tmplText, opts.DryRun)
	}
}

func checkFile(rel, path string) fileResult {
	has, err := injector.HasHeader(path)
	if err != nil {
		return fileResult{Path: rel, Action: "error", Err: err}
	}
	if has {
		return fileResult{Path: rel, Action: "ok"}
	}
	return fileResult{Path: rel, Action: "missing"}
}

func removeFile(rel, path string, lang *header.Lang, dryRun bool) fileResult {
	if dryRun {
		has, err := injector.HasHeader(path)
		if err != nil {
			return fileResult{Path: rel, Action: "error", Err: err}
		}
		if !has {
			return fileResult{Path: rel, Action: "skipped"}
		}
		return fileResult{Path: rel, Action: "would-remove"}
	}
	changed, err := injector.Remove(path, lang.LineComment, lang.BlockOpen, lang.BlockClose)
	if err != nil {
		return fileResult{Path: rel, Action: "error", Err: err}
	}
	if !changed {
		return fileResult{Path: rel, Action: "skipped"}
	}
	return fileResult{Path: rel, Action: "removed"}
}

func updateFile(rel, path string, lang *header.Lang, hData header.Data, tmplText string, dryRun bool) fileResult {
	if dryRun {
		has, err := injector.HasHeader(path)
		if err != nil {
			return fileResult{Path: rel, Action: "error", Err: err}
		}
		if has {
			return fileResult{Path: rel, Action: "would-update"}
		}
		return fileResult{Path: rel, Action: "would-add"}
	}
	changed, err := injector.Remove(path, lang.LineComment, lang.BlockOpen, lang.BlockClose)
	if err != nil {
		return fileResult{Path: rel, Action: "error", Err: err}
	}
	rendered, err := header.Render(tmplText, hData, *lang)
	if err != nil {
		return fileResult{Path: rel, Action: "error", Err: err}
	}
	if err := injector.Inject(path, rendered); err != nil {
		return fileResult{Path: rel, Action: "error", Err: err}
	}
	if changed {
		return fileResult{Path: rel, Action: "updated"}
	}
	return fileResult{Path: rel, Action: "added"}
}

func addFile(rel, path string, lang *header.Lang, hData header.Data, tmplText string, dryRun bool) fileResult {
	has, err := injector.HasHeader(path)
	if err != nil {
		return fileResult{Path: rel, Action: "error", Err: err}
	}
	if has {
		return fileResult{Path: rel, Action: "skipped"}
	}
	if dryRun {
		return fileResult{Path: rel, Action: "would-add"}
	}
	rendered, err := header.Render(tmplText, hData, *lang)
	if err != nil {
		return fileResult{Path: rel, Action: "error", Err: err}
	}
	if err := injector.Inject(path, rendered); err != nil {
		return fileResult{Path: rel, Action: "error", Err: err}
	}
	return fileResult{Path: rel, Action: "added"}
}

func handleResults(results <-chan fileResult, opts config.Options) error {
	useJSON := opts.Format == "json"
	var enc *json.Encoder
	if useJSON {
		enc = json.NewEncoder(os.Stdout)
	}

	var missing []string
	var firstErr error

	for r := range results {
		emit(r, useJSON, enc, opts.Verbose, opts.Quiet)
		switch r.Action {
		case "missing":
			missing = append(missing, r.Path)
		case "error":
			if firstErr == nil {
				firstErr = fmt.Errorf("%s: %w", r.Path, r.Err)
			}
		}
	}

	if firstErr != nil {
		return firstErr
	}
	if len(missing) > 0 {
		return fmt.Errorf("%d file(s) missing license header", len(missing))
	}
	return nil
}

func emit(r fileResult, useJSON bool, enc *json.Encoder, verbose, quiet bool) {
	if useJSON {
		rec := jsonRecord{File: r.Path, Status: r.Action}
		if r.Err != nil {
			rec.Error = r.Err.Error()
		}
		_ = enc.Encode(rec)
		return
	}
	switch r.Action {
	case "added", "removed", "updated":
		if !quiet {
			fmt.Printf("✓ %s\n", r.Path)
		}
	case "would-add", "would-remove", "would-update":
		if !quiet {
			fmt.Printf("[dry-run] %s: %s\n", r.Action, r.Path)
		}
	case "missing":
		if !quiet {
			fmt.Fprintf(os.Stderr, "missing header: %s\n", r.Path)
		}
	case "error":
		fmt.Fprintf(os.Stderr, "error: %s: %v\n", r.Path, r.Err)
	case "ok", "skipped":
		if verbose {
			fmt.Printf("  %s (%s)\n", r.Path, r.Action)
		}
	}
}

func buildCopyrightLine(year int, author string, reuse bool) string {
	prefix := "Copyright"
	if reuse {
		prefix = "SPDX-FileCopyrightText:"
	}
	line := fmt.Sprintf("%s %d", prefix, year)
	if author != "" {
		line += " " + author
	}
	return line
}
