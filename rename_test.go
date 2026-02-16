package main

import (
	"strings"
	"testing"

	"github.com/spf13/afero"
)

func TestRename(t *testing.T) {
	t.Run("replace existing title", func(t *testing.T) {
		setupTestFs(t, testFixture)
		cmd := RenameCmd{Name: "New Deck", File: "slides.md"}
		if err := cmd.Run(); err != nil {
			t.Fatalf("Run() error: %v", err)
		}
		data, _ := afero.ReadFile(appFs, "slides.md")
		d := activeFormat.Parse(data)
		if got := extractMetadataTitle(d.Metadata); got != "New Deck" {
			t.Errorf("got %q, want %q", got, "New Deck")
		}
		// Slides must be untouched.
		titles := getTitles(t)
		want := []string{"A", "B", "C", "D"}
		if !equal(titles, want) {
			t.Errorf("slides got %v, want %v", titles, want)
		}
	})

	t.Run("add title when missing", func(t *testing.T) {
		fixture := "---\ntheme: default\n---\n\n# A\n"
		setupTestFs(t, fixture)
		cmd := RenameCmd{Name: "Added", File: "slides.md"}
		if err := cmd.Run(); err != nil {
			t.Fatalf("Run() error: %v", err)
		}
		data, _ := afero.ReadFile(appFs, "slides.md")
		d := activeFormat.Parse(data)
		if got := extractMetadataTitle(d.Metadata); got != "Added" {
			t.Errorf("got %q, want %q", got, "Added")
		}
	})

	t.Run("directory path", func(t *testing.T) {
		setupTestFs(t, testFixture)
		_ = appFs.MkdirAll("mydir", 0o755)
		writeTestFile(t, "mydir/slides.md", testFixture)
		cmd := RenameCmd{Name: "Dir Deck", File: "mydir/"}
		if err := cmd.Run(); err != nil {
			t.Fatalf("Run() error: %v", err)
		}
		data, _ := afero.ReadFile(appFs, "mydir/slides.md")
		d := activeFormat.Parse(data)
		if got := extractMetadataTitle(d.Metadata); got != "Dir Deck" {
			t.Errorf("got %q, want %q", got, "Dir Deck")
		}
	})

	t.Run("other metadata preserved", func(t *testing.T) {
		setupTestFs(t, testFixture)
		cmd := RenameCmd{Name: "Renamed", File: "slides.md"}
		if err := cmd.Run(); err != nil {
			t.Fatalf("Run() error: %v", err)
		}
		data, _ := afero.ReadFile(appFs, "slides.md")
		d := activeFormat.Parse(data)
		if got := extractMetadataTitle(d.Metadata); got != "Renamed" {
			t.Errorf("title: got %q, want %q", got, "Renamed")
		}
		if got := extractMetadataField(d.Metadata, "theme"); got != "default" {
			t.Errorf("theme: got %q, want %q", got, "default")
		}
	})
}

// extractMetadataTitle returns the value of the title: line in metadata.
func extractMetadataTitle(fm string) string {
	return extractMetadataField(fm, "title")
}

// extractMetadataField returns the value of a top-level key: value line.
func extractMetadataField(fm, key string) string {
	prefix := key + ":"
	for _, line := range strings.Split(fm, "\n") {
		if strings.HasPrefix(line, prefix) {
			return strings.TrimSpace(strings.TrimPrefix(line, prefix))
		}
	}
	return ""
}
