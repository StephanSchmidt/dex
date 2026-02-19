package main

import "fmt"

type InsertCmd struct {
	Pos   string `arg:"" help:"[dir/]N — insert before slide N (1-based, N+1 appends, -1 appends). E.g. 3, -1, acme/2."`
	Title string `arg:"" help:"Title for the new slide (becomes an H1 heading)."`
}

func (c *InsertCmd) Run() error {
	dir, posStr := splitDirRange(c.Pos)
	d, file, err := readDeck(dir)
	if err != nil {
		return err
	}

	idx, err := resolveInsertIndex(posStr, len(d.Slides))
	if err != nil {
		return fmt.Errorf("invalid position %q: %w", posStr, err)
	}

	newSlide := activeFormat.NewSlide(c.Title)
	return writeDeck(file, d.Insert(idx, []slide{newSlide}))
}
