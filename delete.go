package main

import "fmt"

type DeleteCmd struct {
	Expr string `arg:"" help:"[dir/]range — e.g. 3, acme/1, 1,3, 1-3"`
}

func (c *DeleteCmd) Run() error {
	dir, rangeExpr := splitDirRange(c.Expr)
	d, file, err := readDeck(dir)
	if err != nil {
		return err
	}

	indices, err := parseSliceExpr(rangeExpr, len(d.Slides))
	if err != nil {
		return fmt.Errorf("%s: %w", c.Expr, err)
	}

	remove := make(map[int]bool, len(indices))
	for _, i := range indices {
		remove[i] = true
	}

	if len(remove) >= len(d.Slides) {
		return fmt.Errorf("cannot delete all %d slides", len(d.Slides))
	}

	kept := make([]slide, 0, len(d.Slides)-len(remove))
	for i, s := range d.Slides {
		if !remove[i] {
			kept = append(kept, s)
		}
	}
	d.Slides = kept

	return writeDeck(file, d)
}
