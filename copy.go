package main

import "fmt"

type CopyCmd struct {
	Source string `arg:"" help:"Source slides as [dir/]range (e.g. dir1/1-3, 2,4)."`
	Target string `arg:"" help:"Insert before this position in target deck as [dir/]N (1-based, N+1 appends, -1 appends). E.g. dir2/5, -1, 3."`
}

func (c *CopyCmd) Run() error {
	// Parse source.
	srcDir, srcRange := splitDirRange(c.Source)
	src, srcFile, err := readDeck(srcDir)
	if err != nil {
		return err
	}
	srcFmt := activeFormat
	indices, err := parseSliceExpr(srcRange, len(src.Slides))
	if err != nil {
		return fmt.Errorf("%s: %w", c.Source, err)
	}

	// Collect source slides before touching the target.
	copied := make([]slide, len(indices))
	for i, idx := range indices {
		copied[i] = src.Slides[idx]
	}

	// Parse target.
	tgtDir, posStr := splitDirRange(c.Target)
	tgt, tgtFile, err := readDeck(tgtDir)
	if err != nil {
		return err
	}
	tgtFmt := activeFormat
	if srcFmt != tgtFmt {
		return fmt.Errorf("cannot copy across formats: %s is %T but %s is %T", srcFile, srcFmt, tgtFile, tgtFmt)
	}

	idx, err := resolveInsertIndex(posStr, len(tgt.Slides))
	if err != nil {
		return fmt.Errorf("invalid target position %q: %w", posStr, err)
	}

	return writeDeck(tgtFile, tgt.Insert(idx, copied))
}
