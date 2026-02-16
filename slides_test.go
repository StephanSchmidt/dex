package main

import (
	"testing"

	"github.com/spf13/afero"
)

func TestSlides(t *testing.T) {
	t.Run("default file", func(t *testing.T) {
		buf := setupTestFs(t, testFixture)
		cmd := SlidesCmd{}
		if err := cmd.Run(); err != nil {
			t.Fatalf("SlidesCmd.Run() error: %v", err)
		}
		want := "  1  A\n  2  B\n  3  C\n  4  D\n"
		if got := buf.String(); got != want {
			t.Errorf("SlidesCmd output:\ngot:  %q\nwant: %q", got, want)
		}
	})

	t.Run("directory with trailing slash", func(t *testing.T) {
		buf := setupTestFs(t, testFixture)
		afero.WriteFile(appFs, "mydir/slides.md", []byte(testFixture), 0o644)
		cmd := SlidesCmd{Exprs: []string{"mydir/"}}
		if err := cmd.Run(); err != nil {
			t.Fatalf("SlidesCmd.Run() error: %v", err)
		}
		want := "  1  A\n  2  B\n  3  C\n  4  D\n"
		if got := buf.String(); got != want {
			t.Errorf("SlidesCmd output:\ngot:  %q\nwant: %q", got, want)
		}
	})

	t.Run("with range", func(t *testing.T) {
		buf := setupTestFs(t, testFixture)
		cmd := SlidesCmd{Exprs: []string{"1,3"}}
		if err := cmd.Run(); err != nil {
			t.Fatalf("SlidesCmd.Run() error: %v", err)
		}
		want := "  1  A\n  3  C\n"
		if got := buf.String(); got != want {
			t.Errorf("SlidesCmd output:\ngot:  %q\nwant: %q", got, want)
		}
	})

	t.Run("multiple sources", func(t *testing.T) {
		buf := setupTestFs(t, testFixture)
		afero.WriteFile(appFs, "other/slides.md", []byte(testFixture), 0o644)
		cmd := SlidesCmd{Exprs: []string{"1,3", "other/2-4"}}
		if err := cmd.Run(); err != nil {
			t.Fatalf("SlidesCmd.Run() error: %v", err)
		}
		want := "  1  A\n  3  C\n  2  B\n  3  C\n  4  D\n"
		if got := buf.String(); got != want {
			t.Errorf("SlidesCmd output:\ngot:  %q\nwant: %q", got, want)
		}
	})
}
