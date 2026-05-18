# `github.com/laclance/go-sr`

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
    Timeframe string
    Lookback  int
    Mode      Mode
    Tolerance float64
}

func Compute(candles []Candle, opts Options) Levels
func EmptyLevels(timeframe string) Levels
func AggregateCandlesToTimeframe(candles []Candle, fromInterval, toInterval string) []Candle
func WarmupCandles(lookback int, mode Mode) int
func RequiredKlineLimit(baseInterval, targetInterval string, lookback int, mode Mode) int
```

`Tolerance` applies only to `ModeLegacy`. When `Tolerance <= 0`, the fallback remains `0.002`.

## Behavioral Contract

- `Compute` is deterministic for the same candle prefix and options.
- Zone-mode pivots are confirmation-based; no future candles are read beyond the current prefix.
- `AggregateCandlesToTimeframe` uses UTC-aligned buckets and drops leading/trailing partial buckets.
- `RequiredKlineLimit` returns the number of raw candles needed to build a higher-timeframe SR bundle and includes one extra live candle for exchange REST responses.
- Supported interval strings use `<n><unit>` with `m`, `h`, or `d`, and the target interval must be larger than and evenly divisible by the base interval.

## Install

```bash
go get github.com/laclance/go-sr
```

## Quick Start

```go
import sr "github.com/laclance/go-sr"

levels := sr.Compute(candles, sr.Options{
    Timeframe: "5m",
    Lookback:  50,
    Mode:      sr.ModeZones,
})

agg15m := sr.AggregateCandlesToTimeframe(candles, "5m", "15m")
levels15m := sr.Compute(agg15m, sr.Options{
    Timeframe: "15m",
    Lookback:  50,
    Mode:      sr.ModeZones,
})
```

See the runnable examples in `examples_test.go` for minimal workflows covering zone mode, legacy mode, and multi-timeframe aggregation.
