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

// runScaffoldCmd executes the post-scaffold command. Overridable in tests.
var runScaffoldCmd = func(args []string) error {
	cmd := exec.Command(args[0], args[1:]...) // #nosec G204 -- args from trusted format
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
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

	for _, sf := range activeFormat.Scaffold(title, slug) {
		p := filepath.Join(dir, sf.Path)
		if sf.IsDir {
			if err := appFs.MkdirAll(p, 0o755); err != nil {
				return fmt.Errorf("creating %s: %w", p, err)
			}
		} else {
			// Ensure parent directory exists.
			if err := appFs.MkdirAll(filepath.Dir(p), 0o755); err != nil {
				return fmt.Errorf("creating %s: %w", filepath.Dir(p), err)
			}
			if err := afero.WriteFile(appFs, p, []byte(sf.Content), 0o644); err != nil {
				return fmt.Errorf("writing %s: %w", p, err)
			}
		}
	}

	if args := activeFormat.PostScaffoldCmd(); len(args) > 0 {
		fmt.Fprintf(stdout, "Running %s...\n", strings.Join(args, " "))
		if err := runScaffoldCmd(args); err != nil {
			return fmt.Errorf("running %s: %w", strings.Join(args, " "), err)
		}
	}

	fmt.Fprintln(stdout)
	fmt.Fprintf(stdout, "Created %s\n", dir)             // #nosec G705 -- CLI output, not web
	fmt.Fprintf(stdout, "Run: make dev PRES=%s\n", slug) // #nosec G705 -- CLI output, not web
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
