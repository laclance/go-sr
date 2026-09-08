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

Follow the local quality-gate commands and coverage policy in [`CONTRIBUTING.md`](CONTRIBUTING.md). CI is authoritative for exact analyzer versions and verifies both the minimum Go version declared by `go.mod` and the current stable Go release.

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
