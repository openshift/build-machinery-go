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
	_, _ = fmt.Fprintf(os.Stdout, "config: %+v\n", *cfg)

	mergeBase, err := commitchecker.DetermineMergeBase(cfg, commitchecker.FetchMode(opts.FetchMode), opts.End)
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "ERROR: couldn't determine merge base: %v\n", err)
		os.Exit(1)
	}
	start := opts.Start
	if mergeBase != "" {
		_, _ = fmt.Fprintf(os.Stdout, "Determined merge-base with %s/%s@%s at %s\n", cfg.UpstreamOrg, cfg.UpstreamRepo, cfg.UpstreamBranch, mergeBase)
		start = mergeBase
	}

	commits, err := commitchecker.CommitsBetween(start, opts.End)
	if err != nil {
		if err == commitchecker.ErrNotCommit {
			_, _ = fmt.Fprintf(os.Stderr, "WARNING: one of the provided commits does not exist, not a true branch\n")
			os.Exit(0)
		}
		_, _ = fmt.Fprintf(os.Stderr, "ERROR: couldn't find commits from %s..%s: %v\n", opts.Start, opts.End, err)
		os.Exit(1)
	}

	_, _ = fmt.Fprintf(os.Stdout, "Validating %d commits between %s...%s\n", len(commits), start, opts.End)
	var errs []string
	for _, commit := range commits {
		_, _ = fmt.Fprintf(os.Stdout, "Validating commit %+v\n", commit)
		for _, validate := range commitchecker.AllCommitValidators {
			errs = append(errs, validate(commit)...)
		}
	}

	if len(errs) > 0 {
		for _, e := range errs {
			_, _ = fmt.Fprintf(os.Stderr, "%s\n\n", e)
		}

		os.Exit(2)
	}
}
