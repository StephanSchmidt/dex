package format

// Slide represents a single slide in a presentation.
type Slide struct {
	Metadata string // format-specific metadata (empty string if none)
	Content  string // raw slide content
}

// Deck represents a complete presentation.
type Deck struct {
	Metadata string
	Slides   []Slide
}

// Insert returns a new deck with newSlides inserted before position pos (0-based).
func (d Deck) Insert(pos int, newSlides []Slide) Deck {
	result := make([]Slide, 0, len(d.Slides)+len(newSlides))
	result = append(result, d.Slides[:pos]...)
	result = append(result, newSlides...)
	result = append(result, d.Slides[pos:]...)
	return Deck{Metadata: d.Metadata, Slides: result}
}

// ScaffoldFile describes a file or directory to create when scaffolding a new presentation.
type ScaffoldFile struct {
	Path    string
	Content string
	IsDir   bool
}

// Format defines how a presentation file is parsed and rendered.
type Format interface {
	Parse(data []byte) Deck
	Render(d Deck) []byte
	ExtractTitle(s Slide) string
	DefaultFile() string
	RenderSlide(s Slide) string
	RenameSlide(s Slide, name string) Slide
	NewSlide(title string) Slide
	RenameDeck(d Deck, name string) Deck
	Scaffold(title, slug string) []ScaffoldFile
	PostScaffoldCmd() []string
}
