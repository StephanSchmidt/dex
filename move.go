package main

import (
	"fmt"
	"strconv"
	"strings"
)

type MoveCmd struct {
	From int    `arg:"" help:"1-based slide number to move."`
	To   string `arg:"" help:"1-based target position, or +N/-N relative offset (e.g. +1, -2)."`
	File string `arg:"" optional:"" default:"" help:"Path to slides file or its parent directory."`
}

func (c *MoveCmd) Run() error {
	d, file, err := readDeck(c.File)
	if err != nil {
		return err
	}

	n := len(d.Slides)
	if c.From < 1 || c.From > n {
		return fmt.Errorf("from index %d out of range (1..%d)", c.From, n)
	}

	to, err := parseTarget(c.To, c.From)
	if err != nil {
		return err
	}

	if to < 1 || to > n {
		return fmt.Errorf("to index %d out of range (1..%d)", to, n)
	}
	if c.From == to {
		return fmt.Errorf("from and to are the same (%d)", c.From)
	}

	from := c.From - 1
	toIdx := to - 1

	// Remove slide at from.
	s := d.Slides[from]
	d.Slides = append(d.Slides[:from], d.Slides[from+1:]...)

	// Adjust target: if from was before to, decrement to account for removal.
	if from < toIdx {
		toIdx--
	}

	return writeDeck(file, d.Insert(toIdx, []slide{s}))
}

// parseTarget parses a target position string. It accepts an absolute number
// (e.g. "4") or a relative offset prefixed with + or - (e.g. "+2", "-1").
func parseTarget(s string, from int) (int, error) {
	s = strings.TrimSpace(s)
	if strings.HasPrefix(s, "+") || strings.HasPrefix(s, "-") {
		offset, err := strconv.Atoi(s)
		if err != nil {
			return 0, fmt.Errorf("invalid relative offset %q", s)
		}
		if offset == 0 {
			return 0, fmt.Errorf("relative offset must not be zero")
		}
		return from + offset, nil
	}
	v, err := strconv.Atoi(s)
	if err != nil {
		return 0, fmt.Errorf("invalid target position %q", s)
	}
	return v, nil
}
