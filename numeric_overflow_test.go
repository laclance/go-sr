package sr

import (
	"math"
	"testing"
	"time"
)

func TestComputeRejectsNonFiniteDerivedZone(t *testing.T) {
	start := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	candles := make([]Candle, 9)
	for i := range candles {
		openTime := start.Add(time.Duration(i) * 5 * time.Minute)
		candles[i] = Candle{
			OpenTime:  openTime,
			CloseTime: openTime.Add(5 * time.Minute),
			Open:      9e307,
			High:      1e308,
			Low:       8e307,
			Close:     9e307,
			Volume:    1,
		}
	}
	candles[4].High = math.MaxFloat64

	levels, err := Compute(candles, Options{
		Timeframe:   "5m",
		Lookback:    len(candles),
		Mode:        ModeZones,
		MinStrength: 1,
	})
	if err == nil {
		t.Fatal("expected non-finite derived zone to be rejected")
	}
	if levels.Timeframe != "5m" || len(levels.Levels) != 0 || len(levels.RawZones) != 0 {
		t.Fatalf("expected empty levels on arithmetic overflow, got %+v", levels)
	}
}

func TestComputeLegacyAvoidsClusterSumOverflow(t *testing.T) {
	start := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	candles := make([]Candle, 17)
	for i := range candles {
		openTime := start.Add(time.Duration(i) * 5 * time.Minute)
		candles[i] = Candle{
			OpenTime:  openTime,
			CloseTime: openTime.Add(5 * time.Minute),
			Open:      8.5e307,
			High:      9e307,
			Low:       8e307,
			Close:     8.5e307,
			Volume:    1,
		}
	}
	candles[5].High = 1e308
	candles[11].High = 1e308

	levels, err := Compute(candles, Options{
		Timeframe: "5m",
		Lookback:  len(candles),
		Mode:      ModeLegacy,
		Tolerance: 0.002,
	})
	if err != nil {
		t.Fatalf("expected overflow-resistant legacy clustering, got %v", err)
	}
	if len(levels.Levels) != 1 {
		t.Fatalf("expected one merged legacy level, got %d", len(levels.Levels))
	}
	level := levels.Levels[0]
	if !isFiniteFloat(level.Price) || !isFiniteFloat(level.Top) || !isFiniteFloat(level.Bottom) {
		t.Fatalf("expected finite legacy level, got %+v", level)
	}
	if level.Price != 1e308 || level.Strength != 2 {
		t.Fatalf("unexpected merged legacy level: %+v", level)
	}
}

func TestAggregateCandlesToTimeframeRejectsVolumeOverflow(t *testing.T) {
	start := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	candles := make([]Candle, 3)
	for i := range candles {
		openTime := start.Add(time.Duration(i) * 5 * time.Minute)
		candles[i] = Candle{
			OpenTime:  openTime,
			CloseTime: openTime.Add(5 * time.Minute),
			Open:      100,
			High:      110,
			Low:       90,
			Close:     105,
			Volume:    1e308,
		}
	}

	if got := AggregateCandlesToTimeframe(candles, "5m", "15m"); got != nil {
		t.Fatalf("expected nil on aggregated volume overflow, got %+v", got)
	}
}
