package main

import "fmt"

type MoveCmd struct {
	From int    `arg:"" help:"1-based slide number to move."`
	To   int    `arg:"" help:"1-based target position (slide is placed here, others shift)."`
	File string `arg:"" optional:"" default:"slides.md" help:"Path to slides.md file or its parent directory."`
}

func (c *MoveCmd) Run() error {
	d, file, err := readDeck(c.File)
	if err != nil {
		return err
	}

	n := len(d.slides)
	if c.From < 1 || c.From > n {
		return fmt.Errorf("from index %d out of range (1..%d)", c.From, n)
	}
	if c.To < 1 || c.To > n {
		return fmt.Errorf("to index %d out of range (1..%d)", c.To, n)
	}
	if c.From == c.To {
		return fmt.Errorf("from and to are the same (%d)", c.From)
	}

	from := c.From - 1
	to := c.To - 1

	// Remove slide at from.
	s := d.slides[from]
	d.slides = append(d.slides[:from], d.slides[from+1:]...)

	// Adjust target: if from was before to, decrement to account for removal.
	if from < to {
		to--
	}

	return writeDeck(file, d.insert(to, []slide{s}))
}
