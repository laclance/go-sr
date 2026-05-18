# `github.com/laclance/go-sr`

[![CI](https://github.com/laclance/go-sr/actions/workflows/ci.yml/badge.svg)](https://github.com/laclance/go-sr/actions/workflows/ci.yml)

`go-sr` is a standalone Go module for deterministic support/resistance detection and SR-specific multi-timeframe helpers.

## Scope

This module owns:

- Closed-candle SR detection
- Legacy line-based and zone-based SR modes
- Deterministic nearest support/resistance metadata
- SR-specific multi-timeframe helpers for aggregation, warmup sizing, and fetch sizing

This module intentionally does not own:

- Exchange/Binance parsing
- Strategy scoring and trade evaluation
- App-specific timeframe policy like "use `15m` and `1h` as the higher-timeframe bundles"

## Public API

```go
import sr "github.com/laclance/go-sr"

type Mode string

const (
    ModeLegacy Mode = "legacy"
    ModeZones  Mode = "zone"
)

type Options struct {
    Timeframe   string
    Lookback    int
    Mode        Mode
    Tolerance   float64
    MinStrength int
}

func Compute(candles []Candle, opts Options) (Levels, error)
func EmptyLevels(timeframe string) Levels
func AggregateCandlesToTimeframe(candles []Candle, fromInterval, toInterval string) []Candle
func WarmupCandles(lookback int, mode Mode) int
func RequiredKlineLimit(baseInterval, targetInterval string, lookback int, mode Mode) int
```

`Tolerance` applies only to `ModeLegacy`. When `Tolerance <= 0`, the fallback remains `0.002`.

`MinStrength` applies only to `ModeZones`; values <= 0 use the default of 2.

## Behavioral Contract

- `Compute` is deterministic for the same candle prefix and options.
- `Compute` returns an empty level bundle and an error for an unknown `Mode`.
- Zone-mode pivots are confirmation-based; no future candles are read beyond the current prefix.
- `AggregateCandlesToTimeframe` uses UTC-aligned buckets and drops leading/trailing partial buckets.
- `RequiredKlineLimit` returns the number of raw candles needed to build a higher-timeframe SR bundle and includes one extra live candle for exchange REST responses.
- Supported interval strings use `<n><unit>` with `m`, `h`, or `d`, and the target interval must be larger than and evenly divisible by the base interval.
- `NearSupport` / `NearResistance` flag whether the **nearest** level on each side is within a "near" threshold:
  - `ModeZones`: within `2 ×` the zone's half-width (i.e., distance to zone center ≤ zone width). If the zone has zero width, the threshold falls back to `0.1%` of the current price.
  - `ModeLegacy`: within `Tolerance × close` (absolute price distance). A closer-but-out-of-tolerance level overrides a farther within-tolerance one — the flag describes the nearest level, not any level.

## Quality Gate

CI runs on every push and pull request. The gate requires:

- `gofmt`
- `go test ./...`
- `go test -race ./...`
- `go vet ./...`
- `staticcheck ./...`
- `golangci-lint run`
- `100.0%` statement coverage
- 5-second fuzz smoke tests for aggregation and compute invariants

## Manual Chart Inspection

Generate a local HTML chart from the BTC fixture when you want to visually inspect whether SR zones line up with market structure:

```bash
GO_SR_CHART=/tmp/go-sr-btc-5m.html go test -run TestGenerateManualSRChart -count=1 -v
```

Optional overrides:

```bash
GO_SR_CHART_TIMEFRAME=15m
GO_SR_CHART_MODE=legacy
GO_SR_CHART_LOOKBACK=80
GO_SR_CHART_WINDOW=300
GO_SR_CHART_MIN_STRENGTH=1
```

## Install

```bash
go get github.com/laclance/go-sr
```

## Quick Start

```go
import sr "github.com/laclance/go-sr"

levels, err := sr.Compute(candles, sr.Options{
    Timeframe: "5m",
    Lookback:  50,
    Mode:      sr.ModeZones,
})
if err != nil {
    return err
}

agg15m := sr.AggregateCandlesToTimeframe(candles, "5m", "15m")
levels15m, err := sr.Compute(agg15m, sr.Options{
    Timeframe: "15m",
    Lookback:  50,
    Mode:      sr.ModeZones,
})
if err != nil {
    return err
}
```

See the runnable examples in `examples_test.go` for minimal workflows covering zone mode, legacy mode, and multi-timeframe aggregation.
