package cmd

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func TestIsPathInside(t *testing.T) {
	cases := []struct {
		name   string
		child  string
		parent string
		want   bool
	}{
		{"equal", "/a/b", "/a/b", true},
		{"descendant", "/a/b/c", "/a/b", true},
		{"sibling", "/a/c", "/a/b", false},
		{"parent-of", "/a", "/a/b", false},
		{"case-fold", "/A/B", "/a/b", true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := isPathInside(filepath.Clean(tc.child), filepath.Clean(tc.parent))
			if got != tc.want {
				t.Fatalf("isPathInside(%q,%q)=%v want %v",
					tc.child, tc.parent, got, tc.want)
			}
		})
	}
}

func TestEscapeCwdIfInside_NotInside(t *testing.T) {
	target := t.TempDir()
	other := t.TempDir()

	if err := os.Chdir(other); err != nil {
		t.Fatalf("chdir other: %v", err)
	}

	got, err := escapeCwdIfInside(target)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.EqualFold(filepath.Clean(got), filepath.Clean(other)) {
		t.Fatalf("cwd should be unchanged; got %q want %q", got, other)
	}
}

func TestEscapeCwdIfInside_EscapesWhenInside(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("temp-dir symlink resolution differs on Windows CI; behavior covered by integration tests")
	}

	target, err := filepath.EvalSymlinks(t.TempDir())
	if err != nil {
		t.Fatalf("evalsymlinks: %v", err)
	}

	if err := os.Chdir(target); err != nil {
		t.Fatalf("chdir target: %v", err)
	}

	got, err := escapeCwdIfInside(target)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	wantParent := filepath.Dir(target)
	if !strings.EqualFold(filepath.Clean(got), filepath.Clean(wantParent)) {
		t.Fatalf("escape landed in %q want parent %q", got, wantParent)
	}

	cwd, _ := os.Getwd()
	resolved, _ := filepath.EvalSymlinks(cwd)
	if !strings.EqualFold(filepath.Clean(resolved), filepath.Clean(wantParent)) {
		t.Fatalf("os cwd %q (resolved %q) != parent %q", cwd, resolved, wantParent)
	}
}
