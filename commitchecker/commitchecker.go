package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"github.com/openshift/build-machinery-go/commitchecker/pkg/commitchecker"
	"github.com/openshift/build-machinery-go/commitchecker/pkg/commitchecker/validatorloader"
	"github.com/openshift/build-machinery-go/commitchecker/pkg/commitchecker/validatorruntime"
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

	commits, err := commitchecker.CommitsBetween(start, opts.End)
	if err != nil {
		if err == commitchecker.ErrNotCommit {
			_, _ = fmt.Fprintf(os.Stderr, "WARNING: one of the provided commits does not exist, not a true branch\n")
			os.Exit(0)
		}
		_, _ = fmt.Fprintf(os.Stderr, "ERROR: couldn't find commits from %s..%s: %v\n", opts.Start, opts.End, err)
		os.Exit(1)
	}

	var entries []string
	if cliValidators := opts.ValidatorsList(); len(cliValidators) > 0 {
		entries = cliValidators
	} else if cfg != nil && len(cfg.Validators) > 0 {
		entries = cfg.Validators
	}

	validators, err := resolveValidators(entries)
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "ERROR: invalid validators: %v\n", err)
		os.Exit(1)
	}

	_, _ = fmt.Fprintf(os.Stdout, "Validating %d commits between %s...%s\n", len(commits), start, opts.End)
	var errs []error
	for _, commit := range commits {
		_, _ = fmt.Fprintf(os.Stdout, "Validating commit %+v\n", commit)
		for _, v := range validators {
			for _, e := range v.Validate(&commit) {
				errs = append(errs, fmt.Errorf("FAILED: [%s] commit: %s, error: %q, title: %q", v.Name(), commit.Sha, e, commit.Summary))
			}
		}
	}

	if len(errs) > 0 {
		_, _ = fmt.Fprintf(os.Stderr, "--------------------------------\n")
		_, _ = fmt.Fprintf(os.Stderr, "Validation found %d issues:\n", len(errs))
		_, _ = fmt.Fprintf(os.Stderr, "--------------------------------\n")
		for _, e := range errs {
			_, _ = fmt.Fprintf(os.Stderr, "%s\n", e)
		}

		os.Exit(2)
	}

	_, _ = fmt.Fprintf(os.Stdout, "Validation completed successfully\n")
}

func resolveValidators(entries []string) ([]commitchecker.Validator, error) {
	if len(entries) == 0 {
		return commitchecker.ValidatorsForNames(nil)
	}

	engine := validatorruntime.NewEngine()
	var validators []commitchecker.Validator

	for _, entry := range entries {
		if validatorloader.IsDynamicSource(entry) {
			src, err := validatorloader.Load(entry)
			if err != nil {
				return nil, fmt.Errorf("failed to load validator %q: %w", entry, err)
			}
			name := validatorNameFromSource(entry)
			v, err := engine.LoadValidator(name, src)
			if err != nil {
				return nil, fmt.Errorf("failed to initialize validator %q: %w", entry, err)
			}
			validators = append(validators, v)
		} else {
			resolved, err := commitchecker.ValidatorsForNames([]string{entry})
			if err != nil {
				return nil, err
			}
			validators = append(validators, resolved...)
		}
	}

	return validators, nil
}

func validatorNameFromSource(source string) string {
	base := filepath.Base(source)
	ext := filepath.Ext(base)
	if ext != "" {
		return base[:len(base)-len(ext)]
	}
	return base
}
