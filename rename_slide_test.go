package main

import (
	"testing"

	"github.com/spf13/afero"
)

func TestRenameSlide(t *testing.T) {
	t.Run("rename H1 heading", func(t *testing.T) {
		setupTestFs(t, testFixture)
		cmd := RenameSlideCmd{Expr: "1", Name: "NewTitle"}
		if err := cmd.Run(); err != nil {
			t.Fatalf("Run() error: %v", err)
		}
		titles := getTitles(t)
		want := []string{"NewTitle", "B", "C", "D"}
		if !equal(titles, want) {
			t.Errorf("got %v, want %v", titles, want)
		}
	})

	t.Run("rename title attribute", func(t *testing.T) {
		fixture := "---\ntheme: default\n---\n\n<div title=\"Slide One\">\ncontent\n</div>\n\n---\n\n# B\n"
		setupTestFs(t, fixture)
		cmd := RenameSlideCmd{Expr: "1", Name: "New Attr"}
		if err := cmd.Run(); err != nil {
			t.Fatalf("Run() error: %v", err)
		}
		titles := getTitles(t)
		if titles[0] != "New Attr" {
			t.Errorf("got %q, want %q", titles[0], "New Attr")
		}
	})

	t.Run("no title adds H1", func(t *testing.T) {
		fixture := "---\ntheme: default\n---\n\nJust text\n\n---\n\n# B\n"
		setupTestFs(t, fixture)
		cmd := RenameSlideCmd{Expr: "1", Name: "Added"}
		if err := cmd.Run(); err != nil {
			t.Fatalf("Run() error: %v", err)
		}
		titles := getTitles(t)
		if titles[0] != "Added" {
			t.Errorf("got %q, want %q", titles[0], "Added")
		}
	})

	t.Run("preserves frontmatter", func(t *testing.T) {
		setupTestFs(t, testFixture)
		cmd := RenameSlideCmd{Expr: "3", Name: "NewC"}
		if err := cmd.Run(); err != nil {
			t.Fatalf("Run() error: %v", err)
		}
		data, _ := afero.ReadFile(appFs, "slides.md")
		d := parseDeck(data)
		if d.slides[2].frontmatter == "" {
			t.Error("expected frontmatter on slide 3, got empty")
		}
		titles := getTitles(t)
		want := []string{"A", "B", "NewC", "D"}
		if !equal(titles, want) {
			t.Errorf("got %v, want %v", titles, want)
		}
	})

	t.Run("out of range", func(t *testing.T) {
		setupTestFs(t, testFixture)
		if err := (&RenameSlideCmd{Expr: "0", Name: "X"}).Run(); err == nil {
			t.Error("expected error for slide 0")
		}
		if err := (&RenameSlideCmd{Expr: "5", Name: "X"}).Run(); err == nil {
			t.Error("expected error for slide 5 (out of range)")
		}
	})

	t.Run("directory path", func(t *testing.T) {
		setupTestFs(t, testFixture)
		_ = appFs.MkdirAll("mydir", 0o755)
		writeTestFile(t, "mydir/slides.md", testFixture)
		cmd := RenameSlideCmd{Expr: "mydir/2", Name: "Renamed"}
		if err := cmd.Run(); err != nil {
			t.Fatalf("Run() error: %v", err)
		}
		titles := getTitlesFrom(t, "mydir/slides.md")
		want := []string{"A", "Renamed", "C", "D"}
		if !equal(titles, want) {
			t.Errorf("got %v, want %v", titles, want)
		}
	})

	t.Run("multi rename", func(t *testing.T) {
		setupTestFs(t, testFixture)
		cmd := RenameSlideCmd{Expr: "1,3", Name: "Same"}
		if err := cmd.Run(); err != nil {
			t.Fatalf("Run() error: %v", err)
		}
		titles := getTitles(t)
		want := []string{"Same", "B", "Same", "D"}
		if !equal(titles, want) {
			t.Errorf("got %v, want %v", titles, want)
		}
	})

	t.Run("range rename", func(t *testing.T) {
		setupTestFs(t, testFixture)
		cmd := RenameSlideCmd{Expr: "2-4", Name: "Bulk"}
		if err := cmd.Run(); err != nil {
			t.Fatalf("Run() error: %v", err)
		}
		titles := getTitles(t)
		want := []string{"A", "Bulk", "Bulk", "Bulk"}
		if !equal(titles, want) {
			t.Errorf("got %v, want %v", titles, want)
		}
	})
}
