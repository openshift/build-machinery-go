package validatorloader

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestLoadLocal(t *testing.T) {
	dir := t.TempDir()
	content := []byte("package myvalidator\n\nfunc Validate() {}\n")
	path := filepath.Join(dir, "test_validator.go")
	if err := os.WriteFile(path, content, 0644); err != nil {
		t.Fatal(err)
	}

	tt := []struct {
		name        string
		source      string
		expected    string
		expectError bool
	}{
		{
			name:     "absolute path",
			source:   path,
			expected: string(content),
		},
		{
			name:        "nonexistent file",
			source:      filepath.Join(dir, "nonexistent.go"),
			expectError: true,
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			data, err := Load(tc.source)
			if tc.expectError {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if diff := cmp.Diff(tc.expected, string(data)); diff != "" {
				t.Errorf("content mismatch: %s", diff)
			}
		})
	}
}

func TestIsDynamicSource(t *testing.T) {
	tt := []struct {
		entry    string
		expected bool
	}{
		{entry: "author", expected: false},
		{entry: "message", expected: false},
		{entry: "./vendor.go", expected: true},
		{entry: "/opt/validators/vendor.go", expected: true},
		{entry: "vendor.go", expected: true},
		{entry: "https://example.com/validator.go", expected: true},
		{entry: "http://example.com/validator.go", expected: true},
		{entry: "git://github.com/org/repo@main:path/validator.go", expected: true},
	}

	for _, tc := range tt {
		t.Run(tc.entry, func(t *testing.T) {
			got := IsDynamicSource(tc.entry)
			if got != tc.expected {
				t.Errorf("IsDynamicSource(%q) = %v, want %v", tc.entry, got, tc.expected)
			}
		})
	}
}

func TestLoadGitParseErrors(t *testing.T) {
	tt := []struct {
		name   string
		source string
	}{
		{name: "missing @ref", source: "git://github.com/org/repo"},
		{name: "missing :path", source: "git://github.com/org/repo@main"},
		{name: "empty ref", source: "git://github.com/org/repo@:path"},
		{name: "empty path", source: "git://github.com/org/repo@main:"},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			_, err := Load(tc.source)
			if err == nil {
				t.Fatal("expected error for invalid git source, got nil")
			}
		})
	}
}
