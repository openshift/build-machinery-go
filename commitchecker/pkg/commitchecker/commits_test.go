package commitchecker_test

import (
	"os"
	"os/exec"
	"testing"

	"github.com/openshift/build-machinery-go/commitchecker/pkg/commitchecker"
)

// initRepo creates a bare-bones git repo in a temp dir, changes into it, and
// returns a cleanup function that restores the original working directory.
func initRepo(t *testing.T) func() {
	t.Helper()
	origDir, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	dir := t.TempDir()
	if err := os.Chdir(dir); err != nil {
		t.Fatal(err)
	}
	git(t, "init")
	git(t, "config", "user.email", "test@test.com")
	git(t, "config", "user.name", "Test")
	return func() {
		if err := os.Chdir(origDir); err != nil {
			t.Fatal(err)
		}
	}
}

// git runs a git command and fails the test on error.
func git(t *testing.T, args ...string) string {
	t.Helper()
	cmd := exec.Command("git", args...)
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("git %v failed: %s\n%s", args, err, out)
	}
	return string(out)
}

// commit creates an empty commit with the given message and returns its SHA.
func commit(t *testing.T, msg string) string {
	t.Helper()
	git(t, "commit", "--allow-empty", "-m", msg)
	out := git(t, "rev-parse", "HEAD")
	// trim trailing newline
	return out[:len(out)-1]
}

// TestStaleBranch verifies that DirectCommitsBetween returns no commits when
// the PR branch has fallen behind main, while AllCommitsBetween finds them.
//
// DAG:
//
//	main:   A --- B --- C --- D
//	         \                 \
//	          E --- F --- G (merge)
func TestStaleBranch(t *testing.T) {
	cleanup := initRepo(t)
	defer cleanup()

	// A: initial commit on main
	commit(t, "UPSTREAM: <carry>: A initial")
	git(t, "checkout", "-b", "main")

	// Branch off at A for the PR
	git(t, "checkout", "-b", "pr-branch")
	commit(t, "UPSTREAM: <carry>: E pr commit 1")
	commit(t, "UPSTREAM: <carry>: F pr commit 2")

	// Advance main past A
	git(t, "checkout", "main")
	commit(t, "UPSTREAM: <carry>: B advance")
	commit(t, "UPSTREAM: <carry>: C advance")
	startSHA := commit(t, "UPSTREAM: <carry>: D advance")

	// Merge PR onto main (simulating CI merge).
	// The merge commit itself is filtered by --no-merges, so --ancestry-path
	// returns zero non-merge commits when E and F are not descendants of D.
	git(t, "merge", "pr-branch", "--no-ff", "-m", "Merge PR")
	mergeSHA := git(t, "rev-parse", "HEAD")
	mergeSHA = mergeSHA[:len(mergeSHA)-1]

	// AllCommitsBetween should find the PR commits (E, F)
	allCommits, err := commitchecker.AllCommitsBetween(startSHA, mergeSHA)
	if err != nil {
		t.Fatalf("AllCommitsBetween: %v", err)
	}
	if len(allCommits) == 0 {
		t.Error("AllCommitsBetween returned 0 commits; expected PR commits to be found")
	}

	// DirectCommitsBetween should find no non-merge commits because E and F
	// are not descendants of D (the stale branch problem).
	directCommits, err := commitchecker.DirectCommitsBetween(startSHA, mergeSHA)
	if err != nil {
		t.Fatalf("DirectCommitsBetween: %v", err)
	}

	// The key assertion: all commits found, but direct commits missed them.
	if len(allCommits) > 0 && len(directCommits) == 0 {
		t.Logf("Correctly detected stale branch: %d all commits, %d direct commits", len(allCommits), len(directCommits))
	} else if len(directCommits) > 0 {
		t.Errorf("Expected DirectCommitsBetween to return 0 commits on a stale branch, got %d", len(directCommits))
	}
}

// TestUpToDateBranch verifies that both AllCommitsBetween and
// DirectCommitsBetween find the same commits when the branch is up to date.
//
// DAG:
//
//	main:   A --- B(start) --- C --- D(end)
func TestUpToDateBranch(t *testing.T) {
	cleanup := initRepo(t)
	defer cleanup()

	commit(t, "UPSTREAM: <carry>: A initial")
	git(t, "checkout", "-b", "main")
	startSHA := commit(t, "UPSTREAM: <carry>: B start")
	commit(t, "UPSTREAM: <carry>: C middle")
	commit(t, "UPSTREAM: <carry>: D end")

	allCommits, err := commitchecker.AllCommitsBetween(startSHA, "HEAD")
	if err != nil {
		t.Fatalf("AllCommitsBetween: %v", err)
	}

	directCommits, err := commitchecker.DirectCommitsBetween(startSHA, "HEAD")
	if err != nil {
		t.Fatalf("DirectCommitsBetween: %v", err)
	}

	if len(allCommits) != len(directCommits) {
		t.Errorf("Expected same commit count, got all=%d direct=%d", len(allCommits), len(directCommits))
	}
	if len(allCommits) != 2 {
		t.Errorf("Expected 2 commits (C, D), got %d", len(allCommits))
	}
}
