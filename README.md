<p align="center">
  <img src="assets/dex-logo.svg" alt="dex" width="360">
</p>

# dex

A CLI tool for manipulating slide presentations ([Slidev](https://sli.dev) markdown and [reveal.js](https://revealjs.com) HTML).

## Features

| Command | Description | Example |
|---------|-------------|---------|
| `slides` | List slide numbers and titles | `dex slides acme/` |
| `slide` | Print raw slide content | `dex slide 3` |
| `copy` | Copy slides between decks | `dex copy dir1/1-3 dir2/5` |
| `delete` | Delete slides from a deck | `dex delete 3` |
| `insert` | Insert a new blank slide at a position | `dex insert 3 "New Slide"`, `dex insert -1 "End"` |
| `move` | Reorder slides within a deck | `dex move 2 4`, `dex move -1 1` |
| `rename` | Rename the deck title | `dex rename "New Title"` |
| `rename-slide` | Rename a slide's title | `dex rename-slide 3 "Hi"` |
| `swap` | Swap two slides (same or different decks) | `dex swap 1 3` |
| `ui` | Interactive TUI for reordering slides (lipgloss styled) | `dex ui` |
| `new` | Scaffold a new presentation directory | `dex new "My Talk"` |

## Installation

### Homebrew

```sh
brew tap StephanSchmidt/dex
brew install dex
```

### Go

```sh
go install github.com/StephanSchmidt/dex@latest
```

## Usage

Each presentation is a directory containing a `slides.md` (Slidev) or `index.html` (reveal.js) file.
The format is auto-detected from the file extension; use `--format` to override.
`dir/` or `dir` resolves to the default file for the active format; bare ranges use the current directory.

### Addressing

```
3            single slide
1-3          slides 1 through 3
1,2,7        specific slides
-1           last slide
-2           second-to-last slide
1:-1         all slides
dir/1-3      slides 1-3 from dir/slides.md
```

Negative positions work everywhere — for insert/copy targets, `-1` means append at the end, `-2` means before the last slide, etc.

### Examples

```sh
dex slides                  # list all slides in ./slides.md
dex slides acme/            # list slides in acme/slides.md
dex slide 3                 # print raw markdown of slide 3
dex copy dir1/1-3 dir2/5    # copy slides 1-3 from dir1, insert before slide 5 in dir2
dex delete 3                # delete slide 3
dex delete acme/1,3         # delete slides 1 and 3 from acme/
dex insert 3 "New Slide"    # insert a new slide before slide 3
dex insert -1 "End"         # append a new slide at the end
dex insert acme/2 "Intro"   # insert at position 2 in acme/
dex move 2 4                # move slide 2 to position 4
dex move -1 1               # move last slide to the front
dex move 2 +1               # move slide 2 down by one position
dex move 4 -2               # move slide 4 up by two positions
dex rename "New Title"          # rename the deck title in metadata
dex rename "New Title" acme/    # rename deck in acme/slides.md
dex rename-slide 3 "New Title"  # rename slide 3
dex rename-slide acme/3 "Hi"   # rename slide 3 in acme/slides.md
dex rename-slide 1,3 "Same"   # rename slides 1 and 3
dex swap 1 3                # swap slides 1 and 3
dex swap acme/2 other/1     # swap across decks
dex ui                      # interactive slide reordering (styled TUI)
dex ui acme/                # reorder slides in acme/
dex new "My Talk"           # scaffold my-talk/
```

## Build & Test

```sh
make build    # compile binary
make test     # run tests
make lint     # go vet
make install  # go install
```

## License

MIT
