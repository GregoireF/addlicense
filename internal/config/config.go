package config

import (
	"fmt"
	"os"
	"time"

	"gopkg.in/yaml.v3"
)

var DefaultIgnore = []string{
	"vendor", "node_modules", ".git", "dist", "build", "*.pb.go", "*.gen.go",
}

type Options struct {
	License   string
	Author    string
	Template  string
	Ignore    []string
	Year      int
	Paths     []string
	CheckOnly bool
}

// fileConfig is the shape of .addlicenserc.yaml / .addlicenserc.json.
type fileConfig struct {
	License string   `yaml:"license"`
	Author  string   `yaml:"author"`
	Ignore  []string `yaml:"ignore"`
	Year    int      `yaml:"year"`
}

// Load merges file config into opts, but CLI flags already set on opts take precedence.
func Load(opts *Options) error {
	fc, err := findAndParse()
	if err != nil {
		return err
	}
	if fc == nil {
		return nil
	}

	if opts.License == "" && fc.License != "" {
		opts.License = fc.License
	}
	if opts.Author == "" && fc.Author != "" {
		opts.Author = fc.Author
	}
	if len(opts.Ignore) == 0 && len(fc.Ignore) > 0 {
		opts.Ignore = fc.Ignore
	}
	if opts.Year == 0 && fc.Year != 0 {
		opts.Year = fc.Year
	}

	return nil
}

func (o *Options) Normalize() {
	if o.Year == 0 {
		o.Year = time.Now().Year()
	}
	if o.License == "" {
		o.License = "MIT"
	}
	if len(o.Ignore) == 0 {
		o.Ignore = DefaultIgnore
	}
}

var candidates = []string{
	".addlicenserc.yaml",
	".addlicenserc.yml",
	".addlicenserc.json",
	"addlicense.json",
}

func findAndParse() (*fileConfig, error) {
	for _, name := range candidates {
		data, err := os.ReadFile(name)
		if os.IsNotExist(err) {
			continue
		}
		if err != nil {
			return nil, fmt.Errorf("reading %s: %w", name, err)
		}
		var fc fileConfig
		if err := yaml.Unmarshal(data, &fc); err != nil {
			return nil, fmt.Errorf("parsing %s: %w", name, err)
		}
		return &fc, nil
	}
	return nil, nil
}
