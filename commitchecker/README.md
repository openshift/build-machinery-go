# commitchecker

commitchecker validates a range of commits in a git repository and ensures they meet specific requirements:

1. The author's email address does not start with "root@".
2. The message starts with one of:
   1. UPSTREAM: <PR number|carry|drop>: description
   2. UPSTREAM: revert: <normal upstream format>

This is useful for repositories that are downstream forks of upstream repositories.

# History

This comes from
from https://github.com/openshift/kubernetes/tree/0abcd84431df81ed9f2a1846b5045e46d9032cc1/openshift-hack/commitchecker.

That repository can't be vendored, and you can't use `go get` against it (for multiple reasons), so we extracted the
commitchecker code and moved it here.
