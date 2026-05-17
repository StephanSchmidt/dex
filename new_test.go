package main

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"strings"
	"testing"

	"github.com/StephanSchmidt/dex/revealjs"
	"github.com/StephanSchmidt/dex/slidev"
	"github.com/spf13/afero"
)

func TestSlugify(t *testing.T) {
	tests := []struct {
		name  string
		title string
		want  string
	}{
		{"simple", "Hello", "hello"},
		{"spaces", "My Talk", "my-talk"},
		{"mixed case", "FooBarBaz", "foobarbaz"},
		{"digits ok", "Talk 2024", "talk-2024"},
		{"punctuation stripped", "Hello, World!", "hello-world"},
		{"unicode stripped", "Café résumé", "caf-rsum"},
		{"only punctuation", "!!!", ""},
		{"leading/trailing spaces become dashes", " Hi ", "-hi-"},
		{"empty", "", ""},
		{"hyphens preserved", "already-slug", "already-slug"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := slugify(tt.title); got != tt.want {
				t.Errorf("slugify(%q) = %q, want %q", tt.title, got, tt.want)
			}
		})
	}
}

// withTestScaffoldCmd swaps out runScaffoldCmd for the duration of a test and
// records every invocation. The cleanup restores the original function.
func withTestScaffoldCmd(t *testing.T) *[][]string {
	t.Helper()
	var calls [][]string
	orig := runScaffoldCmd
	runScaffoldCmd = func(args []string) error {
		calls = append(calls, args)
		return nil
	}
	t.Cleanup(func() { runScaffoldCmd = orig })
	return &calls
}

// setupNewTest prepares an in-memory filesystem and captures stdout so
// NewCmd.Run can be exercised hermetically. It also restores activeFormat.
func setupNewTest(t *testing.T) *bytes.Buffer {
	t.Helper()
	appFs = afero.NewMemMapFs()
	buf := &bytes.Buffer{}
	stdout = buf
	origFormat := activeFormat
	t.Cleanup(func() { activeFormat = origFormat })
	return buf
}

func TestNewCmdRunSlidev(t *testing.T) {
	buf := setupNewTest(t)
	activeFormat = slidev.Format{}
	calls := withTestScaffoldCmd(t)

	cmd := NewCmd{Title: "My Talk"}
	if err := cmd.Run(); err != nil {
		t.Fatalf("NewCmd.Run() error: %v", err)
	}

	// Directory scaffolded under the slug.
	for _, want := range []string{"my-talk", "my-talk/public", "my-talk/snippets"} {
		info, err := appFs.Stat(want)
		if err != nil {
			t.Errorf("expected %s to exist: %v", want, err)
			continue
		}
		if !info.IsDir() {
			t.Errorf("%s should be a directory", want)
		}
	}

	// Files were written with templated content.
	pkg, err := afero.ReadFile(appFs, "my-talk/package.json")
	if err != nil {
		t.Fatalf("reading package.json: %v", err)
	}
	if !strings.Contains(string(pkg), "presentation-my-talk") {
		t.Errorf("package.json missing slug: %s", pkg)
	}

	slides, err := afero.ReadFile(appFs, "my-talk/slides.md")
	if err != nil {
		t.Fatalf("reading slides.md: %v", err)
	}
	if !strings.Contains(string(slides), "title: My Talk") {
		t.Errorf("slides.md missing title: %s", slides)
	}

	// Post-scaffold command ran exactly once with the expected args.
	if len(*calls) != 1 {
		t.Fatalf("expected 1 post-scaffold invocation, got %d", len(*calls))
	}
	if got := (*calls)[0]; len(got) != 2 || got[0] != "pnpm" || got[1] != "install" {
		t.Errorf("post-scaffold args: got %v, want [pnpm install]", got)
	}

	// User-facing instructions reference the new dir.
	out := buf.String()
	if !strings.Contains(out, "Created my-talk") {
		t.Errorf("missing creation summary: %q", out)
	}
	if !strings.Contains(out, "make dev PRES=my-talk") {
		t.Errorf("missing dev hint: %q", out)
	}
}

func TestNewCmdRunRevealjs(t *testing.T) {
	setupNewTest(t)
	activeFormat = revealjs.Format{}
	calls := withTestScaffoldCmd(t)

	cmd := NewCmd{Title: "Demo Deck"}
	if err := cmd.Run(); err != nil {
		t.Fatalf("NewCmd.Run() error: %v", err)
	}

	if _, err := appFs.Stat("demo-deck/index.html"); err != nil {
		t.Errorf("expected index.html: %v", err)
	}
	if _, err := appFs.Stat("demo-deck/images"); err != nil {
		t.Errorf("expected images dir: %v", err)
	}

	// reveal.js has no post-scaffold step.
	if len(*calls) != 0 {
		t.Errorf("expected no post-scaffold calls, got %v", *calls)
	}
}

func TestNewCmdRunReadsStdinWhenTitleMissing(t *testing.T) {
	setupNewTest(t)
	activeFormat = revealjs.Format{}
	withTestScaffoldCmd(t)

	// Replace os.Stdin with a pipe that emits a title.
	origStdin := os.Stdin
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("pipe: %v", err)
	}
	os.Stdin = r
	t.Cleanup(func() { os.Stdin = origStdin })

	go func() {
		_, _ = io.WriteString(w, "Piped Title\n")
		_ = w.Close()
	}()

	cmd := NewCmd{}
	if err := cmd.Run(); err != nil {
		t.Fatalf("NewCmd.Run() error: %v", err)
	}
	if _, err := appFs.Stat("piped-title/index.html"); err != nil {
		t.Errorf("expected piped-title dir from stdin input: %v", err)
	}
}

func TestNewCmdRunNoStdinInput(t *testing.T) {
	setupNewTest(t)
	activeFormat = revealjs.Format{}
	withTestScaffoldCmd(t)

	origStdin := os.Stdin
	r, w, _ := os.Pipe()
	os.Stdin = r
	t.Cleanup(func() { os.Stdin = origStdin })
	// Close immediately so Scanner.Scan() sees EOF with no data.
	_ = w.Close()

	if err := (&NewCmd{}).Run(); err == nil || !strings.Contains(err.Error(), "no input provided") {
		t.Errorf("expected no-input error, got %v", err)
	}
}

func TestNewCmdRunEmptyStdinTitle(t *testing.T) {
	setupNewTest(t)
	activeFormat = revealjs.Format{}
	withTestScaffoldCmd(t)

	origStdin := os.Stdin
	r, w, _ := os.Pipe()
	os.Stdin = r
	t.Cleanup(func() { os.Stdin = origStdin })
	go func() {
		_, _ = io.WriteString(w, "   \n")
		_ = w.Close()
	}()

	if err := (&NewCmd{}).Run(); err == nil || !strings.Contains(err.Error(), "title cannot be empty") {
		t.Errorf("expected empty-title error, got %v", err)
	}
}

func TestNewCmdRunEmptySlug(t *testing.T) {
	setupNewTest(t)
	activeFormat = revealjs.Format{}
	withTestScaffoldCmd(t)

	if err := (&NewCmd{Title: "!!!"}).Run(); err == nil || !strings.Contains(err.Error(), "empty slug") {
		t.Errorf("expected empty-slug error, got %v", err)
	}
}

func TestNewCmdRunAlreadyExists(t *testing.T) {
	setupNewTest(t)
	activeFormat = revealjs.Format{}
	withTestScaffoldCmd(t)

	if err := appFs.MkdirAll("my-talk", 0o755); err != nil {
		t.Fatalf("setup: %v", err)
	}
	err := (&NewCmd{Title: "My Talk"}).Run()
	if err == nil || !strings.Contains(err.Error(), "already exists") {
		t.Errorf("expected already-exists error, got %v", err)
	}
}

func TestNewCmdRunPostScaffoldError(t *testing.T) {
	setupNewTest(t)
	activeFormat = slidev.Format{}
	orig := runScaffoldCmd
	runScaffoldCmd = func(_ []string) error { return fmt.Errorf("boom") }
	t.Cleanup(func() { runScaffoldCmd = orig })

	err := (&NewCmd{Title: "X"}).Run()
	if err == nil || !strings.Contains(err.Error(), "boom") {
		t.Errorf("expected propagated post-scaffold error, got %v", err)
	}
}
