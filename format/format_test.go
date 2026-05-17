package format

import (
	"reflect"
	"testing"
)

func TestDeckInsert(t *testing.T) {
	mk := func(names ...string) []Slide {
		out := make([]Slide, len(names))
		for i, n := range names {
			out[i] = Slide{Content: n}
		}
		return out
	}

	names := func(slides []Slide) []string {
		out := make([]string, len(slides))
		for i, s := range slides {
			out[i] = s.Content
		}
		return out
	}

	tests := []struct {
		name     string
		existing []Slide
		pos      int
		insert   []Slide
		want     []string
	}{
		{"insert at start", mk("a", "b", "c"), 0, mk("x"), []string{"x", "a", "b", "c"}},
		{"insert in middle", mk("a", "b", "c"), 1, mk("x"), []string{"a", "x", "b", "c"}},
		{"insert at end", mk("a", "b", "c"), 3, mk("x"), []string{"a", "b", "c", "x"}},
		{"insert multiple", mk("a", "b"), 1, mk("x", "y"), []string{"a", "x", "y", "b"}},
		{"insert empty list", mk("a", "b"), 1, mk(), []string{"a", "b"}},
		{"insert into empty deck", mk(), 0, mk("x", "y"), []string{"x", "y"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := Deck{Metadata: "META", Slides: tt.existing}
			got := d.Insert(tt.pos, tt.insert)
			if g := names(got.Slides); !reflect.DeepEqual(g, tt.want) {
				t.Errorf("Insert: got %v, want %v", g, tt.want)
			}
			if got.Metadata != "META" {
				t.Errorf("Insert dropped metadata: got %q, want META", got.Metadata)
			}
		})
	}
}

func TestDeckInsertDoesNotMutateOriginal(t *testing.T) {
	original := Deck{
		Metadata: "M",
		Slides:   []Slide{{Content: "a"}, {Content: "b"}},
	}
	originalLen := len(original.Slides)
	originalFirst := original.Slides[0].Content

	_ = original.Insert(1, []Slide{{Content: "x"}})

	if len(original.Slides) != originalLen {
		t.Errorf("Insert mutated original length: got %d, want %d", len(original.Slides), originalLen)
	}
	if original.Slides[0].Content != originalFirst {
		t.Errorf("Insert mutated original element: got %q, want %q", original.Slides[0].Content, originalFirst)
	}
}
