package main

import "fmt"

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

	indices, err := parseSliceExpr(rangeExpr, len(d.Slides))
	if err != nil {
		return fmt.Errorf("%s: %w", c.Expr, err)
	}

	for _, idx := range indices {
		d.Slides[idx] = activeFormat.RenameSlide(d.Slides[idx], c.Name)
	}

	return writeDeck(file, d)
}
