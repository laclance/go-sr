// Package sr provides deterministic, closed-candle support/resistance detection
// for Go trading systems, backtests, and live strategies.
//
// It is designed for reproducible S/R signals without lookahead: zone-mode
// pivots are confirmation-based, and a given closed-candle prefix plus options
// produces deterministic output. Results include qualified levels, raw zones,
// nearest support/resistance prices, distances, strengths, scores, and proximity
// flags suitable for strategy logic.
//
// Two compute modes are supported:
//   - ModeLegacy: line-based pivots with fixed-tolerance proximity
//   - ModeZones: zone-based detection with composite scoring and raw/qualified output
//
// The package also provides S/R-specific multi-timeframe helpers for aggregating
// candles and calculating warmup/fetch requirements. Exchange parsing,
// app-specific timeframe policy, strategy scoring, and order execution remain
// outside the package.
package sr
