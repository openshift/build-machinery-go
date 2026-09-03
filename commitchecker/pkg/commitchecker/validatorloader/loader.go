package validatorloader

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

func Load(source string) ([]byte, error) {
	switch {
	case strings.HasPrefix(source, "https://") || strings.HasPrefix(source, "http://"):
		return loadHTTP(source)
	case strings.HasPrefix(source, "git://"):
		return loadGit(source)
	default:
		return loadLocal(source)
	}
}

func loadLocal(path string) ([]byte, error) {
	absPath, err := filepath.Abs(path)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve path %q: %w", path, err)
	}
	data, err := os.ReadFile(absPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read validator from %q: %w", absPath, err)
	}
	return data, nil
}

var httpClient = &http.Client{Timeout: 30 * time.Second}

func loadHTTP(url string) ([]byte, error) {
	resp, err := httpClient.Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch validator from %q: %w", url, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to fetch validator from %q: HTTP %d", url, resp.StatusCode)
	}

	const maxSize = 1 << 20 // 1 MB
	data, err := io.ReadAll(io.LimitReader(resp.Body, maxSize))
	if err != nil {
		return nil, fmt.Errorf("failed to read response from %q: %w", url, err)
	}
	if len(data) == maxSize {
		return nil, fmt.Errorf("validator from %q exceeds 1 MB size limit", url)
	}
	return data, nil
}

// loadGit loads a file from a git repository.
// Format: git://host/org/repo@ref:filepath
// Example: git://github.com/openshift/release@main:ci-operator/validators/vendor.go
func loadGit(source string) ([]byte, error) {
	trimmed := strings.TrimPrefix(source, "git://")

	atIdx := strings.Index(trimmed, "@")
	if atIdx < 0 {
		return nil, fmt.Errorf("invalid git source %q: missing @ref (expected git://host/org/repo@ref:path)", source)
	}
	repoURL := "https://" + trimmed[:atIdx]
	rest := trimmed[atIdx+1:]

	colonIdx := strings.Index(rest, ":")
	if colonIdx < 0 {
		return nil, fmt.Errorf("invalid git source %q: missing :path (expected git://host/org/repo@ref:path)", source)
	}
	ref := rest[:colonIdx]
	filePath := rest[colonIdx+1:]

	if ref == "" || filePath == "" {
		return nil, fmt.Errorf("invalid git source %q: ref and path must not be empty", source)
	}

	tmpDir, err := os.MkdirTemp("", "commitchecker-git-*")
	if err != nil {
		return nil, fmt.Errorf("failed to create temp dir: %w", err)
	}
	defer os.RemoveAll(tmpDir)

	cmd := exec.Command("git", "init", tmpDir)
	if output, err := cmd.CombinedOutput(); err != nil {
		return nil, fmt.Errorf("failed to init temp repo: %s: %w", output, err)
	}

	cmd = exec.Command("git", "-C", tmpDir, "fetch", "--depth=1", repoURL, ref)
	if output, err := cmd.CombinedOutput(); err != nil {
		return nil, fmt.Errorf("failed to fetch %s@%s: %s: %w", repoURL, ref, output, err)
	}

	cmd = exec.Command("git", "-C", tmpDir, "show", "FETCH_HEAD:"+filePath)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("failed to read %q from %s@%s: %s: %w", filePath, repoURL, ref, output, err)
	}
	return output, nil
}

func IsDynamicSource(entry string) bool {
	return strings.HasPrefix(entry, "/") ||
		strings.HasPrefix(entry, "./") ||
		strings.HasPrefix(entry, "git://") ||
		strings.HasPrefix(entry, "http://") ||
		strings.HasPrefix(entry, "https://") ||
		strings.HasSuffix(entry, ".go")
}
