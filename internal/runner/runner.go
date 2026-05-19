// SPDX-License-Identifier: MIT
// Copyright 2026 Grégoire Favreau

package runner

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/GregoireF/addlicense/internal/config"
	"github.com/GregoireF/addlicense/internal/header"
	"github.com/GregoireF/addlicense/internal/injector"
	"github.com/GregoireF/addlicense/internal/scanner"
)

// Run executes the add or check operation based on opts.
func Run(opts config.Options) error {
	if err := config.Load(&opts); err != nil {
		return err
	}
	opts.Normalize()

	// Load custom template or fall back to built-in.
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

	copyrightLine := fmt.Sprintf("Copyright %d", opts.Year)
	if opts.Reuse {
		copyrightLine = fmt.Sprintf("SPDX-FileCopyrightText: %d", opts.Year)
	}
	if opts.Author != "" {
		copyrightLine += " " + opts.Author
	}

	data := header.Data{
		Year:          opts.Year,
		Author:        opts.Author,
		License:       opts.License,
		SPDX:          header.SPDX(opts.License),
		CopyrightLine: copyrightLine,
	}

	var missing []string

	for _, f := range files {
		lang := header.LangFor(f.Ext)
		if lang == nil {
			continue
		}

		has, err := injector.HasHeader(f.Path)
		if err != nil {
			return fmt.Errorf("%s: %w", f.Path, err)
		}
		if has {
			continue
		}

		if opts.CheckOnly {
			missing = append(missing, f.Path)
			continue
		}

		rendered, err := header.Render(tmplText, data, *lang)
		if err != nil {
			return fmt.Errorf("%s: %w", f.Path, err)
		}

		if err := injector.Inject(f.Path, rendered); err != nil {
			return fmt.Errorf("%s: %w", f.Path, err)
		}

		rel, _ := filepath.Rel(".", f.Path)
		fmt.Printf("✓ %s\n", rel)
	}

	if len(missing) > 0 {
		for _, p := range missing {
			rel, _ := filepath.Rel(".", p)
			fmt.Fprintf(os.Stderr, "missing header: %s\n", rel)
		}
		return fmt.Errorf("%d file(s) missing license header", len(missing))
	}

	return nil
}
