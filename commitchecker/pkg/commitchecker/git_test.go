package commitchecker

import (
	"testing"
)

func TestUpstreamSummaryPattern(t *testing.T) {
	tt := []struct {
		summary string
		valid   bool
	}{
		// Rule 3.2: base formats
		{valid: true, summary: "UPSTREAM: 12345: a change"},
		{valid: true, summary: "UPSTREAM: <carry>: a change"},
		{valid: true, summary: "UPSTREAM: <drop>: a change"},

		// Rule 3.3: revert formats
		{valid: true, summary: "UPSTREAM: revert: 12345: a change"},
		{valid: true, summary: "UPSTREAM: revert: <carry>: a change"},
		{valid: true, summary: "UPSTREAM: revert: <drop>: a change"},

		// Rule 3.4: module prefix formats
		{valid: true, summary: "UPSTREAM: k8s.io/heapster: 12345: a change"},
		{valid: true, summary: "UPSTREAM: coreos/etcd: <carry>: a change"},
		{valid: true, summary: "UPSTREAM: coreos/etcd: <drop>: a change"},
		{valid: true, summary: "UPSTREAM: gopkg.in/ldap.v2: 51: exposed better API for paged search"},

		// Rule 3.4: module prefix with dots and hyphens
		{valid: true, summary: "UPSTREAM: k8s.io/kube-openapi: 12345: a change"},
		{valid: true, summary: "UPSTREAM: github.com/foo-bar: <carry>: a change"},

		// Rule 3.4: revert + module prefix formats
		{valid: true, summary: "UPSTREAM: revert: k8s.io/heapster: 12345: a change"},
		{valid: true, summary: "UPSTREAM: revert: coreos/etcd: <carry>: a change"},
		{valid: true, summary: "UPSTREAM: revert: coreos/etcd: <drop>: a change"},
		{valid: true, summary: "UPSTREAM: revert: gopkg.in/ldap.v2: 51: a change"},

		// Rule 3.5: invalid summaries
		{valid: false, summary: "UPSTREAM: whoopsie daisy"},
		{valid: false, summary: "upstream: 12345: lowercase prefix"},
		{valid: false, summary: "a]normal commit message"},
		{valid: false, summary: "UPSTREAM: : missing tag"},
		{valid: false, summary: ""},
	}
	for _, tc := range tt {
		t.Run(tc.summary, func(t *testing.T) {
			got := UpstreamSummaryPattern.Match([]byte(tc.summary))

			if tc.valid != got {
				t.Errorf("expected %#v, got %#v", tc.valid, got)
			}
		})
	}
}
