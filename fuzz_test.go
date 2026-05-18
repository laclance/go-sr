package sr

import (
	"math"
	"testing"
	"time"
)

func FuzzAggregateCandlesToTimeframe(f *testing.F) {
	f.Add(12, int64(0))
	f.Add(37, int64(5))

	f.Fuzz(func(t *testing.T, count int, offsetMinutes int64) {
		if count < 0 {
			count = -count
		}
		count = count%180 + 1
		offsetMinutes %= 60
		if offsetMinutes < 0 {
			offsetMinutes = -offsetMinutes
		}

		start := time.Date(2024, 4, 1, 0, 0, 0, 0, time.UTC).Add(time.Duration(offsetMinutes) * time.Minute)
		candles := make([]Candle, count)
		for i := range candles {
			openTime := start.Add(time.Duration(i) * 5 * time.Minute)
			base := 100 + float64(i%17)
			candles[i] = Candle{
				OpenTime:  openTime,
				CloseTime: openTime.Add(5 * time.Minute),
				Open:      base,
				High:      base + 2,
				Low:       base - 2,
				Close:     base + 1,
				Volume:    10 + float64(i%7),
			}
		}

		agg := AggregateCandlesToTimeframe(candles, "5m", "15m")
		for _, candle := range agg {
			if candle.CloseTime.Sub(candle.OpenTime) != 15*time.Minute {
				t.Fatalf("expected 15m aggregated candle, got %s", candle.CloseTime.Sub(candle.OpenTime))
			}
			if candle.High < candle.Open || candle.High < candle.Close {
				t.Fatalf("aggregated high violated OHLC invariants: %+v", candle)
			}
			if candle.Low > candle.Open || candle.Low > candle.Close {
				t.Fatalf("aggregated low violated OHLC invariants: %+v", candle)
			}
		}
	})
}

func FuzzComputeInvariants(f *testing.F) {
	f.Add(120, int64(0), false)
	f.Add(220, int64(11), true)

	f.Fuzz(func(t *testing.T, count int, phaseSeed int64, useLegacy bool) {
		if count < 0 {
			count = -count
		}
		count = count%320 + 1

		candles := make([]Candle, count)
		start := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
		price := 50000.0
		for i := range candles {
			phase := float64((int64(i)+phaseSeed)%31) / 5
			price += math.Sin(phase)*15 + float64((i%5)-2)*6
			if price < 100 {
				price = 100
			}

			open := price - 10
			closePrice := price + 10
			if i%2 == 1 {
				open, closePrice = closePrice, open
			}

			openTime := start.Add(time.Duration(i) * 5 * time.Minute)
			candles[i] = Candle{
				OpenTime:  openTime,
				CloseTime: openTime.Add(5 * time.Minute),
				Open:      open,
				High:      math.Max(open, closePrice) + 5,
				Low:       math.Min(open, closePrice) - 5,
				Close:     closePrice,
				Volume:    100 + float64(i%9),
			}
		}

		mode := ModeZones
		if useLegacy {
			mode = ModeLegacy
		}
		levels, err := Compute(candles, Options{
			Timeframe: "5m",
			Lookback:  50,
			Mode:      mode,
			Tolerance: 0.002,
		})
		if err != nil {
			t.Fatalf("unexpected Compute error: %v", err)
		}

		if len(levels.RawZones) < len(levels.Levels) {
			t.Fatalf("raw zones smaller than qualified levels: raw=%d qualified=%d", len(levels.RawZones), len(levels.Levels))
		}
		if len(candles) == 0 {
			return
		}

		closePrice := candles[len(candles)-1].Close
		if levels.NearestSupport != 0 && levels.NearestSupport > closePrice {
			t.Fatalf("nearest support above price: support=%.2f close=%.2f", levels.NearestSupport, closePrice)
		}
		if levels.NearestResistance != 0 && levels.NearestResistance < closePrice {
			t.Fatalf("nearest resistance below price: resistance=%.2f close=%.2f", levels.NearestResistance, closePrice)
		}
	})
}
