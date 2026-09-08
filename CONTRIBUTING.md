# Contributing

Thanks for helping maintain `github.com/laclance/go-sr`. This repo is a small deterministic Go library, so changes should stay focused and easy to review.

## Local Setup

```bash
git clone https://github.com/laclance/go-sr.git
cd go-sr
go test ./...
```

Static analysis uses Staticcheck and golangci-lint. CI pins the exact analyzer versions; use [`.github/workflows/ci.yml`](.github/workflows/ci.yml) when you need to reproduce those pins locally.

## Go Version Policy

`go.mod` declares the minimum supported Go version. CI runs the normal test suite on that minimum version and on the current stable Go release. The heavier race and static-analysis quality gate runs once on current stable Go so tool versions are not constrained by the compatibility floor.

Do not raise the `go.mod` minimum solely to accommodate development tooling.

## Tests and Quality Gate

Before opening a PR, run:

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

Run `gofmt -w .` first if the formatting check prints files. Coverage must stay at the documented `100.0%` statement coverage target.

## Issues

Open an issue with a minimal reproduction, Go version, OS, module version or commit, expected behavior, and actual behavior. Include sample candles or input when relevant, especially for determinism or lookahead-bias concerns.

## Pull Requests

- Use Conventional Commits such as `feat:`, `fix:`, `docs:`, `test:`, `refactor:`, `chore:`, and `ci:`.
- Keep PRs small and focused.
- Add regression tests for bug fixes.
- Update README examples or docs if public behavior or API usage changes.
- Do not add dependencies unless the PR explains why they are needed.

AI-assisted contributions are welcome. Generated PRs must pass CI and be reviewed by a maintainer before merge.

## Contribution License

Unless explicitly stated otherwise, contributions intentionally submitted for inclusion in `go-sr` are provided under the Apache License 2.0, consistent with the repository [`LICENSE`](LICENSE).
