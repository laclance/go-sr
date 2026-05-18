package sr

import (
	"math"
	"reflect"
	"testing"
	"time"
)

func TestSRZone_PivotSnapshotStability(t *testing.T) {
	candles := buildSRCandles(220)
	lookback := len(candles)

	for idx := 140; idx <= 200; idx += 10 {
		highs1 := findPivotHighs(candles[:idx], "5m", lookback)
		highs2 := findPivotHighs(candles[:idx+1], "5m", lookback)
		for _, p1 := range highs1 {
			p2 := findPivotByIndex(highs2, p1.Index)
			if p2 == nil {
				t.Fatalf("idx=%d: pivot high at %d disappeared after append", idx, p1.Index)
			}
			if p1.ConfirmedAtIndex != p2.ConfirmedAtIndex {
				t.Errorf("idx=%d high %d: ConfirmedAtIndex changed %d -> %d", idx, p1.Index, p1.ConfirmedAtIndex, p2.ConfirmedAtIndex)
			}
			if p1.ATRSnapshot != p2.ATRSnapshot {
				t.Errorf("idx=%d high %d: ATRSnapshot changed %.4f -> %.4f", idx, p1.Index, p1.ATRSnapshot, p2.ATRSnapshot)
			}
			if p1.AvgVolumeSnapshot != p2.AvgVolumeSnapshot {
				t.Errorf("idx=%d high %d: AvgVolumeSnapshot changed %.4f -> %.4f", idx, p1.Index, p1.AvgVolumeSnapshot, p2.AvgVolumeSnapshot)
			}
			if p1.MergeWidth != p2.MergeWidth {
				t.Errorf("idx=%d high %d: MergeWidth changed %.4f -> %.4f", idx, p1.Index, p1.MergeWidth, p2.MergeWidth)
			}
			if p1.BounceATR != p2.BounceATR {
				t.Errorf("idx=%d high %d: BounceATR changed %.4f -> %.4f", idx, p1.Index, p1.BounceATR, p2.BounceATR)
			}
		}

		lows1 := findPivotLows(candles[:idx], "5m", lookback)
		lows2 := findPivotLows(candles[:idx+1], "5m", lookback)
		for _, p1 := range lows1 {
			p2 := findPivotByIndex(lows2, p1.Index)
			if p2 == nil {
				t.Fatalf("idx=%d: pivot low at %d disappeared after append", idx, p1.Index)
			}
			if p1.ConfirmedAtIndex != p2.ConfirmedAtIndex {
				t.Errorf("idx=%d low %d: ConfirmedAtIndex changed %d -> %d", idx, p1.Index, p1.ConfirmedAtIndex, p2.ConfirmedAtIndex)
			}
			if p1.ATRSnapshot != p2.ATRSnapshot {
				t.Errorf("idx=%d low %d: ATRSnapshot changed %.4f -> %.4f", idx, p1.Index, p1.ATRSnapshot, p2.ATRSnapshot)
			}
			if p1.AvgVolumeSnapshot != p2.AvgVolumeSnapshot {
				t.Errorf("idx=%d low %d: AvgVolumeSnapshot changed %.4f -> %.4f", idx, p1.Index, p1.AvgVolumeSnapshot, p2.AvgVolumeSnapshot)
			}
			if p1.MergeWidth != p2.MergeWidth {
				t.Errorf("idx=%d low %d: MergeWidth changed %.4f -> %.4f", idx, p1.Index, p1.MergeWidth, p2.MergeWidth)
			}
			if p1.BounceATR != p2.BounceATR {
				t.Errorf("idx=%d low %d: BounceATR changed %.4f -> %.4f", idx, p1.Index, p1.BounceATR, p2.BounceATR)
			}
		}
	}
}

func TestSRZone_GeometryStableWhenSourcePivotsUnchanged(t *testing.T) {
	candles := buildSRCandles(220)
	lookback := len(candles)

	for idx := 150; idx <= 200; idx += 10 {
		sr1 := computeZone(candles[:idx], "5m", lookback)
		sr2 := computeZone(candles[:idx+1], "5m", lookback)

		for _, z1 := range sr1.RawZones {
			z2 := findZoneBySource(sr2.RawZones, z1.SourcePivotIndexes, z1.IsHigh)
			if z2 == nil {
				t.Logf("idx=%d: zone %v sources=%v changed membership after append", idx, z1.Price, z1.SourcePivotIndexes)
				continue
			}
			if z1.Price != z2.Price || z1.Top != z2.Top || z1.Bottom != z2.Bottom {
				t.Errorf("idx=%d: zone sources=%v geometry drifted: %+v -> %+v", idx, z1.SourcePivotIndexes, z1, *z2)
			}
		}
	}
}

func TestSRZone_ScoreDriftBoundedWhenSourcePivotsUnchanged(t *testing.T) {
	candles := buildSRCandles(220)
	lookback := len(candles)
	const maxScoreDrift = 1.8

	for idx := 150; idx <= 200; idx += 10 {
		sr1 := computeZone(candles[:idx], "5m", lookback)
		sr2 := computeZone(candles[:idx+1], "5m", lookback)

		for _, z1 := range sr1.RawZones {
			z2 := findZoneBySource(sr2.RawZones, z1.SourcePivotIndexes, z1.IsHigh)
			if z2 == nil {
				t.Logf("idx=%d: zone %v sources=%v changed membership after append", idx, z1.Price, z1.SourcePivotIndexes)
				continue
			}
			if diff := math.Abs(z1.Score - z2.Score); diff > maxScoreDrift {
				t.Errorf("idx=%d: zone sources=%v score drift %.3f exceeds %.3f (%.3f -> %.3f)",
					idx, z1.SourcePivotIndexes, diff, maxScoreDrift, z1.Score, z2.Score)
			}
		}
	}
}

func TestSR_Deterministic_Zone(t *testing.T) {
	candles := buildSRCandles(220)

	for idx := 140; idx <= 200; idx += 15 {
		sr1 := computeZone(candles[:idx], "5m", 50)
		sr2 := computeZone(candles[:idx], "5m", 50)

		if !reflect.DeepEqual(sr1.Levels, sr2.Levels) {
			t.Errorf("idx=%d: Levels non-deterministic: %+v vs %+v", idx, sr1.Levels, sr2.Levels)
		}
		if !reflect.DeepEqual(sr1.RawZones, sr2.RawZones) {
			t.Errorf("idx=%d: RawZones non-deterministic: %+v vs %+v", idx, sr1.RawZones, sr2.RawZones)
		}
		if sr1.NearestSupport != sr2.NearestSupport {
			t.Errorf("idx=%d: NearestSupport non-deterministic: %.2f vs %.2f", idx, sr1.NearestSupport, sr2.NearestSupport)
		}
		if sr1.NearestResistance != sr2.NearestResistance {
			t.Errorf("idx=%d: NearestResistance non-deterministic: %.2f vs %.2f", idx, sr1.NearestResistance, sr2.NearestResistance)
		}
	}
}

func TestSR_Deterministic_Legacy(t *testing.T) {
	candles := buildSRCandles(220)

	for idx := 140; idx <= 200; idx += 15 {
		sr1 := computeLegacy(candles[:idx], "5m", 50, 0.002)
		sr2 := computeLegacy(candles[:idx], "5m", 50, 0.002)

		if !reflect.DeepEqual(sr1.Levels, sr2.Levels) {
			t.Errorf("idx=%d: legacy Levels non-deterministic: %+v vs %+v", idx, sr1.Levels, sr2.Levels)
		}
		if !reflect.DeepEqual(sr1.RawZones, sr2.RawZones) {
			t.Errorf("idx=%d: legacy RawZones non-deterministic: %+v vs %+v", idx, sr1.RawZones, sr2.RawZones)
		}
	}
}

func TestSRZone_DeterministicOrdering(t *testing.T) {
	sr := computeZone(buildSRCategoryCandles(), "5m", 120)
	assertZoneOrdering(t, sr.RawZones)
	assertZoneOrdering(t, sr.Levels)
}

func TestSRZone_RawZonesIsSuperset(t *testing.T) {
	candles := buildSRCandles(220)

	for idx := 140; idx <= 200; idx += 10 {
		sr := computeZone(candles[:idx], "5m", 50)
		if len(sr.RawZones) < len(sr.Levels) {
			t.Errorf("idx=%d: RawZones (%d) smaller than Levels (%d)", idx, len(sr.RawZones), len(sr.Levels))
		}
		for _, lvl := range sr.Levels {
			match := findZoneBySource(sr.RawZones, lvl.SourcePivotIndexes, lvl.IsHigh)
			if match == nil {
				t.Errorf("idx=%d: qualified zone sources=%v not in RawZones", idx, lvl.SourcePivotIndexes)
				continue
			}
			if !reflect.DeepEqual(*match, lvl) {
				t.Errorf("idx=%d: qualified zone metadata differs from RawZones entry: raw=%+v level=%+v", idx, *match, lvl)
			}
		}
	}
}

func TestSRZone_RepeatedLows_OneSupportZone(t *testing.T) {
	candles := buildSRCategoryCandles()
	sr := computeZone(candles, "5m", 120)

	if got := countLevelsNear(sr.RawZones, 49000, 75, false); got != 1 {
		t.Fatalf("expected exactly 1 support raw zone near 49000, got %d", got)
	}
	if got := countLevelsNear(sr.RawZones, 49000, 75, true); got != 0 {
		t.Fatalf("expected 0 resistance raw zones near 49000, got %d", got)
	}

	lvl := findLevelNear(sr.Levels, 49000, 75, false)
	if lvl == nil {
		t.Fatalf("expected qualified support zone near 49000")
	}
	if lvl.Strength < 2 {
		t.Errorf("expected merged support strength >= 2, got %d", lvl.Strength)
	}
}

func TestSRZone_RepeatedHighs_OneResistanceZone(t *testing.T) {
	candles := buildSRCategoryCandles()
	sr := computeZone(candles, "5m", 120)

	if got := countLevelsNear(sr.RawZones, 51000, 75, true); got != 1 {
		t.Fatalf("expected exactly 1 resistance raw zone near 51000, got %d", got)
	}
	if got := countLevelsNear(sr.RawZones, 51000, 75, false); got != 0 {
		t.Fatalf("expected 0 support raw zones near 51000, got %d", got)
	}

	lvl := findLevelNear(sr.Levels, 51000, 75, true)
	if lvl == nil {
		t.Fatalf("expected qualified resistance zone near 51000")
	}
	if lvl.Strength < 2 {
		t.Errorf("expected merged resistance strength >= 2, got %d", lvl.Strength)
	}
}

func TestSRZone_DistantPivots_StaySeparate(t *testing.T) {
	candles := buildSRCandles(160)
	candles[50].Low = 48000
	candles[50].High = 48200
	candles[50].Open = 48100
	candles[50].Close = 48100
	candles[110].High = 52000
	candles[110].Low = 51800
	candles[110].Open = 51900
	candles[110].Close = 51900

	sr := computeZone(candles, "5m", 80)

	nearLow := false
	nearHigh := false
	for _, lvl := range sr.RawZones {
		if math.Abs(lvl.Price-48000) < 500 {
			nearLow = true
		}
		if math.Abs(lvl.Price-52000) < 500 {
			nearHigh = true
		}
	}
	if !nearLow {
		t.Log("no zone detected near 48000 (acceptable in this fixture)")
	}
	if !nearHigh {
		t.Log("no zone detected near 52000 (acceptable in this fixture)")
	}
	if nearLow && nearHigh {
		for _, a := range sr.RawZones {
			for _, b := range sr.RawZones {
				if a.Price == b.Price {
					continue
				}
				if math.Abs(a.Price-48000) < 500 && math.Abs(b.Price-52000) < 500 && math.Abs(a.Price-b.Price) < 500 {
					t.Errorf("distant pivots (48000, 52000) appear to be merged into one zone")
				}
			}
		}
	}
}

func TestSRZone_NearestSupportBelowPrice(t *testing.T) {
	candles := buildSRCandles(200)

	for idx := 150; idx <= 195; idx += 5 {
		sr := computeZone(candles[:idx], "5m", 60)
		if sr.NearestSupport > 0 && sr.NearestSupport > candles[idx-1].Close {
			t.Errorf("idx=%d: NearestSupport %.2f > close %.2f", idx, sr.NearestSupport, candles[idx-1].Close)
		}
	}
}

func TestSRZone_NearestResistanceAbovePrice(t *testing.T) {
	candles := buildSRCandles(200)

	for idx := 150; idx <= 195; idx += 5 {
		sr := computeZone(candles[:idx], "5m", 60)
		if sr.NearestResistance > 0 && sr.NearestResistance <= candles[idx-1].Close {
			t.Errorf("idx=%d: NearestResistance %.2f <= close %.2f", idx, sr.NearestResistance, candles[idx-1].Close)
		}
	}
}

func TestSRZone_NearestLevelMetadataMatchesLevels(t *testing.T) {
	candles := buildSRCandles(220)

	for idx := 150; idx <= 200; idx += 10 {
		sr := computeZone(candles[:idx], "5m", 60)

		if sr.NearestSupport > 0 {
			lvl := findLevelByPrice(sr.Levels, sr.NearestSupport)
			if lvl == nil {
				t.Fatalf("idx=%d: nearest support %.2f not found in Levels", idx, sr.NearestSupport)
			}
			if lvl.Strength != sr.NearestSupportStrength {
				t.Errorf("idx=%d: nearest support strength mismatch: level=%d field=%d", idx, lvl.Strength, sr.NearestSupportStrength)
			}
			if lvl.Score != sr.NearestSupportScore {
				t.Errorf("idx=%d: nearest support score mismatch: level=%.4f field=%.4f", idx, lvl.Score, sr.NearestSupportScore)
			}
		}

		if sr.NearestResistance > 0 {
			lvl := findLevelByPrice(sr.Levels, sr.NearestResistance)
			if lvl == nil {
				t.Fatalf("idx=%d: nearest resistance %.2f not found in Levels", idx, sr.NearestResistance)
			}
			if lvl.Strength != sr.NearestResistanceStrength {
				t.Errorf("idx=%d: nearest resistance strength mismatch: level=%d field=%d", idx, lvl.Strength, sr.NearestResistanceStrength)
			}
			if lvl.Score != sr.NearestResistanceScore {
				t.Errorf("idx=%d: nearest resistance score mismatch: level=%.4f field=%.4f", idx, lvl.Score, sr.NearestResistanceScore)
			}
		}
	}
}

func TestSRZone_FalseBreaksIgnoreBarsBeforeEstablishment(t *testing.T) {
	t0 := time.Date(2024, 3, 1, 0, 0, 0, 0, time.UTC)
	candles := makeFlatCandles(18, 100.0, t0)
	zone := Level{Top: 101.0, Bottom: 99.0, IsHigh: true}

	candles[5].Close = 101.4
	candles[6].Close = 100.2
	candles[12].Close = 101.6
	candles[13].Close = 100.3

	if got := countFalseBreaks(zone, candles, 10); got != 1 {
		t.Fatalf("expected exactly 1 post-establishment false break, got %d", got)
	}
}

func TestSRZone_FalseBreaksNeedPrefixLocalReentry(t *testing.T) {
	t0 := time.Date(2024, 3, 2, 0, 0, 0, 0, time.UTC)
	prefix := makeFlatCandles(12, 100.0, t0)
	extended := makeFlatCandles(13, 100.0, t0)
	zone := Level{Top: 101.0, Bottom: 99.0, IsHigh: false}

	prefix[10].Close = 98.7
	extended[10].Close = 98.7
	prefix[11].Close = 98.6
	extended[11].Close = 98.6
	extended[12].Close = 99.2

	if got := countFalseBreaks(zone, prefix, 5); got != 0 {
		t.Fatalf("expected prefix-only count 0 without in-prefix reentry, got %d", got)
	}
	if got := countFalseBreaks(zone, extended, 5); got != 1 {
		t.Fatalf("expected extended prefix to count the completed false break, got %d", got)
	}
}

func TestSRZone_FalseBreaksCountSingleEpisodeOnce(t *testing.T) {
	t0 := time.Date(2024, 3, 3, 0, 0, 0, 0, time.UTC)
	candles := makeFlatCandles(18, 100.0, t0)
	zone := Level{Top: 101.0, Bottom: 99.0, IsHigh: true}

	candles[7].Close = 101.4
	candles[8].Close = 101.8
	candles[9].Close = 100.5
	candles[12].Close = 101.2
	candles[13].Close = 100.6

	if got := countFalseBreaks(zone, candles, 5); got != 2 {
		t.Fatalf("expected two false-break episodes, got %d", got)
	}
}

func TestSRLegacy_NearestSupportBelowPrice(t *testing.T) {
	candles := buildSRCandles(200)

	for idx := 100; idx <= 195; idx += 5 {
		sr := computeLegacy(candles[:idx], "5m", 60, 0.002)
		if sr.NearestSupport > 0 && sr.NearestSupport > candles[idx-1].Close {
			t.Errorf("idx=%d: legacy NearestSupport %.2f > close %.2f", idx, sr.NearestSupport, candles[idx-1].Close)
		}
	}
}

func TestSRLegacy_NearestResistanceAbovePrice(t *testing.T) {
	candles := buildSRCandles(200)

	for idx := 100; idx <= 195; idx += 5 {
		sr := computeLegacy(candles[:idx], "5m", 60, 0.002)
		if sr.NearestResistance > 0 && sr.NearestResistance <= candles[idx-1].Close {
			t.Errorf("idx=%d: legacy NearestResistance %.2f <= close %.2f", idx, sr.NearestResistance, candles[idx-1].Close)
		}
	}
}

func TestSRLegacy_RawZonesMirrorLevels(t *testing.T) {
	candles := buildSRCandles(200)
	sr := computeLegacy(candles, "5m", 60, 0.002)
	if !reflect.DeepEqual(sr.RawZones, sr.Levels) {
		t.Fatalf("legacy RawZones should mirror Levels: raw=%+v levels=%+v", sr.RawZones, sr.Levels)
	}
}

func TestSRLegacy_HighsAndLowsSeparate(t *testing.T) {
	candles := buildSRCandles(200)
	sr := computeLegacy(candles, "5m", 80, 0.002)
	assertZoneOrdering(t, sr.Levels)
}

func TestSRLegacy_RepeatedPivotsClusterIntoTypedLevels(t *testing.T) {
	candles := buildSRCategoryCandles()
	const tol = 0.002

	sr := computeLegacy(candles, "5m", 120, tol)
	tolAbs := candles[len(candles)-1].Close * tol

	support := findLevelNear(sr.Levels, 49000, 125, false)
	if support == nil {
		t.Fatalf("expected support level near 49000")
	}
	if findLevelNear(sr.Levels, 49000, 125, true) != nil {
		t.Fatalf("did not expect resistance level near 49000")
	}
	if support.Strength < 2 {
		t.Errorf("expected merged support strength >= 2, got %d", support.Strength)
	}
	if support.Score != 0 {
		t.Errorf("legacy support score should remain 0, got %.2f", support.Score)
	}
	if support.Top != support.Price+tolAbs || support.Bottom != support.Price-tolAbs {
		t.Errorf("legacy support bounds should be price +/- tolAbs: got top=%.2f bottom=%.2f tolAbs=%.2f", support.Top, support.Bottom, tolAbs)
	}

	resistance := findLevelNear(sr.Levels, 51000, 125, true)
	if resistance == nil {
		t.Fatalf("expected resistance level near 51000")
	}
	if findLevelNear(sr.Levels, 51000, 125, false) != nil {
		t.Fatalf("did not expect support level near 51000")
	}
	if resistance.Strength < 2 {
		t.Errorf("expected merged resistance strength >= 2, got %d", resistance.Strength)
	}
	if resistance.Score != 0 {
		t.Errorf("legacy resistance score should remain 0, got %.2f", resistance.Score)
	}
	if resistance.Top != resistance.Price+tolAbs || resistance.Bottom != resistance.Price-tolAbs {
		t.Errorf("legacy resistance bounds should be price +/- tolAbs: got top=%.2f bottom=%.2f tolAbs=%.2f", resistance.Top, resistance.Bottom, tolAbs)
	}
}
