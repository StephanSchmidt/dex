package main

import "fmt"

type SlidesCmd struct {
	Exprs []string `arg:"" optional:"" help:"[dir/][range] — e.g. acme/, acme/1-5"`
}

func (c *SlidesCmd) Run() error {
	exprs := c.Exprs
	if len(exprs) == 0 {
		exprs = []string{""}
	}

	for _, expr := range exprs {
		dir, rangeExpr := splitDirRange(expr)
		d, _, err := readDeck(dir)
		if err != nil {
			return err
		}

		if title := activeFormat.DeckTitle(d); title != "" {
			fmt.Fprintf(stdout, "%s\n", title)
		}

		if rangeExpr == "" {
			for i, s := range d.Slides {
				fmt.Fprintf(stdout, "%3d  %s\n", i+1, displayTitle(s))
			}
		} else {
			indices, err := parseSliceExpr(rangeExpr, len(d.Slides))
			if err != nil {
				return fmt.Errorf("%s: %w", expr, err)
			}
			for _, idx := range indices {
				fmt.Fprintf(stdout, "%3d  %s\n", idx+1, displayTitle(d.Slides[idx]))
			}
		}
	}
	return nil
}
