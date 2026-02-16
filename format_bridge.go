package main

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/StephanSchmidt/dex/format"
	"github.com/StephanSchmidt/dex/slidev"
	"github.com/spf13/afero"
)

type slide = format.Slide
type deck = format.Deck

var activeFormat format.Format = slidev.Format{}

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
func readDeck(dirOrFile string) (deck, string, error) {
	file := resolveFile(dirOrFile)
	data, err := afero.ReadFile(appFs, file)
	if err != nil {
		return deck{}, "", fmt.Errorf("reading %s: %w", file, err)
	}
	return activeFormat.Parse(data), file, nil
}

// writeDeck renders a deck and writes it to the given file.
func writeDeck(file string, d deck) error {
	if err := afero.WriteFile(appFs, file, activeFormat.Render(d), 0o644); err != nil {
		return fmt.Errorf("writing %s: %w", file, err)
	}
	return nil
}
