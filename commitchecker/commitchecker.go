package main

import (
	"flag"
	"fmt"
	"os"

	commitchecker2 "github.com/openshift/build-machinery-go/commitchecker/pkg/commitchecker"
	"github.com/openshift/build-machinery-go/commitchecker/pkg/version"
)

func main() {
	_, _ = fmt.Fprintf(os.Stdout, "commitchecker verson %v\n", version.Get().String())
	opts := commitchecker2.DefaultOptions()
	opts.Bind(flag.CommandLine)
	flag.Parse()

	if err := opts.Validate(); err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "ERROR: invalid flags: %v\n", err)
		os.Exit(1)
	}

	cfg, err := commitchecker2.Load(opts.ConfigFile)
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "ERROR: couldn't load config: %v\n", err)
		os.Exit(1)
	}

	mergeBase, err := commitchecker2.DetermineMergeBase(cfg, commitchecker2.FetchMode(opts.FetchMode), opts.End)
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "ERROR: couldn't determine merge base: %v\n", err)
		os.Exit(1)
	}
	start := opts.Start
	if mergeBase != "" {
		start = mergeBase
	}

	commits, err := commitchecker2.CommitsBetween(start, opts.End)
	if err != nil {
		if err == commitchecker2.ErrNotCommit {
			_, _ = fmt.Fprintf(os.Stderr, "WARNING: one of the provided commits does not exist, not a true branch\n")
			os.Exit(0)
		}
		_, _ = fmt.Fprintf(os.Stderr, "ERROR: couldn't find commits from %s..%s: %v\n", opts.Start, opts.End, err)
		os.Exit(1)
	}

	var errs []string
	for _, validate := range commitchecker2.AllCommitValidators {
		for _, commit := range commits {
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
