// SPDX-License-Identifier: MIT
// Copyright 2026 Grégoire Favreau

package runner_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/GregoireF/addlicense/internal/config"
	"github.com/GregoireF/addlicense/internal/runner"
)

func makeProject(t *testing.T, files map[string]string) string {
	t.Helper()
	root := t.TempDir()
	for rel, content := range files {
		path := filepath.Join(root, rel)
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	return root
}

func readFile(t *testing.T, path string) string {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	return string(data)
}

func TestRun_AddsMITHeader(t *testing.T) {
	root := makeProject(t, map[string]string{
		"main.go":    "package main\n",
		"app.ts":     "export default {}\n",
		"script.sh":  "#!/bin/bash\necho hello\n",
		"config.yml": "key: value\n",
	})

	opts := config.Options{
		License: "MIT",
		Author:  "Grégoire",
		Year:    2026,
		Ignore:  config.DefaultIgnore,
		Paths:   []string{root},
	}

	if err := runner.Run(opts); err != nil {
		t.Fatalf("Run: %v", err)
	}

	cases := map[string][]string{
		"main.go":    {"// SPDX-License-Identifier: MIT", "// Copyright 2026 Grégoire"},
		"app.ts":     {"// SPDX-License-Identifier: MIT"},
		"script.sh":  {"# SPDX-License-Identifier: MIT"},
		"config.yml": {"# SPDX-License-Identifier: MIT"},
	}

	for rel, want := range cases {
		content := readFile(t, filepath.Join(root, rel))
		for _, line := range want {
			if !strings.Contains(content, line) {
				t.Errorf("%s: missing %q\ngot:\n%s", rel, line, content)
			}
		}
	}
}

func TestRun_IdempotentOnSecondCall(t *testing.T) {
	root := makeProject(t, map[string]string{
		"main.go": "package main\n",
	})

	opts := config.Options{
		License: "MIT",
		Author:  "Grégoire",
		Year:    2026,
		Ignore:  config.DefaultIgnore,
		Paths:   []string{root},
	}

	if err := runner.Run(opts); err != nil {
		t.Fatal(err)
	}
	first := readFile(t, filepath.Join(root, "main.go"))

	if err := runner.Run(opts); err != nil {
		t.Fatal(err)
	}
	second := readFile(t, filepath.Join(root, "main.go"))

	if first != second {
		t.Errorf("second run modified the file:\nbefore:\n%s\nafter:\n%s", first, second)
	}
}

func TestRun_PreservesShebang(t *testing.T) {
	root := makeProject(t, map[string]string{
		"deploy.sh": "#!/usr/bin/env bash\nset -euo pipefail\n",
	})

	opts := config.Options{
		License: "Apache-2.0",
		Author:  "Grégoire",
		Year:    2026,
		Ignore:  config.DefaultIgnore,
		Paths:   []string{root},
	}

	if err := runner.Run(opts); err != nil {
		t.Fatal(err)
	}

	content := readFile(t, filepath.Join(root, "deploy.sh"))
	if !strings.HasPrefix(content, "#!/usr/bin/env bash") {
		t.Errorf("shebang not on first line:\n%s", content)
	}
	if !strings.Contains(content, "SPDX-License-Identifier: Apache-2.0") {
		t.Errorf("missing SPDX header:\n%s", content)
	}
}

func TestRun_CheckMode_DetectsOK(t *testing.T) {
	root := makeProject(t, map[string]string{
		"main.go": "// SPDX-License-Identifier: MIT\n// Copyright 2026 Grégoire\npackage main\n",
	})

	opts := config.Options{
		License:   "MIT",
		Ignore:    config.DefaultIgnore,
		Paths:     []string{root},
		CheckOnly: true,
	}

	if err := runner.Run(opts); err != nil {
		t.Errorf("check mode on licensed file should pass, got: %v", err)
	}
}

func TestRun_CheckMode_FailsOnMissing(t *testing.T) {
	root := makeProject(t, map[string]string{
		"main.go": "package main\n",
	})

	opts := config.Options{
		License:   "MIT",
		Ignore:    config.DefaultIgnore,
		Paths:     []string{root},
		CheckOnly: true,
	}

	if err := runner.Run(opts); err == nil {
		t.Error("check mode should fail when headers are missing")
	}
}

func TestRun_IgnoresVendorAndNodeModules(t *testing.T) {
	root := makeProject(t, map[string]string{
		"main.go":              "package main\n",
		"vendor/lib.go":        "package lib\n",
		"node_modules/pkg.js":  "module.exports = {}\n",
	})

	opts := config.Options{
		License: "MIT",
		Author:  "Grégoire",
		Year:    2026,
		Ignore:  config.DefaultIgnore,
		Paths:   []string{root},
	}

	if err := runner.Run(opts); err != nil {
		t.Fatal(err)
	}

	vendorContent := readFile(t, filepath.Join(root, "vendor", "lib.go"))
	if strings.Contains(vendorContent, "SPDX-License-Identifier") {
		t.Error("vendor/lib.go should not have been modified")
	}
}

func TestRun_CustomTemplate(t *testing.T) {
	root := makeProject(t, map[string]string{
		"main.go": "package main\n",
	})

	tmplPath := filepath.Join(t.TempDir(), "header.txt")
	if err := os.WriteFile(tmplPath, []byte("Copyright {{.Year}} {{.Author}} — All rights reserved."), 0o644); err != nil {
		t.Fatal(err)
	}

	opts := config.Options{
		License:  "MIT",
		Author:   "ACME Corp",
		Year:     2026,
		Template: tmplPath,
		Ignore:   config.DefaultIgnore,
		Paths:    []string{root},
	}

	if err := runner.Run(opts); err != nil {
		t.Fatal(err)
	}

	content := readFile(t, filepath.Join(root, "main.go"))
	if !strings.Contains(content, "Copyright 2026 ACME Corp — All rights reserved.") {
		t.Errorf("custom template not applied:\n%s", content)
	}
}

func TestRun_ConfigFileOverride(t *testing.T) {
	root := makeProject(t, map[string]string{
		"main.go":            "package main\n",
		".addlicenserc.yaml": "license: Apache-2.0\nauthor: ConfigAuthor\nyear: 2025\n",
	})

	// opts with no license set — should pick up from config file
	opts := config.Options{
		Ignore: config.DefaultIgnore,
		Paths:  []string{root},
	}

	if err := runner.Run(opts); err != nil {
		t.Fatal(err)
	}

	content := readFile(t, filepath.Join(root, "main.go"))
	if !strings.Contains(content, "SPDX-License-Identifier: Apache-2.0") {
		t.Errorf("expected Apache-2.0 from config, got:\n%s", content)
	}
	if !strings.Contains(content, "ConfigAuthor") {
		t.Errorf("expected ConfigAuthor from config, got:\n%s", content)
	}
}
