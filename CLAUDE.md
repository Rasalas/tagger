# cli-template

Short description of what your tool does.

## Tech Stack

- Go 1.25, Cobra CLI
- No other external dependencies

## Project Structure

- `cmd/` — Cobra commands (root)
- `internal/term/` — Color palette and terminal output helpers

## Conventions

- Color palette: Teal (#2DD4BF) as primary
- NO_COLOR support via `init()` in `internal/term/colors.go`
- Output via `term.W` (io.Writer), not stdout directly — enables testability
- Exit codes: 0=pass, 1=errors, 2=config error
- Tests next to code (`_test.go` in same package)
- Documentation always in English

## Development

```bash
go test ./...    # Tests
go run .         # Run
```
