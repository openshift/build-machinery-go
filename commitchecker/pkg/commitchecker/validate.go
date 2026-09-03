package commitchecker

import (
	"fmt"
	"strings"
)

func init() {
	// Register the built-in validators.
	RegisterValidator(NewBuiltinValidator("author", ValidateCommitAuthor))
	RegisterValidator(NewBuiltinValidator("message", ValidateCommitMessage))
}

func ValidateCommitAuthor(commit *Commit) []error {
	var allErrors []error

	if strings.HasPrefix(commit.Email, "root@") {
		allErrors = append(allErrors, fmt.Errorf("Commit %s has invalid email %q", commit.Sha, commit.Email))
	}

	return allErrors
}

func ValidateCommitMessage(commit *Commit) []error {
	if commit.MatchesMergeSummaryPattern() {
		// Ignore merges
		return nil
	}

	var allErrors []error

	if !commit.MatchesUpstreamSummaryPattern() {
		allErrors = append(allErrors, fmt.Errorf(
			"invalid commit summary; expected format: UPSTREAM: <PR number|carry|drop>: description. More on this: https://github.com/openshift/build-machinery-go/blob/master/commitchecker/COMMITS.md"))
		return allErrors
	}

	return allErrors
}
