# cli-template

A Go CLI project template for use with [`gonew`](https://pkg.go.dev/golang.org/x/tools/cmd/gonew).

## Usage

```bash
gonew github.com/rasalas/cli-template github.com/rasalas/my-tool
cd my-tool
go run .
```

## What's included

- Cobra CLI with `--version` flag and exit code handling
- `internal/term/` color system with NO_COLOR support
- `.goreleaser.yml` for cross-platform releases
- `CLAUDE.md` template for AI-assisted development
- `docs/decisions/` directory for ADRs

## After scaffolding

1. Update `CLAUDE.md` with your project's details
2. Update `cmd/root.go` — change `Use`, `Short`, `Long`, and `Version`
3. Pick a primary color in `internal/term/colors.go`
4. Update binary name in `.goreleaser.yml`
5. Update `.gitignore` (replace `cli-template` with your binary name)
