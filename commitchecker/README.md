# commitchecker

commitchecker validates a range of commits in a git repository and ensures they meet specific requirements.
It ships with built-in validators and supports loading additional validators dynamically at runtime.

Commits to validate are enumerated using `git log --no-merges --ancestry-path`, so only non-merge commits are validated. This means that if upstream commits are brought in via a single merge (as it is expected), they are automatically excluded from the evaluated range.

## Built-in validators

- **author**: rejects commits where the author email starts with `root@`.
- **message**: requires commit messages to follow the [UPSTREAM Fork Commit Format](COMMITS.md):
  - `UPSTREAM: <PR number|carry|drop>: description`
  - `UPSTREAM: revert: <normal upstream format>`

This is useful for repositories that are downstream forks of upstream projects.

## Usage

### Command line

```
commitchecker [flags]

Flags:
  -start string
        The start of the revision range for analysis (default "main")
  -end string
        The end of the revision range for analysis (default "HEAD")
  -config string
        The configuration file to use, optional (default "./commitchecker.yaml")
  -fetch-mode string
        Method to use for fetching from git remotes: "https" or "ssh" (default "https")
  -validators string
        Comma-separated list of validators to run (default: all built-in validators)
```

### Selecting validators

By default, all built-in validators run. Use `--validators` to filter built-in validators or to add dynamic ones:

```bash
# Run only the author built-in validator
commitchecker --validators=author

# Run built-in validators plus a custom validator from a local file
commitchecker --validators=author,message,./validators/vendor.go

# Load a dynamic validator from a URL
commitchecker --validators=author,message,https://example.com/validators/vendor.go
```

Validators run in the order they are listed.

### Configuration file

Validators can also be specified in `commitchecker.yaml`:

```yaml
upstreamOrg: kubernetes
upstreamRepo: kubernetes
upstreamBranch: master
validators:
  - author
  - message
  - ./validators/vendor.go
```

The `--validators` flag takes precedence over the config file. When neither is set, all built-in validators run.

## Dynamic validators

Dynamic validators are Go files that are loaded and interpreted at runtime using [Yaegi](https://github.com/traefik/yaegi).
They must export a single `Validate` function:

```go
package nofixup

import (
    "errors"
    "strings"

    "github.com/openshift/build-machinery-go/commitchecker/pkg/commitchecker"
)

func Validate(commit *commitchecker.Commit) []error {
    if strings.HasPrefix(commit.Summary, "fixup! ") {
        return []error{errors.New("fixup commit must be squashed before merging")}
    }
    return nil
}
```

Validators can inspect the list of changed files via `commit.Files()`:

```go
package novendorincarry

import (
    "errors"
    "regexp"
    "strings"

    "github.com/openshift/build-machinery-go/commitchecker/pkg/commitchecker"
)

var carryOrPRPattern = regexp.MustCompile(`^UPSTREAM: (revert: )?(([\w.-]+/[\w-.-]+)?: )?(\d+:|<carry>:)`)

func Validate(commit *commitchecker.Commit) []error {
    if !carryOrPRPattern.MatchString(commit.Summary) {
        return nil
    }

    files, err := commit.Files()
    if err != nil {
        return []error{err}
    }

    for _, f := range files {
        if strings.HasPrefix(f, "vendor/") || strings.Contains(f, "/vendor/") {
            return []error{errors.New("carry/PR commit must not touch vendor files; use a separate UPSTREAM: <drop>: commit")}
        }
    }

    return nil
}
```

The function receives a `*Commit` with the following API:

- `Sha` - the commit hash
- `Summary` - the first line of the commit message
- `Email` - the author email
- `Files()` - returns the list of changed file paths (lazily loaded, cached)
- `FileContent(path)` - returns file content at this commit's revision (lazily loaded, cached per path)

### Supported sources

Dynamic validators can be loaded from:

| Source | Example |
|--------|---------|
| Relative path | `./validators/vendor.go` |
| Absolute path | `/opt/validators/vendor.go` |
| HTTPS URL | `https://raw.githubusercontent.com/org/repo/main/validator.go` |
| Git reference | `git://github.com/org/repo@main:path/to/validator.go` |

The `git://` format clones the repository at the specified ref (branch, tag, or SHA) and reads the file from the clone.

### Sandbox

Dynamic validators run in a restricted environment.
They have access to a limited set of standard library packages
(`fmt`, `strings`, `regexp`, `path`, `strconv`, `sort`, `math`, `unicode`, `bytes`, `errors`) and to the `Commit` API.
Packages like `os`, `os/exec`, `net`, and `syscall` are not available.
File access is provided exclusively through `commit.Files()` and `commit.FileContent(path)`,
which are scoped to the commit being validated.

## In OpenShift CI

In your repository configuration (`github.com/openshift/release/ci-operator/config/$org/$repo/*.yaml`):

1. Import the commitchecker image from the `ci` namespace:

```yaml
base_images:
  commitchecker:
    name: commitchecker
    namespace: ci
    tag: "latest"
```

2. Add a `verify-commits` presubmit CI job:

```yaml
tests:
- as: verify-commits
  commands: |
    commitchecker --start ${PULL_BASE_SHA:-main}
  container:
    from: commitchecker
```

To include a dynamic validator in CI, you can fetch it from a URL or include it in your repository:

```yaml
tests:
- as: verify-commits
  commands: |
    commitchecker --start ${PULL_BASE_SHA:-main} --validators=author,message,./validators/vendor.go
  container:
    from: commitchecker
```

## History

This was originally extracted from https://github.com/openshift/kubernetes/tree/0abcd84431df81ed9f2a1846b5045e46d9032cc1/openshift-hack/commitchecker. That repository cannot be vendored and `go get` does not work against it, so the commitchecker code was moved here.
