# BBGO closed-kline adapter

This example shows the adapter code that a BBGO strategy can use to feed closed 5m klines into `go-sr`. The directory is an isolated Go module so CI can compile the real adapter against BBGO without making BBGO a dependency of the `go-sr` root module.

The compatibility module pins BBGO v1.64.2 and uses the repository's current `go-sr` source through a local `replace` directive. From this directory, `go test ./...` compiles the adapter and its BBGO integration contract. Normal root-module commands such as `go test ./...` do not traverse this nested module.

Copy the relevant fields and methods from [`adapter.go`](adapter.go) into your BBGO strategy package. If your strategy already defines `Strategy`, `Subscribe`, or `Run`, merge the shown fields and callback into those definitions instead of adding duplicates.

The integration follows four rules:

1. Subscribe to BBGO's 5m kline channel.
2. Append data only inside `MarketDataStream.OnKLineClosed(func(kline types.KLine))` and defensively reject any event whose `Closed` flag is false.
3. Convert BBGO timestamps with `.Time()` and fixed-point OHLCV values with `.Float64()`.
4. Keep at most 240 closed candles, use `Lookback: lookback` for the S/R scan window, and wait until `sr.WarmupCandles(lookback, sr.ModeZones)` closed candles are available before calling `sr.Compute`.

The 120-candle `Lookback` controls how much history the S/R calculation scans; it is not the readiness threshold. `WarmupCandles` determines when enough closed history exists for a fully warmed calculation in the selected mode. For REST bootstrap sizing or higher-timeframe aggregation, use `sr.RequiredKlineLimit` rather than duplicating the warmup and alignment math.

The callback includes the newly closed kline in each computation. It must never append the current still-open kline: do not build this slice from `OnKLine`, a live-kline cache, or a REST response that includes an unfinished final candle.

The result exposes `NearestSupport`, `NearestResistance`, `NearSupport`, and `NearResistance` for strategy decisions. A zero nearest price means no qualifying level was found on that side.

This remains integration guidance rather than a complete runnable BBGO strategy. BBGO is intentionally confined to this example module and is not a dependency of `github.com/laclance/go-sr`. BBGO is licensed under AGPL-3.0; review its license requirements for your use case.
