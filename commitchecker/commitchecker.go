package main

import (
	"flag"
	"fmt"
	"os"
)

type FetchMode string

const (
	HTTPS FetchMode = "https"
	SSH   FetchMode = "ssh"
)

type Options struct {
	Start      string
	End        string
	ConfigFile string
	FetchMode  string
}

func DefaultOptions() Options {
	return Options{
		Start:      "main",
		End:        "HEAD",
		ConfigFile: "./commitchecker.yaml",
		FetchMode:  string(HTTPS),
	}
}

func (o *Options) Bind(fs *flag.FlagSet) {
	fs.StringVar(&o.Start, "start", o.Start, "The start of the revision range for analysis")
	fs.StringVar(&o.End, "end", o.End, "The end of the revision range for analysis")
	fs.StringVar(&o.ConfigFile, "config", o.ConfigFile, "The configuration file to use (optional).")
	fs.StringVar(&o.FetchMode, "fetch-mode", o.FetchMode, "Method to use for fetching from git remotes.")
}

func (o *Options) Validate() error {
	switch FetchMode(o.FetchMode) {
	case SSH, HTTPS:
	default:
		return fmt.Errorf("--fetch-mode must be one of %v", []FetchMode{HTTPS, SSH})
	}
	return nil
}

func main() {
	opts := DefaultOptions()
	opts.Bind(flag.CommandLine)
	flag.Parse()

	if err := opts.Validate(); err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "ERROR: invalid flags: %v\n", err)
		os.Exit(1)
	}

	cfg, err := Load(opts.ConfigFile)
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "ERROR: couldn't load config: %v\n", err)
		os.Exit(1)
	}

	mergeBase, err := DetermineMergeBase(cfg, FetchMode(opts.FetchMode), opts.End)
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "ERROR: couldn't determine merge base: %v\n", err)
		os.Exit(1)
	}
	start := opts.Start
	if mergeBase != "" {
		start = mergeBase
	}

	commits, err := CommitsBetween(start, opts.End)
	if err != nil {
		if err == ErrNotCommit {
			_, _ = fmt.Fprintf(os.Stderr, "WARNING: one of the provided commits does not exist, not a true branch\n")
			os.Exit(0)
		}
		_, _ = fmt.Fprintf(os.Stderr, "ERROR: couldn't find commits from %s..%s: %v\n", opts.Start, opts.End, err)
		os.Exit(1)
	}

	var errs []string
	for _, validate := range allCommitValidators {
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
