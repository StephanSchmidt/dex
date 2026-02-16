package main

import (
	"fmt"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

	"github.com/spf13/afero"
)

var headingRe = regexp.MustCompile(`^# (.+)`)
var titleAttrRe = regexp.MustCompile(`title="([^"]+)"`)

type slide struct {
	frontmatter string // raw YAML (empty string if none)
	content     string // raw markdown content
}

type deck struct {
	frontmatter string
	slides      []slide
}

// parseDeck splits a Slidev markdown file into its document frontmatter
// and individual slides.
func parseDeck(data []byte) deck {
	// Split into blocks separated by "---" lines.
	//   blocks[0] = before first --- (empty)
	//   blocks[1] = document frontmatter
	//   blocks[2+] = slide content or per-slide frontmatter
	var blocks []string
	var cur strings.Builder
	for _, line := range strings.Split(string(data), "\n") {
		if strings.TrimSpace(line) == "---" {
			blocks = append(blocks, cur.String())
			cur.Reset()
		} else {
			cur.WriteString(line)
			cur.WriteString("\n")
		}
	}
	blocks = append(blocks, cur.String())

	var d deck
	if len(blocks) > 1 {
		d.frontmatter = blocks[1]
	}

	// From block 2 onward, each block is either per-slide frontmatter
	// (followed by a content block) or slide content directly.
	for i := 2; i < len(blocks); i++ {
		if i+1 < len(blocks) && isFrontmatter(blocks[i]) {
			d.slides = append(d.slides, slide{
				frontmatter: blocks[i],
				content:     blocks[i+1],
			})
			i++ // skip content block
		} else {
			d.slides = append(d.slides, slide{
				content: blocks[i],
			})
		}
	}
	return d
}

// render reconstructs a valid Slidev markdown file from its parts.
func (d deck) render() []byte {
	var buf strings.Builder
	buf.WriteString("---\n")
	buf.WriteString(d.frontmatter)
	for _, s := range d.slides {
		buf.WriteString("---\n")
		if s.frontmatter != "" {
			buf.WriteString(s.frontmatter)
			buf.WriteString("---\n")
		}
		buf.WriteString(s.content)
	}
	return []byte(buf.String())
}

// insert returns a new deck with newSlides inserted before position pos (0-based).
func (d deck) insert(pos int, newSlides []slide) deck {
	result := make([]slide, 0, len(d.slides)+len(newSlides))
	result = append(result, d.slides[:pos]...)
	result = append(result, newSlides...)
	result = append(result, d.slides[pos:]...)
	return deck{frontmatter: d.frontmatter, slides: result}
}

// readDeck resolves a directory-or-file path, reads it, and parses it.
func readDeck(dirOrFile string) (deck, string, error) {
	file := resolveFile(dirOrFile)
	data, err := afero.ReadFile(appFs, file)
	if err != nil {
		return deck{}, "", fmt.Errorf("reading %s: %w", file, err)
	}
	return parseDeck(data), file, nil
}

// writeDeck renders a deck and writes it to the given file.
func writeDeck(file string, d deck) error {
	if err := afero.WriteFile(appFs, file, d.render(), 0o644); err != nil {
		return fmt.Errorf("writing %s: %w", file, err)
	}
	return nil
}

// resolveFile turns a path into a slides.md file path. If the path ends
// with "/" or points to an existing directory, "slides.md" is appended.
func resolveFile(path string) string {
	if path == "" {
		return "slides.md"
	}
	if strings.HasSuffix(path, "/") {
		return filepath.Join(path, "slides.md")
	}
	if info, err := appFs.Stat(path); err == nil && info.IsDir() {
		return filepath.Join(path, "slides.md")
	}
	return path
}

// splitDirRange splits "dir/range" into ("dir/", "range"). If there is
// no "/" the dir part is empty and the whole string is the range.
func splitDirRange(expr string) (dir, rangeExpr string) {
	i := strings.LastIndex(expr, "/")
	if i < 0 {
		return "", expr
	}
	return expr[:i+1], expr[i+1:]
}

// parseSliceExpr parses a comma-separated list of 1-based slide selectors
// into 0-based indices. Each item is either a single number or a range
// (using : or -). Negative numbers count from the end.
func parseSliceExpr(expr string, length int) ([]int, error) {
	var result []int
	for _, part := range strings.Split(expr, ",") {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}

		// Try colon range first, then dash range.
		start, end, isRange := "", "", false
		if idx := strings.Index(part, ":"); idx >= 0 {
			start, end = part[:idx], part[idx+1:]
			isRange = true
		} else {
			// For dash: only treat as range separator when preceded by a digit.
			for j := 1; j < len(part); j++ {
				if part[j] == '-' && part[j-1] >= '0' && part[j-1] <= '9' {
					start, end = part[:j], part[j+1:]
					isRange = true
					break
				}
			}
		}

		if isRange {
			s, err := resolveIndex(start, length)
			if err != nil {
				return nil, fmt.Errorf("invalid range start %q: %w", start, err)
			}
			e, err := resolveIndex(end, length)
			if err != nil {
				return nil, fmt.Errorf("invalid range end %q: %w", end, err)
			}
			if s > e {
				return nil, fmt.Errorf("invalid range: %d > %d", s+1, e+1)
			}
			for i := s; i <= e; i++ {
				result = append(result, i)
			}
		} else {
			idx, err := resolveIndex(part, length)
			if err != nil {
				return nil, err
			}
			result = append(result, idx)
		}
	}
	if len(result) == 0 {
		return nil, fmt.Errorf("empty slide expression")
	}
	return result, nil
}

// resolveIndex converts a 1-based (possibly negative) index string to a
// 0-based index. Negative values count from the end (-1 = last).
func resolveIndex(s string, length int) (int, error) {
	s = strings.TrimSpace(s)
	n, err := strconv.Atoi(s)
	if err != nil {
		return 0, fmt.Errorf("invalid number %q", s)
	}
	if n == 0 {
		return 0, fmt.Errorf("index 0 is not valid (1-based)")
	}
	var idx int
	if n > 0 {
		idx = n - 1
	} else {
		idx = length + n
	}
	if idx < 0 || idx >= length {
		return 0, fmt.Errorf("index %d out of range (1..%d)", n, length)
	}
	return idx, nil
}

// isFrontmatter checks whether a block between --- lines looks like YAML
// frontmatter rather than slide content. Empty blocks (from consecutive
// --- lines) count as empty frontmatter.
func isFrontmatter(block string) bool {
	trimmed := strings.TrimSpace(block)
	if trimmed == "" {
		return true
	}
	for _, line := range strings.Split(trimmed, "\n") {
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

func extractTitle(content string) string {
	for _, line := range strings.Split(content, "\n") {
		if m := headingRe.FindStringSubmatch(line); m != nil {
			return m[1]
		}
	}
	for _, line := range strings.Split(content, "\n") {
		if m := titleAttrRe.FindStringSubmatch(line); m != nil {
			return m[1]
		}
	}
	return "(untitled)"
}
