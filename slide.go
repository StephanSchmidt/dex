package main

import "fmt"

type SlideCmd struct {
	Exprs []string `arg:"" required:"" help:"[dir/]range — e.g. 1-3, acme/1, dir1/1:3 dir2/5"`
}

func (c *SlideCmd) Run() error {
	for _, expr := range c.Exprs {
		dir, rangeExpr := splitDirRange(expr)
		d, _, err := readDeck(dir)
		if err != nil {
			return err
		}

		indices, err := parseSliceExpr(rangeExpr, len(d.slides))
		if err != nil {
			return fmt.Errorf("%s: %w", expr, err)
		}

		for _, idx := range indices {
			s := d.slides[idx]
			fmt.Fprint(stdout, "---\n")
			if s.frontmatter != "" {
				fmt.Fprint(stdout, s.frontmatter)
				fmt.Fprint(stdout, "---\n")
			}
			fmt.Fprint(stdout, s.content)
		}
	}
	return nil
}
