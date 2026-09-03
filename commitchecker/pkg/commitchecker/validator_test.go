package commitchecker

import (
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestValidatorsForNames(t *testing.T) {
	tt := []struct {
		name           string
		names          []string
		expectedNames  []string
		expectedErrMsg string
	}{
		{
			name:          "nil returns all registered validators in order",
			names:         nil,
			expectedNames: []string{"author", "message"},
		},
		{
			name:          "empty returns all registered validators in order",
			names:         []string{},
			expectedNames: []string{"author", "message"},
		},
		{
			name:          "single validator",
			names:         []string{"author"},
			expectedNames: []string{"author"},
		},
		{
			name:          "single validator message only",
			names:         []string{"message"},
			expectedNames: []string{"message"},
		},
		{
			name:          "explicit order is preserved",
			names:         []string{"message", "author"},
			expectedNames: []string{"message", "author"},
		},
		{
			name:           "unknown validator returns error",
			names:          []string{"nonexistent"},
			expectedErrMsg: `unknown validator "nonexistent"; available validators: [author message]`,
		},
		{
			name:           "mix of known and unknown returns error",
			names:          []string{"author", "nonexistent"},
			expectedErrMsg: `unknown validator "nonexistent"; available validators: [author message]`,
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			validators, err := ValidatorsForNames(tc.names)

			if tc.expectedErrMsg != "" {
				if err == nil {
					t.Fatalf("expected error, got nil")
				}
				if diff := cmp.Diff(tc.expectedErrMsg, err.Error()); diff != "" {
					t.Errorf("error mismatch: %s", diff)
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			var gotNames []string
			for _, v := range validators {
				gotNames = append(gotNames, v.Name())
			}
			if diff := cmp.Diff(tc.expectedNames, gotNames); diff != "" {
				t.Errorf("validator names mismatch: %s", diff)
			}
		})
	}
}
