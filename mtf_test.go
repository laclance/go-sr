package sr

import (
	"testing"
	"time"
)

func TestAggregateCandlesToTimeframe_5mTo15mOHLCV(t *testing.T) {
	start := time.Date(2024, 4, 1, 0, 0, 0, 0, time.UTC)
	candles := make5mCandles(start, []struct {
		open   float64
		high   float64
		low    float64
		close  float64
		volume float64
	}{
		{100, 101, 99, 100.5, 10},
		{100.5, 102, 100, 101.5, 20},
		{101.5, 103, 101, 102.5, 30},
		{102.5, 104, 102, 103.0, 40},
		{103.0, 105, 102.5, 104.5, 50},
		{104.5, 106, 104, 105.5, 60},
	})

	agg := AggregateCandlesToTimeframe(candles, "5m", "15m")
	if len(agg) != 2 {
		t.Fatalf("expected 2 aggregated candles, got %d", len(agg))
	}

	first := agg[0]
	if !first.OpenTime.Equal(start) || !first.CloseTime.Equal(start.Add(15*time.Minute)) {
		t.Fatalf("unexpected 15m bucket times: %s -> %s", first.OpenTime, first.CloseTime)
	}
	if first.Open != 100 || first.High != 103 || first.Low != 99 || first.Close != 102.5 || first.Volume != 60 {
		t.Fatalf("unexpected first 15m OHLCV: %+v", first)
	}
}

func TestAggregateCandlesToTimeframe_AcceptsBinanceInclusiveSourceCloseTimes(t *testing.T) {
	start := time.Date(2024, 4, 1, 0, 0, 0, 0, time.UTC)
	values := []struct {
		open   float64
		high   float64
		low    float64
		close  float64
		volume float64
	}{
		{100, 101, 99, 100.5, 10},
		{100.5, 102, 100, 101.5, 20},
		{101.5, 103, 101, 102.5, 30},
	}
	candles := make5mCandles(start, values)
	for i := range candles {
		candles[i].CloseTime = candles[i].CloseTime.Add(-time.Millisecond)
	}

	agg := AggregateCandlesToTimeframe(candles, "5m", "15m")
	if len(agg) != 1 {
		t.Fatalf("expected one complete bucket from Binance-style closes, got %d", len(agg))
	}
	if !agg[0].CloseTime.Equal(start.Add(15 * time.Minute)) {
		t.Fatalf("aggregated close time = %s, want exclusive bucket end %s", agg[0].CloseTime, start.Add(15*time.Minute))
	}
}

func TestAggregateCandlesToTimeframe_5mTo1hOHLCV(t *testing.T) {
	start := time.Date(2024, 4, 1, 0, 0, 0, 0, time.UTC)
	values := make([]struct {
		open   float64
		high   float64
		low    float64
		close  float64
		volume float64
	}, 12)
	for i := range values {
		values[i] = struct {
			open   float64
			high   float64
			low    float64
			close  float64
			volume float64
		}{
			open:   100 + float64(i),
			high:   101 + float64(i),
			low:    99 + float64(i),
			close:  100.5 + float64(i),
			volume: 10 + float64(i),
		}
	}

	agg := AggregateCandlesToTimeframe(make5mCandles(start, values), "5m", "1h")
	if len(agg) != 1 {
		t.Fatalf("expected 1 aggregated candle, got %d", len(agg))
	}

	first := agg[0]
	if !first.OpenTime.Equal(start) || !first.CloseTime.Equal(start.Add(time.Hour)) {
		t.Fatalf("unexpected 1h bucket times: %s -> %s", first.OpenTime, first.CloseTime)
	}
	if first.Open != 100 || first.High != 112 || first.Low != 99 || first.Close != 111.5 || first.Volume != 186 {
		t.Fatalf("unexpected first 1h OHLCV: %+v", first)
	}
}

func TestAggregateCandlesToTimeframe_DropsTrailingPartialBucket(t *testing.T) {
	start := time.Date(2024, 4, 1, 0, 0, 0, 0, time.UTC)
	candles := make5mCandles(start, []struct {
		open   float64
		high   float64
		low    float64
		close  float64
		volume float64
	}{
		{100, 101, 99, 100.5, 10},
		{100.5, 102, 100, 101.5, 20},
		{101.5, 103, 101, 102.5, 30},
		{102.5, 104, 102, 103.0, 40},
	})

	agg := AggregateCandlesToTimeframe(candles, "5m", "15m")
	if len(agg) != 1 {
		t.Fatalf("expected trailing partial bucket to be dropped, got %d buckets", len(agg))
	}
}

func TestAggregateCandlesToTimeframe_UsesUTCBucketAlignment(t *testing.T) {
	start := time.Date(2024, 4, 1, 0, 5, 0, 0, time.UTC)
	values := make([]struct {
		open   float64
		high   float64
		low    float64
		close  float64
		volume float64
	}, 7)
	for i := range values {
		values[i] = struct {
			open   float64
			high   float64
			low    float64
			close  float64
			volume float64
		}{100, 101, 99, 100.5, 10}
	}

	agg := AggregateCandlesToTimeframe(make5mCandles(start, values), "5m", "15m")
	if len(agg) != 1 {
		t.Fatalf("expected exactly one fully aligned 15m bucket, got %d", len(agg))
	}
	if want := time.Date(2024, 4, 1, 0, 15, 0, 0, time.UTC); !agg[0].OpenTime.Equal(want) {
		t.Fatalf("expected aligned bucket to start at %s, got %s", want, agg[0].OpenTime)
	}
	if want := time.Date(2024, 4, 1, 0, 30, 0, 0, time.UTC); !agg[0].CloseTime.Equal(want) {
		t.Fatalf("expected aligned bucket to close at %s, got %s", want, agg[0].CloseTime)
	}
}

func TestAggregateCandlesToTimeframe_InvalidIntervalsReturnNil(t *testing.T) {
	candles := makeFlatCandles(12, 100, time.Date(2024, 4, 1, 0, 0, 0, 0, time.UTC))
	cases := []struct {
		name string
		from string
		to   string
	}{
		{name: "invalid source", from: "x", to: "1h"},
		{name: "invalid target", from: "5m", to: "y"},
		{name: "same timeframe", from: "5m", to: "5m"},
		{name: "smaller target", from: "1h", to: "15m"},
		{name: "non divisible", from: "15m", to: "1h30m"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := AggregateCandlesToTimeframe(candles, tc.from, tc.to); got != nil {
				t.Fatalf("expected nil for %s -> %s, got %d candles", tc.from, tc.to, len(got))
			}
		})
	}
}

func TestWarmupCandles_ModeAware(t *testing.T) {
	if got := WarmupCandles(50, ModeLegacy); got != 70 {
		t.Fatalf("legacy warmup: got %d want 70", got)
	}
	if got := WarmupCandles(50, ModeZones); got != 68 {
		t.Fatalf("zone warmup: got %d want 68", got)
	}
	if got := WarmupCandles(-10, ModeLegacy); got != 20 {
		t.Fatalf("negative lookback should clamp to zero in legacy mode: got %d want 20", got)
	}
	if got := WarmupCandles(-10, ModeZones); got != 18 {
		t.Fatalf("negative lookback should clamp to zero in zone mode: got %d want 18", got)
	}
}

func TestRequiredKlineLimit_ModeAware(t *testing.T) {
	if got := RequiredKlineLimit("5m", "1h", 50, ModeLegacy); got != 841 {
		t.Fatalf("legacy required limit: got %d want 841", got)
	}
	if got := RequiredKlineLimit("5m", "1h", 50, ModeZones); got != 817 {
		t.Fatalf("zone required limit: got %d want 817", got)
	}
	if got := RequiredKlineLimit("15m", "1h", 50, ModeLegacy); got != 281 {
		t.Fatalf("legacy 15m->1h required limit: got %d want 281", got)
	}
}

func TestRequiredKlineLimit_InvalidCombosReturnZero(t *testing.T) {
	cases := []struct {
		name string
		base string
		to   string
	}{
		{name: "invalid base", base: "x", to: "1h"},
		{name: "invalid target", base: "5m", to: "y"},
		{name: "same interval", base: "5m", to: "5m"},
		{name: "smaller target", base: "1h", to: "15m"},
		{name: "not divisible", base: "40m", to: "1h"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := RequiredKlineLimit(tc.base, tc.to, 50, ModeZones); got != 0 {
				t.Fatalf("expected zero required limit for %s -> %s, got %d", tc.base, tc.to, got)
			}
		})
	}
}
