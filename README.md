# tagger

Semantic Git tagging based on Conventional Commits.

Tagger reads commits since the last tag, determines the next semver bump (major/minor/patch), and creates the tag. No changelog, no release notes — just the tag.

## Usage

```bash
# Preview what would happen
tagger bump --dry-run

# Create the next tag (with confirmation prompt)
tagger bump

# Create and push, skip confirmation
tagger bump --push --yes

# Force a specific bump level
tagger bump --major
tagger bump --minor
tagger bump --patch
```

## Flags

| Flag | Short | Default | Description |
|------|-------|---------|-------------|
| `--push` | `-p` | `false` | Push tag after creation |
| `--yes` | `-y` | `false` | Skip confirmation prompt |
| `--dry-run` | — | `false` | Show what would happen without making changes |
| `--major` | — | `false` | Force a major bump |
| `--minor` | — | `false` | Force a minor bump |
| `--patch` | — | `false` | Force a patch bump |
| `--prefix` | — | `v` | Tag prefix |

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
