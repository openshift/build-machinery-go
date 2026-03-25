# `--ancestry-path` Git Semantics

## What `git log A..B` does normally

`git log A..B` shows all commits reachable from `B` that are **not** reachable from
`A`. It walks backwards from `B` through all parent pointers (including merge parents)
until it hits commits that are ancestors of `A`.

In a simple linear history this is intuitive:

```
A --- C --- D --- B
```

`A..B` = {C, D, B}

## What `--ancestry-path` adds

`--ancestry-path` further restricts the output to commits that are on a **direct path
of descent** from `A` to `B`. A commit is included only if it is both:

1. A **descendant** of `A` — `A` is reachable by walking backwards through its parents
2. An **ancestor** of `B` — the commit is reachable by walking backwards from `B`

In other words: the commit must sit on a chain `A -> ... -> commit -> ... -> B`.

## Why the commitchecker needs it

OpenShift's downstream rebase process creates a complex DAG. During a rebase, the
downstream repo effectively resets to the upstream state and then re-applies carry
patches. This creates a history like:

```
upstream:    U1 --- U2 --- U3 --- U4
                                   \
downstream:  ... --- M (merge) --- carry1 --- carry2 --- HEAD
                     |
                merge-base
```

Without `--ancestry-path`, `git log merge-base..HEAD` walks all parent pointers from
`HEAD`. Through the merge commit `M`, it reaches back into upstream history and finds
`U1, U2, U3, U4` — commits that don't follow the `UPSTREAM:` convention because they
are actual upstream commits, not downstream carries. The checker would flag all of them
as invalid.

With `--ancestry-path`, only commits on the direct descendant path from `merge-base`
to `HEAD` are shown: `carry1, carry2`. The upstream commits reached via side paths
through merges are excluded. This is the correct set to validate.

## How `--ancestry-path` can silently drop commits

The problem arises in CI when a PR branch has fallen behind `main`.

Consider this scenario:

```
main:        A --- B --- C --- D        (D = current main tip = PULL_BASE_SHA)
              \
PR branch:    E --- F --- HEAD          (PR commits forked from A)
```

CI merges the PR onto main, producing:

```
              A --- B --- C --- D
              \                  \
               E --- F --- G (merge of PR onto main)
```

Now, `git log --ancestry-path D..G`:

For a commit to be on the ancestry path from `D` to `G`, it must be a **descendant of
`D`**. Let's check each commit:

- **`G` (the merge)** — descendant of `D`? Yes (D is a parent of G). Ancestor of `G`?
  Trivially yes. **Included** (but filtered by `--no-merges`).
- **`E`** — descendant of `D`? **No.** `E`'s parent is `A`, not `D`. Walking backwards
  from `E` gives `E -> A`, which never reaches `D`. **Excluded.**
- **`F`** — descendant of `D`? **No.** `F`'s parent is `E`, whose parent is `A`.
  **Excluded.**

Result: `--ancestry-path` returns **zero non-merge commits**. If used alone, the
checker would validate nothing and exit successfully — a false positive.

Without `--ancestry-path`, `git log D..G` correctly returns `{E, F}` because they are
reachable from `G` (via its second merge parent) but not from `D`.

### The key insight

`--ancestry-path` requires the commit to be a **descendant** of the start ref — meaning
the start must be reachable by walking **backwards** from that commit through its parent
chain. For `E`, walking backwards gives `E -> A`, never reaching `D`. The commit is
reachable from `G` but not descended from `D`.

## The trade-off

| Scenario | `--ancestry-path` | Without `--ancestry-path` |
|---|---|---|
| Rebase PRs (complex merges) | Correct: filters out upstream commits | Wrong: includes upstream commits |
| Stale PR branches | Wrong: drops PR commits | Correct: finds all PR commits |

No single `git log` invocation handles both cases correctly. The commitchecker
addresses this by running both: `--ancestry-path` for enforcement (to handle rebases
correctly) and without `--ancestry-path` to find all reachable commits. If commits
exist but none are on the direct ancestry path, the checker fails with an error
telling the user to rebase.