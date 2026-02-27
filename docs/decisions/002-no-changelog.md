# 002. No Changelog or Release Notes

**Status:** Accepted

## Context

Many versioning tools also generate changelogs, release notes, or GitHub releases. This adds complexity: templates, formatting options, filtering, grouping by type/scope, and output targets.

## Decision

Tagger only determines the bump and creates the git tag. It does not generate changelogs, release notes, or any other artifacts.

## Consequences

- **Minimal scope:** The tool does one thing well — semver tagging.
- **No template engine:** No need for Go templates, Markdown formatting, or output configuration.
- **Composable:** Tagger can be chained with dedicated changelog tools (`git-cliff`, `conventional-changelog`) if needed.
- **Fewer dependencies:** No file I/O beyond git operations, no additional output formats to maintain.
- **Trade-off:** Users who want an all-in-one solution need to combine tagger with other tools.
