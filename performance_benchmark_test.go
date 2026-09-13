package sr

import (
	"math"
	"testing"
	"time"
)

var (
	benchmarkComputeLevels Levels
	benchmarkAggregated    []Candle
)

func BenchmarkComputeZonesEndToEnd(b *testing.B) {
	cases := []struct {
		name     string
		size     int
		lookback int
	}{
		{name: "bounded_120", size: WarmupCandles(120, ModeZones), lookback: 120},
		{name: "bounded_2000", size: WarmupCandles(2000, ModeZones), lookback: 2000},
		{name: "all_history_10000", size: 10000, lookback: 0},
	}

	for _, tc := range cases {
		candles := benchmarkSwingCandles(tc.size)
		b.Run(tc.name, func(b *testing.B) {
			opts := Options{
				Timeframe:   "5m",
				Lookback:    tc.lookback,
				Mode:        ModeZones,
				MinStrength: 2,
			}
			b.ReportAllocs()
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				levels, err := Compute(candles, opts)
				if err != nil {
					b.Fatal(err)
				}
				benchmarkComputeLevels = levels
			}
		})
	}
}

func BenchmarkComputeLegacyEndToEnd(b *testing.B) {
	cases := []struct {
		name     string
		size     int
		lookback int
	}{
		{name: "bounded_120", size: WarmupCandles(120, ModeLegacy), lookback: 120},
		{name: "all_history_5000", size: 5000, lookback: 0},
	}

	for _, tc := range cases {
		candles := benchmarkSwingCandles(tc.size)
		b.Run(tc.name, func(b *testing.B) {
			opts := Options{
				Timeframe: "5m",
				Lookback:  tc.lookback,
				Mode:      ModeLegacy,
			}
			b.ReportAllocs()
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				levels, err := Compute(candles, opts)
				if err != nil {
					b.Fatal(err)
				}
				benchmarkComputeLevels = levels
			}
		})
	}
}

func BenchmarkAggregateCandlesToTimeframeEndToEnd(b *testing.B) {
	cases := []struct {
		name string
		size int
	}{
		{name: "normal_120", size: 120},
		{name: "historical_12000", size: 12000},
		{name: "historical_100k", size: 100008},
	}

	for _, tc := range cases {
		candles := benchmarkSwingCandles(tc.size)
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
	candles := make([]Candle, n)
	start := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	previousClose := 100.0

	for i := range candles {
		x := float64(i)
		center := 100 + 2.5*math.Sin(2*math.Pi*x/233) + 0.75*math.Sin(2*math.Pi*x/997)
		closePrice := center + 4.5*math.Sin(2*math.Pi*x/24) + 1.2*math.Sin(2*math.Pi*x/53)
		openPrice := previousClose
		wick := 0.55 + 0.15*(1+math.Sin(2*math.Pi*x/17))
		high := math.Max(openPrice, closePrice) + wick
		low := math.Min(openPrice, closePrice) - wick*0.9
		volume := 900 + 180*(1+math.Sin(2*math.Pi*x/31)) + 25*float64(i%7)
		openTime := start.Add(time.Duration(i) * 5 * time.Minute)

		candles[i] = Candle{
			OpenTime:  openTime,
			CloseTime: openTime.Add(5 * time.Minute),
			Open:      openPrice,
			High:      high,
			Low:       low,
			Close:     closePrice,
			Volume:    volume,
		}
		previousClose = closePrice
	}

	return candles
}
