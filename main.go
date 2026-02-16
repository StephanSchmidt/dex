package main

import (
	"fmt"
	"io"
	"os"

	"github.com/alecthomas/kong"
	"github.com/spf13/afero"
)

var appFs afero.Fs = afero.NewOsFs()
var stdout io.Writer = os.Stdout

var cli struct {
	New         NewCmd         `cmd:"" help:"Scaffold a new presentation directory."`
	Copy        CopyCmd        `cmd:"" help:"Copy slides from source deck into target deck before a given position. Source is unchanged."`
	Delete      DeleteCmd      `cmd:"" help:"Delete one or more slides from a deck."`
	Move        MoveCmd        `cmd:"" help:"Move a slide to a different position within one deck file."`
	Rename      RenameCmd      `cmd:"" help:"Rename the deck title in frontmatter."`
	RenameSlide RenameSlideCmd `cmd:"rename-slide" help:"Rename a slide's title."`
	Slide       SlideCmd       `cmd:"" help:"Print raw slide markdown to stdout (read-only)."`
	Slides      SlidesCmd      `cmd:"" help:"List slide numbers and titles to stdout (read-only)."`
}

func main() {
	ctx := kong.Parse(&cli,
		kong.Name("dex"),
		kong.Description(`Manipulate Slidev markdown presentations.
Each presentation is a directory with a slides.md file.
"dir/" or "dir" resolves to "dir/slides.md"; bare ranges use ./slides.md.

Addressing — [dir/]range:
  3            single slide
  1-3          slides 1 through 3
  1,2,7        specific slides
  -1           last slide
  1:-1         all slides
  dir/1-3      slides 1-3 from dir/slides.md
  ../dir/2,4   relative path

Examples:
  dex slides                    list all slides in ./slides.md
  dex slides acme/              list slides in acme/slides.md
  dex slides 1-5                list only slides 1 through 5
  dex slide 3                   print raw markdown of slide 3
  dex slide dir1/1-3            print slides 1-3 from dir1
  dex copy dir1/1-3 dir2/5      copy slides 1-3 from dir1, insert before slide 5 in dir2
  dex copy 1 3                  duplicate slide 1 before slide 3 (same file)
  dex copy dir1/2 dir2/4        append to dir2 (position 4 = after last of 3 slides)
  dex delete 3                  delete slide 3 from ./slides.md
  dex delete acme/1,3           delete slides 1 and 3 from acme/slides.md
  dex move 2 4                  move slide 2 to position 4 in ./slides.md
  dex move 3 1 acme/            move slide 3 to the front in acme/slides.md
  dex rename "New Title"           rename the deck title in frontmatter
  dex rename "New Title" acme/     rename deck in acme/slides.md
  dex rename-slide 3 "New Title"   rename slide 3
  dex rename-slide acme/3 "Hi"    rename slide 3 in acme/slides.md
  dex rename-slide 1,3 "Same"     rename slides 1 and 3
  dex new "My Talk"             scaffold my-talk/`),
		kong.UsageOnError(),
	)
	if err := ctx.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
