// Package e2e provides shared fixture builders and run helpers for
// the commit-in end-to-end test suites (Steps 9–12 of the commit-in
// implementation plan).
//
// All builders shell out to the real `git` binary so the tests cover
// the same code paths users hit in production. Tests that need a fast
// fake-runner path should stay in their respective sub-packages
// (walk, replay, etc.) — this package is intentionally heavy.
package e2e

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// Repo is a real on-disk git repo built inside t.TempDir(). Cleanup is
// automatic via t.Cleanup; callers never need to RemoveAll.
type Repo struct {
	t    *testing.T
	Path string
}

// NewRepo creates an empty repo at <tempDir>/<name> with a deterministic
// initial branch (`main`) and identity. Skips the test if `git` is not
// in PATH so CI without git degrades cleanly.
func NewRepo(t *testing.T, name string) *Repo {
	t.Helper()
	if _, err := exec.LookPath("git"); err != nil {
		t.Skipf("git not available: %v", err)
	}
	dir := filepath.Join(t.TempDir(), name)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatalf("mkdir %s: %v", dir, err)
	}
	r := &Repo{t: t, Path: dir}
	r.git("init", "-b", "main")
	r.git("config", "user.email", "e2e@gitmap.test")
	r.git("config", "user.name", "E2E Bot")
	r.git("config", "commit.gpgsign", "false")
	return r
}

// Commit writes `path` with `body`, stages it, and commits with
// `message` at `when` (both author and committer date). Returns the
// new commit SHA so callers can assert on dedupe behavior.
func (r *Repo) Commit(path, body, message string, when time.Time) string {
	r.t.Helper()
	full := filepath.Join(r.Path, path)
	if err := os.MkdirAll(filepath.Dir(full), 0o755); err != nil {
		r.t.Fatalf("mkdir parent of %s: %v", full, err)
	}
	if err := os.WriteFile(full, []byte(body), 0o644); err != nil {
		r.t.Fatalf("write %s: %v", full, err)
	}
	// Use plumbing `update-index --add` instead of porcelain `add` so
	// the harness works inside sandboxes that block `git add`.
	r.git("update-index", "--add", path)
	stamp := when.UTC().Format(time.RFC3339)
	cmd := exec.Command("git", "commit", "-m", message, "--date", stamp)
	cmd.Dir = r.Path
	cmd.Env = append(os.Environ(),
		"GIT_AUTHOR_DATE="+stamp,
		"GIT_COMMITTER_DATE="+stamp,
		"GIT_AUTHOR_NAME=E2E Bot",
		"GIT_AUTHOR_EMAIL=e2e@gitmap.test",
		"GIT_COMMITTER_NAME=E2E Bot",
		"GIT_COMMITTER_EMAIL=e2e@gitmap.test",
	)
	if out, err := cmd.CombinedOutput(); err != nil {
		r.t.Fatalf("git commit %q: %v\n%s", message, err, out)
	}
	return r.headSha()
}

// git runs a git subcommand inside r.Path and fatals on error. Used
// for setup-only commands where output is uninteresting.
func (r *Repo) git(args ...string) {
	r.t.Helper()
	cmd := exec.Command("git", args...)
	cmd.Dir = r.Path
	if out, err := cmd.CombinedOutput(); err != nil {
		r.t.Fatalf("git %s: %v\n%s", strings.Join(args, " "), err, out)
	}
}

// headSha returns the current HEAD SHA; fatals on error.
func (r *Repo) headSha() string {
	r.t.Helper()
	cmd := exec.Command("git", "rev-parse", "HEAD")
	cmd.Dir = r.Path
	out, err := cmd.Output()
	if err != nil {
		r.t.Fatalf("rev-parse HEAD: %v", err)
	}
	return strings.TrimSpace(string(out))
}

// MustExist fatals if `rel` (relative to r.Path) is missing.
func (r *Repo) MustExist(rel string) {
	r.t.Helper()
	if _, err := os.Stat(filepath.Join(r.Path, rel)); err != nil {
		r.t.Fatalf("expected %s to exist: %v", rel, err)
	}
}

// String renders for %v formatting — useful in test failure output.
func (r *Repo) String() string { return fmt.Sprintf("Repo(%s)", r.Path) }
