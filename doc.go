// Package sr provides deterministic support/resistance detection for OHLCV
// candles plus a small set of generic multi-timeframe helpers used to build
// higher-timeframe SR bundles from a lower-timeframe stream.
//
// The package exposes Candle, PivotInfo, Level, Levels, Mode, Options,
// Compute, EmptyLevels, AggregateCandlesToTimeframe, WarmupCandles, and
// RequiredKlineLimit.
//
// Two compute modes are supported:
//   - ModeLegacy: line-based pivots with fixed-tolerance proximity
//   - ModeZones: zone-based detection with composite scoring and raw/qualified output
//
// The package is intentionally limited to SR detection and SR-specific
// timeframe helpers. Exchange parsing, app-specific timeframe policy, strategy
// scoring, and trade evaluation remain outside this package.
package sr
