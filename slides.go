package main

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/spf13/afero"
)

type SlidesCmd struct {
	Exprs []string `arg:"" optional:"" help:"[dir/][range] — e.g. acme/, acme/1-5, talks/* for all decks under talks/"`
}

func (c *SlidesCmd) Run() error {
	exprs := c.Exprs
	if len(exprs) == 0 {
		exprs = []string{""}
	}

	expanded, err := expandSlideExprs(exprs)
	if err != nil {
		return err
	}

	for i, expr := range expanded {
		if i > 0 {
			fmt.Fprintln(stdout)
		}
		dir, rangeExpr := splitDirRange(expr)
		d, _, err := readDeck(dir)
		if err != nil {
			return err
		}

		if title := activeFormat.DeckTitle(d); title != "" {
			fmt.Fprintf(stdout, "%s\n", title)
		}

		if rangeExpr == "" {
			for j, s := range d.Slides {
				fmt.Fprintf(stdout, "%3d  %s\n", j+1, displayTitle(s))
			}
		} else {
			indices, err := parseSliceExpr(rangeExpr, len(d.Slides))
			if err != nil {
				return fmt.Errorf("%s: %w", expr, err)
			}
			for _, idx := range indices {
				fmt.Fprintf(stdout, "%3d  %s\n", idx+1, displayTitle(d.Slides[idx]))
			}
		}
	}
	return nil
}

// expandSlideExprs expands any glob patterns and filters expanded matches to
// presentation-like paths. Non-glob exprs (including the empty string for the
// implicit current directory) pass through unchanged.
func expandSlideExprs(exprs []string) ([]string, error) {
	var out []string
	for _, expr := range exprs {
		if !strings.ContainsAny(expr, "*?[") {
			out = append(out, expr)
			continue
		}
		matches, err := afero.Glob(appFs, expr)
		if err != nil {
			return nil, fmt.Errorf("glob %q: %w", expr, err)
		}
		for _, m := range matches {
			if isDeckPath(m) {
				out = append(out, m)
			}
		}
	}
	return out, nil
}

// isDeckPath reports whether a path looks like a presentation: either a
// directory containing slides.md / index.html, or one of those convention
// files directly. Arbitrary .md / .html files are not treated as decks so
// that `dex slides talks/*` skips notes/README files alongside real decks.
func isDeckPath(path string) bool {
	info, err := appFs.Stat(path)
	if err != nil {
		return false
	}
	if !info.IsDir() {
		base := strings.ToLower(filepath.Base(path))
		return base == "slides.md" || base == "index.html"
	}
	for _, name := range []string{"slides.md", "index.html"} {
		if _, err := appFs.Stat(filepath.Join(path, name)); err == nil {
			return true
		}
	}
	return false
}
