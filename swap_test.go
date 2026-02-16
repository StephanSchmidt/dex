package main

import "testing"

func TestSwap(t *testing.T) {
	t.Run("same file", func(t *testing.T) {
		setupTestFs(t, testFixture)

		cmd := SwapCmd{A: "1", B: "3"}
		if err := cmd.Run(); err != nil {
			t.Fatalf("SwapCmd.Run() error: %v", err)
		}

		got := getTitles(t)
		want := []string{"C", "B", "A", "D"}
		if !equal(got, want) {
			t.Errorf("got %v, want %v", got, want)
		}
	})

	t.Run("same file with dir", func(t *testing.T) {
		setupTestFs(t, testFixture)
		writeTestFile(t, "mydir/slides.md", testFixture)

		cmd := SwapCmd{A: "mydir/1", B: "mydir/4"}
		if err := cmd.Run(); err != nil {
			t.Fatalf("SwapCmd.Run() error: %v", err)
		}

		got := getTitlesFrom(t, "mydir/slides.md")
		want := []string{"D", "B", "C", "A"}
		if !equal(got, want) {
			t.Errorf("got %v, want %v", got, want)
		}
	})

	t.Run("cross-file", func(t *testing.T) {
		setupTestFs(t, testFixture)
		writeTestFile(t, "dir1/slides.md", testFixture)
		writeTestFile(t, "dir2/slides.md", testFixture2)

		cmd := SwapCmd{A: "dir1/1", B: "dir2/2"}
		if err := cmd.Run(); err != nil {
			t.Fatalf("SwapCmd.Run() error: %v", err)
		}

		got1 := getTitlesFrom(t, "dir1/slides.md")
		want1 := []string{"Y", "B", "C", "D"}
		if !equal(got1, want1) {
			t.Errorf("dir1 got %v, want %v", got1, want1)
		}

		got2 := getTitlesFrom(t, "dir2/slides.md")
		want2 := []string{"X", "A", "Z"}
		if !equal(got2, want2) {
			t.Errorf("dir2 got %v, want %v", got2, want2)
		}
	})

	t.Run("error: same slide", func(t *testing.T) {
		setupTestFs(t, testFixture)

		cmd := SwapCmd{A: "2", B: "2"}
		if err := cmd.Run(); err == nil {
			t.Fatal("expected error when swapping a slide with itself")
		}
	})

	t.Run("error: range instead of single", func(t *testing.T) {
		setupTestFs(t, testFixture)

		cmd := SwapCmd{A: "1-3", B: "4"}
		if err := cmd.Run(); err == nil {
			t.Fatal("expected error for range argument")
		}
	})

	t.Run("error: out of range", func(t *testing.T) {
		setupTestFs(t, testFixture)

		cmd := SwapCmd{A: "1", B: "10"}
		if err := cmd.Run(); err == nil {
			t.Fatal("expected error for out-of-range index")
		}
	})
}
