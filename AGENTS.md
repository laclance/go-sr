# AI Contributor Guide

## Project Summary

`github.com/laclance/go-sr` is a deterministic Go support/resistance detection library. It focuses on closed-candle SR detection, zone and legacy modes, and SR-specific multi-timeframe helpers.

## Rules

- Preserve deterministic behavior for the same candle prefix and options.
- Never introduce lookahead bias.
- Keep the public API small and stable.
- Do not add dependencies unless the PR clearly justifies them.
- Prefer idiomatic, simple Go over clever abstractions.
- Add regression tests for every bug fix.
- Update README examples when public behavior changes.

## Required Checks Before PR

Run the full quality gate before opening a PR:

```bash
test -z "$(gofmt -l .)"
go test ./...
go test -race ./...
go vet ./...
staticcheck ./...
golangci-lint run
go test -coverprofile=/tmp/go-sr-coverage.out ./...
go tool cover -func=/tmp/go-sr-coverage.out
go test -run=^$ -fuzz=FuzzAggregateCandlesToTimeframe -fuzztime=5s
go test -run=^$ -fuzz=FuzzComputeInvariants -fuzztime=5s
```

Run `gofmt -w .` first if the formatting check prints files. Coverage must remain at the documented `100.0%` statement coverage target. Run any fuzz smoke tests present in the repo.

## Commit Style

Use Conventional Commits:

- `feat:`
- `fix:`
- `docs:`
- `test:`
- `refactor:`
- `chore:`
- `ci:`

## PR Expectations

- Keep PRs small and focused.
- Explain what changed and why.
- List the tests and checks run.
- Do not merge without maintainer approval.
