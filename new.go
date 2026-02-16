package main

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/spf13/afero"
)

type NewCmd struct {
	Title string `arg:"" optional:"" help:"Presentation title (e.g. 'My Talk')."`
}

func (c *NewCmd) Run() error {
	title := c.Title
	if title == "" {
		scanner := bufio.NewScanner(os.Stdin)
		fmt.Fprint(stdout, "Presentation title (e.g. 'My Talk'): ")
		if !scanner.Scan() {
			return fmt.Errorf("no input provided")
		}
		title = strings.TrimSpace(scanner.Text())
		if title == "" {
			return fmt.Errorf("title cannot be empty")
		}
	}

	slug := slugify(title)
	if slug == "" {
		return fmt.Errorf("title produces an empty slug")
	}

	dir := slug

	if _, err := appFs.Stat(dir); err == nil {
		return fmt.Errorf("%s already exists", dir)
	}

	for _, sub := range []string{"public", "snippets"} {
		if err := appFs.MkdirAll(filepath.Join(dir, sub), 0o755); err != nil {
			return fmt.Errorf("creating directories: %w", err)
		}
	}

	pkgJSON := fmt.Sprintf(packageJSONTemplate, slug)
	if err := afero.WriteFile(appFs, filepath.Join(dir, "package.json"), []byte(pkgJSON), 0o644); err != nil {
		return fmt.Errorf("writing package.json: %w", err)
	}

	slidesMD := fmt.Sprintf(slidesMDTemplate, title, title)
	if err := afero.WriteFile(appFs, filepath.Join(dir, "slides.md"), []byte(slidesMD), 0o644); err != nil {
		return fmt.Errorf("writing slides.md: %w", err)
	}

	fmt.Fprintln(stdout, "Running pnpm install...")
	cmd := exec.Command("pnpm", "install")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("running pnpm install: %w", err)
	}

	fmt.Fprintln(stdout)
	fmt.Fprintf(stdout, "Created %s\n", dir)
	fmt.Fprintf(stdout, "Run: make dev PRES=%s\n", slug)
	return nil
}

func slugify(title string) string {
	s := strings.ToLower(title)
	s = strings.ReplaceAll(s, " ", "-")
	var b strings.Builder
	for _, c := range s {
		if (c >= 'a' && c <= 'z') || (c >= '0' && c <= '9') || c == '-' {
			b.WriteRune(c)
		}
	}
	return b.String()
}

const packageJSONTemplate = `{
  "name": "presentation-%s",
  "type": "module",
  "private": true,
  "scripts": {
    "build": "slidev build",
    "dev": "slidev --open",
    "export": "slidev export"
  },
  "dependencies": {
    "@slidev/cli": "^52.11.5",
    "@slidev/theme-default": "latest",
    "@slidev/theme-seriph": "latest"
  },
  "devDependencies": {
    "playwright-chromium": "^1.58.1"
  }
}
`

const slidesMDTemplate = `---
theme: default
title: %s
class: text-center
---

# %s

`
