package revealjs

import (
	"strings"
	"testing"

	"github.com/StephanSchmidt/dex/format"
)

// fixture is based on the official reveal.js index.html from
// https://github.com/hakimel/reveal.js/blob/master/index.html
// with extra slides and a data-background-color attribute from
// https://revealjs.com/backgrounds/.
const fixture = `<!doctype html>
<html lang="en">
	<head>
		<meta charset="utf-8">
		<meta name="viewport" content="width=device-width, initial-scale=1.0, maximum-scale=1.0, user-scalable=no">

		<title>reveal.js</title>

		<link rel="stylesheet" href="dist/reset.css">
		<link rel="stylesheet" href="dist/reveal.css">
		<link rel="stylesheet" href="dist/theme/black.css">

		<!-- Theme used for syntax highlighted code -->
		<link rel="stylesheet" href="plugin/highlight/monokai.css">
	</head>
	<body>
		<div class="reveal">
			<div class="slides">
				<section><h1>Slide 1</h1></section>
				<section><h2>Slide 2</h2></section>
				<section data-background-color="aquamarine"><h2>Slide 3</h2></section>
				<section><h1>Slide 4</h1></section>
			</div>
		</div>

		<script src="dist/reveal.js"></script>
		<script src="plugin/notes/notes.js"></script>
		<script src="plugin/markdown/markdown.js"></script>
		<script src="plugin/highlight/highlight.js"></script>
		<script>
			Reveal.initialize({
				hash: true,
				plugins: [ RevealMarkdown, RevealHighlight, RevealNotes ]
			});
		</script>
	</body>
</html>
`

// fixture2 uses the minimal markup from https://revealjs.com/markup/.
const fixture2 = `<!doctype html>
<html lang="en">
	<head>
		<meta charset="utf-8">
		<title>Minimal</title>
		<link rel="stylesheet" href="dist/reveal.css">
		<link rel="stylesheet" href="dist/theme/white.css">
	</head>
	<body>
		<div class="reveal">
			<div class="slides">
				<section><h1>Hello</h1></section>
				<section><h1>World</h1></section>
				<section><h1>Bye</h1></section>
			</div>
		</div>
		<script src="dist/reveal.js"></script>
		<script>Reveal.initialize();</script>
	</body>
</html>
`

var f Format

func TestParse(t *testing.T) {
	t.Run("slide count", func(t *testing.T) {
		d := f.Parse([]byte(fixture))
		if got := len(d.Slides); got != 4 {
			t.Fatalf("got %d slides, want 4", got)
		}
	})

	t.Run("deck metadata has sentinel", func(t *testing.T) {
		d := f.Parse([]byte(fixture))
		if !strings.Contains(d.Metadata, sentinel) {
			t.Errorf("deck metadata missing sentinel %q", sentinel)
		}
	})

	t.Run("plain slide content", func(t *testing.T) {
		d := f.Parse([]byte(fixture))
		if got := d.Slides[0].Content; !strings.Contains(got, "Slide 1") {
			t.Errorf("slide 1 content missing title: got %q", got)
		}
		if d.Slides[0].Metadata != "" {
			t.Errorf("slide 1 should have no metadata, got %q", d.Slides[0].Metadata)
		}
	})

	t.Run("slide with data attributes", func(t *testing.T) {
		d := f.Parse([]byte(fixture))
		s := d.Slides[2] // slide 3 has data-background-color
		if !strings.Contains(s.Metadata, `data-background-color="aquamarine"`) {
			t.Errorf("slide 3 metadata: got %q, want data-background-color attr", s.Metadata)
		}
		if !strings.Contains(s.Content, "Slide 3") {
			t.Errorf("slide 3 content missing title: got %q", s.Content)
		}
	})

	t.Run("second fixture", func(t *testing.T) {
		d := f.Parse([]byte(fixture2))
		if got := len(d.Slides); got != 3 {
			t.Fatalf("got %d slides, want 3", got)
		}
	})

	t.Run("falls back to .slides when div.slides missing", func(t *testing.T) {
		// No <div class="slides"> — the container is found by class only.
		input := `<!doctype html><html><body><section class="slides"><section><h1>X</h1></section><section><h1>Y</h1></section></section></body></html>`
		d := f.Parse([]byte(input))
		if got := len(d.Slides); got != 2 {
			t.Fatalf("got %d slides, want 2", got)
		}
	})

	t.Run("non-section children ignored", func(t *testing.T) {
		// A <div> mixed in with sections should be skipped.
		input := `<!doctype html><html><body><div class="slides"><section><h1>A</h1></section><div>not a slide</div><section><h1>B</h1></section></div></body></html>`
		d := f.Parse([]byte(input))
		if got := len(d.Slides); got != 2 {
			t.Fatalf("got %d slides, want 2", got)
		}
	})
}

func TestRender(t *testing.T) {
	t.Run("round-trip idempotent", func(t *testing.T) {
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
		{"h1", format.Slide{Content: "<h1>Hello World</h1>"}, "Hello World"},
		{"h2", format.Slide{Content: "<h2>Subtitle</h2>"}, "Subtitle"},
		{"h1 precedence over h2", format.Slide{Content: "<h2>Second</h2><h1>First</h1>"}, "First"},
		{"no heading", format.Slide{Content: "<p>Just text</p>"}, "(untitled)"},
		{"empty", format.Slide{}, "(untitled)"},
		{"br becomes space", format.Slide{Content: "<h2>The new<br>roles.</h2>"}, "The new roles."},
		{"br self-closing", format.Slide{Content: "<h2>One<br/>two<br />three</h2>"}, "One two three"},
		{"br with inline tags", format.Slide{Content: "<h2>Worked example:<br>a <em>checkout team</em>.</h2>"}, "Worked example: a checkout team."},
		{"collapses whitespace", format.Slide{Content: "<h1>  spaced   out  </h1>"}, "spaced out"},
		{
			name:  "data-menu-title fallback",
			slide: format.Slide{Metadata: `class="quote-slide" data-menu-title="Quote intro"`, Content: "<blockquote>...</blockquote>"},
			want:  "Quote intro",
		},
		{
			name:  "aria-label fallback",
			slide: format.Slide{Metadata: `class="stat" aria-label="Big number"`, Content: "<p>42%</p>"},
			want:  "Big number",
		},
		{
			name:  "title attribute fallback",
			slide: format.Slide{Metadata: `title="Tooltip title"`, Content: "<p>body</p>"},
			want:  "Tooltip title",
		},
		{
			name:  "heading wins over data-menu-title",
			slide: format.Slide{Metadata: `data-menu-title="From metadata"`, Content: "<h2>From heading</h2>"},
			want:  "From heading",
		},
		{
			name:  "data-menu-title wins over aria-label",
			slide: format.Slide{Metadata: `data-menu-title="Menu" aria-label="Aria"`, Content: "<p>body</p>"},
			want:  "Menu",
		},
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
	if got := f.DefaultFile(); got != "index.html" {
		t.Errorf("DefaultFile() = %q, want %q", got, "index.html")
	}
}

func TestRenderSlide(t *testing.T) {
	t.Run("without attributes", func(t *testing.T) {
		s := format.Slide{Content: "<h1>A</h1>"}
		got := f.RenderSlide(s)
		want := "<section><h1>A</h1></section>\n"
		if got != want {
			t.Errorf("RenderSlide():\ngot:  %q\nwant: %q", got, want)
		}
	})

	t.Run("with attributes", func(t *testing.T) {
		s := format.Slide{Metadata: `data-background-color="aquamarine"`, Content: "<h2>Slide 3</h2>"}
		got := f.RenderSlide(s)
		want := `<section data-background-color="aquamarine"><h2>Slide 3</h2></section>` + "\n"
		if got != want {
			t.Errorf("RenderSlide():\ngot:  %q\nwant: %q", got, want)
		}
	})
}

func TestRenameSlide(t *testing.T) {
	t.Run("replace h1", func(t *testing.T) {
		s := format.Slide{Content: "<h1>Old Title</h1>\n<p>Body text</p>"}
		got := f.RenameSlide(s, "New Title")
		if !strings.Contains(got.Content, "New Title") {
			t.Errorf("expected h1 replaced, got %q", got.Content)
		}
		if strings.Contains(got.Content, "Old Title") {
			t.Errorf("old title still present: %q", got.Content)
		}
	})

	t.Run("replace h2", func(t *testing.T) {
		s := format.Slide{Content: "<h2>Old Sub</h2>\n<p>Body</p>"}
		got := f.RenameSlide(s, "New Sub")
		if !strings.Contains(got.Content, "New Sub") {
			t.Errorf("expected h2 replaced, got %q", got.Content)
		}
		if strings.Contains(got.Content, "Old Sub") {
			t.Errorf("old title still present: %q", got.Content)
		}
	})

	t.Run("prepend when no heading", func(t *testing.T) {
		s := format.Slide{Content: "<p>Just text</p>"}
		got := f.RenameSlide(s, "Added")
		if !strings.HasPrefix(got.Content, "<h1>Added</h1>") {
			t.Errorf("expected h1 prepended, got %q", got.Content)
		}
	})

	t.Run("h1 priority over h2", func(t *testing.T) {
		s := format.Slide{Content: "<h2>Sub</h2><h1>Main</h1>"}
		got := f.RenameSlide(s, "Replaced")
		if !strings.Contains(got.Content, "Replaced") {
			t.Errorf("expected h1 replaced, got %q", got.Content)
		}
		if !strings.Contains(got.Content, "Sub") {
			t.Errorf("h2 should be untouched, got %q", got.Content)
		}
	})
}

func TestNewSlide(t *testing.T) {
	got := f.NewSlide("Hello World")
	if !strings.Contains(got.Content, "<h1>Hello World</h1>") {
		t.Errorf("expected h1 heading, got %q", got.Content)
	}
}

func TestDeckTitle(t *testing.T) {
	tests := []struct {
		name string
		deck format.Deck
		want string
	}{
		{"title element", format.Deck{Metadata: "<html><head><title>My Talk</title></head><body>" + sentinel + "</body></html>"}, "My Talk"},
		{"collapses whitespace", format.Deck{Metadata: "<html><head><title>  spaced   out  </title></head><body></body></html>"}, "spaced out"},
		{"no title element", format.Deck{Metadata: "<html><head></head><body>" + sentinel + "</body></html>"}, ""},
		{"empty metadata", format.Deck{}, ""},
		{"first title wins", format.Deck{Metadata: "<html><head><title>First</title><title>Second</title></head><body></body></html>"}, "First"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := f.DeckTitle(tt.deck); got != tt.want {
				t.Errorf("DeckTitle() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestRenameDeck(t *testing.T) {
	t.Run("replace existing title", func(t *testing.T) {
		d := format.Deck{Metadata: "<html><head><title>Old</title></head><body>" + sentinel + "</body></html>"}
		got := f.RenameDeck(d, "New")
		if !strings.Contains(got.Metadata, "<title>New</title>") {
			t.Errorf("expected title replaced, got %q", got.Metadata)
		}
		if strings.Contains(got.Metadata, "Old") {
			t.Errorf("old title still present: %q", got.Metadata)
		}
	})

	t.Run("add when missing", func(t *testing.T) {
		d := format.Deck{Metadata: "<html><head></head><body>" + sentinel + "</body></html>"}
		got := f.RenameDeck(d, "Added")
		if !strings.Contains(got.Metadata, "<title>Added</title>") {
			t.Errorf("expected title added, got %q", got.Metadata)
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

	if dirs != 1 {
		t.Errorf("expected 1 dir, got %d", dirs)
	}
	if regular != 1 {
		t.Errorf("expected 1 file, got %d", regular)
	}
	for _, want := range []string{"images", "index.html"} {
		if !names[want] {
			t.Errorf("missing %q in scaffold output", want)
		}
	}

	for _, sf := range files {
		if sf.Path == "index.html" && !strings.Contains(sf.Content, "My Talk") {
			t.Errorf("index.html doesn't contain title: %q", sf.Content)
		}
	}
}

func TestPostScaffoldCmd(t *testing.T) {
	got := f.PostScaffoldCmd()
	if got != nil {
		t.Errorf("PostScaffoldCmd() = %v, want nil", got)
	}
}
