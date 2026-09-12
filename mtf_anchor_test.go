package sr

import (
	"testing"
	"time"
)

func TestAggregateCandlesToTimeframe_7hUsesUnixEpochAnchor(t *testing.T) {
	location := time.FixedZone("UTC+2", 2*60*60)
	start := time.Date(2024, 4, 1, 2, 0, 0, 0, location) // 2024-04-01 00:00 UTC.
	candles := makeIntervalCandles(start, 16, time.Hour)

	agg := AggregateCandlesToTimeframe(candles, "1h", "7h")
	if len(agg) != 2 {
		t.Fatalf("expected 2 complete 7h buckets, got %d", len(agg))
	}

	want := [][2]time.Time{
		{
			time.Date(2024, 4, 1, 2, 0, 0, 0, time.UTC),
			time.Date(2024, 4, 1, 9, 0, 0, 0, time.UTC),
		},
		{
			time.Date(2024, 4, 1, 9, 0, 0, 0, time.UTC),
			time.Date(2024, 4, 1, 16, 0, 0, 0, time.UTC),
		},
	}
	for i := range agg {
		if !agg[i].OpenTime.Equal(want[i][0]) || !agg[i].CloseTime.Equal(want[i][1]) {
			t.Fatalf("bucket %d times = %s -> %s, want %s -> %s", i, agg[i].OpenTime, agg[i].CloseTime, want[i][0], want[i][1])
		}
		if agg[i].OpenTime.Location() != time.UTC || agg[i].CloseTime.Location() != time.UTC {
			t.Fatalf("bucket %d times are not normalized to UTC: %s -> %s", i, agg[i].OpenTime, agg[i].CloseTime)
		}
	}
}

func TestAggregateCandlesToTimeframe_3dUsesUnixEpochAnchor(t *testing.T) {
	start := time.Date(2024, 3, 30, 0, 0, 0, 0, time.UTC)
	candles := makeIntervalCandles(start, 6, 24*time.Hour)

	agg := AggregateCandlesToTimeframe(candles, "1d", "3d")
	if len(agg) != 2 {
		t.Fatalf("expected 2 complete 3d buckets, got %d", len(agg))
	}

	want := [][2]time.Time{
		{
			time.Date(2024, 3, 30, 0, 0, 0, 0, time.UTC),
			time.Date(2024, 4, 2, 0, 0, 0, 0, time.UTC),
		},
		{
			time.Date(2024, 4, 2, 0, 0, 0, 0, time.UTC),
			time.Date(2024, 4, 5, 0, 0, 0, 0, time.UTC),
		},
	}
	for i := range agg {
		if !agg[i].OpenTime.Equal(want[i][0]) || !agg[i].CloseTime.Equal(want[i][1]) {
			t.Fatalf("bucket %d times = %s -> %s, want %s -> %s", i, agg[i].OpenTime, agg[i].CloseTime, want[i][0], want[i][1])
		}
	}
}

func makeIntervalCandles(start time.Time, count int, interval time.Duration) []Candle {
	candles := make([]Candle, count)
	for i := range candles {
		openTime := start.Add(time.Duration(i) * interval)
		candles[i] = Candle{
			OpenTime:  openTime,
			CloseTime: openTime.Add(interval),
			Open:      100,
			High:      101,
			Low:       99,
			Close:     100,
			Volume:    1,
		}
	}
	return candles
}
