# BBGO closed-kline adapter

This example shows the adapter code that a BBGO strategy can use to feed closed 5m klines into `go-sr`. The source is intentionally excluded from this repository's build with `//go:build ignore`, so BBGO remains a dependency of the consuming strategy rather than the `go-sr` root module.

Copy the relevant fields and methods from [`adapter.go`](adapter.go) into your BBGO strategy package, then remove the build constraint. If your strategy already defines `Strategy`, `Subscribe`, or `Run`, merge the shown fields and callback into those definitions instead of adding duplicates.

The integration follows four rules:

1. Subscribe to BBGO's 5m kline channel.
2. Append data only inside `MarketDataStream.OnKLineClosed(func(kline types.KLine))` and defensively reject any event whose `Closed` flag is false.
3. Convert BBGO timestamps with `.Time()` and fixed-point OHLCV values with `.Float64()`.
4. Keep at most 240 closed candles, wait for the 120-candle lookback, and then call `sr.Compute` in zone mode.

The callback includes the newly closed kline in each computation. It must never append the current still-open kline: do not build this slice from `OnKLine`, a live-kline cache, or a REST response that includes an unfinished final candle.

The result exposes `NearestSupport`, `NearestResistance`, `NearSupport`, and `NearResistance` for strategy decisions. A zero nearest price means no qualifying level was found on that side.

This example is integration guidance, not a runnable package in `go-sr`. Build and run the adapted code from the BBGO module that owns your strategy and its BBGO dependency. BBGO is licensed under AGPL-3.0; review its license requirements for your use case.
