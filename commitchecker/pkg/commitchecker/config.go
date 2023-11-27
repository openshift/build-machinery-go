package commitchecker

import (
	"fmt"
	"os"

	"sigs.k8s.io/yaml"
)

// Config determines how the commit-checker runs
type Config struct {
	// UpstreamOrg is the organization of the upstream repository
	UpstreamOrg string `json:"upstreamOrg,omitempty"`
	// UpstreamRepo is the repo name of the upstream repository
	UpstreamRepo string `json:"upstreamRepo,omitempty"`
	// UpstreamBranch is the branch from the upstream repository we're tracking
	UpstreamBranch string `json:"upstreamBranch,omitempty"`
	// ExpectedMergeBase is the latest commit from the upstream that is expected to be present in this downstream
	ExpectedMergeBase string `json:"expectedMergeBase,omitempty"`
}

// Load returns a configuration from the path, if one exists. If no configuration file is present,
// there is no error.
func Load(path string) (*Config, error) {
	_, err := os.Stat(path)
	if err != nil && !os.IsNotExist(err) {
		return nil, fmt.Errorf("failed to stat configfile %q: %w", path, err)
	}
	if os.IsNotExist(err) {
		return nil, nil
	}

	raw, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read configfile %q: %w", path, err)
	}

	var cfg Config
	if err := yaml.Unmarshal(raw, &cfg); err != nil {
		return nil, fmt.Errorf("failed to unmarshal configfile %q: %w", path, err)
	}
	return &cfg, nil
}
