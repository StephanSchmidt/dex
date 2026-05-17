package main

import "testing"

func TestSlides(t *testing.T) {
	t.Run("default file", func(t *testing.T) {
		buf := setupTestFs(t, testFixture)
		cmd := SlidesCmd{}
		if err := cmd.Run(); err != nil {
			t.Fatalf("SlidesCmd.Run() error: %v", err)
		}
		want := "Test\n  1  A\n  2  B\n  3  C\n  4  D\n"
		if got := buf.String(); got != want {
			t.Errorf("SlidesCmd output:\ngot:  %q\nwant: %q", got, want)
		}
	})

	t.Run("directory with trailing slash", func(t *testing.T) {
		buf := setupTestFs(t, testFixture)
		writeTestFile(t, "mydir/slides.md", testFixture)
		cmd := SlidesCmd{Exprs: []string{"mydir/"}}
		if err := cmd.Run(); err != nil {
			t.Fatalf("SlidesCmd.Run() error: %v", err)
		}
		want := "Test\n  1  A\n  2  B\n  3  C\n  4  D\n"
		if got := buf.String(); got != want {
			t.Errorf("SlidesCmd output:\ngot:  %q\nwant: %q", got, want)
		}
	})

	t.Run("with range", func(t *testing.T) {
		buf := setupTestFs(t, testFixture)
		cmd := SlidesCmd{Exprs: []string{"1,3"}}
		if err := cmd.Run(); err != nil {
			t.Fatalf("SlidesCmd.Run() error: %v", err)
		}
		want := "Test\n  1  A\n  3  C\n"
		if got := buf.String(); got != want {
			t.Errorf("SlidesCmd output:\ngot:  %q\nwant: %q", got, want)
		}
	})

	t.Run("multiple sources", func(t *testing.T) {
		buf := setupTestFs(t, testFixture)
		writeTestFile(t, "other/slides.md", testFixture)
		cmd := SlidesCmd{Exprs: []string{"1,3", "other/2-4"}}
		if err := cmd.Run(); err != nil {
			t.Fatalf("SlidesCmd.Run() error: %v", err)
		}
		want := "Test\n  1  A\n  3  C\n\nTest\n  2  B\n  3  C\n  4  D\n"
		if got := buf.String(); got != want {
			t.Errorf("SlidesCmd output:\ngot:  %q\nwant: %q", got, want)
		}
	})

	t.Run("directory without slash", func(t *testing.T) {
		buf := setupTestFs(t, testFixture)
		writeTestFile(t, "mydir/slides.md", testFixture)
		cmd := SlidesCmd{Exprs: []string{"mydir"}}
		if err := cmd.Run(); err != nil {
			t.Fatalf("SlidesCmd.Run() error: %v", err)
		}
		want := "Test\n  1  A\n  2  B\n  3  C\n  4  D\n"
		if got := buf.String(); got != want {
			t.Errorf("SlidesCmd output:\ngot:  %q\nwant: %q", got, want)
		}
	})

	t.Run("glob lists all decks under dir", func(t *testing.T) {
		buf := setupTestFs(t, testFixture)
		writeTestFile(t, "talks/talk1/slides.md", testFixture)
		writeTestFile(t, "talks/talk2/slides.md", testFixture2)
		writeTestFile(t, "talks/README.md", "# notes\n")
		_ = appFs.MkdirAll("talks/empty", 0o755)
		cmd := SlidesCmd{Exprs: []string{"talks/*"}}
		if err := cmd.Run(); err != nil {
			t.Fatalf("SlidesCmd.Run() error: %v", err)
		}
		// README.md and empty/ are skipped; the two real decks render their
		// titles. Order is glob-sorted alphabetically.
		got := buf.String()
		want := "Test\n  1  A\n  2  B\n  3  C\n  4  D\n\nTest2\n  1  X\n  2  Y\n  3  Z\n"
		if got != want {
			t.Errorf("SlidesCmd output:\ngot:  %q\nwant: %q", got, want)
		}
	})
}
