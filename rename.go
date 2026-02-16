package main

type RenameCmd struct {
	Name string `arg:"" help:"New title for the deck."`
	File string `arg:"" optional:"" default:"" help:"Path to slides file or its parent directory."`
}

func (c *RenameCmd) Run() error {
	d, file, err := readDeck(c.File)
	if err != nil {
		return err
	}

	d = activeFormat.RenameDeck(d, c.Name)

	return writeDeck(file, d)
}
