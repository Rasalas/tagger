# 003. NUL Byte Delimiter for Commit Parsing

**Status:** Accepted

## Context

Tagger calls `git log --format=%B` to retrieve commit messages. Multi-line messages (with body, footers, breaking change trailers) contain newlines, so newline cannot serve as a delimiter between commits.

Common alternatives:
- **Line-based separators** (`---`, `===`) — could appear in commit messages.
- **Multiple git calls** — one per commit, slow on large histories.
- **NUL byte** (`\x00`) — cannot appear in commit messages, natively supported by git.

## Decision

Use `--format=%B%x00` and split on `\x00`. Each commit's full message (subject + body) is cleanly separated regardless of its internal structure.

```go
out, err := run("git", "log", rangeSpec, "--format=%B%x00")
// ...
for _, part := range strings.Split(out, "\x00") {
    msg := strings.TrimSpace(part)
    if msg != "" {
        messages = append(messages, msg)
    }
}
```

## Consequences

- **Robust:** No false splits on multi-line messages, special characters, or Markdown content in commits.
- **Simple:** Single `git log` call, single `strings.Split` — no complex state machine.
- **Standard:** `%x00` is documented git format syntax. The same pattern is used by `git log -z` and `xargs -0`.
- **Trade-off:** NUL bytes require care in languages with C-style strings, but Go handles them natively.
