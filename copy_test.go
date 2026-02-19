package main

import "testing"

func TestCopy(t *testing.T) {
	t.Run("cross-file copy", func(t *testing.T) {
		setupTestFs(t, testFixture)
		writeTestFile(t, "dir1/slides.md", testFixture)
		writeTestFile(t, "dir2/slides.md", testFixture2)

		cmd := CopyCmd{Source: "dir1/1-2", Target: "dir2/3"}
		if err := cmd.Run(); err != nil {
			t.Fatalf("CopyCmd.Run() error: %v", err)
		}

		got := getTitlesFrom(t, "dir2/slides.md")
		want := []string{"X", "Y", "A", "B", "Z"}
		if !equal(got, want) {
			t.Errorf("dir2 got %v, want %v", got, want)
		}

		// Source must be unchanged.
		src := getTitlesFrom(t, "dir1/slides.md")
		wantSrc := []string{"A", "B", "C", "D"}
		if !equal(src, wantSrc) {
			t.Errorf("dir1 got %v, want %v (source modified!)", src, wantSrc)
		}
	})

	t.Run("same-file copy", func(t *testing.T) {
		setupTestFs(t, testFixture)

		cmd := CopyCmd{Source: "1", Target: "3"}
		if err := cmd.Run(); err != nil {
			t.Fatalf("CopyCmd.Run() error: %v", err)
		}

		got := getTitles(t)
		want := []string{"A", "B", "A", "C", "D"}
		if !equal(got, want) {
			t.Errorf("got %v, want %v", got, want)
		}
	})

	t.Run("copy to end (n+1)", func(t *testing.T) {
		setupTestFs(t, testFixture)
		writeTestFile(t, "dir1/slides.md", testFixture)
		writeTestFile(t, "dir2/slides.md", testFixture2)

		cmd := CopyCmd{Source: "dir1/1", Target: "dir2/4"}
		if err := cmd.Run(); err != nil {
			t.Fatalf("CopyCmd.Run() error: %v", err)
		}

		got := getTitlesFrom(t, "dir2/slides.md")
		want := []string{"X", "Y", "Z", "A"}
		if !equal(got, want) {
			t.Errorf("got %v, want %v", got, want)
		}
	})

	t.Run("error: position out of range", func(t *testing.T) {
		setupTestFs(t, testFixture)
		writeTestFile(t, "dir2/slides.md", testFixture2)

		cmd := CopyCmd{Source: "1", Target: "dir2/5"}
		if err := cmd.Run(); err == nil {
			t.Fatal("expected error for out-of-range position")
		}
	})

	t.Run("relative path", func(t *testing.T) {
		setupTestFs(t, testFixture)
		writeTestFile(t, "../other/slides.md", testFixture2)

		cmd := CopyCmd{Source: "../other/1-2", Target: "3"}
		if err := cmd.Run(); err != nil {
			t.Fatalf("CopyCmd.Run() error: %v", err)
		}

		got := getTitles(t)
		want := []string{"A", "B", "X", "Y", "C", "D"}
		if !equal(got, want) {
			t.Errorf("got %v, want %v", got, want)
		}
	})

	t.Run("copy to end with -1", func(t *testing.T) {
		setupTestFs(t, testFixture)
		writeTestFile(t, "dir1/slides.md", testFixture)
		writeTestFile(t, "dir2/slides.md", testFixture2)

		cmd := CopyCmd{Source: "dir1/1", Target: "dir2/-1"}
		if err := cmd.Run(); err != nil {
			t.Fatalf("CopyCmd.Run() error: %v", err)
		}

		got := getTitlesFrom(t, "dir2/slides.md")
		want := []string{"X", "Y", "Z", "A"}
		if !equal(got, want) {
			t.Errorf("got %v, want %v", got, want)
		}
	})

	t.Run("copy before last with -2", func(t *testing.T) {
		setupTestFs(t, testFixture)
		writeTestFile(t, "dir2/slides.md", testFixture2)

		cmd := CopyCmd{Source: "1", Target: "dir2/-2"}
		if err := cmd.Run(); err != nil {
			t.Fatalf("CopyCmd.Run() error: %v", err)
		}

		got := getTitlesFrom(t, "dir2/slides.md")
		want := []string{"X", "Y", "A", "Z"}
		if !equal(got, want) {
			t.Errorf("got %v, want %v", got, want)
		}
	})

	t.Run("error: position 0", func(t *testing.T) {
		setupTestFs(t, testFixture)

		cmd := CopyCmd{Source: "1", Target: "0"}
		if err := cmd.Run(); err == nil {
			t.Fatal("expected error for position 0")
		}
	})
}
