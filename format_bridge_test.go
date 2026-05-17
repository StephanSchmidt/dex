package main

import (
	"testing"

	"github.com/StephanSchmidt/dex/revealjs"
	"github.com/StephanSchmidt/dex/slidev"
	"github.com/spf13/afero"
)

func TestDetectFormat(t *testing.T) {
	tests := []struct {
		name string
		path string
		want string // type name
	}{
		{"md is slidev", "deck.md", "slidev.Format"},
		{"html is revealjs", "deck.html", "revealjs.Format"},
		{"htm is revealjs", "deck.htm", "revealjs.Format"},
		{"uppercase HTML is revealjs", "deck.HTML", "revealjs.Format"},
		{"unknown extension defaults to slidev", "deck.foo", "slidev.Format"},
		{"no extension defaults to slidev", "deck", "slidev.Format"},
		{"directory path defaults to slidev", "some/dir/", "slidev.Format"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := detectFormat(tt.path)
			switch got.(type) {
			case slidev.Format:
				if tt.want != "slidev.Format" {
					t.Errorf("detectFormat(%q) = slidev, want %s", tt.path, tt.want)
				}
			case revealjs.Format:
				if tt.want != "revealjs.Format" {
					t.Errorf("detectFormat(%q) = revealjs, want %s", tt.path, tt.want)
				}
			default:
				t.Errorf("unexpected format type %T for %q", got, tt.path)
			}
		})
	}
}

func TestResolveFile(t *testing.T) {
	origFs, origFmt, origExplicit := appFs, activeFormat, formatExplicit
	t.Cleanup(func() { appFs, activeFormat, formatExplicit = origFs, origFmt, origExplicit })

	appFs = afero.NewMemMapFs()
	activeFormat = slidev.Format{}
	formatExplicit = false
	_ = appFs.MkdirAll("mydir", 0o755)
	_ = afero.WriteFile(appFs, "mydir/slides.md", []byte("---\n"), 0o644)
	_ = afero.WriteFile(appFs, "explicit.md", []byte("---\n"), 0o644)

	tests := []struct {
		name string
		in   string
		want string
	}{
		{"empty uses default", "", "slides.md"},
		{"trailing slash appends default", "mydir/", "mydir/slides.md"},
		{"existing directory appends default", "mydir", "mydir/slides.md"},
		{"existing file returned as-is", "explicit.md", "explicit.md"},
		{"non-existent path returned as-is", "missing.md", "missing.md"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := resolveFile(tt.in); got != tt.want {
				t.Errorf("resolveFile(%q) = %q, want %q", tt.in, got, tt.want)
			}
		})
	}
}

func TestResolveFileAutoDetectsOtherFormat(t *testing.T) {
	origFs, origFmt, origExplicit := appFs, activeFormat, formatExplicit
	t.Cleanup(func() { appFs, activeFormat, formatExplicit = origFs, origFmt, origExplicit })

	appFs = afero.NewMemMapFs()
	activeFormat = slidev.Format{} // default
	formatExplicit = false

	// Directory contains only index.html — the active format's default
	// (slides.md) doesn't exist there. resolveFile should fall through
	// to the other registered format's default rather than blindly
	// returning a path that won't exist.
	_ = appFs.MkdirAll("htmldir", 0o755)
	_ = afero.WriteFile(appFs, "htmldir/index.html", []byte("<html></html>"), 0o644)

	for _, in := range []string{"htmldir", "htmldir/"} {
		got := resolveFile(in)
		if got != "htmldir/index.html" {
			t.Errorf("resolveFile(%q) = %q, want htmldir/index.html", in, got)
		}
	}
}

func TestResolveFileBothPresentPrefersActive(t *testing.T) {
	origFs, origFmt, origExplicit := appFs, activeFormat, formatExplicit
	t.Cleanup(func() { appFs, activeFormat, formatExplicit = origFs, origFmt, origExplicit })

	appFs = afero.NewMemMapFs()
	activeFormat = slidev.Format{}
	formatExplicit = false
	_ = appFs.MkdirAll("both", 0o755)
	_ = afero.WriteFile(appFs, "both/slides.md", []byte("---\n"), 0o644)
	_ = afero.WriteFile(appFs, "both/index.html", []byte("<html></html>"), 0o644)

	if got := resolveFile("both/"); got != "both/slides.md" {
		t.Errorf("resolveFile = %q, want both/slides.md (active format wins)", got)
	}
}

func TestResolveFileExplicitFormatNoProbe(t *testing.T) {
	origFs, origFmt, origExplicit := appFs, activeFormat, formatExplicit
	t.Cleanup(func() { appFs, activeFormat, formatExplicit = origFs, origFmt, origExplicit })

	appFs = afero.NewMemMapFs()
	activeFormat = slidev.Format{}
	formatExplicit = true // user pinned slidev
	_ = appFs.MkdirAll("htmldir", 0o755)
	_ = afero.WriteFile(appFs, "htmldir/index.html", []byte("<html></html>"), 0o644)

	// User said --format slidev, so we should NOT detect the index.html
	// and instead return the (missing) slides.md path for a clean error.
	if got := resolveFile("htmldir/"); got != "htmldir/slides.md" {
		t.Errorf("resolveFile = %q, want htmldir/slides.md (explicit format)", got)
	}
}

func TestReadDeckErrorOnMissingFile(t *testing.T) {
	origFs, origFmt, origExplicit := appFs, activeFormat, formatExplicit
	t.Cleanup(func() { appFs, activeFormat, formatExplicit = origFs, origFmt, origExplicit })

	appFs = afero.NewMemMapFs()
	activeFormat = slidev.Format{}
	formatExplicit = false

	_, _, err := readDeck("nope.md")
	if err == nil {
		t.Fatal("expected error reading missing file")
	}
}

func TestReadDeckAutoDetectsFormat(t *testing.T) {
	origFs, origFmt, origExplicit := appFs, activeFormat, formatExplicit
	t.Cleanup(func() { appFs, activeFormat, formatExplicit = origFs, origFmt, origExplicit })

	appFs = afero.NewMemMapFs()
	activeFormat = slidev.Format{} // start as slidev
	formatExplicit = false

	html := `<!doctype html><html><head><title>X</title></head><body><div class="slides"><section><h1>A</h1></section></div></body></html>`
	_ = afero.WriteFile(appFs, "deck.html", []byte(html), 0o644)

	if _, _, err := readDeck("deck.html"); err != nil {
		t.Fatalf("readDeck: %v", err)
	}
	if _, ok := activeFormat.(revealjs.Format); !ok {
		t.Errorf("expected activeFormat to switch to revealjs, got %T", activeFormat)
	}
}

func TestReadDeckRespectsExplicitFormat(t *testing.T) {
	origFs, origFmt, origExplicit := appFs, activeFormat, formatExplicit
	t.Cleanup(func() { appFs, activeFormat, formatExplicit = origFs, origFmt, origExplicit })

	appFs = afero.NewMemMapFs()
	activeFormat = slidev.Format{}
	formatExplicit = true // user pinned slidev

	// File extension says html, but we should NOT auto-detect because
	// formatExplicit is set.
	_ = afero.WriteFile(appFs, "deck.html", []byte("---\ntitle: ok\n---\n"), 0o644)

	if _, _, err := readDeck("deck.html"); err != nil {
		t.Fatalf("readDeck: %v", err)
	}
	if _, ok := activeFormat.(slidev.Format); !ok {
		t.Errorf("expected activeFormat to stay slidev when explicit, got %T", activeFormat)
	}
}

func TestDisplayTitle(t *testing.T) {
	origFmt := activeFormat
	t.Cleanup(func() { activeFormat = origFmt })
	activeFormat = slidev.Format{}

	tests := []struct {
		name string
		s    slide
		want string
	}{
		{"plain slide", slide{Content: "# Hello\n"}, "Hello"},
		{"detail slide gets > prefix", slide{Content: "# Hello\n", Metadata: "nav: detail\n"}, "> Hello"},
		{"non-detail metadata", slide{Content: "# Hello\n", Metadata: "layout: center\n"}, "Hello"},
		{"untitled detail slide", slide{Content: "text\n", Metadata: "nav: detail\n"}, "> (untitled)"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := displayTitle(tt.s); got != tt.want {
				t.Errorf("displayTitle = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestWriteDeckErrorOnReadOnlyFs(t *testing.T) {
	origFs, origFmt := appFs, activeFormat
	t.Cleanup(func() { appFs, activeFormat = origFs, origFmt })

	activeFormat = slidev.Format{}
	appFs = afero.NewReadOnlyFs(afero.NewMemMapFs())

	if err := writeDeck("deck.md", deck{Metadata: "x\n"}); err == nil {
		t.Fatal("expected error on read-only fs")
	}
}
