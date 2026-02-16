package main

import "fmt"

type SwapCmd struct {
	A string `arg:"" help:"First slide as [dir/]N (1-based)."`
	B string `arg:"" help:"Second slide as [dir/]N (1-based)."`
}

func (c *SwapCmd) Run() error {
	dirA, rangeA := splitDirRange(c.A)
	dirB, rangeB := splitDirRange(c.B)

	dA, fileA, err := readDeck(dirA)
	if err != nil {
		return err
	}
	idxA, err := parseSliceExpr(rangeA, len(dA.Slides))
	if err != nil {
		return fmt.Errorf("%s: %w", c.A, err)
	}
	if len(idxA) != 1 {
		return fmt.Errorf("%s: swap requires a single slide, got %d", c.A, len(idxA))
	}

	dB, fileB, err := readDeck(dirB)
	if err != nil {
		return err
	}
	idxB, err := parseSliceExpr(rangeB, len(dB.Slides))
	if err != nil {
		return fmt.Errorf("%s: %w", c.B, err)
	}
	if len(idxB) != 1 {
		return fmt.Errorf("%s: swap requires a single slide, got %d", c.B, len(idxB))
	}

	a, b := idxA[0], idxB[0]

	if fileA == fileB {
		if a == b {
			return fmt.Errorf("cannot swap a slide with itself")
		}
		dA.Slides[a], dA.Slides[b] = dA.Slides[b], dA.Slides[a]
		return writeDeck(fileA, dA)
	}

	dA.Slides[a], dB.Slides[b] = dB.Slides[b], dA.Slides[a]
	if err := writeDeck(fileA, dA); err != nil {
		return err
	}
	return writeDeck(fileB, dB)
}
