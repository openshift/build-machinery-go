package main

import (
	"bytes"
	"fmt"
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

func CommitsBetween(a, b string) ([]Commit, error) {
	var commits []Commit
	stdout, stderr, err := run("git", "log", "--oneline", fmt.Sprintf("%s..%s", a, b))
	if err != nil {
		if !IsCommit(a) || !IsCommit(b) {
			return nil, ErrNotCommit
		}
		return nil, fmt.Errorf("error executing git log: %s: %s", stderr, err)
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
	if err != nil {
		return commit, err
	}
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
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err := cmd.Run()
	return stdout.String(), stderr.String(), err
}
