package main

import (
	"testing"

	"github.com/spf13/afero"
)

func TestDelete(t *testing.T) {
	t.Run("delete single middle slide", func(t *testing.T) {
		setupTestFs(t, testFixture)
		cmd := DeleteCmd{Expr: "2"}
		if err := cmd.Run(); err != nil {
			t.Fatalf("Run() error: %v", err)
		}
		titles := getTitles(t)
		want := []string{"A", "C", "D"}
		if !equal(titles, want) {
			t.Errorf("got %v, want %v", titles, want)
		}
	})

	t.Run("delete multiple via comma", func(t *testing.T) {
		setupTestFs(t, testFixture)
		cmd := DeleteCmd{Expr: "1,3"}
		if err := cmd.Run(); err != nil {
			t.Fatalf("Run() error: %v", err)
		}
		titles := getTitles(t)
		want := []string{"B", "D"}
		if !equal(titles, want) {
			t.Errorf("got %v, want %v", titles, want)
		}
	})

	t.Run("delete range", func(t *testing.T) {
		setupTestFs(t, testFixture)
		cmd := DeleteCmd{Expr: "2-4"}
		if err := cmd.Run(); err != nil {
			t.Fatalf("Run() error: %v", err)
		}
		titles := getTitles(t)
		want := []string{"A"}
		if !equal(titles, want) {
			t.Errorf("got %v, want %v", titles, want)
		}
	})

	t.Run("preserves frontmatter on remaining slides", func(t *testing.T) {
		setupTestFs(t, testFixture)
		// Slide 3 has frontmatter (layout: center); delete slide 2 and check slide 3 keeps it.
		cmd := DeleteCmd{Expr: "2"}
		if err := cmd.Run(); err != nil {
			t.Fatalf("Run() error: %v", err)
		}
		data, _ := afero.ReadFile(appFs, "slides.md")
		d := activeFormat.Parse(data)
		// After deleting B (index 1), C moves to index 1.
		if d.Slides[1].Frontmatter == "" {
			t.Error("expected frontmatter on slide C, got empty")
		}
	})

	t.Run("out of range", func(t *testing.T) {
		setupTestFs(t, testFixture)
		if err := (&DeleteCmd{Expr: "0"}).Run(); err == nil {
			t.Error("expected error for slide 0")
		}
		if err := (&DeleteCmd{Expr: "5"}).Run(); err == nil {
			t.Error("expected error for slide 5 (out of range)")
		}
	})

	t.Run("delete all slides returns error", func(t *testing.T) {
		setupTestFs(t, testFixture)
		cmd := DeleteCmd{Expr: "1-4"}
		if err := cmd.Run(); err == nil {
			t.Error("expected error when deleting all slides")
		}
	})

	t.Run("directory path", func(t *testing.T) {
		setupTestFs(t, testFixture)
		_ = appFs.MkdirAll("mydir", 0o755)
		writeTestFile(t, "mydir/slides.md", testFixture)
		cmd := DeleteCmd{Expr: "mydir/2"}
		if err := cmd.Run(); err != nil {
			t.Fatalf("Run() error: %v", err)
		}
		titles := getTitlesFrom(t, "mydir/slides.md")
		want := []string{"A", "C", "D"}
		if !equal(titles, want) {
			t.Errorf("got %v, want %v", titles, want)
		}
	})
}
