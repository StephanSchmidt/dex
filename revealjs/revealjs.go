package revealjs

import (
	"html"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/StephanSchmidt/dex/format"
)

const sentinel = "<!--DEX:SLIDES-->"

// Format implements format.Format for reveal.js HTML presentations.
type Format struct{}

// Parse splits a reveal.js HTML file into its deck metadata (the full
// HTML shell with a sentinel where slides go) and individual slides.
func (Format) Parse(data []byte) format.Deck {
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(string(data)))
	if err != nil {
		return format.Deck{}
	}

	slidesContainer := doc.Find("div.slides")
	if slidesContainer.Length() == 0 {
		slidesContainer = doc.Find(".slides")
	}

	var slides []format.Slide
	slidesContainer.Children().Each(func(_ int, s *goquery.Selection) {
		if !s.Is("section") {
			return
		}
		// Extract attributes as metadata.
		var attrs []string
		for _, attr := range s.Get(0).Attr {
			attrs = append(attrs, attr.Key+`="`+attr.Val+`"`)
		}
		fm := strings.Join(attrs, " ")

		content, _ := s.Html()
		slides = append(slides, format.Slide{
			Metadata: fm,
			Content:  content,
		})
	})

	// Replace slides content with sentinel.
	slidesContainer.SetHtml(sentinel)
	shell, _ := doc.Html()

	return format.Deck{
		Metadata: shell,
		Slides:   slides,
	}
}

// Render reconstructs a valid reveal.js HTML file from its parts.
func (Format) Render(d format.Deck) []byte {
	var sections strings.Builder
	for _, s := range d.Slides {
		sections.WriteString(renderSection(s))
	}
	result := strings.Replace(d.Metadata, sentinel, sections.String(), 1)
	return []byte(result)
}

// ExtractTitle returns the title from slide content, looking for h1
// first, then h2.
func (Format) ExtractTitle(s format.Slide) string {
	doc, err := goquery.NewDocumentFromReader(strings.NewReader("<div>" + s.Content + "</div>"))
	if err != nil {
		return "(untitled)"
	}
	if h1 := doc.Find("h1").First(); h1.Length() > 0 {
		return h1.Text()
	}
	if h2 := doc.Find("h2").First(); h2.Length() > 0 {
		return h2.Text()
	}
	return "(untitled)"
}

// DefaultFile returns the default presentation filename for reveal.js.
func (Format) DefaultFile() string {
	return "index.html"
}

// RenderSlide outputs a single slide as a reveal.js <section>.
func (Format) RenderSlide(s format.Slide) string {
	return renderSection(s)
}

// RenameSlide returns a copy of the slide with its title changed to name.
func (Format) RenameSlide(s format.Slide, name string) format.Slide {
	doc, err := goquery.NewDocumentFromReader(strings.NewReader("<div>" + s.Content + "</div>"))
	if err != nil {
		s.Content = "<h1>" + html.EscapeString(name) + "</h1>\n" + s.Content
		return s
	}

	if h1 := doc.Find("h1").First(); h1.Length() > 0 {
		h1.SetText(name)
		result, _ := doc.Find("div").First().Html()
		s.Content = result
		return s
	}

	if h2 := doc.Find("h2").First(); h2.Length() > 0 {
		h2.SetText(name)
		result, _ := doc.Find("div").First().Html()
		s.Content = result
		return s
	}

	s.Content = "<h1>" + html.EscapeString(name) + "</h1>\n" + s.Content
	return s
}

// NewSlide returns a new slide with an h1 heading for the given title.
func (Format) NewSlide(title string) format.Slide {
	return format.Slide{Content: "<h1>" + html.EscapeString(title) + "</h1>\n"}
}

// RenameDeck returns a copy of the deck with the <title> element updated.
func (Format) RenameDeck(d format.Deck, name string) format.Deck {
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(d.Metadata))
	if err != nil {
		return d
	}

	titleEl := doc.Find("title").First()
	if titleEl.Length() > 0 {
		titleEl.SetText(name)
	} else {
		doc.Find("head").AppendHtml("<title>" + html.EscapeString(name) + "</title>")
	}

	d.Metadata, _ = doc.Html()
	return d
}

// Scaffold returns the list of files and directories needed for a new
// reveal.js presentation.
func (Format) Scaffold(title, slug string) []format.ScaffoldFile {
	content := strings.ReplaceAll(indexHTMLTemplate, "{{TITLE}}", html.EscapeString(title))
	return []format.ScaffoldFile{
		{Path: "images", IsDir: true},
		{Path: "index.html", Content: content},
	}
}

// PostScaffoldCmd returns nil since the CDN-based template needs no install step.
func (Format) PostScaffoldCmd() []string {
	return nil
}

func renderSection(s format.Slide) string {
	var buf strings.Builder
	buf.WriteString("<section")
	if s.Metadata != "" {
		buf.WriteString(" ")
		buf.WriteString(s.Metadata)
	}
	buf.WriteString(">")
	buf.WriteString(s.Content)
	buf.WriteString("</section>\n")
	return buf.String()
}

const indexHTMLTemplate = `<!doctype html>
<html lang="en">
<head>
<meta charset="utf-8">
<meta name="viewport" content="width=device-width, initial-scale=1.0">
<title>{{TITLE}}</title>
<link rel="stylesheet" href="https://cdn.jsdelivr.net/npm/reveal.js@5/dist/reset.css">
<link rel="stylesheet" href="https://cdn.jsdelivr.net/npm/reveal.js@5/dist/reveal.css">
<link rel="stylesheet" href="https://cdn.jsdelivr.net/npm/reveal.js@5/dist/theme/black.css">
</head>
<body>
<div class="reveal">
<div class="slides">
<section><h1>{{TITLE}}</h1></section>
</div>
</div>
<script src="https://cdn.jsdelivr.net/npm/reveal.js@5/dist/reveal.js"></script>
<script>Reveal.initialize();</script>
</body>
</html>
`
