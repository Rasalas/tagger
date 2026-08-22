# 001. Conventional Commits Only

**Status:** Accepted

## Context

Tagger needs to determine the semver bump level from commit history. There are several approaches:

- **Keyword heuristics** — scan messages for words like "add", "remove", "breaking" and guess intent.
- **Configuration-based** — let users define custom patterns or mappings.
- **Conventional Commits** — rely on the structured `type(scope): summary` format from the spec.

## Decision

Tagger exclusively uses the Conventional Commits specification to classify commits. Only recognized types (`feat`, `fix`, `refactor`, `perf`, `revert`) produce a bump. Non-conforming messages are silently ignored (`None`). Commit types are matched case-insensitively and normalized to lowercase (`Feat:` counts as `feat`), mirroring common tooling behavior.

Breaking changes are detected via:
- `!` suffix before the colon (e.g., `feat!: remove API`)
- `BREAKING CHANGE:` / `BREAKING-CHANGE:` trailer in the body

GitLab merge commit subjects are accepted when the source branch starts with a
Conventional Commit prefix followed by `/`. For example,
`Merge branch 'feat/split-gitlab-ci' into 'main'` is interpreted as
`feat: split-gitlab-ci`.

## Consequences

- **Deterministic:** The same commit always produces the same bump level — no ambiguity.
- **Zero configuration:** No config files, no custom patterns, no mapping tables.
- **Explicit opt-in:** Teams must adopt Conventional Commits. Non-conforming commits are simply not counted.
- **Composable:** Other tools in the ecosystem (changelog generators, linters) understand the same format.
