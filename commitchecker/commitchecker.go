package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/openshift/build-machinery-go/commitchecker/pkg/commitchecker"
	"github.com/openshift/build-machinery-go/commitchecker/pkg/version"
)

func main() {
	_, _ = fmt.Fprintf(os.Stdout, "commitchecker verson %v\n", version.Get().String())
	opts := commitchecker.DefaultOptions()
	_, _ = fmt.Fprintf(os.Stdout, "default options: %+v\n", opts)
	opts.Bind(flag.CommandLine)
	flag.Parse()

	if err := opts.Validate(); err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "ERROR: invalid flags: %v\n", err)
		os.Exit(1)
	}

	cfg, err := commitchecker.Load(opts.ConfigFile)
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "ERROR: couldn't load config: %v\n", err)
		os.Exit(1)
	}

	_, _ = fmt.Fprintf(os.Stdout, "post-argument options: %+v\n", opts)
	if cfg != nil {
		_, _ = fmt.Fprintf(os.Stdout, "config: %+v\n", cfg)
	} else {
		_, _ = fmt.Fprintf(os.Stdout, "config: (no config file found)\n")
	}

	mergeBase, err := commitchecker.DetermineMergeBase(cfg, commitchecker.FetchMode(opts.FetchMode), opts.End)
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "ERROR: couldn't determine merge base: %v\n", err)
		os.Exit(1)
	}
	start := opts.Start
	if mergeBase != "" && cfg != nil {
		_, _ = fmt.Fprintf(os.Stdout, "Determined merge-base with %s/%s@%s at %s\n", cfg.UpstreamOrg, cfg.UpstreamRepo, cfg.UpstreamBranch, mergeBase)
		start = mergeBase
	}

	// Diagnostic: find all commits reachable from end but not start. This catches
	// every PR commit regardless of branch topology, so we can detect when the
	// stricter direct-ancestry check below silently drops commits on stale branches.
	allCommits := listCommits(commitchecker.AllCommitsBetween(opts.Start, opts.End))
	_, _ = fmt.Fprintf(os.Stdout, "Validating all %d commits between %s..%s\n", len(allCommits), opts.Start, opts.End)
	validateCommits(allCommits)

	// Enforced: find only commits on the direct descent path from the upstream
	// merge-base to end. This excludes upstream commits reached via side branches
	// in the DAG created by the downstream rebase process.
	directCommits := listCommits(commitchecker.DirectCommitsBetween(start, opts.End))
	_, _ = fmt.Fprintf(os.Stdout, "Validating %d direct commits between %s..%s\n", len(directCommits), start, opts.End)
	if errs := validateCommits(directCommits); len(errs) > 0 {
		os.Exit(2)
	}
}

func listCommits(commits []commitchecker.Commit, err error) []commitchecker.Commit {
	if err != nil {
		if err == commitchecker.ErrNotCommit {
			_, _ = fmt.Fprintf(os.Stderr, "WARNING: one of the provided commits does not exist, not a true branch\n")
			os.Exit(0)
		}
		_, _ = fmt.Fprintf(os.Stderr, "ERROR: couldn't list commits: %v\n", err)
		os.Exit(1)
	}
	return commits
}

func validateCommits(commits []commitchecker.Commit) []string {
	var errs []string
	for _, commit := range commits {
		_, _ = fmt.Fprintf(os.Stdout, "Validating commit %+v\n", commit)
		for _, validate := range commitchecker.AllCommitValidators {
			for _, e := range validate(commit) {
				_, _ = fmt.Fprintf(os.Stderr, "%s\n\n", e)
				errs = append(errs, e)
			}
		}
	}
	return errs
}
