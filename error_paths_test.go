package main

import (
	"strings"
	"testing"

	"github.com/spf13/afero"
)

// These tests exist to cover the trivial `if err != nil { return err }`
// branches in every command's Run method — without them coverage shows
// holes that are misleading: the propagation logic is real and worth
// pinning down once.

func TestCopyErrorPaths(t *testing.T) {
	t.Run("source read error", func(t *testing.T) {
		setupTestFs(t, testFixture)
		cmd := CopyCmd{Source: "missing/1", Target: "3"}
		if err := cmd.Run(); err == nil {
			t.Fatal("expected error for missing source dir")
		}
	})

	t.Run("source range parse error", func(t *testing.T) {
		setupTestFs(t, testFixture)
		cmd := CopyCmd{Source: "abc", Target: "1"}
		if err := cmd.Run(); err == nil {
			t.Fatal("expected error for invalid source range")
		}
	})

	t.Run("target read error", func(t *testing.T) {
		setupTestFs(t, testFixture)
		cmd := CopyCmd{Source: "1", Target: "missing/1"}
		if err := cmd.Run(); err == nil {
			t.Fatal("expected error for missing target dir")
		}
	})

	t.Run("cross-format copy errors out", func(t *testing.T) {
		origFs, origFmt, origExplicit := appFs, activeFormat, formatExplicit
		t.Cleanup(func() { appFs, activeFormat, formatExplicit = origFs, origFmt, origExplicit })
		appFs = afero.NewMemMapFs()
		formatExplicit = false
		_ = afero.WriteFile(appFs, "md/slides.md", []byte(testFixture), 0o644)
		_ = afero.WriteFile(appFs, "html/index.html", []byte(`<html><head><title>X</title></head><body><div class="slides"><section><h1>A</h1></section></div></body></html>`), 0o644)

		err := (&CopyCmd{Source: "md/1", Target: "html/1"}).Run()
		if err == nil || !strings.Contains(err.Error(), "across formats") {
			t.Errorf("expected cross-format error, got %v", err)
		}
	})
}

func TestDeleteErrorPaths(t *testing.T) {
	t.Run("read error", func(t *testing.T) {
		setupTestFs(t, testFixture)
		if err := (&DeleteCmd{Expr: "missing/1"}).Run(); err == nil {
			t.Fatal("expected error for missing dir")
		}
	})

	t.Run("invalid range", func(t *testing.T) {
		setupTestFs(t, testFixture)
		if err := (&DeleteCmd{Expr: "abc"}).Run(); err == nil {
			t.Fatal("expected error for non-numeric range")
		}
	})
}

func TestInsertErrorPaths(t *testing.T) {
	t.Run("read error", func(t *testing.T) {
		setupTestFs(t, testFixture)
		if err := (&InsertCmd{Pos: "missing/1", Title: "X"}).Run(); err == nil {
			t.Fatal("expected error for missing dir")
		}
	})

	t.Run("invalid position", func(t *testing.T) {
		setupTestFs(t, testFixture)
		if err := (&InsertCmd{Pos: "abc", Title: "X"}).Run(); err == nil {
			t.Fatal("expected error for non-numeric position")
		}
	})
}

func TestRenameErrorPath(t *testing.T) {
	setupTestFs(t, testFixture)
	cmd := RenameCmd{Name: "X", File: "missing.md"}
	if err := cmd.Run(); err == nil {
		t.Fatal("expected error for missing file")
	}
}

func TestRenameSlideErrorPaths(t *testing.T) {
	t.Run("read error", func(t *testing.T) {
		setupTestFs(t, testFixture)
		if err := (&RenameSlideCmd{Expr: "missing/1", Name: "X"}).Run(); err == nil {
			t.Fatal("expected error for missing dir")
		}
	})
}

func TestSlideErrorPaths(t *testing.T) {
	t.Run("read error", func(t *testing.T) {
		setupTestFs(t, testFixture)
		if err := (&SlideCmd{Exprs: []string{"missing/1"}}).Run(); err == nil {
			t.Fatal("expected error for missing dir")
		}
	})

	t.Run("invalid range", func(t *testing.T) {
		setupTestFs(t, testFixture)
		if err := (&SlideCmd{Exprs: []string{"abc"}}).Run(); err == nil {
			t.Fatal("expected error for non-numeric range")
		}
	})
}

func TestSlidesErrorPaths(t *testing.T) {
	t.Run("read error", func(t *testing.T) {
		setupTestFs(t, testFixture)
		if err := (&SlidesCmd{Exprs: []string{"missing/"}}).Run(); err == nil {
			t.Fatal("expected error for missing dir")
		}
	})

	t.Run("invalid range", func(t *testing.T) {
		setupTestFs(t, testFixture)
		if err := (&SlidesCmd{Exprs: []string{"abc"}}).Run(); err == nil {
			t.Fatal("expected error for non-numeric range")
		}
	})

	t.Run("bad glob pattern", func(t *testing.T) {
		setupTestFs(t, testFixture)
		// `[` opens a character class that's never closed — filepath.Glob
		// returns ErrBadPattern.
		if err := (&SlidesCmd{Exprs: []string{"[invalid"}}).Run(); err == nil {
			t.Fatal("expected error for bad glob pattern")
		}
	})
}

func TestSwapErrorPaths(t *testing.T) {
	t.Run("read error for A", func(t *testing.T) {
		setupTestFs(t, testFixture)
		cmd := SwapCmd{A: "missing/1", B: "2"}
		if err := cmd.Run(); err == nil {
			t.Fatal("expected error reading A")
		}
	})

	t.Run("invalid A expression", func(t *testing.T) {
		setupTestFs(t, testFixture)
		cmd := SwapCmd{A: "abc", B: "2"}
		if err := cmd.Run(); err == nil {
			t.Fatal("expected error for invalid A")
		}
	})

	t.Run("read error for B", func(t *testing.T) {
		setupTestFs(t, testFixture)
		cmd := SwapCmd{A: "1", B: "missing/1"}
		if err := cmd.Run(); err == nil {
			t.Fatal("expected error reading B")
		}
	})

	t.Run("invalid B expression", func(t *testing.T) {
		setupTestFs(t, testFixture)
		cmd := SwapCmd{A: "1", B: "abc"}
		if err := cmd.Run(); err == nil {
			t.Fatal("expected error for invalid B")
		}
	})

	t.Run("range for B not allowed", func(t *testing.T) {
		setupTestFs(t, testFixture)
		cmd := SwapCmd{A: "1", B: "2-3"}
		if err := cmd.Run(); err == nil {
			t.Fatal("expected error for range B")
		}
	})

	t.Run("cross-format swap errors out", func(t *testing.T) {
		origFs, origFmt, origExplicit := appFs, activeFormat, formatExplicit
		t.Cleanup(func() { appFs, activeFormat, formatExplicit = origFs, origFmt, origExplicit })
		appFs = afero.NewMemMapFs()
		formatExplicit = false
		_ = afero.WriteFile(appFs, "md/slides.md", []byte(testFixture), 0o644)
		_ = afero.WriteFile(appFs, "html/index.html", []byte(`<html><head><title>X</title></head><body><div class="slides"><section><h1>A</h1></section></div></body></html>`), 0o644)

		err := (&SwapCmd{A: "md/1", B: "html/1"}).Run()
		if err == nil || !strings.Contains(err.Error(), "across formats") {
			t.Errorf("expected cross-format error, got %v", err)
		}
	})

	t.Run("cross-file write error propagates", func(t *testing.T) {
		// Layer a read-only FS over an FS we can pre-populate.
		base := afero.NewMemMapFs()
		_ = afero.WriteFile(base, "a/slides.md", []byte(testFixture), 0o644)
		_ = afero.WriteFile(base, "b/slides.md", []byte(testFixture2), 0o644)
		origFs, origFmt, origStdout := appFs, activeFormat, stdout
		t.Cleanup(func() { appFs, activeFormat, stdout = origFs, origFmt, origStdout })
		appFs = afero.NewReadOnlyFs(base)
		// activeFormat stays at slidev; stdout doesn't matter here.

		if err := (&SwapCmd{A: "a/1", B: "b/2"}).Run(); err == nil {
			t.Fatal("expected write error on read-only fs")
		}
	})
}

func TestIsDeckPathStatError(t *testing.T) {
	origFs := appFs
	t.Cleanup(func() { appFs = origFs })
	appFs = afero.NewMemMapFs()

	if isDeckPath("does-not-exist") {
		t.Error("expected false for non-existent path")
	}
}
