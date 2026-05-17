package main

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/StephanSchmidt/dex/format"
	"github.com/StephanSchmidt/dex/revealjs"
	"github.com/StephanSchmidt/dex/slidev"
	"github.com/spf13/afero"
)

type slide = format.Slide
type deck = format.Deck

var activeFormat format.Format = slidev.Format{}
var formatExplicit bool // true when --format was given

var formatRegistry = map[string]format.Format{
	"slidev":   slidev.Format{},
	"revealjs": revealjs.Format{},
}

// detectFormat returns the format matching a file extension.
func detectFormat(path string) format.Format {
	switch strings.ToLower(filepath.Ext(path)) {
	case ".html", ".htm":
		return revealjs.Format{}
	default:
		return slidev.Format{}
	}
}

// resolveFile turns a path into a presentation file path. If the path ends
// with "/", is empty (implicit current directory), or points to an existing
// directory, a default filename is appended. When the format was not
// explicitly chosen via --format, the directory is probed for any registered
// format's default file so e.g. `dex slides somedir/` works for a reveal.js
// deck (index.html) without needing --format.
func resolveFile(path string) string {
	dir := path
	switch {
	case dir == "":
		dir = "."
	case strings.HasSuffix(dir, "/"):
		// keep as-is
	default:
		info, err := appFs.Stat(dir)
		if err != nil || !info.IsDir() {
			return path
		}
	}

	// Probe candidates: active format first, then other registered formats
	// (unless --format was explicit, in which case stick to active).
	candidates := []string{activeFormat.DefaultFile()}
	if !formatExplicit {
		seen := map[string]bool{candidates[0]: true}
		for _, f := range formatRegistry {
			d := f.DefaultFile()
			if !seen[d] {
				candidates = append(candidates, d)
				seen[d] = true
			}
		}
	}
	for _, name := range candidates {
		c := filepath.Join(dir, name)
		if _, err := appFs.Stat(c); err == nil {
			return c
		}
	}
	// Nothing exists; fall back to active format's default so the caller
	// surfaces a sensible "no such file" error.
	return filepath.Join(dir, activeFormat.DefaultFile())
}

// readDeck resolves a directory-or-file path, reads it, and parses it.
// If the format was not explicitly set via --format, it auto-detects from
// the resolved file extension.
func readDeck(dirOrFile string) (deck, string, error) {
	file := resolveFile(dirOrFile)

	if !formatExplicit {
		activeFormat = detectFormat(file)
	}

	data, err := afero.ReadFile(appFs, file)
	if err != nil {
		return deck{}, "", fmt.Errorf("reading %s: %w", file, err)
	}
	return activeFormat.Parse(data), file, nil
}

// displayTitle returns the slide title, prefixed with "> " for detail slides.
func displayTitle(s slide) string {
	title := activeFormat.ExtractTitle(s)
	if strings.Contains(s.Metadata, "nav: detail") {
		return "> " + title
	}
	return title
}

// writeDeck renders a deck and writes it to the given file.
func writeDeck(file string, d deck) error {
	if err := afero.WriteFile(appFs, file, activeFormat.Render(d), 0o644); err != nil {
		return fmt.Errorf("writing %s: %w", file, err)
	}
	return nil
}
