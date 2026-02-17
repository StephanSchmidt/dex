package slidev

import (
	"strings"
	"testing"

	"github.com/StephanSchmidt/dex/format"
)

const fixture = `---
theme: default
title: Test
---

# A

---

# B

---
layout: center
---

# C

---

# D
`

const fixture2 = `---
theme: default
title: Test2
---

# X

---

# Y

---

# Z
`

var f Format

func TestParse(t *testing.T) {
	t.Run("slide count", func(t *testing.T) {
		d := f.Parse([]byte(fixture))
		if got := len(d.Slides); got != 4 {
			t.Fatalf("got %d slides, want 4", got)
		}
	})

	t.Run("deck metadata", func(t *testing.T) {
		d := f.Parse([]byte(fixture))
		want := "theme: default\ntitle: Test\n"
		if d.Metadata != want {
			t.Errorf("deck metadata:\ngot:  %q\nwant: %q", d.Metadata, want)
		}
	})

	t.Run("plain slide content", func(t *testing.T) {
		d := f.Parse([]byte(fixture))
		if got := d.Slides[0].Content; got != "\n# A\n\n" {
			t.Errorf("slide 1 content: got %q", got)
		}
		if d.Slides[0].Metadata != "" {
			t.Errorf("slide 1 should have no metadata, got %q", d.Slides[0].Metadata)
		}
	})

	t.Run("slide with metadata", func(t *testing.T) {
		d := f.Parse([]byte(fixture))
		s := d.Slides[2] // slide C has layout: center
		if s.Metadata != "layout: center\n" {
			t.Errorf("slide 3 metadata: got %q, want %q", s.Metadata, "layout: center\n")
		}
		if got := s.Content; got != "\n# C\n\n" {
			t.Errorf("slide 3 content: got %q", got)
		}
	})

	t.Run("second fixture", func(t *testing.T) {
		d := f.Parse([]byte(fixture2))
		if got := len(d.Slides); got != 3 {
			t.Fatalf("got %d slides, want 3", got)
		}
	})
}

func TestRender(t *testing.T) {
	t.Run("round-trip idempotent", func(t *testing.T) {
		// Render→parse→render must be stable (idempotent after first pass).
		d := f.Parse([]byte(fixture))
		r1 := f.Render(d)
		r2 := f.Render(f.Parse(r1))
		if string(r1) != string(r2) {
			t.Errorf("not idempotent:\nfirst:  %q\nsecond: %q", r1, r2)
		}
	})

	t.Run("round-trip preserves structure", func(t *testing.T) {
		d := f.Parse([]byte(fixture))
		d2 := f.Parse(f.Render(d))
		if len(d.Slides) != len(d2.Slides) {
			t.Fatalf("slide count: got %d, want %d", len(d2.Slides), len(d.Slides))
		}
		if d.Metadata != d2.Metadata {
			t.Errorf("metadata: got %q, want %q", d2.Metadata, d.Metadata)
		}
		for i := range d.Slides {
			if d.Slides[i].Metadata != d2.Slides[i].Metadata {
				t.Errorf("slide %d metadata: got %q, want %q", i+1, d2.Slides[i].Metadata, d.Slides[i].Metadata)
			}
		}
	})

	t.Run("round-trip fixture2 idempotent", func(t *testing.T) {
		d := f.Parse([]byte(fixture2))
		r1 := f.Render(d)
		r2 := f.Render(f.Parse(r1))
		if string(r1) != string(r2) {
			t.Errorf("not idempotent:\nfirst:  %q\nsecond: %q", r1, r2)
		}
	})
}

func TestExtractTitle(t *testing.T) {
	tests := []struct {
		name  string
		slide format.Slide
		want  string
	}{
		{"H1 heading", format.Slide{Content: "\n# Hello World\n\n"}, "Hello World"},
		{"title attribute", format.Slide{Content: "\n<div title=\"Slide One\">\ncontent\n</div>\n\n"}, "Slide One"},
		{"H1 takes precedence", format.Slide{Content: "\n# Heading\n<div title=\"Attr\">\n</div>\n"}, "Heading"},
		{"metadata title", format.Slide{Metadata: "layout: quote\ntitle: From Metadata\n"}, "From Metadata"},
		{"H1 over metadata", format.Slide{Content: "\n# Heading\n", Metadata: "title: Meta\n"}, "Heading"},
		{"no title", format.Slide{Content: "\nJust text\n\n"}, "(untitled)"},
		{"empty", format.Slide{}, "(untitled)"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := f.ExtractTitle(tt.slide); got != tt.want {
				t.Errorf("ExtractTitle() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestDefaultFile(t *testing.T) {
	if got := f.DefaultFile(); got != "slides.md" {
		t.Errorf("DefaultFile() = %q, want %q", got, "slides.md")
	}
}

func TestRenderSlide(t *testing.T) {
	t.Run("without metadata", func(t *testing.T) {
		s := format.Slide{Content: "\n# A\n\n"}
		got := f.RenderSlide(s)
		want := "---\n\n# A\n\n"
		if got != want {
			t.Errorf("RenderSlide():\ngot:  %q\nwant: %q", got, want)
		}
	})

	t.Run("with metadata", func(t *testing.T) {
		s := format.Slide{Metadata: "layout: center\n", Content: "\n# C\n\n"}
		got := f.RenderSlide(s)
		want := "---\nlayout: center\n---\n\n# C\n\n"
		if got != want {
			t.Errorf("RenderSlide():\ngot:  %q\nwant: %q", got, want)
		}
	})
}

func TestIsFrontmatter(t *testing.T) {
	tests := []struct {
		name  string
		block string
		want  bool
	}{
		{"empty", "", true},
		{"yaml key-value", "layout: center\n", true},
		{"yaml list", "- item\n", true},
		{"heading", "# Title\n", false},
		{"html", "<div>\n</div>\n", false},
		{"plain text", "just some text\n", false},
		{"multi-line yaml", "layout: center\nclass: text-center\n", true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := isFrontmatter(tt.block); got != tt.want {
				t.Errorf("isFrontmatter(%q) = %v, want %v", tt.block, got, tt.want)
			}
		})
	}
}

func TestRenameSlide(t *testing.T) {
	t.Run("replace H1 heading", func(t *testing.T) {
		s := format.Slide{Content: "\n# Old Title\n\nBody text\n"}
		got := f.RenameSlide(s, "New Title")
		if !strings.Contains(got.Content, "# New Title") {
			t.Errorf("expected H1 replaced, got %q", got.Content)
		}
		if strings.Contains(got.Content, "Old Title") {
			t.Errorf("old title still present: %q", got.Content)
		}
	})

	t.Run("replace title attribute", func(t *testing.T) {
		s := format.Slide{Content: "\n<div title=\"Slide One\">\ncontent\n</div>\n"}
		got := f.RenameSlide(s, "New Attr")
		if !strings.Contains(got.Content, `title="New Attr"`) {
			t.Errorf("expected title attr replaced, got %q", got.Content)
		}
	})

	t.Run("prepend H1 when no title", func(t *testing.T) {
		s := format.Slide{Content: "\nJust text\n"}
		got := f.RenameSlide(s, "Added")
		if !strings.HasPrefix(got.Content, "# Added\n") {
			t.Errorf("expected H1 prepended, got %q", got.Content)
		}
	})

	t.Run("H1 takes priority over title attr", func(t *testing.T) {
		s := format.Slide{Content: "\n# Heading\n<div title=\"Attr\">\n</div>\n"}
		got := f.RenameSlide(s, "Replaced")
		if !strings.Contains(got.Content, "# Replaced") {
			t.Errorf("expected H1 replaced, got %q", got.Content)
		}
		if !strings.Contains(got.Content, `title="Attr"`) {
			t.Errorf("title attr should be untouched, got %q", got.Content)
		}
	})
}

func TestNewSlide(t *testing.T) {
	got := f.NewSlide("Hello World")
	if !strings.Contains(got.Content, "# Hello World") {
		t.Errorf("expected H1 heading, got %q", got.Content)
	}
}

func TestRenameDeck(t *testing.T) {
	t.Run("replace existing title", func(t *testing.T) {
		d := format.Deck{Metadata: "theme: default\ntitle: Old\n"}
		got := f.RenameDeck(d, "New")
		if !strings.Contains(got.Metadata, "title: New") {
			t.Errorf("expected title replaced, got %q", got.Metadata)
		}
		if strings.Contains(got.Metadata, "title: Old") {
			t.Errorf("old title still present: %q", got.Metadata)
		}
	})

	t.Run("add title when missing", func(t *testing.T) {
		d := format.Deck{Metadata: "theme: default\n"}
		got := f.RenameDeck(d, "Added")
		if !strings.Contains(got.Metadata, "title: Added") {
			t.Errorf("expected title added, got %q", got.Metadata)
		}
	})

	t.Run("preserve other keys", func(t *testing.T) {
		d := format.Deck{Metadata: "theme: seriph\ntitle: Old\nclass: text-center\n"}
		got := f.RenameDeck(d, "New")
		if !strings.Contains(got.Metadata, "theme: seriph") {
			t.Errorf("theme lost: %q", got.Metadata)
		}
		if !strings.Contains(got.Metadata, "class: text-center") {
			t.Errorf("class lost: %q", got.Metadata)
		}
	})
}

func TestScaffold(t *testing.T) {
	files := f.Scaffold("My Talk", "my-talk")

	dirs := 0
	regular := 0
	names := make(map[string]bool)
	for _, sf := range files {
		names[sf.Path] = true
		if sf.IsDir {
			dirs++
		} else {
			regular++
		}
	}

	if dirs != 2 {
		t.Errorf("expected 2 dirs, got %d", dirs)
	}
	if regular != 2 {
		t.Errorf("expected 2 files, got %d", regular)
	}
	for _, want := range []string{"public", "snippets", "package.json", "slides.md"} {
		if !names[want] {
			t.Errorf("missing %q in scaffold output", want)
		}
	}

	// Check package.json contains the slug.
	for _, sf := range files {
		if sf.Path == "package.json" && !strings.Contains(sf.Content, "my-talk") {
			t.Errorf("package.json doesn't contain slug: %q", sf.Content)
		}
		if sf.Path == "slides.md" && !strings.Contains(sf.Content, "My Talk") {
			t.Errorf("slides.md doesn't contain title: %q", sf.Content)
		}
	}
}

func TestPostScaffoldCmd(t *testing.T) {
	got := f.PostScaffoldCmd()
	if len(got) != 2 || got[0] != "pnpm" || got[1] != "install" {
		t.Errorf("PostScaffoldCmd() = %v, want [pnpm install]", got)
	}
}
