package slidev

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/StephanSchmidt/dex/format"
)

var headingRe = regexp.MustCompile(`^# (.+)`)
var titleAttrRe = regexp.MustCompile(`title="([^"]+)"`)

// Format implements format.Format for Slidev markdown presentations.
type Format struct{}

// Parse splits a Slidev markdown file into its document metadata
// and individual slides.
func (Format) Parse(data []byte) format.Deck {
	// Split into blocks separated by "---" lines.
	//   blocks[0] = before first --- (empty)
	//   blocks[1] = document metadata
	//   blocks[2+] = slide content or per-slide metadata
	var blocks []string
	var cur strings.Builder
	lines := strings.Split(string(data), "\n")
	// strings.Split produces a trailing empty element for \n-terminated input;
	// drop it so the last slide doesn't accumulate an extra \n per round-trip.
	if len(lines) > 0 && lines[len(lines)-1] == "" {
		lines = lines[:len(lines)-1]
	}
	for _, line := range lines {
		if strings.TrimSpace(line) == "---" {
			blocks = append(blocks, cur.String())
			cur.Reset()
		} else {
			cur.WriteString(line)
			cur.WriteString("\n")
		}
	}
	blocks = append(blocks, cur.String())

	var d format.Deck
	if len(blocks) > 1 {
		d.Metadata = blocks[1]
	}

	// From block 2 onward, each block is either per-slide metadata
	// (followed by a content block) or slide content directly.
	for i := 2; i < len(blocks); i++ {
		if i+1 < len(blocks) && isFrontmatter(blocks[i]) {
			d.Slides = append(d.Slides, format.Slide{
				Metadata: blocks[i],
				Content:  blocks[i+1],
			})
			i++ // skip content block
		} else {
			d.Slides = append(d.Slides, format.Slide{
				Content: blocks[i],
			})
		}
	}
	return d
}

// Render reconstructs a valid Slidev markdown file from its parts.
func (Format) Render(d format.Deck) []byte {
	var buf strings.Builder
	buf.WriteString("---\n")
	buf.WriteString(d.Metadata)
	for _, s := range d.Slides {
		buf.WriteString("---\n")
		if s.Metadata != "" {
			buf.WriteString(s.Metadata)
			buf.WriteString("---\n")
		}
		buf.WriteString(s.Content)
	}
	return []byte(buf.String())
}

var metaTitleRe = regexp.MustCompile(`(?m)^title:\s*(.+)`)

// ExtractTitle returns the title from a slide, looking for H1 headings
// first, then title attributes in content, then title: in metadata.
func (Format) ExtractTitle(s format.Slide) string {
	for line := range strings.SplitSeq(s.Content, "\n") {
		if m := headingRe.FindStringSubmatch(line); m != nil {
			return m[1]
		}
	}
	for line := range strings.SplitSeq(s.Content, "\n") {
		if m := titleAttrRe.FindStringSubmatch(line); m != nil {
			return m[1]
		}
	}
	if m := metaTitleRe.FindStringSubmatch(s.Metadata); m != nil {
		return strings.TrimSpace(m[1])
	}
	return "(untitled)"
}

// DefaultFile returns the default presentation filename for Slidev.
func (Format) DefaultFile() string {
	return "slides.md"
}

// RenderSlide outputs a single slide as Slidev markdown.
func (Format) RenderSlide(s format.Slide) string {
	var buf strings.Builder
	buf.WriteString("---\n")
	if s.Metadata != "" {
		buf.WriteString(s.Metadata)
		buf.WriteString("---\n")
	}
	buf.WriteString(s.Content)
	return buf.String()
}

// RenameSlide returns a copy of the slide with its title changed to name.
// It replaces an existing H1 heading first, then a title attribute, or
// prepends an H1 if no title is found.
func (Format) RenameSlide(s format.Slide, name string) format.Slide {
	lines := strings.Split(s.Content, "\n")

	for i, line := range lines {
		if headingRe.MatchString(line) {
			lines[i] = "# " + name
			s.Content = strings.Join(lines, "\n")
			return s
		}
	}

	for i, line := range lines {
		if loc := titleAttrRe.FindStringIndex(line); loc != nil {
			lines[i] = line[:loc[0]] + `title="` + name + `"` + line[loc[1]:]
			s.Content = strings.Join(lines, "\n")
			return s
		}
	}

	s.Content = "# " + name + "\n" + s.Content
	return s
}

// NewSlide returns a new slide with an H1 heading for the given title.
func (Format) NewSlide(title string) format.Slide {
	return format.Slide{Content: "\n# " + title + "\n\n"}
}

// RenameDeck returns a copy of the deck with the metadata title: field
// set to name. If no title: line exists, one is appended.
func (Format) RenameDeck(d format.Deck, name string) format.Deck {
	lines := strings.Split(d.Metadata, "\n")
	replaced := false
	for i, line := range lines {
		if strings.HasPrefix(line, "title:") {
			lines[i] = "title: " + name
			replaced = true
			break
		}
	}
	if replaced {
		d.Metadata = strings.Join(lines, "\n")
	} else {
		d.Metadata += "title: " + name + "\n"
	}
	return d
}

// Scaffold returns the list of files and directories needed for a new
// Slidev presentation.
func (Format) Scaffold(title, slug string) []format.ScaffoldFile {
	return []format.ScaffoldFile{
		{Path: "public", IsDir: true},
		{Path: "snippets", IsDir: true},
		{Path: "package.json", Content: fmt.Sprintf(packageJSONTemplate, slug)},
		{Path: "slides.md", Content: fmt.Sprintf(slidesMDTemplate, title, title)},
	}
}

// PostScaffoldCmd returns the command to run after scaffolding a Slidev project.
func (Format) PostScaffoldCmd() []string {
	return []string{"pnpm", "install"}
}

const packageJSONTemplate = `{
  "name": "presentation-%s",
  "type": "module",
  "private": true,
  "scripts": {
    "build": "slidev build",
    "dev": "slidev --open",
    "export": "slidev export"
  },
  "dependencies": {
    "@slidev/cli": "^52.11.5",
    "@slidev/theme-default": "latest",
    "@slidev/theme-seriph": "latest"
  },
  "devDependencies": {
    "playwright-chromium": "^1.58.1"
  }
}
`

const slidesMDTemplate = `---
theme: default
title: %s
class: text-center
---

# %s

`

// isFrontmatter checks whether a block between --- lines looks like YAML
// frontmatter rather than slide content. Empty blocks (from consecutive
// --- lines) count as empty frontmatter.
func isFrontmatter(block string) bool {
	trimmed := strings.TrimSpace(block)
	if trimmed == "" {
		return true
	}
	for line := range strings.SplitSeq(trimmed, "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		if strings.HasPrefix(line, "# ") || strings.HasPrefix(line, "<") {
			return false
		}
		if !strings.Contains(line, ":") && !strings.HasPrefix(line, "-") {
			return false
		}
	}
	return true
}
