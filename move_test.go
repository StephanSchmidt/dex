package main

import "testing"

func TestMove(t *testing.T) {
	t.Run("move 2 to 4", func(t *testing.T) {
		setupTestFs(t, testFixture)
		cmd := MoveCmd{From: 2, To: 4, File: "slides.md"}
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
		cmd := MoveCmd{From: 4, To: 2, File: "slides.md"}
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
		cmd := MoveCmd{From: 3, To: 1, File: "slides.md"}
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
		cmd := MoveCmd{From: 5, To: 1, File: "slides.md"}
		if err := cmd.Run(); err == nil {
			t.Fatal("expected error for out-of-range index")
		}
	})

	t.Run("same position", func(t *testing.T) {
		setupTestFs(t, testFixture)
		cmd := MoveCmd{From: 2, To: 2, File: "slides.md"}
		if err := cmd.Run(); err == nil {
			t.Fatal("expected error for same position")
		}
	})
}
