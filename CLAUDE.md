# CLAUDE.md

## Project

`wen` is a natural language date parser (Go library) and interactive terminal calendar (Bubble Tea TUI). The parser is zero-dependency; the calendar depends on charmbracelet and fsnotify.

## Build & Test

```bash
make check          # run tests + lint (preferred)
make test           # go test -race -count=1 ./...
make lint           # golangci-lint run
go build ./cmd/wen  # build binary
```

## Architecture

- `wen.go`, `lexer.go`, `parser.go`, `token.go`, `errors.go` — core library (zero deps beyond stdlib)
- `calendar/` — Bubble Tea TUI: model, view, render, config, styles, highlight, watcher
- `cmd/wen/` — CLI entry point, subcommand routing

The parser is a standalone library. The calendar depends on the parser but not vice versa. The CLI is a thin wrapper.

## Code Style

- Follow existing patterns. Table-driven tests with descriptive subtest names.
- Run `make check` before committing. All lint issues must be resolved.
- `.golangci.yml` configures errcheck, govet, staticcheck, gocritic, revive. Do not add `//nolint` directives without justification.
- Unexported functions don't need docstrings unless the logic is non-obvious.
- Exported functions and types always need godoc comments.
- Use `time.UTC` for map keys and date comparisons. Use `time.Local` for display.

## Testing

- Tests live next to the code they test (`foo_test.go` in same package).
- Use `t.Parallel()` for tests that don't share state.
- CLI integration tests in `cmd/wen/main_test.go` build the binary once via `TestMain` and run it via `exec.Command`.
- In-process tests use `run(&buf, args)` to test the `run()` function directly.
- `runCalendar` can't be tested in-process (requires TUI). Test calendar logic via `calendar/` package tests instead.

## Commit Conventions

- `feat:` new features, `fix:` bug fixes, `refactor:` restructuring, `docs:` documentation, `chore:` deps/CI
- Keep commits focused. One logical change per commit.

## Key Patterns

- **Functional options**: `ModelOption`, `wen.Option` — use this pattern for new optional configuration.
- **Fail silently for highlights**: `LoadHighlightedDates` returns nil on any error. The watcher continues on errors. Missing highlights are not fatal.
- **Value receivers on Model**: Bubble Tea requires value receivers for `Init()` and `Update()`. Do not store mutable pointers on Model unless they survive copies (use pointer indirection like `activeWatcher`).
- **Tilde expansion**: Always use `expandTilde()` before passing paths to fsnotify or other OS APIs that don't understand `~`.
