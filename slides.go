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

		if rangeExpr == "" {
			for i, s := range d.slides {
				fmt.Fprintf(stdout, "%3d  %s\n", i+1, extractTitle(s.content))
			}
		} else {
			indices, err := parseSliceExpr(rangeExpr, len(d.slides))
			if err != nil {
				return fmt.Errorf("%s: %w", expr, err)
			}
			for _, idx := range indices {
				fmt.Fprintf(stdout, "%3d  %s\n", idx+1, extractTitle(d.slides[idx].content))
			}
		}
	}
	return nil
}
