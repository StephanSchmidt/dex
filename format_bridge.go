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
// with "/" or points to an existing directory, the default filename is appended.
func resolveFile(path string) string {
	if path == "" {
		return activeFormat.DefaultFile()
	}
	if strings.HasSuffix(path, "/") {
		return filepath.Join(path, activeFormat.DefaultFile())
	}
	if info, err := appFs.Stat(path); err == nil && info.IsDir() {
		return filepath.Join(path, activeFormat.DefaultFile())
	}
	return path
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
