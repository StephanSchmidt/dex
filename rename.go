package main

import "strings"

type RenameCmd struct {
	Name string `arg:"" help:"New title for the deck."`
	File string `arg:"" optional:"" default:"slides.md" help:"Path to slides.md file or its parent directory."`
}

func (c *RenameCmd) Run() error {
	d, file, err := readDeck(c.File)
	if err != nil {
		return err
	}

	lines := strings.Split(d.frontmatter, "\n")
	replaced := false
	for i, line := range lines {
		if strings.HasPrefix(line, "title:") {
			lines[i] = "title: " + c.Name
			replaced = true
			break
		}
	}
	if replaced {
		d.frontmatter = strings.Join(lines, "\n")
	} else {
		d.frontmatter += "title: " + c.Name + "\n"
	}

	return writeDeck(file, d)
}
