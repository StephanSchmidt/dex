package main

import (
	"fmt"
	"io"
	"os"

	"github.com/alecthomas/kong"
	"github.com/spf13/afero"
)

var version = "dev"

var appFs afero.Fs = afero.NewOsFs()
var stdout io.Writer = os.Stdout

var cli struct {
	Format      string           `help:"Presentation format (slidev, revealjs). Auto-detected from file extension if omitted." enum:"slidev,revealjs," default:""`
	Version     kong.VersionFlag `help:"Print version and exit." short:"v"`
	New         NewCmd           `cmd:"" help:"Scaffold a new presentation directory."`
	Copy        CopyCmd          `cmd:"" help:"Copy slides from source deck into target deck before a given position. Source is unchanged."`
	Delete      DeleteCmd        `cmd:"" help:"Delete one or more slides from a deck."`
	Insert      InsertCmd        `cmd:"" help:"Insert a new blank slide with the given title at a position."`
	Move        MoveCmd          `cmd:"" help:"Move a slide to a different position within one deck file."`
	UI          UICmd            `cmd:"" help:"Interactive TUI for reordering slides."`
	Rename      RenameCmd        `cmd:"" help:"Rename the deck title in metadata."`
	RenameSlide RenameSlideCmd   `cmd:"rename-slide" help:"Rename a slide's title."`
	Slide       SlideCmd         `cmd:"" help:"Print raw slide content to stdout (read-only)."`
	Slides      SlidesCmd        `cmd:"" help:"List slide numbers and titles to stdout (read-only)."`
	Swap        SwapCmd          `cmd:"" help:"Swap two slides (same or different decks)."`
}

func main() {
	ctx := kong.Parse(&cli,
		kong.Name("dex"),
		kong.Vars{"version": version},
		kong.Description(`Manipulate slide presentations (Slidev markdown or reveal.js HTML).
Format is auto-detected from file extension (.md → Slidev, .html → reveal.js).
Use --format to override. Each presentation is a directory with a default file.

Addressing — [dir/]range:
  3            single slide
  1-3          slides 1 through 3
  1,2,7        specific slides
  -1           last slide
  1:-1         all slides
  dir/1-3      slides 1-3 from dir/
  ../dir/2,4   relative path

Examples:
  dex slides                    list all slides in ./slides.md
  dex slides acme/              list slides in acme/
  dex slides 1-5                list only slides 1 through 5
  dex slide 3                   print raw content of slide 3
  dex slide dir1/1-3            print slides 1-3 from dir1
  dex copy dir1/1-3 dir2/5      copy slides 1-3 from dir1, insert before slide 5 in dir2
  dex copy 1 3                  duplicate slide 1 before slide 3 (same file)
  dex copy dir1/2 dir2/4        append to dir2 (position 4 = after last of 3 slides)
  dex delete 3                  delete slide 3
  dex delete acme/1,3           delete slides 1 and 3 from acme/
  dex insert 3 "New Slide"      insert a new slide before slide 3
  dex insert -1 "End"           append a new slide at the end
  dex insert acme/2 "Intro"     insert a new slide at position 2 in acme/
  dex move 2 4                  move slide 2 to position 4
  dex move -1 1                 move last slide to the front
  dex move 2 +1                 move slide 2 down by one position
  dex move 4 -2                 move slide 4 up by two positions
  dex move 3 1 acme/            move slide 3 to the front in acme/
  dex ui                        interactive slide reordering
  dex ui acme/                  reorder slides in acme/
  dex rename "New Title"           rename the deck title
  dex rename "New Title" acme/     rename deck in acme/
  dex rename-slide 3 "New Title"   rename slide 3
  dex rename-slide acme/3 "Hi"    rename slide 3 in acme/
  dex rename-slide 1,3 "Same"     rename slides 1 and 3
  dex swap 1 3                  swap slides 1 and 3
  dex swap acme/2 other/1       swap slide 2 in acme/ with slide 1 in other/
  dex new "My Talk"             scaffold my-talk/ (Slidev)
  dex new --format revealjs "My Talk"  scaffold my-talk/ (reveal.js)`),
		kong.UsageOnError(),
	)

	if cli.Format != "" {
		if f, ok := formatRegistry[cli.Format]; ok {
			activeFormat = f
			formatExplicit = true
		}
	}

	if err := ctx.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
