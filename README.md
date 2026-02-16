# dex

A CLI tool for manipulating [Slidev](https://sli.dev) markdown presentations.

## Features

| Command | Description |
|---------|-------------|
| `slides` | List slide numbers and titles |
| `slide` | Print raw slide markdown |
| `copy` | Copy slides between decks |
| `move` | Reorder slides within a deck |
| `rename` | Rename the deck title in frontmatter |
| `rename-slide` | Rename a slide's title |
| `new` | Scaffold a new presentation directory |

## Installation

```sh
go install github.com/StephanSchmidt/dex@latest
```

## Usage

Each presentation is a directory containing a `slides.md` file.
`dir/` or `dir` resolves to `dir/slides.md`; bare ranges use `./slides.md`.

### Addressing

```
3            single slide
1-3          slides 1 through 3
1,2,7        specific slides
-1           last slide
1:-1         all slides
dir/1-3      slides 1-3 from dir/slides.md
```

### Examples

```sh
dex slides                  # list all slides in ./slides.md
dex slides acme/            # list slides in acme/slides.md
dex slide 3                 # print raw markdown of slide 3
dex copy dir1/1-3 dir2/5    # copy slides 1-3 from dir1, insert before slide 5 in dir2
dex move 2 4                # move slide 2 to position 4
dex rename "New Title"          # rename the deck title in frontmatter
dex rename "New Title" acme/    # rename deck in acme/slides.md
dex rename-slide 3 "New Title"  # rename slide 3
dex rename-slide acme/3 "Hi"   # rename slide 3 in acme/slides.md
dex rename-slide 1,3 "Same"   # rename slides 1 and 3
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
