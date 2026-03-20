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

	// failingChecks lists the modes whose errors cause a non-zero exit.
	// To fail on both checks, add commitchecker.NoAncestryPath: true to the map.
	failingChecks := map[commitchecker.CheckMode]bool{commitchecker.AncestryPath: true}

	var failErrs []string
	for _, mode := range []commitchecker.CheckMode{commitchecker.NoAncestryPath, commitchecker.AncestryPath} {
		// NoAncestryPath uses the CLI-provided start (e.g. PULL_BASE_SHA) so that we
		// find PR commits even when the branch has fallen behind main.
		// AncestryPath uses the computed start (upstream merge base) to walk the correct
		// linear carry path through the downstream repo.
		checkStart := opts.Start
		if mode == commitchecker.AncestryPath {
			checkStart = start
		}
		commits, err := commitchecker.CommitsBetween(checkStart, opts.End, mode)
		if err != nil {
			if err == commitchecker.ErrNotCommit {
				_, _ = fmt.Fprintf(os.Stderr, "WARNING: one of the provided commits does not exist, not a true branch\n")
				os.Exit(0)
			}
			_, _ = fmt.Fprintf(os.Stderr, "ERROR: couldn't find commits from %s..%s (%s): %v\n", checkStart, opts.End, mode, err)
			os.Exit(1)
		}

		_, _ = fmt.Fprintf(os.Stdout, "Validating %d commits between %s...%s (%s)\n", len(commits), checkStart, opts.End, mode)
		for _, commit := range commits {
			_, _ = fmt.Fprintf(os.Stdout, "Validating commit %+v (%s)\n", commit, mode)
			for _, validate := range commitchecker.AllCommitValidators {
				for _, e := range validate(commit) {
					msg := fmt.Sprintf("[%s] %s", mode, e)
					_, _ = fmt.Fprintf(os.Stderr, "%s\n\n", msg)
					if failingChecks[mode] {
						failErrs = append(failErrs, msg)
					}
				}
			}
		}
	}

	if len(failErrs) > 0 {
		os.Exit(2)
	}
}
