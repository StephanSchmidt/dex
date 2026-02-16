package main

import (
	"strings"
	"testing"
)

func TestSlide(t *testing.T) {
	t.Run("single slide", func(t *testing.T) {
		buf := setupTestFs(t, testFixture)
		cmd := SlideCmd{Exprs: []string{"1"}}
		if err := cmd.Run(); err != nil {
			t.Fatalf("SlideCmd.Run() error: %v", err)
		}
		out := buf.String()
		if !strings.Contains(out, "# A") {
			t.Errorf("expected slide A, got:\n%s", out)
		}
		if strings.Contains(out, "theme: default") {
			t.Errorf("should not contain document frontmatter, got:\n%s", out)
		}
	})

	t.Run("range", func(t *testing.T) {
		buf := setupTestFs(t, testFixture)
		cmd := SlideCmd{Exprs: []string{"2:3"}}
		if err := cmd.Run(); err != nil {
			t.Fatalf("SlideCmd.Run() error: %v", err)
		}
		out := buf.String()
		if !strings.Contains(out, "# B") {
			t.Errorf("expected slide B, got:\n%s", out)
		}
		if !strings.Contains(out, "# C") {
			t.Errorf("expected slide C, got:\n%s", out)
		}
		if !strings.Contains(out, "layout: center") {
			t.Errorf("expected per-slide frontmatter for C, got:\n%s", out)
		}
	})

	t.Run("negative index", func(t *testing.T) {
		buf := setupTestFs(t, testFixture)
		cmd := SlideCmd{Exprs: []string{"-1"}}
		if err := cmd.Run(); err != nil {
			t.Fatalf("SlideCmd.Run() error: %v", err)
		}
		out := buf.String()
		if !strings.Contains(out, "# D") {
			t.Errorf("expected slide D, got:\n%s", out)
		}
	})

	t.Run("directory prefix", func(t *testing.T) {
		buf := setupTestFs(t, testFixture)
		writeTestFile(t, "mydir/slides.md", testFixture)
		cmd := SlideCmd{Exprs: []string{"mydir/2"}}
		if err := cmd.Run(); err != nil {
			t.Fatalf("SlideCmd.Run() error: %v", err)
		}
		out := buf.String()
		if !strings.Contains(out, "# B") {
			t.Errorf("expected slide B, got:\n%s", out)
		}
	})

	t.Run("multiple sources", func(t *testing.T) {
		buf := setupTestFs(t, testFixture)
		writeTestFile(t, "other/slides.md", testFixture)
		cmd := SlideCmd{Exprs: []string{"1:2", "other/4"}}
		if err := cmd.Run(); err != nil {
			t.Fatalf("SlideCmd.Run() error: %v", err)
		}
		out := buf.String()
		if !strings.Contains(out, "# A") {
			t.Errorf("expected slide A, got:\n%s", out)
		}
		if !strings.Contains(out, "# B") {
			t.Errorf("expected slide B, got:\n%s", out)
		}
		if !strings.Contains(out, "# D") {
			t.Errorf("expected slide D from other dir, got:\n%s", out)
		}
	})
}
