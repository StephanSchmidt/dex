package main

import (
	"bytes"
	"testing"

	"github.com/spf13/afero"
)

const testFixture = `---
theme: default
title: Test
---

# A

---

# B

---
layout: center
---

# C

---

# D
`

const testFixture2 = `---
theme: default
title: Test2
---

# X

---

# Y

---

# Z
`

func setupTestFs(t *testing.T, content string) *bytes.Buffer {
	t.Helper()
	appFs = afero.NewMemMapFs()
	buf := &bytes.Buffer{}
	stdout = buf
	writeTestFile(t, "slides.md", content)
	return buf
}

func writeTestFile(t *testing.T, path, content string) {
	t.Helper()
	if err := afero.WriteFile(appFs, path, []byte(content), 0o644); err != nil {
		t.Fatalf("writing %s: %v", path, err)
	}
}

func getTitles(t *testing.T) []string {
	t.Helper()
	return getTitlesFrom(t, "slides.md")
}

func getTitlesFrom(t *testing.T, file string) []string {
	t.Helper()
	data, err := afero.ReadFile(appFs, file)
	if err != nil {
		t.Fatalf("reading %s: %v", file, err)
	}
	d := parseDeck(data)
	var titles []string
	for _, s := range d.slides {
		titles = append(titles, extractTitle(s.content))
	}
	return titles
}

func equal(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}
