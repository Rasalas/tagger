# tagger

Semantic Git tagging based on Conventional Commits.

Tagger reads commits since the last tag, determines the next semver bump (major/minor/patch), and creates the tag. No changelog, no release notes — just the tag.

## Usage

```bash
# Preview what would happen
tagger bump --dry-run

# Suggest the next tag without creating it
tagger status
tagger suggest

# Create the next tag (with confirmation prompt)
tagger bump

# Create and push, skip confirmation
tagger bump --yes

# Create without pushing
tagger bump --no-push

# Force a specific bump level
tagger bump --major
tagger bump --minor
tagger bump --patch
```

## Flags

| Flag | Short | Default | Description |
|------|-------|---------|-------------|
| `--no-push` | — | `false` | Skip pushing the tag to origin |
| `--yes` | `-y` | `false` | Skip confirmation prompt |
| `--dry-run` | — | `false` | Show what would happen without making changes |
| `--major` | — | `false` | Force a major bump |
| `--minor` | — | `false` | Force a minor bump |
| `--patch` | — | `false` | Force a patch bump |
| `--prefix` | — | `v` | Tag prefix |

`tagger status` has the same analysis flags (`--major`, `--minor`, `--patch`,
`--prefix`) and never creates or pushes a tag. `tagger suggest` is an alias.

The release baseline is the highest valid version tag matching `--prefix` that
is reachable from `HEAD`. Tags on branches that have not been merged into the
current history are ignored.

## Build suffixes

Tagger also supports numeric build suffixes such as `v3.4.0-58`. When the
latest tag uses that format, the semver part is bumped from the commits and the
build number is incremented for the next tag:

| Latest | Commit | Next |
|--------|--------|------|
| `v3.4.0-58` | `fix: ...` | `v3.4.1-59` |
| `v3.4.0-58` | `feat: ...` | `v3.5.0-59` |
| `v3.4.0-58` | `feat!: ...` | `v4.0.0-59` |

## Conventional Commits

Tagger uses [Conventional Commits](https://www.conventionalcommits.org/) to determine the bump level:

| Commit | Bump |
|--------|------|
| `feat: ...` | minor |
| `fix: ...` | patch |
| `refactor: ...` | patch |
| `perf: ...` | patch |
| `revert: ...` | patch |
| `feat!: ...` or `BREAKING CHANGE:` in body | major |
| anything else | none (ignored) |

GitLab merge commits with a Conventional Commit source branch are also recognized.
For example, `Merge branch 'feat/split-gitlab-ci' into 'main'` is treated like
`feat: split-gitlab-ci`.

## Install

```bash
go install github.com/rasalas/tagger@latest
```
