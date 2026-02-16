package main

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

func key(k string) tea.KeyMsg {
	return tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(k)}
}

func specialKey(t tea.KeyType) tea.KeyMsg {
	return tea.KeyMsg{Type: t}
}

func setupUIModel(t *testing.T) uiModel {
	t.Helper()
	setupTestFs(t, testFixture)
	d, file, err := readDeck("")
	if err != nil {
		t.Fatalf("readDeck: %v", err)
	}
	return newUIModel(d.Slides, d.Metadata, file)
}

func update(m uiModel, msg tea.Msg) (uiModel, tea.Cmd) {
	model, cmd := m.Update(msg)
	return model.(uiModel), cmd
}

func TestUINavigation(t *testing.T) {
	t.Run("down moves cursor", func(t *testing.T) {
		m := setupUIModel(t)
		m, _ = update(m, key("j"))
		if m.cursor != 1 {
			t.Errorf("cursor = %d, want 1", m.cursor)
		}
	})

	t.Run("up moves cursor", func(t *testing.T) {
		m := setupUIModel(t)
		m, _ = update(m, key("j"))
		m, _ = update(m, key("k"))
		if m.cursor != 0 {
			t.Errorf("cursor = %d, want 0", m.cursor)
		}
	})

	t.Run("up at top clamps", func(t *testing.T) {
		m := setupUIModel(t)
		m, _ = update(m, key("k"))
		if m.cursor != 0 {
			t.Errorf("cursor = %d, want 0", m.cursor)
		}
	})

	t.Run("down at bottom clamps", func(t *testing.T) {
		m := setupUIModel(t)
		for range 10 {
			m, _ = update(m, key("j"))
		}
		if m.cursor != 3 {
			t.Errorf("cursor = %d, want 3", m.cursor)
		}
	})

	t.Run("arrow keys work", func(t *testing.T) {
		m := setupUIModel(t)
		m, _ = update(m, specialKey(tea.KeyDown))
		if m.cursor != 1 {
			t.Errorf("cursor = %d, want 1", m.cursor)
		}
		m, _ = update(m, specialKey(tea.KeyUp))
		if m.cursor != 0 {
			t.Errorf("cursor = %d, want 0", m.cursor)
		}
	})
}

func TestUINudge(t *testing.T) {
	t.Run("nudge down", func(t *testing.T) {
		m := setupUIModel(t)
		m, _ = update(m, key("J"))
		if m.cursor != 1 {
			t.Errorf("cursor = %d, want 1", m.cursor)
		}
		titles := getTitles(t)
		want := []string{"B", "A", "C", "D"}
		if !equal(titles, want) {
			t.Errorf("got %v, want %v", titles, want)
		}
		if !equal(m.titles, want) {
			t.Errorf("model titles = %v, want %v", m.titles, want)
		}
	})

	t.Run("nudge up", func(t *testing.T) {
		m := setupUIModel(t)
		m, _ = update(m, key("j"))
		m, _ = update(m, key("K"))
		if m.cursor != 0 {
			t.Errorf("cursor = %d, want 0", m.cursor)
		}
		titles := getTitles(t)
		want := []string{"B", "A", "C", "D"}
		if !equal(titles, want) {
			t.Errorf("got %v, want %v", titles, want)
		}
	})

	t.Run("nudge down at bottom is noop", func(t *testing.T) {
		m := setupUIModel(t)
		for range 3 {
			m, _ = update(m, key("j"))
		}
		m, _ = update(m, key("J"))
		if m.cursor != 3 {
			t.Errorf("cursor = %d, want 3", m.cursor)
		}
		titles := getTitles(t)
		want := []string{"A", "B", "C", "D"}
		if !equal(titles, want) {
			t.Errorf("got %v, want %v", titles, want)
		}
	})

	t.Run("nudge up at top is noop", func(t *testing.T) {
		m := setupUIModel(t)
		m, _ = update(m, key("K"))
		if m.cursor != 0 {
			t.Errorf("cursor = %d, want 0", m.cursor)
		}
		titles := getTitles(t)
		want := []string{"A", "B", "C", "D"}
		if !equal(titles, want) {
			t.Errorf("got %v, want %v", titles, want)
		}
	})
}

func TestUIDelete(t *testing.T) {
	t.Run("delete middle slide", func(t *testing.T) {
		m := setupUIModel(t)
		m, _ = update(m, key("j"))
		m, _ = update(m, key("d"))
		if m.cursor != 1 {
			t.Errorf("cursor = %d, want 1", m.cursor)
		}
		titles := getTitles(t)
		want := []string{"A", "C", "D"}
		if !equal(titles, want) {
			t.Errorf("got %v, want %v", titles, want)
		}
	})

	t.Run("delete last slide clamps cursor", func(t *testing.T) {
		m := setupUIModel(t)
		for range 3 {
			m, _ = update(m, key("j"))
		}
		m, _ = update(m, key("d"))
		if m.cursor != 2 {
			t.Errorf("cursor = %d, want 2", m.cursor)
		}
		titles := getTitles(t)
		want := []string{"A", "B", "C"}
		if !equal(titles, want) {
			t.Errorf("got %v, want %v", titles, want)
		}
	})

	t.Run("delete only slide refused", func(t *testing.T) {
		setupTestFs(t, `---
theme: default
---

# Solo
`)
		d, file, err := readDeck("")
		if err != nil {
			t.Fatalf("readDeck: %v", err)
		}
		m := newUIModel(d.Slides, d.Metadata, file)
		m, _ = update(m, key("d"))
		if m.err == nil {
			t.Error("expected error when deleting only slide")
		}
		if len(m.slides) != 1 {
			t.Errorf("slides len = %d, want 1", len(m.slides))
		}
	})

	t.Run("x also deletes", func(t *testing.T) {
		m := setupUIModel(t)
		m, _ = update(m, key("x"))
		titles := getTitles(t)
		want := []string{"B", "C", "D"}
		if !equal(titles, want) {
			t.Errorf("got %v, want %v", titles, want)
		}
		if m.cursor != 0 {
			t.Errorf("cursor = %d, want 0", m.cursor)
		}
	})
}

func TestUISelectMove(t *testing.T) {
	t.Run("select and move", func(t *testing.T) {
		m := setupUIModel(t)
		// Select slide 0 (A)
		m, _ = update(m, key(" "))
		if m.selected != 0 {
			t.Fatalf("selected = %d, want 0", m.selected)
		}
		// Move cursor to 2
		m, _ = update(m, key("j"))
		m, _ = update(m, key("j"))
		// Press m to move
		m, _ = update(m, key("m"))
		if m.selected != -1 {
			t.Errorf("selected = %d, want -1", m.selected)
		}
		titles := getTitles(t)
		want := []string{"B", "C", "A", "D"}
		if !equal(titles, want) {
			t.Errorf("got %v, want %v", titles, want)
		}
	})

	t.Run("move to same position is noop", func(t *testing.T) {
		m := setupUIModel(t)
		m, _ = update(m, key(" "))
		m, _ = update(m, key("m"))
		if m.selected != -1 {
			t.Errorf("selected = %d, want -1", m.selected)
		}
		titles := getTitles(t)
		want := []string{"A", "B", "C", "D"}
		if !equal(titles, want) {
			t.Errorf("got %v, want %v", titles, want)
		}
	})
}

func TestUISelectCopy(t *testing.T) {
	t.Run("select and copy", func(t *testing.T) {
		m := setupUIModel(t)
		// Select slide 0 (A)
		m, _ = update(m, key(" "))
		// Move cursor to 2
		m, _ = update(m, key("j"))
		m, _ = update(m, key("j"))
		// Press c to copy
		m, _ = update(m, key("c"))
		if m.selected != -1 {
			t.Errorf("selected = %d, want -1", m.selected)
		}
		titles := getTitles(t)
		want := []string{"A", "B", "A", "C", "D"}
		if !equal(titles, want) {
			t.Errorf("got %v, want %v", titles, want)
		}
	})
}

func TestUISelectSwap(t *testing.T) {
	t.Run("select and swap", func(t *testing.T) {
		m := setupUIModel(t)
		// Select slide 0 (A)
		m, _ = update(m, key(" "))
		// Move cursor to 2
		m, _ = update(m, key("j"))
		m, _ = update(m, key("j"))
		// Press s to swap
		m, _ = update(m, key("s"))
		if m.selected != -1 {
			t.Errorf("selected = %d, want -1", m.selected)
		}
		titles := getTitles(t)
		want := []string{"C", "B", "A", "D"}
		if !equal(titles, want) {
			t.Errorf("got %v, want %v", titles, want)
		}
	})

	t.Run("swap same position is noop", func(t *testing.T) {
		m := setupUIModel(t)
		m, _ = update(m, key(" "))
		m, _ = update(m, key("s"))
		if m.selected != -1 {
			t.Errorf("selected = %d, want -1", m.selected)
		}
		titles := getTitles(t)
		want := []string{"A", "B", "C", "D"}
		if !equal(titles, want) {
			t.Errorf("got %v, want %v", titles, want)
		}
	})
}

func TestUISelectCancel(t *testing.T) {
	t.Run("esc deselects", func(t *testing.T) {
		m := setupUIModel(t)
		m, _ = update(m, key(" "))
		if m.selected != 0 {
			t.Fatalf("selected = %d, want 0", m.selected)
		}
		m, _ = update(m, specialKey(tea.KeyEscape))
		if m.selected != -1 {
			t.Errorf("selected = %d, want -1", m.selected)
		}
		// No write should have occurred
		titles := getTitles(t)
		want := []string{"A", "B", "C", "D"}
		if !equal(titles, want) {
			t.Errorf("got %v, want %v", titles, want)
		}
	})
}

func TestUIQuit(t *testing.T) {
	t.Run("q quits", func(t *testing.T) {
		m := setupUIModel(t)
		_, cmd := m.Update(key("q"))
		if cmd == nil {
			t.Fatal("expected tea.Quit cmd, got nil")
		}
	})

	t.Run("esc quits in normal mode", func(t *testing.T) {
		m := setupUIModel(t)
		_, cmd := m.Update(specialKey(tea.KeyEscape))
		if cmd == nil {
			t.Fatal("expected tea.Quit cmd, got nil")
		}
	})

	t.Run("ctrl+c quits", func(t *testing.T) {
		m := setupUIModel(t)
		_, cmd := m.Update(specialKey(tea.KeyCtrlC))
		if cmd == nil {
			t.Fatal("expected tea.Quit cmd, got nil")
		}
	})
}

func TestUIScrolling(t *testing.T) {
	t.Run("scroll follows cursor down", func(t *testing.T) {
		m := setupUIModel(t)
		m.height = 9 // visible = 9 - 6 = 3
		for range 3 {
			m, _ = update(m, key("j"))
		}
		// cursor=3, visible=3 → offset should be 1
		if m.offset != 1 {
			t.Errorf("offset = %d, want 1", m.offset)
		}
		if m.cursor != 3 {
			t.Errorf("cursor = %d, want 3", m.cursor)
		}
	})

	t.Run("scroll follows cursor up", func(t *testing.T) {
		m := setupUIModel(t)
		m.height = 9
		m.cursor = 3
		m.offset = 1
		m, _ = update(m, key("k"))
		m, _ = update(m, key("k"))
		// cursor=1, offset should still be 1
		if m.offset != 1 {
			t.Errorf("offset = %d, want 1", m.offset)
		}
		m, _ = update(m, key("k"))
		// cursor=0, offset should adjust to 0
		if m.offset != 0 {
			t.Errorf("offset = %d, want 0", m.offset)
		}
	})

	t.Run("WindowSizeMsg updates height and width", func(t *testing.T) {
		m := setupUIModel(t)
		m, _ = update(m, tea.WindowSizeMsg{Width: 80, Height: 10})
		if m.height != 10 {
			t.Errorf("height = %d, want 10", m.height)
		}
		if m.width != 80 {
			t.Errorf("width = %d, want 80", m.width)
		}
	})

	t.Run("no height yet shows all slides", func(t *testing.T) {
		m := setupUIModel(t)
		// height=0 by default, all slides should be visible
		view := m.View()
		for i, title := range m.titles {
			if !strings.Contains(view, title) {
				t.Errorf("slide %d title %q not found in view", i, title)
			}
		}
	})
}

func TestUIIntegration(t *testing.T) {
	t.Run("nudge writes to disk", func(t *testing.T) {
		m := setupUIModel(t)
		_, _ = update(m, key("J"))
		titles := getTitles(t)
		want := []string{"B", "A", "C", "D"}
		if !equal(titles, want) {
			t.Errorf("disk titles = %v, want %v", titles, want)
		}
	})

	t.Run("select enter works", func(t *testing.T) {
		m := setupUIModel(t)
		m, _ = update(m, specialKey(tea.KeyEnter))
		if m.selected != 0 {
			t.Errorf("selected = %d, want 0", m.selected)
		}
	})
}
