package main

import "testing"

func TestMove(t *testing.T) {
	t.Run("move 2 to 4", func(t *testing.T) {
		setupTestFs(t, testFixture)
		cmd := MoveCmd{From: "2", To: "4", File: "slides.md"}
		if err := cmd.Run(); err != nil {
			t.Fatalf("MoveCmd.Run() error: %v", err)
		}
		titles := getTitles(t)
		want := []string{"A", "C", "B", "D"}
		if !equal(titles, want) {
			t.Errorf("got %v, want %v", titles, want)
		}
	})

	t.Run("move 4 to 2", func(t *testing.T) {
		setupTestFs(t, testFixture)
		cmd := MoveCmd{From: "4", To: "2", File: "slides.md"}
		if err := cmd.Run(); err != nil {
			t.Fatalf("MoveCmd.Run() error: %v", err)
		}
		titles := getTitles(t)
		want := []string{"A", "D", "B", "C"}
		if !equal(titles, want) {
			t.Errorf("got %v, want %v", titles, want)
		}
	})

	t.Run("move 3 to 1", func(t *testing.T) {
		setupTestFs(t, testFixture)
		cmd := MoveCmd{From: "3", To: "1", File: "slides.md"}
		if err := cmd.Run(); err != nil {
			t.Fatalf("MoveCmd.Run() error: %v", err)
		}
		titles := getTitles(t)
		want := []string{"C", "A", "B", "D"}
		if !equal(titles, want) {
			t.Errorf("got %v, want %v", titles, want)
		}
	})

	t.Run("out of range", func(t *testing.T) {
		setupTestFs(t, testFixture)
		cmd := MoveCmd{From: "5", To: "1", File: "slides.md"}
		if err := cmd.Run(); err == nil {
			t.Fatal("expected error for out-of-range index")
		}
	})

	t.Run("same position", func(t *testing.T) {
		setupTestFs(t, testFixture)
		cmd := MoveCmd{From: "2", To: "2", File: "slides.md"}
		if err := cmd.Run(); err == nil {
			t.Fatal("expected error for same position")
		}
	})

	t.Run("relative move 2 +2", func(t *testing.T) {
		setupTestFs(t, testFixture)
		cmd := MoveCmd{From: "2", To: "+2", File: "slides.md"}
		if err := cmd.Run(); err != nil {
			t.Fatalf("MoveCmd.Run() error: %v", err)
		}
		titles := getTitles(t)
		want := []string{"A", "C", "B", "D"}
		if !equal(titles, want) {
			t.Errorf("got %v, want %v", titles, want)
		}
	})

	t.Run("relative move 4 -2", func(t *testing.T) {
		setupTestFs(t, testFixture)
		cmd := MoveCmd{From: "4", To: "-2", File: "slides.md"}
		if err := cmd.Run(); err != nil {
			t.Fatalf("MoveCmd.Run() error: %v", err)
		}
		titles := getTitles(t)
		want := []string{"A", "D", "B", "C"}
		if !equal(titles, want) {
			t.Errorf("got %v, want %v", titles, want)
		}
	})

	t.Run("relative move 3 -1", func(t *testing.T) {
		setupTestFs(t, testFixture)
		cmd := MoveCmd{From: "3", To: "-1", File: "slides.md"}
		if err := cmd.Run(); err != nil {
			t.Fatalf("MoveCmd.Run() error: %v", err)
		}
		titles := getTitles(t)
		want := []string{"A", "C", "B", "D"}
		if !equal(titles, want) {
			t.Errorf("got %v, want %v", titles, want)
		}
	})

	t.Run("relative offset out of range", func(t *testing.T) {
		setupTestFs(t, testFixture)
		cmd := MoveCmd{From: "1", To: "-1", File: "slides.md"}
		if err := cmd.Run(); err == nil {
			t.Fatal("expected error for relative offset out of range")
		}
	})

	t.Run("relative offset +0", func(t *testing.T) {
		setupTestFs(t, testFixture)
		cmd := MoveCmd{From: "2", To: "+0", File: "slides.md"}
		if err := cmd.Run(); err == nil {
			t.Fatal("expected error for +0 offset")
		}
	})

	t.Run("negative from: move last to 1", func(t *testing.T) {
		setupTestFs(t, testFixture)
		cmd := MoveCmd{From: "-1", To: "1", File: "slides.md"}
		if err := cmd.Run(); err != nil {
			t.Fatalf("MoveCmd.Run() error: %v", err)
		}
		titles := getTitles(t)
		want := []string{"D", "A", "B", "C"}
		if !equal(titles, want) {
			t.Errorf("got %v, want %v", titles, want)
		}
	})

	t.Run("negative from: move second-to-last to 1", func(t *testing.T) {
		setupTestFs(t, testFixture)
		cmd := MoveCmd{From: "-2", To: "1", File: "slides.md"}
		if err := cmd.Run(); err != nil {
			t.Fatalf("MoveCmd.Run() error: %v", err)
		}
		titles := getTitles(t)
		want := []string{"C", "A", "B", "D"}
		if !equal(titles, want) {
			t.Errorf("got %v, want %v", titles, want)
		}
	})

	t.Run("to end", func(t *testing.T) {
		setupTestFs(t, testFixture)
		cmd := MoveCmd{From: "1", To: "end", File: "slides.md"}
		if err := cmd.Run(); err != nil {
			t.Fatalf("MoveCmd.Run() error: %v", err)
		}
		titles := getTitles(t)
		want := []string{"B", "C", "D", "A"}
		if !equal(titles, want) {
			t.Errorf("got %v, want %v", titles, want)
		}
	})

	t.Run("combined from:to form", func(t *testing.T) {
		setupTestFs(t, testFixture)
		cmd := MoveCmd{From: "1:end", File: "slides.md"}
		if err := cmd.Run(); err != nil {
			t.Fatalf("MoveCmd.Run() error: %v", err)
		}
		titles := getTitles(t)
		want := []string{"B", "C", "D", "A"}
		if !equal(titles, want) {
			t.Errorf("got %v, want %v", titles, want)
		}
	})

	t.Run("combined from:to with file in To slot", func(t *testing.T) {
		setupTestFs(t, testFixture)
		writeTestFile(t, "mydir/slides.md", testFixture)
		cmd := MoveCmd{From: "1:end", To: "mydir/"}
		if err := cmd.Run(); err != nil {
			t.Fatalf("MoveCmd.Run() error: %v", err)
		}
	})

	t.Run("to end-1", func(t *testing.T) {
		setupTestFs(t, testFixture)
		cmd := MoveCmd{From: "1", To: "end-1", File: "slides.md"}
		if err := cmd.Run(); err != nil {
			t.Fatalf("MoveCmd.Run() error: %v", err)
		}
		titles := getTitles(t)
		want := []string{"B", "C", "A", "D"}
		if !equal(titles, want) {
			t.Errorf("got %v, want %v", titles, want)
		}
	})

	t.Run("combined form with both To and File errors", func(t *testing.T) {
		// From:to with non-empty To AND File is ambiguous and should error.
		setupTestFs(t, testFixture)
		cmd := MoveCmd{From: "1:2", To: "ignored", File: "slides.md"}
		if err := cmd.Run(); err == nil {
			t.Fatal("expected error when from:to is combined with separate to + file")
		}
	})

	t.Run("missing target", func(t *testing.T) {
		setupTestFs(t, testFixture)
		cmd := MoveCmd{From: "1", File: "slides.md"}
		if err := cmd.Run(); err == nil {
			t.Fatal("expected error for missing target")
		}
	})

	t.Run("read error on missing file", func(t *testing.T) {
		setupTestFs(t, testFixture)
		cmd := MoveCmd{From: "1", To: "2", File: "nope.md"}
		if err := cmd.Run(); err == nil {
			t.Fatal("expected error for missing file")
		}
	})

	t.Run("invalid from", func(t *testing.T) {
		setupTestFs(t, testFixture)
		cmd := MoveCmd{From: "abc", To: "2", File: "slides.md"}
		if err := cmd.Run(); err == nil {
			t.Fatal("expected error for non-numeric from")
		}
	})

	t.Run("invalid end- offset", func(t *testing.T) {
		setupTestFs(t, testFixture)
		cmd := MoveCmd{From: "1", To: "end-abc", File: "slides.md"}
		if err := cmd.Run(); err == nil {
			t.Fatal("expected error for non-numeric end- offset")
		}
	})

	t.Run("invalid relative offset", func(t *testing.T) {
		setupTestFs(t, testFixture)
		cmd := MoveCmd{From: "1", To: "+abc", File: "slides.md"}
		if err := cmd.Run(); err == nil {
			t.Fatal("expected error for non-numeric relative offset")
		}
	})

	t.Run("invalid absolute target", func(t *testing.T) {
		setupTestFs(t, testFixture)
		cmd := MoveCmd{From: "1", To: "abc", File: "slides.md"}
		if err := cmd.Run(); err == nil {
			t.Fatal("expected error for non-numeric absolute target")
		}
	})
}
