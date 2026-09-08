package sr

import (
	"math"
	"strconv"
	"testing"
	"time"
)

func TestWarmupCandlesOverflow(t *testing.T) {
	for _, mode := range []Mode{ModeLegacy, ModeZones} {
		if got := WarmupCandles(math.MaxInt, mode); got != 0 {
			t.Fatalf("WarmupCandles(math.MaxInt, %q) = %d, want 0", mode, got)
		}
	}
}

func TestWarmupCandlesBoundary(t *testing.T) {
	cases := []struct {
		mode  Mode
		extra int
	}{
		{ModeLegacy, 20},
		{ModeZones, 18},
	}
	for _, tc := range cases {
		lookback := math.MaxInt - tc.extra
		if got := WarmupCandles(lookback, tc.mode); got != math.MaxInt {
			t.Fatalf("WarmupCandles(%d, %q) = %d, want %d", lookback, tc.mode, got, math.MaxInt)
		}
		if got := WarmupCandles(lookback+1, tc.mode); got != 0 {
			t.Fatalf("WarmupCandles overflow for %q = %d, want 0", tc.mode, got)
		}
	}
}

func TestRequiredKlineLimitBoundary(t *testing.T) {
	const extra = 18
	maxWarmup := math.MaxInt/2 - 1
	lookback := maxWarmup - extra
	want := (maxWarmup + 1) * 2
	if got := RequiredKlineLimit("1m", "2m", lookback, ModeZones); got != want {
		t.Fatalf("RequiredKlineLimit boundary = %d, want %d", got, want)
	}
	if got := RequiredKlineLimit("1m", "2m", lookback+1, ModeZones); got != 0 {
		t.Fatalf("RequiredKlineLimit overflow = %d, want 0", got)
	}
	if got := RequiredKlineLimit("5m", "1h", math.MaxInt, ModeZones); got != 0 {
		t.Fatalf("RequiredKlineLimit(math.MaxInt) = %d, want 0", got)
	}
}

func TestIntervalDurationBoundary(t *testing.T) {
	cases := []struct {
		suffix string
		unit   time.Duration
	}{
		{"m", time.Minute},
		{"h", time.Hour},
		{"d", 24 * time.Hour},
	}
	for _, tc := range cases {
		maxCount := int64(math.MaxInt64) / int64(tc.unit)
		good := strconv.FormatInt(maxCount, 10) + tc.suffix
		want := time.Duration(maxCount) * tc.unit
		if got := intervalDuration(good); got != want {
			t.Fatalf("intervalDuration(%q) = %s, want %s", good, got, want)
		}
		bad := strconv.FormatInt(maxCount+1, 10) + tc.suffix
		if got := intervalDuration(bad); got != 0 {
			t.Fatalf("intervalDuration(%q) = %s, want 0", bad, got)
		}
	}
}

func TestIntervalDurationSubsecondOverflowRejected(t *testing.T) {
	for _, interval := range []string{"3749353613647811m", "7498707227295622m"} {
		if got := intervalDuration(interval); got != 0 {
			t.Fatalf("intervalDuration(%q) = %s, want 0", interval, got)
		}
	}
}

func TestAggregateCandlesToTimeframeOverflowIntervalsReturnNil(t *testing.T) {
	candles := []Candle{{
		OpenTime:  time.Unix(0, 0).UTC(),
		CloseTime: time.Unix(60, 0).UTC(),
		Open:      1,
		High:      1,
		Low:       1,
		Close:     1,
		Volume:    1,
	}}
	if got := AggregateCandlesToTimeframe(candles, "3749353613647811m", "7498707227295622m"); got != nil {
		t.Fatalf("expected nil for overflowed interval pair, got %+v", got)
	}
}

func TestAggregateBucketStartRejectsInvalidDuration(t *testing.T) {
	if _, ok := aggregateBucketStart(time.Unix(0, 0), time.Nanosecond); ok {
		t.Fatal("expected sub-second bucket duration to be rejected")
	}
	if _, ok := aggregateBucketStart(time.Unix(math.MinInt64, 0), time.Minute); ok {
		t.Fatal("expected overflowing negative bucket floor to be rejected")
	}
	got, ok := aggregateBucketStart(time.Unix(-1, 0), time.Minute)
	if !ok || got.Unix() != -60 {
		t.Fatalf("negative timestamp floor = %v, ok=%v; want Unix -60", got, ok)
	}
}
