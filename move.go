package main

import (
	"fmt"
	"strconv"
	"strings"
)

type MoveCmd struct {
	From string `arg:"" help:"1-based slide number to move (negative counts from end, e.g. -1 = last). Also accepts the combined form from:to, e.g. 10:end."`
	To   string `arg:"" optional:"" help:"1-based target position, +N/-N relative offset (e.g. +1, -2), or 'end' for the last position. Omit when using the combined from:to form."`
	File string `arg:"" optional:"" default:"" help:"Path to slides file or its parent directory."`
}

func (c *MoveCmd) Run() error {
	fromArg, toArg, fileArg := c.From, c.To, c.File
	// Allow the combined "from:to" form (e.g. "10:end" or "10:-1"), which
	// also dodges Kong's flag parsing for targets like -1.
	if i := strings.Index(fromArg, ":"); i >= 0 {
		if toArg != "" && fileArg == "" {
			fileArg = toArg // shift: the slot that was To is actually the file
		} else if toArg != "" && fileArg != "" {
			return fmt.Errorf("cannot use from:to form together with a separate to argument")
		}
		fromArg, toArg = fromArg[:i], fromArg[i+1:]
	}
	if toArg == "" {
		return fmt.Errorf("missing target position")
	}

	d, file, err := readDeck(fileArg)
	if err != nil {
		return err
	}

	n := len(d.Slides)
	fromIdx, err := resolveIndex(fromArg, n)
	if err != nil {
		return fmt.Errorf("from: %w", err)
	}
	fromPos := fromIdx + 1 // 1-based for parseTarget and error messages

	to, err := parseTarget(toArg, fromPos, n)
	if err != nil {
		return err
	}

	if to < 1 || to > n+1 {
		return fmt.Errorf("to index %d out of range (1..%d)", to, n+1)
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
// (e.g. "4"), a relative offset prefixed with + or - (e.g. "+2", "-1"), or
// the keyword "end" (optionally with a -N offset, e.g. "end-1") to refer to
// a position relative to the end of the deck. n is the total slide count.
func parseTarget(s string, from, n int) (int, error) {
	s = strings.TrimSpace(s)
	if s == "end" {
		return n + 1, nil
	}
	if rest, ok := strings.CutPrefix(s, "end-"); ok {
		offset, err := strconv.Atoi(rest)
		if err != nil || offset < 0 {
			return 0, fmt.Errorf("invalid end offset %q", s)
		}
		return n + 1 - offset, nil
	}
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
