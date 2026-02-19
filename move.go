package main

import (
	"fmt"
	"strconv"
	"strings"
)

type MoveCmd struct {
	From string `arg:"" help:"1-based slide number to move (negative counts from end, e.g. -1 = last)."`
	To   string `arg:"" help:"1-based target position, or +N/-N relative offset (e.g. +1, -2)."`
	File string `arg:"" optional:"" default:"" help:"Path to slides file or its parent directory."`
}

func (c *MoveCmd) Run() error {
	d, file, err := readDeck(c.File)
	if err != nil {
		return err
	}

	n := len(d.Slides)
	fromIdx, err := resolveIndex(c.From, n)
	if err != nil {
		return fmt.Errorf("from: %w", err)
	}
	fromPos := fromIdx + 1 // 1-based for parseTarget and error messages

	to, err := parseTarget(c.To, fromPos)
	if err != nil {
		return err
	}

	if to < 1 || to > n {
		return fmt.Errorf("to index %d out of range (1..%d)", to, n)
	}
	if fromPos == to {
		return fmt.Errorf("from and to are the same (%d)", fromPos)
	}

	from := fromIdx
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
