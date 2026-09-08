package sr

import (
	"testing"
	"time"
)

func TestWarmupCandles_ExactMinimum(t *testing.T) {
	zonePadding := max(pivotWindow, max(rsiPeriod, avgVolPeriod)-pivotWindow)
	if zonePadding != 16 {
		t.Fatalf("zone warmup padding = %d, want 16 for current algorithm constants", zonePadding)
	}

	cases := []struct {
		name     string
		lookback int
		mode     Mode
		padding  int
	}{
		{name: "legacy small", lookback: 1, mode: ModeLegacy, padding: legacyPivotWindow},
		{name: "legacy representative", lookback: 50, mode: ModeLegacy, padding: legacyPivotWindow},
		{name: "legacy larger", lookback: 120, mode: ModeLegacy, padding: legacyPivotWindow},
		{name: "zones small", lookback: 1, mode: ModeZones, padding: zonePadding},
		{name: "zones representative", lookback: 50, mode: ModeZones, padding: zonePadding},
		{name: "zones larger", lookback: 120, mode: ModeZones, padding: zonePadding},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			want := tc.lookback + tc.padding
			if got := WarmupCandles(tc.lookback, tc.mode); got != want {
				t.Fatalf("WarmupCandles(%d, %q) = %d, want %d", tc.lookback, tc.mode, got, want)
			}
		})
	}
}

func TestWarmupCandles_LegacyBoundaryHasEarliestPivotContext(t *testing.T) {
	const lookback = 10
	closeTime := time.Date(2024, 4, 20, 0, 0, 0, 0, time.UTC)

	exactLen := WarmupCandles(lookback, ModeLegacy)
	exact := makeFlatCandles(exactLen, 100, closeTime)
	exactPivot := exactLen - lookback
	exact[exactPivot].High = 120

	got := computeSRLegacy(exact, "5m", lookback, 0.002)
	if len(got.Levels) != 1 || got.Levels[0].Price != 120 {
		t.Fatalf("exact warmup should retain earliest bounded pivot, got %+v", got.Levels)
	}

	shortLen := exactLen - 1
	short := makeFlatCandles(shortLen, 100, closeTime)
	shortPivot := shortLen - lookback
	short[shortPivot].High = 120

	got = computeSRLegacy(short, "5m", lookback, 0.002)
	if len(got.Levels) != 0 {
		t.Fatalf("one-less history should lack left context for earliest bounded pivot, got %+v", got.Levels)
	}
}

func TestWarmupCandles_ZonesBoundaryPopulatesFullAverageVolumeSnapshot(t *testing.T) {
	const lookback = 10
	closeTime := time.Date(2024, 4, 20, 0, 0, 0, 0, time.UTC)

	exactLen := WarmupCandles(lookback, ModeZones)
	exactPivot := exactLen - lookback
	exact := makeWarmupPivotCandles(exactLen, exactPivot, closeTime)
	// With the exact warmup, this candle is the 20th sample immediately before
	// the earliest pivot's confirmation candle and must be included.
	exact[avgVolPeriod-1].Volume = float64(avgVolPeriod + 1)

	exactPivotInfo := requirePivotAtIndex(t, findPivotHighs(exact, "5m", lookback), exactPivot)
	if exactPivotInfo.ConfirmedAtIndex != avgVolPeriod {
		t.Fatalf("exact earliest pivot confirmation = %d, want %d", exactPivotInfo.ConfirmedAtIndex, avgVolPeriod)
	}
	if exactPivotInfo.AvgVolumeSnapshot != 2 {
		t.Fatalf("exact average-volume snapshot = %v, want full %d-period average 2", exactPivotInfo.AvgVolumeSnapshot, avgVolPeriod)
	}
	if exactPivotInfo.ATRSnapshot <= 0 {
		t.Fatalf("exact ATR snapshot should be populated, got %v", exactPivotInfo.ATRSnapshot)
	}

	shortLen := exactLen - 1
	shortPivot := shortLen - lookback
	short := makeWarmupPivotCandles(shortLen, shortPivot, closeTime)
	// One candle less makes this the confirmation candle itself, which
	// computeAvgVolume excludes; only 19 prior samples are available.
	short[avgVolPeriod-1].Volume = float64(avgVolPeriod + 1)

	shortPivotInfo := requirePivotAtIndex(t, findPivotHighs(short, "5m", lookback), shortPivot)
	if shortPivotInfo.ConfirmedAtIndex != avgVolPeriod-1 {
		t.Fatalf("short earliest pivot confirmation = %d, want %d", shortPivotInfo.ConfirmedAtIndex, avgVolPeriod-1)
	}
	if shortPivotInfo.AvgVolumeSnapshot != 1 {
		t.Fatalf("one-less average-volume snapshot = %v, want partial-history average 1", shortPivotInfo.AvgVolumeSnapshot)
	}
}

func makeWarmupPivotCandles(n, pivotIndex int, closeTime time.Time) []Candle {
	candles := makeFlatCandles(n, 100, closeTime)
	for i := range candles {
		candles[i].Volume = 1
	}
	candles[pivotIndex].High = 120
	return candles
}

func requirePivotAtIndex(t *testing.T, pivots []srPivot, index int) srPivot {
	t.Helper()
	for _, pivot := range pivots {
		if pivot.Index == index {
			return pivot
		}
	}
	t.Fatalf("pivot at index %d not found in %+v", index, pivots)
	return srPivot{}
}
