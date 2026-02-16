package main

import (
	"strings"
	"testing"

	"github.com/spf13/afero"
)

func TestInsert(t *testing.T) {
	t.Run("insert at beginning", func(t *testing.T) {
		setupTestFs(t, testFixture)

		cmd := InsertCmd{Pos: "1", Title: "New"}
		if err := cmd.Run(); err != nil {
			t.Fatalf("InsertCmd.Run() error: %v", err)
		}

		got := getTitles(t)
		want := []string{"New", "A", "B", "C", "D"}
		if !equal(got, want) {
			t.Errorf("got %v, want %v", got, want)
		}
	})

	t.Run("insert in middle", func(t *testing.T) {
		setupTestFs(t, testFixture)

		cmd := InsertCmd{Pos: "3", Title: "Mid"}
		if err := cmd.Run(); err != nil {
			t.Fatalf("InsertCmd.Run() error: %v", err)
		}

		got := getTitles(t)
		want := []string{"A", "B", "Mid", "C", "D"}
		if !equal(got, want) {
			t.Errorf("got %v, want %v", got, want)
		}
	})

	t.Run("append at end", func(t *testing.T) {
		setupTestFs(t, testFixture)

		cmd := InsertCmd{Pos: "5", Title: "Last"}
		if err := cmd.Run(); err != nil {
			t.Fatalf("InsertCmd.Run() error: %v", err)
		}

		got := getTitles(t)
		want := []string{"A", "B", "C", "D", "Last"}
		if !equal(got, want) {
			t.Errorf("got %v, want %v", got, want)
		}
	})

	t.Run("directory path", func(t *testing.T) {
		setupTestFs(t, testFixture)
		writeTestFile(t, "mydir/slides.md", testFixture2)

		cmd := InsertCmd{Pos: "mydir/2", Title: "Ins"}
		if err := cmd.Run(); err != nil {
			t.Fatalf("InsertCmd.Run() error: %v", err)
		}

		got := getTitlesFrom(t, "mydir/slides.md")
		want := []string{"X", "Ins", "Y", "Z"}
		if !equal(got, want) {
			t.Errorf("got %v, want %v", got, want)
		}
	})

	t.Run("error: position 0", func(t *testing.T) {
		setupTestFs(t, testFixture)

		cmd := InsertCmd{Pos: "0", Title: "Bad"}
		if err := cmd.Run(); err == nil {
			t.Fatal("expected error for position 0")
		}
	})

	t.Run("error: position too high", func(t *testing.T) {
		setupTestFs(t, testFixture)

		cmd := InsertCmd{Pos: "6", Title: "Bad"}
		if err := cmd.Run(); err == nil {
			t.Fatal("expected error for position N+2")
		}
	})

	t.Run("title appears as H1", func(t *testing.T) {
		setupTestFs(t, testFixture)

		cmd := InsertCmd{Pos: "1", Title: "Hello World"}
		if err := cmd.Run(); err != nil {
			t.Fatalf("InsertCmd.Run() error: %v", err)
		}

		data, err := afero.ReadFile(appFs, "slides.md")
		if err != nil {
			t.Fatalf("reading slides.md: %v", err)
		}
		if !strings.Contains(string(data), "# Hello World") {
			t.Error("expected H1 heading with title in output")
		}
	})

	t.Run("frontmatter preserved", func(t *testing.T) {
		setupTestFs(t, testFixture)

		cmd := InsertCmd{Pos: "2", Title: "New"}
		if err := cmd.Run(); err != nil {
			t.Fatalf("InsertCmd.Run() error: %v", err)
		}

		data, err := afero.ReadFile(appFs, "slides.md")
		if err != nil {
			t.Fatalf("reading slides.md: %v", err)
		}
		d := parseDeck(data)
		if !strings.Contains(d.frontmatter, "title: Test") {
			t.Error("deck frontmatter not preserved")
		}
	})
}
