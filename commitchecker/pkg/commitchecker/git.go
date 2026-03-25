package commitchecker

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"strings"
)

var (
	MergeSummaryPattern    = regexp.MustCompile(`^Merge commit .*`)
	UpstreamSummaryPattern = regexp.MustCompile(`^UPSTREAM: (revert: )?(([\w.-]+/[\w-.-]+)?: )?(\d+:|<carry>:|<drop>:)`)
)

type Commit struct {
	Sha     string
	Summary string
	Email   string
}

func (c Commit) MatchesMergeSummaryPattern() bool {
	return MergeSummaryPattern.MatchString(c.Summary)
}

func (c Commit) MatchesUpstreamSummaryPattern() bool {
	return UpstreamSummaryPattern.MatchString(c.Summary)
}

func IsCommit(a string) bool {
	if _, _, err := run("git", "rev-parse", a); err != nil {
		return false
	}
	return true
}

var ErrNotCommit = fmt.Errorf("one or both of the provided commits was not a valid commit")
var ErrNoCommitsFound = fmt.Errorf("no commits found between start and end, but they resolve to different commits")

// SameCommit returns true if a and b resolve to the same commit SHA.
func SameCommit(a, b string) (bool, error) {
	shaA, stderr, err := run("git", "rev-parse", a)
	if err != nil {
		return false, fmt.Errorf("rev-parse %s: %s: %w", a, stderr, err)
	}
	shaB, stderr, err := run("git", "rev-parse", b)
	if err != nil {
		return false, fmt.Errorf("rev-parse %s: %s: %w", b, stderr, err)
	}
	return strings.TrimSpace(shaA) == strings.TrimSpace(shaB), nil
}

// DirectCommitsBetween returns commits on the direct ancestry path from a to b,
// using --ancestry-path to exclude commits reachable via side branches.
func DirectCommitsBetween(a, b string) ([]Commit, error) {
	return gitLog(a, b, "--no-merges", "--ancestry-path")
}

// AllCommitsBetween returns all non-merge commits reachable from b but not from a.
func AllCommitsBetween(a, b string) ([]Commit, error) {
	return gitLog(a, b, "--no-merges")
}

func gitLog(a, b string, extraArgs ...string) ([]Commit, error) {
	var commits []Commit
	args := []string{"git", "log", "--oneline"}
	args = append(args, extraArgs...)
	args = append(args, fmt.Sprintf("%s..%s", a, b))
	stdout, stderr, err := run(args...)
	if err != nil {
		if !IsCommit(a) || !IsCommit(b) {
			return nil, ErrNotCommit
		}
		return nil, fmt.Errorf("error executing %s: %s: %s", strings.Join(args, " "), stderr, err)
	}
	for _, log := range strings.Split(stdout, "\n") {
		if len(log) == 0 {
			continue
		}
		commit, err := NewCommitFromOnelineLog(log)
		if err != nil {
			return nil, err
		}
		commits = append(commits, commit)
	}
	return commits, nil
}

func NewCommitFromOnelineLog(log string) (Commit, error) {
	var commit Commit
	var err error
	parts := strings.Split(log, " ")
	if len(parts) < 2 {
		return commit, fmt.Errorf("invalid log entry: %s", log)
	}
	commit.Sha = parts[0]
	commit.Summary = strings.Join(parts[1:], " ")
	commit.Email, err = emailInCommit(commit.Sha)
	if err != nil {
		return commit, err
	}
	return commit, nil
}

func emailInCommit(sha string) (string, error) {
	stdout, stderr, err := run("git", "show", `--format=%ae`, "-s", sha)
	if err != nil {
		return "", fmt.Errorf("%s: %s", stderr, err)
	}
	return strings.TrimSpace(stdout), nil
}

func run(args ...string) (string, string, error) {
	cmd := exec.Command(args[0], args[1:]...)
	_, _ = fmt.Fprintf(os.Stdout, "running: %+v\n", cmd)
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err := cmd.Run()
	return stdout.String(), stderr.String(), err
}

// DetermineMergeBase will determine the merge-base between the upstream in configuration
// and end, if we can. This processing is best-effort; if an expected merge-base is present
// in the configuration, we validate that it matches.
func DetermineMergeBase(cfg *Config, fetchMode FetchMode, end string) (string, error) {
	if cfg == nil || cfg.UpstreamOrg == "" || cfg.UpstreamRepo == "" || cfg.UpstreamBranch == "" {
		return "", nil
	}

	var upstreamRemote string
	switch fetchMode {
	case SSH:
		upstreamRemote = "git@github.com:" + cfg.UpstreamOrg + "/" + cfg.UpstreamRepo + ".git"
	case HTTPS:
		upstreamRemote = "https://github.com/" + cfg.UpstreamOrg + "/" + cfg.UpstreamRepo + ".git"
	default:
		return "", fmt.Errorf("unexpected fetch mode %s", fetchMode)
	}
	{
		stdout, stderr, err := run("git", "fetch", upstreamRemote, cfg.UpstreamBranch)
		if err != nil {
			return "", fmt.Errorf("failed to fetch upstream: %s, %s: %w", stdout, stderr, err)
		}
	}

	var mergeBase string
	{
		stdout, stderr, err := run("git", "rev-parse", end, "FETCH_HEAD")
		if err != nil {
			return "", fmt.Errorf("failed to rev-parse %s FETCH_HEAD: %s, %s: %w", end, stdout, stderr, err)
		}
		_, _ = fmt.Fprintf(os.Stdout, "output: %s\n", strings.ReplaceAll(strings.Trim(stdout, "\n"), "\n", "\n        "))
		stdout, stderr, err = run("git", "merge-base", end, "FETCH_HEAD")
		if err != nil {
			return "", fmt.Errorf("failed to fetch upstream: %s, %s: %w", stdout, stderr, err)
		}
		_, _ = fmt.Fprintf(os.Stdout, "output: %s\n", strings.ReplaceAll(strings.Trim(stdout, "\n"), "\n", "\n        "))
		mergeBase = strings.TrimSpace(stdout)
	}

	if cfg.ExpectedMergeBase != "" && mergeBase != cfg.ExpectedMergeBase {
		return mergeBase, fmt.Errorf("detected merge-base %q, expected %q", mergeBase, cfg.ExpectedMergeBase)
	}

	return mergeBase, nil
}
