package sr

import (
	"testing"
	"time"
)

func TestScoreZone_FalseBreakPenaltyReducesScore(t *testing.T) {
	baseCandles := makeFlatCandles(12, 100, time.Date(2024, 5, 1, 0, 0, 0, 0, time.UTC))
	falseBreakCandles := append([]Candle(nil), baseCandles...)
	falseBreakCandles[6].Close = 101.5
	falseBreakCandles[7].Close = 100.5

	zone := Level{Top: 101, Bottom: 99, Strength: 2, IsHigh: true, LastTouchIndex: 11}
	members := []srPivot{
		{ConfirmedAtIndex: 2},
		{ConfirmedAtIndex: 3},
	}

	withoutFalseBreak := scoreZone(zone, members, baseCandles, len(baseCandles))
	withFalseBreak := scoreZone(zone, members, falseBreakCandles, len(falseBreakCandles))
	if withFalseBreak >= withoutFalseBreak {
		t.Fatalf("false break should reduce zone score: without=%v with=%v", withoutFalseBreak, withFalseBreak)
	}
}

func TestDetectZoneProximity_ZeroWidthFallbackHasSmallPriceScale(t *testing.T) {
	zone := Level{Price: 100, Top: 100, Bottom: 100, Strength: 2, Score: 4, IsHigh: false}

	nearSup, nearRes, nearestSup, _, supDist, _, _, _, _, _ := detectZoneProximity([]Level{zone}, 100.25)
	if nearSup || nearRes {
		t.Fatalf("0.25%%-distant price should be outside zero-width 0.1%% fallback proximity, got support=%t resistance=%t", nearSup, nearRes)
	}
	if nearestSup != 100 || !almostEqual(supDist, 0.25) {
		t.Fatalf("unexpected nearest support result: price=%v distance=%v", nearestSup, supDist)
	}
}

func TestDetectZoneProximity_IncludesExactZoneBoundary(t *testing.T) {
	tests := []struct {
		name      string
		zone      Level
		price     float64
		wantNear  func(bool, bool) bool
		nearLabel string
	}{
		{
			name:      "support",
			zone:      Level{Price: 100, Top: 101, Bottom: 99, IsHigh: false},
			price:     102,
			wantNear:  func(nearSup, nearRes bool) bool { return nearSup && !nearRes },
			nearLabel: "support",
		},
		{
			name:      "resistance",
			zone:      Level{Price: 100, Top: 101, Bottom: 99, IsHigh: true},
			price:     98,
			wantNear:  func(nearSup, nearRes bool) bool { return !nearSup && nearRes },
			nearLabel: "resistance",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			nearSup, nearRes, _, _, _, _, _, _, _, _ := detectZoneProximity([]Level{tc.zone}, tc.price)
			if !tc.wantNear(nearSup, nearRes) {
				t.Fatalf("expected exact 2x half-width boundary to count as near %s, got support=%t resistance=%t", tc.nearLabel, nearSup, nearRes)
			}
		})
	}
}

func TestFindPivotHigh_VolumeRatioUsesAverageAsDivisor(t *testing.T) {
	candles := makeFlatCandles(25, 100, time.Date(2024, 5, 3, 0, 0, 0, 0, time.UTC))
	for i := range candles {
		candles[i].Volume = 100
	}
	candles[20].High = 110
	candles[20].Volume = 200

	pivot := findPivotByIndex(findPivotHighs(candles, "5m", 0), 20)
	if pivot == nil {
		t.Fatal("expected deterministic pivot high at index 20")
	}
	if pivot.AvgVolumeSnapshot != 105 {
		t.Fatalf("average-volume snapshot: got %v want 105", pivot.AvgVolumeSnapshot)
	}
	wantRatio := 200.0 / 105.0
	if !almostEqual(pivotInfo(*pivot).VolumeRatio, wantRatio) {
		t.Fatalf("volume ratio: got %v want %v", pivot.VolumeRatio, wantRatio)
	}
}

func TestDetectZoneProximity_ZeroWidthResistanceFallbackHasSmallPriceScale(t *testing.T) {
	zone := Level{Price: 100, Top: 100, Bottom: 100, Strength: 2, Score: 4, IsHigh: true}

	nearSup, nearRes, _, nearestRes, _, resDist, _, _, _, _ := detectZoneProximity([]Level{zone}, 99.75)
	if nearSup || nearRes {
		t.Fatalf("0.25%%-distant price should be outside zero-width 0.1%% fallback proximity, got support=%t resistance=%t", nearSup, nearRes)
	}
	if nearestRes != 100 || !almostEqual(resDist, 0.25) {
		t.Fatalf("unexpected nearest resistance result: price=%v distance=%v", nearestRes, resDist)
	}
}

func TestBuildZones_EvenMedianWidthControlsMergeThreshold(t *testing.T) {
	candles := makeFlatCandles(20, 100, time.Date(2024, 5, 2, 0, 0, 0, 0, time.UTC))
	pivots := []srPivot{
		{Index: 5, ConfirmedAtIndex: 9, Price: 100, MergeWidth: 2, IsHigh: false, Timeframe: "5m"},
		{Index: 6, ConfirmedAtIndex: 10, Price: 101, MergeWidth: 6, IsHigh: false, Timeframe: "5m"},
		{Index: 7, ConfirmedAtIndex: 11, Price: 103.5, MergeWidth: 1, IsHigh: false, Timeframe: "5m"},
	}

	zones := buildZones(pivots, candles, len(candles))
	if len(zones) != 1 || zones[0].Strength != len(pivots) {
		t.Fatalf("expected even median width 4 to keep all pivots in one zone, got %+v", zones)
	}
}

func TestComputeATR_MinimumHistoryBoundary(t *testing.T) {
	start := time.Date(2024, 5, 4, 0, 0, 0, 0, time.UTC)
	if got := computeATR(makeFlatCandles(rsiPeriod, 100, start), rsiPeriod); got != 0 {
		t.Fatalf("expected zero ATR with only period candles, got %v", got)
	}

	got := computeATR(makeFlatCandles(rsiPeriod+1, 100, start), rsiPeriod)
	if got <= 0 {
		t.Fatalf("expected positive ATR with exactly period+1 candles, got %v", got)
	}
}
