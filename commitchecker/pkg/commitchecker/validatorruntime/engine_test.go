package validatorruntime

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/openshift/build-machinery-go/commitchecker/pkg/commitchecker"
)

func errMessages(errs []error) []string {
	if errs == nil {
		return nil
	}
	msgs := make([]string, len(errs))
	for i, e := range errs {
		msgs[i] = e.Error()
	}
	return msgs
}

func TestLoadValidator(t *testing.T) {
	engine := NewEngine()

	src := []byte(`package testvalidator

import (
	"fmt"

	"github.com/openshift/build-machinery-go/commitchecker/pkg/commitchecker"
)

func Validate(commit *commitchecker.Commit) []error {
	if commit.Email == "bad@example.com" {
		return []error{fmt.Errorf("commit %s has disallowed email", commit.Sha)}
	}
	return nil
}
`)

	v, err := engine.LoadValidator("test", src)
	if err != nil {
		t.Fatalf("failed to load validator: %v", err)
	}

	if v.Name() != "test" {
		t.Errorf("expected name %q, got %q", "test", v.Name())
	}

	tt := []struct {
		name         string
		commit       *commitchecker.Commit
		expectedMsgs []string
	}{
		{
			name: "passes for allowed email",
			commit: &commitchecker.Commit{
				Sha:   "abc1234",
				Email: "dev@redhat.com",
			},
			expectedMsgs: nil,
		},
		{
			name: "fails for disallowed email",
			commit: &commitchecker.Commit{
				Sha:   "def5678",
				Email: "bad@example.com",
			},
			expectedMsgs: []string{"commit def5678 has disallowed email"},
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			errs := v.Validate(tc.commit)
			if diff := cmp.Diff(tc.expectedMsgs, errMessages(errs)); diff != "" {
				t.Errorf("error mismatch: %s", diff)
			}
		})
	}
}

func TestLoadValidatorMissingValidateFunc(t *testing.T) {
	engine := NewEngine()

	src := []byte(`package testvalidator

func NotValidate() []error {
	return nil
}
`)

	_, err := engine.LoadValidator("bad", src)
	if err == nil {
		t.Fatal("expected error for validator without Validate function, got nil")
	}
}

func TestLoadValidatorWrongSignature(t *testing.T) {
	engine := NewEngine()

	src := []byte(`package testvalidator

func Validate() string {
	return ""
}
`)

	_, err := engine.LoadValidator("bad", src)
	if err == nil {
		t.Fatal("expected error for validator with wrong Validate signature, got nil")
	}
}
