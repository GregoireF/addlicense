// SPDX-License-Identifier: MIT
// Copyright 2026 Grégoire Favreau

package runner

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"

	"github.com/GregoireF/addlicense/internal/config"
	"github.com/GregoireF/addlicense/internal/dep5"
	"github.com/GregoireF/addlicense/internal/header"
	"github.com/GregoireF/addlicense/internal/injector"
	"github.com/GregoireF/addlicense/internal/scanner"
)

// Action values for fileResult.
const (
	actionAdded       = "added"
	actionRemoved     = "removed"
	actionUpdated     = "updated"
	actionSkipped     = "skipped"
	actionMissing     = "missing"
	actionOK          = "ok"
	actionError       = "error"
	actionWouldAdd    = "would-add"
	actionWouldRemove = "would-remove"
	actionWouldUpdate = "would-update"
	actionDiffAdd     = "diff-add"
	actionDiffUpdate  = "diff-update"
	actionDiffRemove  = "diff-remove"
)

// fileResult holds the outcome of processing one file.
type fileResult struct {
	Path   string
	Action string
	Header string // populated by --diff: the rendered header that would be written
	Err    error
}

// jsonRecord is the JSON Lines shape for --format json and --diff.
type jsonRecord struct {
	File   string `json:"file"`
	Status string `json:"status"`
	Header string `json:"header,omitempty"`
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

	cwd, err := os.Getwd()
	if err != nil {
		cwd = "."
	}

	hData := header.Data{
		Year:          opts.Year,
		Author:        opts.Author,
		License:       opts.License,
		SPDX:          header.SPDX(opts.License),
		CopyrightLine: buildCopyrightLine(0, opts.Year, opts.Author, opts.Reuse),
	}

	if opts.Dep5 {
		if err := writeDep5(files, opts, cwd); err != nil {
			return err
		}
	}

	results := runParallel(files, opts, hData, tmplText, cwd)
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
	if opts.Diff && opts.CheckOnly {
		return fmt.Errorf("--diff and --check are mutually exclusive")
	}
	return nil
}

func runParallel(files []scanner.File, opts config.Options, hData header.Data, tmplText, cwd string) <-chan fileResult {
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
				out <- processFile(f, opts, hData, tmplText, cwd)
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

func processFile(f scanner.File, opts config.Options, hData header.Data, tmplText, cwd string) fileResult {
	rel, err := filepath.Rel(cwd, f.Path)
	if err != nil {
		rel = f.Path
	}
	lang := header.LangFor(f.Ext)
	if lang == nil {
		return fileResult{Path: rel, Action: actionSkipped}
	}
	switch {
	case opts.Diff:
		return diffFile(rel, f.Path, lang, hData, tmplText, opts.Update, opts.Remove)
	case opts.CheckOnly:
		return checkFile(rel, f.Path)
	case opts.Update:
		return updateFile(rel, f.Path, lang, hData, tmplText, opts.DryRun, opts.YearRange, opts.Reuse)
	case opts.Remove:
		return removeFile(rel, f.Path, lang, opts.DryRun)
	default:
		return addFile(rel, f.Path, lang, hData, tmplText, opts.DryRun)
	}
}

func checkFile(rel, path string) fileResult {
	has, err := injector.HasHeader(path)
	if err != nil {
		return fileResult{Path: rel, Action: actionError, Err: err}
	}
	if has {
		return fileResult{Path: rel, Action: actionOK}
	}
	return fileResult{Path: rel, Action: actionMissing}
}

func removeFile(rel, path string, lang *header.Lang, dryRun bool) fileResult {
	if dryRun {
		has, err := injector.HasHeader(path)
		if err != nil {
			return fileResult{Path: rel, Action: actionError, Err: err}
		}
		if !has {
			return fileResult{Path: rel, Action: actionSkipped}
		}
		return fileResult{Path: rel, Action: actionWouldRemove}
	}
	changed, err := injector.Remove(path, lang.LineComment, lang.BlockOpen, lang.BlockClose)
	if err != nil {
		return fileResult{Path: rel, Action: actionError, Err: err}
	}
	if !changed {
		return fileResult{Path: rel, Action: actionSkipped}
	}
	return fileResult{Path: rel, Action: actionRemoved}
}

func updateFile(rel, path string, lang *header.Lang, hData header.Data, tmplText string, dryRun, yearRange, reuse bool) fileResult {
	if dryRun {
		has, err := injector.HasHeader(path)
		if err != nil {
			return fileResult{Path: rel, Action: actionError, Err: err}
		}
		if has {
			return fileResult{Path: rel, Action: actionWouldUpdate}
		}
		return fileResult{Path: rel, Action: actionWouldAdd}
	}
	if yearRange {
		if from := injector.ExtractYear(path); from > 0 && from < hData.Year {
			hData.CopyrightLine = buildCopyrightLine(from, hData.Year, hData.Author, reuse)
		}
	}
	changed, err := injector.Remove(path, lang.LineComment, lang.BlockOpen, lang.BlockClose)
	if err != nil {
		return fileResult{Path: rel, Action: actionError, Err: err}
	}
	rendered, err := header.Render(tmplText, hData, *lang)
	if err != nil {
		return fileResult{Path: rel, Action: actionError, Err: err}
	}
	if err := injector.Inject(path, rendered); err != nil {
		return fileResult{Path: rel, Action: actionError, Err: err}
	}
	if changed {
		return fileResult{Path: rel, Action: actionUpdated}
	}
	return fileResult{Path: rel, Action: actionAdded}
}

func addFile(rel, path string, lang *header.Lang, hData header.Data, tmplText string, dryRun bool) fileResult {
	has, err := injector.HasHeader(path)
	if err != nil {
		return fileResult{Path: rel, Action: actionError, Err: err}
	}
	if has {
		return fileResult{Path: rel, Action: actionSkipped}
	}
	if dryRun {
		return fileResult{Path: rel, Action: actionWouldAdd}
	}
	rendered, err := header.Render(tmplText, hData, *lang)
	if err != nil {
		return fileResult{Path: rel, Action: actionError, Err: err}
	}
	if err := injector.Inject(path, rendered); err != nil {
		return fileResult{Path: rel, Action: actionError, Err: err}
	}
	return fileResult{Path: rel, Action: actionAdded}
}

// diffFile reports what would be written without modifying the file.
// It always emits JSON regardless of --format (diff is inherently structured).
func diffFile(rel, path string, lang *header.Lang, hData header.Data, tmplText string, update, remove bool) fileResult {
	has, err := injector.HasHeader(path)
	if err != nil {
		return fileResult{Path: rel, Action: actionError, Err: err}
	}

	switch {
	case remove:
		if !has {
			return fileResult{Path: rel, Action: actionSkipped}
		}
		return fileResult{Path: rel, Action: actionDiffRemove}

	case update:
		rendered, err := header.Render(tmplText, hData, *lang)
		if err != nil {
			return fileResult{Path: rel, Action: actionError, Err: err}
		}
		if has {
			return fileResult{Path: rel, Action: actionDiffUpdate, Header: rendered}
		}
		return fileResult{Path: rel, Action: actionDiffAdd, Header: rendered}

	default: // add mode
		if has {
			return fileResult{Path: rel, Action: actionSkipped}
		}
		rendered, err := header.Render(tmplText, hData, *lang)
		if err != nil {
			return fileResult{Path: rel, Action: actionError, Err: err}
		}
		return fileResult{Path: rel, Action: actionDiffAdd, Header: rendered}
	}
}

func handleResults(results <-chan fileResult, opts config.Options) error {
	useJSON := opts.Format == "json" || opts.Diff
	var enc *json.Encoder
	if useJSON {
		enc = json.NewEncoder(os.Stdout)
	}

	var missing []string
	var firstErr error
	var diffChanged int

	for r := range results {
		emit(r, useJSON, enc, opts.Verbose, opts.Quiet)
		switch r.Action {
		case actionMissing:
			missing = append(missing, r.Path)
		case actionError:
			if firstErr == nil {
				firstErr = fmt.Errorf("%s: %w", r.Path, r.Err)
			}
		case actionDiffAdd, actionDiffUpdate, actionDiffRemove:
			diffChanged++
		}
	}

	if firstErr != nil {
		return firstErr
	}
	if len(missing) > 0 {
		return fmt.Errorf("%d file(s) missing license header", len(missing))
	}
	if diffChanged > 0 {
		return fmt.Errorf("%d file(s) would be modified", diffChanged)
	}
	return nil
}

func emit(r fileResult, useJSON bool, enc *json.Encoder, verbose, quiet bool) {
	if useJSON {
		rec := jsonRecord{File: r.Path, Status: r.Action, Header: r.Header}
		if r.Err != nil {
			rec.Error = r.Err.Error()
		}
		_ = enc.Encode(rec)
		return
	}
	switch r.Action {
	case actionAdded, actionRemoved, actionUpdated:
		if !quiet {
			fmt.Printf("✓ %s\n", r.Path)
		}
	case actionWouldAdd, actionWouldRemove, actionWouldUpdate:
		if !quiet {
			fmt.Printf("[dry-run] %s: %s\n", r.Action, r.Path)
		}
	case actionMissing:
		if !quiet {
			fmt.Fprintf(os.Stderr, "missing header: %s\n", r.Path)
		}
	case actionError:
		fmt.Fprintf(os.Stderr, "error: %s: %v\n", r.Path, r.Err)
	case actionOK, actionSkipped:
		if verbose {
			fmt.Printf("  %s (%s)\n", r.Path, r.Action)
		}
	}
}

func writeDep5(files []scanner.File, opts config.Options, cwd string) error {
	// Root for dep5: first scan path when unambiguous, otherwise CWD.
	dep5Root := cwd
	if len(opts.Paths) == 1 {
		if abs, err := filepath.Abs(opts.Paths[0]); err == nil {
			dep5Root = abs
		}
	}

	unhandled := make([]string, 0, len(files))
	for _, f := range files {
		if header.LangFor(f.Ext) != nil {
			continue
		}
		rel, err := filepath.Rel(dep5Root, f.Path)
		if err != nil {
			rel = f.Path
		}
		unhandled = append(unhandled, filepath.ToSlash(rel))
	}

	if opts.DryRun {
		if !opts.Quiet {
			fmt.Printf("[dry-run] would-write: .reuse/dep5 (%d unhandled file(s))\n", len(unhandled))
		}
		return nil
	}

	content := dep5.Build(unhandled, opts.Year, opts.Author, opts.License)
	dir := filepath.Join(dep5Root, ".reuse")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("creating .reuse directory: %w", err)
	}
	return os.WriteFile(filepath.Join(dir, "dep5"), []byte(content), 0o644)
}

func buildCopyrightLine(fromYear, year int, author string, reuse bool) string {
	prefix := "Copyright"
	if reuse {
		prefix = "SPDX-FileCopyrightText:"
	}
	yearStr := fmt.Sprintf("%d", year)
	if fromYear > 0 && fromYear < year {
		yearStr = fmt.Sprintf("%d-%d", fromYear, year)
	}

	// Support comma-separated authors: "Alice, Bob" → two copyright lines.
	authors := splitAuthors(author)
	if len(authors) == 0 {
		return fmt.Sprintf("%s %s", prefix, yearStr)
	}
	lines := make([]string, len(authors))
	for i, a := range authors {
		lines[i] = fmt.Sprintf("%s %s %s", prefix, yearStr, a)
	}
	return strings.Join(lines, "\n")
}

func splitAuthors(author string) []string {
	if author == "" {
		return nil
	}
	parts := strings.Split(author, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		if t := strings.TrimSpace(p); t != "" {
			out = append(out, t)
		}
	}
	return out
}
