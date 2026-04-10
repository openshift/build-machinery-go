# build-machinery-go — AI Agent Guidelines

This file provides guidance to AI agents working in any OpenShift Go
repository that vendors
[build-machinery-go](https://github.com/openshift/build-machinery-go).

## What This Repo Provides

Reusable GNU Make fragments and helper scripts in three layers:

| Stack        | Entry file          | Use case                              |
|--------------|---------------------|---------------------------------------|
| **Golang**   | `make/golang.mk`    | Pure Go projects                      |
| **Default**  | `make/default.mk`   | OpenShift Go (+ images, bindata, codegen) |
| **Operator** | `make/operator.mk`  | OpenShift operators (extends Default) |

Component repos include the appropriate `*.mk` from the vendored path:

```makefile
include $(addprefix vendor/github.com/openshift/build-machinery-go/make/, \
    default.mk \
)
```

## Build & Test Commands

Run `make help` to see all available targets. Common targets across all stacks:

```bash
make build                    # Build Go binaries
make test-unit                # Run unit tests
make verify                   # All verification checks (gofmt, govet, codegen, bindata, deps)
make update                   # Regenerate all generated files
go mod tidy && go mod vendor  # Update vendored dependencies
```

**Default** stack adds: `make images`, `make verify-codegen`, `make verify-bindata`,
`make update-codegen`, `make update-bindata`, `make verify-deps`.

**Operator** stack adds: `make test-operator-integration`,
`make verify-profile-manifests`, `make update-profile-manifests`, `make telepresence`.

E2E tests (if defined by the component repo) require `KUBECONFIG`:

```bash
make test-e2e
```

## Anti-Patterns — Explicitly Forbidden

| Anti-Pattern | Why |
|---|---|
| Editing files under `vendor/` | Lost on next `go mod vendor`; fix upstream |
| Hardcoding namespace strings | Use constants from `operatorclient` or equivalent |
| `time.Sleep` in operator/controller code | Use `wait.PollImmediate` or controller requeue |
| `_ = foo()` (ignoring errors silently) | Return or log; silent failures are undebuggable |
| `context.TODO()` in production code | Pass proper context from caller |
| Mixing unrelated changes in one commit | Makes `git bisect` impossible |
| Adding dependencies without justification | See Dependency Policy below |
| Skipping `defer` cleanup in e2e tests | Leaves dirty cluster state |
| Hand-editing generated files (`zz_generated.*`, `bindata.go`) | Run `make update` instead |

## AI Agent Behavior Rules

- **Always explain WHY** — not just what changed, but the reasoning and problem context.
- **Prefer minimal diffs** — change only what's necessary; resist "cleanup while you're there."
- **No unrelated refactoring** — don't improve surrounding code unless explicitly requested.
- **Match existing style** — follow patterns in surrounding code.
- **When unsure, ASK** — ask clarifying questions instead of guessing.
- **Verify before submitting** — `make build && make test-unit && make verify`.

**Do NOT suggest code that:**
- Assumes perfect network conditions
- Ignores mid-operation operator restarts
- Requires manual intervention to recover
- Breaks during cluster upgrades
- Depends on resources that may not exist

## Dependency Policy

**Do NOT add new dependencies without strong justification.** Before adding one:
1. Can stdlib or `library-go` do it?
2. How many transitive deps? (`go mod graph | grep <dep> | wc -l`)
3. Is it actively maintained with a compatible license?

## PR / Commit Conventions

### Git Commit Structure
- Exactly **2 commits** (code + deps/generated) OR **1 commit** (docs-only)
- Commits based on `upstream/main`, not fork
- Generated artifacts in commit 2 only
- No merge commits or unrelated fork changes
- Verify: `git log --oneline upstream/main..HEAD`

**Code commits** — functional changes only. Exclude `go.mod`, `go.sum`,
`vendor/`, and generated files (`zz_generated.*`, `bindata.go`).

**Generated/vendor commit (if needed)** — single final commit:
- Vendor only: `vendor: bump(*)`
- Generated only: `update generated`
- Both: `vendor: bump(*), update generated`

### Commit Message Format
```
<component>: <short description>

<Why this change is needed. Jira/Bugzilla link if applicable.>
<Risk level (low/medium/high based on disruption potential).>
```

### PR Description Must Include
- **What** and **why** (Jira/Bugzilla link)
- **Testing strategy** — what tests run, coverage added
- **Impact analysis** — rollout triggers, disruption, upgrade considerations
- **Risk level** — Low (test/docs), Medium (config changes), High (static pod, certs, RBAC)
- **Dependency justification** — if adding deps

### PR Generation — Avoid
- Large refactors (small focused changes are safer)
- Mixed concerns (one problem per PR)
- Speculative improvements (don't add unrequested features)
- Formatting-only changes mixed with logic changes

### Conflict Resolution During Rebase
1. `go.mod`/`go.sum`: accept upstream, re-apply with `go get` + `go mod tidy`
2. `vendor/`: accept upstream, regenerate with `go mod vendor`
3. Generated files: accept upstream, re-run `make update`
4. Never manually resolve conflicts in vendor/ or generated files

## PR Review Checklist

### Build and Tests
- [ ] `make build` succeeds
- [ ] `make test-unit` passes
- [ ] `make verify` passes
- [ ] New/modified code has unit tests

### Git Commit Structure
- [ ] Exactly 2 commits (code + deps/generated) OR 1 commit (docs-only)
- [ ] Commits based on `upstream/main`, not fork
- [ ] Generated artifacts in commit 2 only
- [ ] No merge commits or unrelated fork changes

### Dependency Changes
- [ ] New dependency strongly justified
- [ ] Transitive impact measured: `go mod graph | grep <package>`
- [ ] `go mod tidy && go mod vendor` produces no diff
- [ ] No edits under `vendor/`
- [ ] Deps committed separately in commit 2

### Go Conventions
- [ ] No duplicate constants — extract to shared location
- [ ] No repeated boilerplate — use/create helpers
- [ ] Errors returned up chain (operator code) or `require.NoError` (tests)
- [ ] No hardcoded namespaces — use constants

## Testing Conventions

### Unit Tests
- **Table-driven tests** with descriptive `name` fields
- Cover: happy path, error path, nil/empty input, idempotency
- Use `require.NoError` for fatal checks, `assert.*` for non-fatal

### E2E Tests
- Require `KUBECONFIG` to a running OpenShift cluster
- Always `defer` cleanup with stability waits inside the defer
- Use `wait.PollImmediate(interval, timeout, fn)` — never `time.Sleep`
- Account for rollout times (≥20 min per control-plane rollout)
- Don't assume revision numbers, pod counts, or timing
- New tests go in OTE format

### Feature Gate Testing
- Use `featuregates.NewHardcodedFeatureGateAccess` in unit tests
- Test both enabled and disabled paths
- Feature gates defined in `openshift/api`

## Code Style

- Uses `openshift/library-go` controller framework, not raw controller-runtime
- Error handling: return errors up chain in operator code; `require.NoError` in tests
- Namespace references via constants, never string literals
- Generated files — never hand-edit; use `make update`
- Controllers: return error from `Sync()` → framework retries with backoff; no custom loops
- Logging: `klog.Infof` / `klog.Errorf` with context (namespace, resource, conditions)
