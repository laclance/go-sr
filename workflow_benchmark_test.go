package sr

import (
	"math"
	"testing"
	"time"
)

var (
	benchmarkComputedLevels Levels
	benchmarkAggregated     []Candle
)

func BenchmarkComputeZones(b *testing.B) {
	cases := []struct {
		name     string
		candles  int
		lookback int
	}{
		{name: "bounded_120", candles: 120, lookback: 120},
		{name: "bounded_2000", candles: 2000, lookback: 2000},
		{name: "all_history_10000", candles: 10000, lookback: 0},
	}

	for _, tc := range cases {
		candles := benchmarkSwingCandles(tc.candles)
		opts := Options{
			Timeframe: "5m",
			Lookback:  tc.lookback,
			Mode:      ModeZones,
		}
		if _, err := Compute(candles, opts); err != nil {
			b.Fatalf("preflight Compute failed: %v", err)
		}

		b.Run(tc.name, func(b *testing.B) {
			b.ReportAllocs()
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				levels, err := Compute(candles, opts)
				if err != nil {
					b.Fatal(err)
				}
				benchmarkComputedLevels = levels
			}
		})
	}
}

func BenchmarkComputeLegacy(b *testing.B) {
	cases := []struct {
		name     string
		candles  int
		lookback int
	}{
		{name: "bounded_120", candles: 120, lookback: 120},
		{name: "all_history_10000", candles: 10000, lookback: 0},
	}

	for _, tc := range cases {
		candles := benchmarkSwingCandles(tc.candles)
		opts := Options{
			Timeframe: "5m",
			Lookback:  tc.lookback,
			Mode:      ModeLegacy,
		}
		if _, err := Compute(candles, opts); err != nil {
			b.Fatalf("preflight Compute failed: %v", err)
		}

		b.Run(tc.name, func(b *testing.B) {
			b.ReportAllocs()
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				levels, err := Compute(candles, opts)
				if err != nil {
					b.Fatal(err)
				}
				benchmarkComputedLevels = levels
			}
		})
	}
}

func BenchmarkAggregateCandlesToTimeframe(b *testing.B) {
	cases := []struct {
		name    string
		candles int
	}{
		{name: "5m_to_1h/normal_120", candles: 120},
		{name: "5m_to_1h/medium_12000", candles: 12000},
		{name: "5m_to_1h/large_100000", candles: 100000},
	}

	for _, tc := range cases {
		candles := benchmarkSwingCandles(tc.candles)
		if got := AggregateCandlesToTimeframe(candles, "5m", "1h"); len(got) == 0 {
			b.Fatalf("preflight aggregation returned no candles for %s", tc.name)
		}

		b.Run(tc.name, func(b *testing.B) {
			b.ReportAllocs()
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				benchmarkAggregated = AggregateCandlesToTimeframe(candles, "5m", "1h")
			}
		})
	}
}

func benchmarkSwingCandles(n int) []Candle {
	const interval = 5 * time.Minute
	start := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	candles := make([]Candle, n)
	previousClose := 100.0

	for i := range candles {
		x := float64(i)
		center := 100 + 10*math.Sin(x*0.18) + 4*math.Sin(x*0.047) + 0.0008*x
		close := center + 0.9*math.Sin(x*0.63)
		open := previousClose
		wick := 0.8 + 0.25*(1+math.Sin(x*0.29))
		high := math.Max(open, close) + wick
		low := math.Min(open, close) - wick*(0.9+0.1*math.Sin(x*0.11))
		volume := 1000 + 180*(1+math.Sin(x*0.23)) + 40*(1+math.Sin(x*0.071))
		openTime := start.Add(time.Duration(i) * interval)

		candles[i] = Candle{
			OpenTime:  openTime,
			CloseTime: openTime.Add(interval),
			Open:      open,
			High:      high,
			Low:       low,
			Close:     close,
			Volume:    volume,
		}
		previousClose = close
	}

	return candles
}
