package commitchecker

import (
	"testing"

	"github.com/google/go-cmp/cmp"
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

func TestValidateCommitAuthor(t *testing.T) {
	tt := []struct {
		name         string
		commit       *Commit
		expectedMsgs []string
	}{
		// Rule 7: author requirements
		{
			name: "fails on root@locahost",
			commit: &Commit{
				Sha:     "aaa0000",
				Summary: "a summary",
				Email:   "root@localhost",
			},
			expectedMsgs: []string{
				"Commit aaa0000 has invalid email \"root@localhost\"",
			},
		},
		// Rule 7: valid author
		{
			name: "succeeds for deads2k@redhat.com",
			commit: &Commit{
				Sha:     "aaa0000",
				Summary: "a summary",
				Email:   "deads2k@redhat.com",
			},
			expectedMsgs: nil,
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			gotErrs := ValidateCommitAuthor(tc.commit)
			if diff := cmp.Diff(tc.expectedMsgs, errMessages(gotErrs)); diff != "" {
				t.Errorf("error mismatch: %s", diff)
			}
		})
	}
}

func TestValidateCommitMessage(t *testing.T) {
	tt := []struct {
		name         string
		commit       *Commit
		expectedMsgs []string
	}{
		// Rule 3.1: invalid summary
		{
			name: "modifying k8s without UPSTREAM commit fails",
			commit: &Commit{
				Sha:     "aaa0000",
				Summary: "wrong summary",
			},
			expectedMsgs: []string{
				"invalid commit summary; expected format: UPSTREAM: <PR number|carry|drop>: description. More on this: https://github.com/openshift/build-machinery-go/blob/master/commitchecker/COMMITS.md",
			},
		},
		// Rule 3.6: merge commits are exempt
		{
			name: "merge commit is exempt from message validation",
			commit: &Commit{
				Sha:     "aaa0000",
				Summary: "Merge commit abc1234",
			},
			expectedMsgs: nil,
		},
		// Rule 3.2: valid base format
		{
			name: "modifying k8s with UPSTREAM commit succeeds",
			commit: &Commit{
				Sha:     "aaa0000",
				Summary: "UPSTREAM: 42: Fix kube",
			},
			expectedMsgs: nil,
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			gotErrs := ValidateCommitMessage(tc.commit)
			if diff := cmp.Diff(tc.expectedMsgs, errMessages(gotErrs)); diff != "" {
				t.Errorf("error mismatch: %s", diff)
			}
		})
	}
}
