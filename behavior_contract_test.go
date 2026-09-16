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
	if got := computeATR(makeFlatCandles(atrPeriod, 100, start), atrPeriod); got != 0 {
		t.Fatalf("expected zero ATR with only period candles, got %v", got)
	}

	got := computeATR(makeFlatCandles(atrPeriod+1, 100, start), atrPeriod)
	if got <= 0 {
		t.Fatalf("expected positive ATR with exactly period+1 candles, got %v", got)
	}
}

func TestCandleDirection_DojiIsNeitherBullishNorBearish(t *testing.T) {
	doji := Candle{Open: 100, Close: 100}
	if doji.IsBullish() || doji.IsBearish() {
		t.Fatalf("doji should be neither bullish nor bearish: %+v", doji)
	}
}

func TestCompute_PivotPlateauEqualityDoesNotCreatePivot(t *testing.T) {
	tests := []struct {
		name  string
		apply func([]Candle)
	}{
		{
			name: "equal highs",
			apply: func(candles []Candle) {
				candles[8].High = 110
				candles[9].High = 110
			},
		},
		{
			name: "equal lows",
			apply: func(candles []Candle) {
				candles[8].Low = 90
				candles[9].Low = 90
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			for _, mode := range []Mode{ModeLegacy, ModeZones} {
				t.Run(string(mode), func(t *testing.T) {
					candles := makeFlatCandles(20, 100, time.Date(2024, 5, 5, 0, 0, 0, 0, time.UTC))
					tc.apply(candles)

					got, err := Compute(candles, Options{
						Timeframe:   "5m",
						Lookback:    0,
						Mode:        mode,
						Tolerance:   0.002,
						MinStrength: 1,
					})
					if err != nil {
						t.Fatalf("unexpected Compute error: %v", err)
					}
					if len(got.RawZones) != 0 {
						t.Fatalf("equal-price plateau should not produce a pivot in %s mode, got %+v", mode, got.RawZones)
					}
				})
			}
		})
	}
}

func TestCountFalseBreaks_StartsAfterZoneEstablishment(t *testing.T) {
	candles := makeFlatCandles(6, 100, time.Date(2024, 5, 6, 0, 0, 0, 0, time.UTC))
	candles[2].Close = 102
	candles[3].Close = 100

	if got := countFalseBreaks(Level{Top: 101, IsHigh: true}, candles, 2); got != 0 {
		t.Fatalf("breakout on establishment bar should not count as a false break, got %d", got)
	}
}

func TestCountFalseBreaks_ExactBoundaryIsInsideAndCountsAsReentry(t *testing.T) {
	tests := []struct {
		name          string
		zone          Level
		boundaryTouch []float64
		reentryTouch  []float64
	}{
		{
			name:          "resistance",
			zone:          Level{Top: 101, IsHigh: true},
			boundaryTouch: []float64{100, 101, 100, 100, 100, 100},
			reentryTouch:  []float64{100, 102, 101, 102, 102, 102},
		},
		{
			name:          "support",
			zone:          Level{Bottom: 99, IsHigh: false},
			boundaryTouch: []float64{100, 99, 100, 100, 100, 100},
			reentryTouch:  []float64{100, 98, 99, 98, 98, 98},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			atBoundary := makeFlatCandles(len(tc.boundaryTouch), 100, time.Date(2024, 5, 7, 0, 0, 0, 0, time.UTC))
			for i, closePrice := range tc.boundaryTouch {
				atBoundary[i].Close = closePrice
			}
			if got := countFalseBreaks(tc.zone, atBoundary, 0); got != 0 {
				t.Fatalf("exact boundary touch should remain inside the zone, got %d false breaks", got)
			}

			reentered := makeFlatCandles(len(tc.reentryTouch), 100, time.Date(2024, 5, 8, 0, 0, 0, 0, time.UTC))
			for i, closePrice := range tc.reentryTouch {
				reentered[i].Close = closePrice
			}
			if got := countFalseBreaks(tc.zone, reentered, 0); got != 1 {
				t.Fatalf("exact boundary touch should count as reentry after breakout, got %d false breaks", got)
			}
		})
	}
}
