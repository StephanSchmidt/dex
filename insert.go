package main

import (
	"fmt"
	"strconv"
)

type InsertCmd struct {
	Pos   string `arg:"" help:"[dir/]N — insert before slide N (1-based, N+1 appends). E.g. 3, acme/2."`
	Title string `arg:"" help:"Title for the new slide (becomes an H1 heading)."`
}

func (c *InsertCmd) Run() error {
	dir, posStr := splitDirRange(c.Pos)
	d, file, err := readDeck(dir)
	if err != nil {
		return err
	}

	pos, err := strconv.Atoi(posStr)
	if err != nil {
		return fmt.Errorf("invalid position %q: %w", posStr, err)
	}
	if pos < 1 || pos > len(d.slides)+1 {
		return fmt.Errorf("position %d out of range (1..%d)", pos, len(d.slides)+1)
	}

	newSlide := slide{content: "\n# " + c.Title + "\n\n"}
	return writeDeck(file, d.insert(pos-1, []slide{newSlide}))
}
