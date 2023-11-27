package commitchecker

import (
	"flag"
	"fmt"
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
