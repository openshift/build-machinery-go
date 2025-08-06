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

# Usage

## On command line

```
commitchecker:
  -end string
	The end of the revision range for analysis (default "HEAD")
  -start string
	The start of the revision range for analysis (default "main")
```

## In OpenShift CI

In your repository configuration (`github.com/openshift/release/ci-operator/config/$org/$repo/*.yaml`):

1. Import commitchecker image from `ci` namespace:

```yaml
base_images:
...
  commitchecker:
    name: commitchecker
    namespace: ci
    tag: "latest"
```

2. Add `verify-commits` presubmit CI job that clones your repository, applies the PR-under-test and runs `commitchecker` image with the git clone inside:

```yaml
tests:
...
- as: verify-commits
  commands: |
    commitchecker --start ${PULL_BASE_SHA:-main} # Or -master, whatever is the main branch of your repo
  container:
    from: commitchecker
```

There is no code or `<carry>` patch needed in your repository!
