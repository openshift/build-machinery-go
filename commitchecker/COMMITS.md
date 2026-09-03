# UPSTREAM Fork Commit Format

## 1. Overview

This specification defines the commit message format for repositories that are downstream forks of upstream projects. All downstream-added commits (those not already present in the upstream) MUST conform to this format to enable correct classification and handling during rebases. Commits that originate from the upstream repository are not subject to this specification.

## 2. Terminology

The key words "MUST", "MUST NOT", "SHOULD", "SHOULD NOT", and "MAY" in this document are to be interpreted as described in [RFC 2119](https://www.rfc-editor.org/rfc/rfc2119).

### 2.1 Definitions

- **Summary**: the first line of a commit message, also known as the commit title.
- **Downstream fork**: a repository that is a fork of an upstream project and carries additional patches on top of it.
- **Upstream**: the original project being forked.
- **Rebase**: the process of replaying downstream commits onto a newer version of the upstream.
- **Vendor file**: any file whose path begins with `vendor/` or contains `/vendor/` as a path segment. In Go projects, the `vendor/` directory holds a local copy of all module dependencies, managed by `go mod vendor`. Some repositories contain nested `vendor/` directories (e.g. `staging/src/k8s.io/api/vendor/`) when they host multiple modules.
- **Merge commit**: a commit whose summary matches `Merge commit *`.

### 2.2 Notation

In this document, `{curly braces}` denote placeholders to be substituted with actual values. `<angle brackets>` denote literal strings that MUST appear as-is in the commit message (e.g. `<carry>`, `<drop>`).

## 3. Commit summary format

### 3.1 General rule

Every non-merge commit MUST have a summary matching the following structure:

```
UPSTREAM: [revert: ][{org/repo}: ]{tag}: {description}
```

The `UPSTREAM:` prefix MUST be uppercase. The match is case-sensitive.

Where `{tag}` is one of:

- A numeric PR number (e.g. `12345`)
- The literal string `<carry>`
- The literal string `<drop>`

The commit body (everything after the summary) is unconstrained by this specification.

### 3.2 Base formats

```
UPSTREAM: {PR number}: {description}
UPSTREAM: <carry>: {description}
UPSTREAM: <drop>: {description}
```

### 3.3 Revert formats

```
UPSTREAM: revert: {PR number}: {description}
UPSTREAM: revert: <carry>: {description}
UPSTREAM: revert: <drop>: {description}
```

### 3.4 Module prefix formats

```
UPSTREAM: {org/repo}: {PR number}: {description}
UPSTREAM: {org/repo}: <carry>: {description}
UPSTREAM: {org/repo}: <drop>: {description}
UPSTREAM: revert: {org/repo}: {PR number}: {description}
UPSTREAM: revert: {org/repo}: <carry>: {description}
UPSTREAM: revert: {org/repo}: <drop>: {description}
```

### 3.5 Validation pattern

Commit summaries MUST match the following regular expression:

```
^UPSTREAM: (revert: )?(([\w.-]+/[\w-.-]+)?: )?(\d+:|<carry>:|<drop>:)
```

### 3.6 Merge commits

Merge commits (summaries matching `Merge commit *`) are exempt from all rules in this specification.

## 4. Commit types

### 4.1 PR reference (`UPSTREAM: {PR number}: ...`)

A cherry-pick of a specific upstream pull request.

1. The PR number MUST correspond to a pull request in the upstream repository.
2. During a rebase, a commit referencing a PR number MUST be compared against the new upstream base. If the pull request is included in the upstream base, meaning it has been merged upstream and its merge commit can be found in the branch the downstream is being rebased onto, then the cherry-pick is redundant and MUST be dropped. If the pull request is not yet in the upstream base, the commit MUST be carried forward onto the new base.

### 4.2 Carry (`UPSTREAM: <carry>: ...`)

A downstream-only change that is carried forward across rebases.

1. Carry commits MUST be preserved during rebases.
2. Carry commits SHOULD NOT modify vendor files (see section 5).

### 4.3 Drop (`UPSTREAM: <drop>: ...`)

A commit that is dropped and not carried forward during a rebase.

1. Drop commits SHOULD be used for vendor directory updates, generated manifests, or other artifacts that are regenerated after each rebase.
2. Commits that exclusively modify vendor files SHOULD use this tag.

### 4.4 Revert (`UPSTREAM: revert: ...`)

A revert of a previous UPSTREAM commit.

1. The `revert:` keyword MUST be followed by a valid UPSTREAM format (PR number, `<carry>`, or `<drop>`, with an optional module prefix).
2. During a rebase, a revert is handled the same way as the commit type it reverts: a revert of a `<carry>` is itself carried forward; a revert of a `<drop>` is itself dropped.

Examples:

```
UPSTREAM: revert: 12345: undo the kube fix
UPSTREAM: revert: <carry>: undo a carried change
UPSTREAM: revert: <drop>: undo a dropped change
UPSTREAM: revert: k8s.io/heapster: 12345: undo a heapster fix
```

## 5. Vendor directory rules

1. Commits tagged as `<carry>` or referencing a PR number SHOULD NOT modify vendor files.
2. Vendor changes SHOULD be isolated in their own `UPSTREAM: <drop>:` commit so they can be cleanly dropped and regenerated.

## 6. Upstream module prefix

When a downstream fork tracks multiple upstream modules, an optional `{org/repo}:` prefix MAY appear before the tag to identify which upstream repository the commit refers to.

1. The prefix MUST match the pattern `{org}/{repo}` where both segments may contain word characters, dots, and hyphens.
2. When no prefix is present, the commit is assumed to refer to the primary upstream repository.

Examples:

```
UPSTREAM: k8s.io/heapster: 12345: fix memory leak in heapster
UPSTREAM: coreos/etcd: <carry>: add TLS configuration
UPSTREAM: coreos/etcd: <drop>: regenerate protobuf
UPSTREAM: gopkg.in/ldap.v2: 51: expose better API for paged search
```

## 7. Author requirements

1. Commits with an author email starting with `root@` MUST be rejected.
