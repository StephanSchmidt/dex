package main

import (
	"fmt"
	"strings"
)

type RenameSlideCmd struct {
	Expr string `arg:"" help:"[dir/]range — e.g. 3, acme/1, 1,3, 1-3"`
	Name string `arg:"" help:"New title for the slide(s)."`
}

func (c *RenameSlideCmd) Run() error {
	dir, rangeExpr := splitDirRange(c.Expr)
	d, file, err := readDeck(dir)
	if err != nil {
		return err
	}

	indices, err := parseSliceExpr(rangeExpr, len(d.slides))
	if err != nil {
		return fmt.Errorf("%s: %w", c.Expr, err)
	}

	for _, idx := range indices {
		renameSlide(&d.slides[idx], c.Name)
	}

	return writeDeck(file, d)
}

func renameSlide(s *slide, name string) {
	lines := strings.Split(s.content, "\n")

	// Try replacing H1 heading first (matches extractTitle priority).
	for i, line := range lines {
		if headingRe.MatchString(line) {
			lines[i] = "# " + name
			s.content = strings.Join(lines, "\n")
			return
		}
	}

	// Fall back to title attribute.
	for i, line := range lines {
		if loc := titleAttrRe.FindStringIndex(line); loc != nil {
			lines[i] = line[:loc[0]] + `title="` + name + `"` + line[loc[1]:]
			s.content = strings.Join(lines, "\n")
			return
		}
	}

	// No title found; prepend an H1 heading.
	s.content = "# " + name + "\n" + s.content
}
