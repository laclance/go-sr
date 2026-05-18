package sr

import (
	"reflect"
	"testing"
	"time"
)

func TestZoneProximity_NearSupport_WhenPriceJustAboveZone(t *testing.T) {
	zones := []Level{{Price: 49950, Top: 50000, Bottom: 49900, Strength: 2, Score: 5.0, IsHigh: false}}
	nearSup, nearRes, nearestSup, _, supDist, _, supStr, _, supScore, _ := detectZoneProximity(zones, 50000)
	if !nearSup {
		t.Error("expected nearSupport=true")
	}
	if nearRes {
		t.Error("expected nearResistance=false")
	}
	if nearestSup != 49950 {
		t.Errorf("nearestSupport: want 49950, got %v", nearestSup)
	}
	if supDist != 50 {
		t.Errorf("supportDistance: want 50, got %v", supDist)
	}
	if supStr != 2 {
		t.Errorf("supportStrength: want 2, got %v", supStr)
	}
	if supScore != 5.0 {
		t.Errorf("supportScore: want 5.0, got %v", supScore)
	}
}

func TestZoneProximity_NearResistance_WhenPriceJustBelowZone(t *testing.T) {
	zones := []Level{{Price: 50050, Top: 50100, Bottom: 50000, Strength: 3, Score: 7.0, IsHigh: true}}
	_, nearRes, _, nearestRes, _, resDist, _, resStr, _, resScore := detectZoneProximity(zones, 50000)
	if !nearRes {
		t.Error("expected nearResistance=true")
	}
	if nearestRes != 50050 {
		t.Errorf("nearestResistance: want 50050, got %v", nearestRes)
	}
	if resDist != 50 {
		t.Errorf("resistanceDistance: want 50, got %v", resDist)
	}
	if resStr != 3 {
		t.Errorf("resistanceStrength: want 3, got %v", resStr)
	}
	if resScore != 7.0 {
		t.Errorf("resistanceScore: want 7.0, got %v", resScore)
	}
}

func TestZoneProximity_ExactResistanceCenter_UsesZoneSide(t *testing.T) {
	zones := []Level{{Price: 50000, Top: 50100, Bottom: 49900, Strength: 3, Score: 7.0, IsHigh: true}}
	nearSup, nearRes, nearestSup, nearestRes, supDist, resDist, _, _, _, _ := detectZoneProximity(zones, 50000)
	if nearSup {
		t.Error("expected exact resistance-center touch not to be classified as support")
	}
	if !nearRes {
		t.Error("expected nearResistance=true")
	}
	if nearestSup != 0 || supDist != 0 {
		t.Errorf("nearest support should remain unset, got price=%v dist=%v", nearestSup, supDist)
	}
	if nearestRes != 50000 {
		t.Errorf("nearestResistance: want 50000, got %v", nearestRes)
	}
	if resDist != 0 {
		t.Errorf("resistanceDistance: want 0, got %v", resDist)
	}
}

func TestZoneProximity_ExactSupportCenter_UsesZoneSide(t *testing.T) {
	zones := []Level{{Price: 50000, Top: 50100, Bottom: 49900, Strength: 3, Score: 7.0, IsHigh: false}}
	nearSup, nearRes, nearestSup, nearestRes, supDist, resDist, _, _, _, _ := detectZoneProximity(zones, 50000)
	if !nearSup {
		t.Error("expected nearSupport=true")
	}
	if nearRes {
		t.Error("expected exact support-center touch not to be classified as resistance")
	}
	if nearestSup != 50000 {
		t.Errorf("nearestSupport: want 50000, got %v", nearestSup)
	}
	if supDist != 0 {
		t.Errorf("supportDistance: want 0, got %v", supDist)
	}
	if nearestRes != 0 || resDist != 0 {
		t.Errorf("nearest resistance should remain unset, got price=%v dist=%v", nearestRes, resDist)
	}
}

func TestZoneProximity_TooFarAway_NotNear(t *testing.T) {
	zones := []Level{{Price: 48000, Top: 48050, Bottom: 47950, Strength: 2, Score: 5.0, IsHigh: false}}
	nearSup, nearRes, _, _, _, _, _, _, _, _ := detectZoneProximity(zones, 50000)
	if nearSup || nearRes {
		t.Error("expected neither nearSupport nor nearResistance for distant zone")
	}
}

func TestZoneProximity_EmptyZones_AllFalse(t *testing.T) {
	nearSup, nearRes, ns, nr, nsd, nrd, ss, rs, _, _ := detectZoneProximity(nil, 50000)
	if nearSup || nearRes || ns != 0 || nr != 0 || nsd != 0 || nrd != 0 || ss != 0 || rs != 0 {
		t.Error("expected all zeros/false for empty zones")
	}
}

func TestBuildZones_AdjacentPivots_MergeIntoOneZone(t *testing.T) {
	candles := makeFlatCandles(40, 50000, time.Date(2024, 4, 1, 0, 0, 0, 0, time.UTC))
	pivots := []srPivot{
		{Index: 10, ConfirmedAtIndex: 14, Time: candles[10].CloseTime, Price: 50000, IsHigh: false, Timeframe: "5m", MergeWidth: 100, BounceATR: 0.5},
		{Index: 20, ConfirmedAtIndex: 24, Time: candles[20].CloseTime, Price: 50050, IsHigh: false, Timeframe: "5m", MergeWidth: 100, BounceATR: 0.6},
	}
	zones := buildZones(pivots, candles, 30)
	if len(zones) != 1 {
		t.Fatalf("expected 1 zone, got %d", len(zones))
	}
	if zones[0].Strength != 2 {
		t.Errorf("merged strength: want 2, got %d", zones[0].Strength)
	}
}

func TestBuildZones_DistantPivots_StaySeparate(t *testing.T) {
	candles := makeFlatCandles(40, 50000, time.Date(2024, 4, 2, 0, 0, 0, 0, time.UTC))
	pivots := []srPivot{
		{Index: 10, ConfirmedAtIndex: 14, Time: candles[10].CloseTime, Price: 50000, IsHigh: true, Timeframe: "5m", MergeWidth: 100, BounceATR: 0.5},
		{Index: 20, ConfirmedAtIndex: 24, Time: candles[20].CloseTime, Price: 51000, IsHigh: true, Timeframe: "5m", MergeWidth: 100, BounceATR: 0.5},
	}
	zones := buildZones(pivots, candles, 30)
	if len(zones) != 2 {
		t.Errorf("expected 2 zones, got %d", len(zones))
	}
}

func TestBuildZones_MergedPrice_IsAverage(t *testing.T) {
	candles := makeFlatCandles(40, 50000, time.Date(2024, 4, 3, 0, 0, 0, 0, time.UTC))
	pivots := []srPivot{
		{Index: 10, ConfirmedAtIndex: 14, Time: candles[10].CloseTime, Price: 50000, IsHigh: true, Timeframe: "5m", MergeWidth: 100, BounceATR: 0.5},
		{Index: 20, ConfirmedAtIndex: 24, Time: candles[20].CloseTime, Price: 50050, IsHigh: true, Timeframe: "5m", MergeWidth: 100, BounceATR: 0.5},
	}
	zones := buildZones(pivots, candles, 30)
	if len(zones) != 1 {
		t.Fatalf("expected 1 zone")
	}
	want := 50025.0
	if zones[0].Price < want-1 || zones[0].Price > want+1 {
		t.Errorf("merged center: want ~%.0f, got %.4f", want, zones[0].Price)
	}
}

func TestBuildZones_ZoneHasTopAndBottom(t *testing.T) {
	candles := makeFlatCandles(30, 50000, time.Date(2024, 4, 4, 0, 0, 0, 0, time.UTC))
	pivots := []srPivot{{Index: 10, ConfirmedAtIndex: 14, Time: candles[10].CloseTime, Price: 50000, IsHigh: false, Timeframe: "5m", MergeWidth: 100, BounceATR: 0.5}}
	zones := buildZones(pivots, candles, 20)
	if len(zones) != 1 {
		t.Fatalf("expected 1 zone")
	}
	z := zones[0]
	if z.Top != z.Price+50 || z.Bottom != z.Price-50 {
		t.Errorf("zone boundaries: want %.0f/%.0f, got %.0f/%.0f",
			z.Price+50, z.Price-50, z.Top, z.Bottom)
	}
}

func TestBuildZones_HighsAndLowsClusteredSeparately(t *testing.T) {
	t0 := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	candles := makeFlatCandles(40, 100.0, t0)

	highPivots := []srPivot{
		{Index: 15, ConfirmedAtIndex: 19, Time: candles[15].CloseTime, Price: 105.1, IsHigh: true, Timeframe: "5m", MergeWidth: 0.5, BounceATR: 2.0},
		{Index: 5, ConfirmedAtIndex: 9, Time: candles[5].CloseTime, Price: 105.0, IsHigh: true, Timeframe: "5m", MergeWidth: 0.5, BounceATR: 2.0},
	}
	lowPivots := []srPivot{
		{Index: 20, ConfirmedAtIndex: 24, Time: candles[20].CloseTime, Price: 95.05, IsHigh: false, Timeframe: "5m", MergeWidth: 0.5, BounceATR: 2.0},
		{Index: 10, ConfirmedAtIndex: 14, Time: candles[10].CloseTime, Price: 95.0, IsHigh: false, Timeframe: "5m", MergeWidth: 0.5, BounceATR: 2.0},
	}

	highZones := buildZones(highPivots, candles, 25)
	lowZones := buildZones(lowPivots, candles, 25)

	if len(highZones) != 1 {
		t.Errorf("expected 1 high zone, got %d", len(highZones))
	}
	if len(highZones) > 0 && !highZones[0].IsHigh {
		t.Error("high zone must retain IsHigh=true")
	}
	if len(highZones) > 0 && highZones[0].Strength != 2 {
		t.Errorf("merged strength should be 2, got %d", highZones[0].Strength)
	}
	if len(highZones) > 0 && !reflect.DeepEqual(highZones[0].SourcePivotIndexes, []int{5, 15}) {
		t.Errorf("high zone source pivots should be sorted by index, got %v", highZones[0].SourcePivotIndexes)
	}
	if len(highZones) > 0 && len(highZones[0].Pivots) == 2 && (highZones[0].Pivots[0].Index != 5 || highZones[0].Pivots[1].Index != 15) {
		t.Errorf("high zone pivot diagnostics should be sorted by index, got %+v", highZones[0].Pivots)
	}

	if len(lowZones) != 1 {
		t.Errorf("expected 1 low zone, got %d", len(lowZones))
	}
	if len(lowZones) > 0 && lowZones[0].IsHigh {
		t.Error("low zone must retain IsHigh=false")
	}
	if len(lowZones) > 0 && !reflect.DeepEqual(lowZones[0].SourcePivotIndexes, []int{10, 20}) {
		t.Errorf("low zone source pivots should be sorted by index, got %v", lowZones[0].SourcePivotIndexes)
	}
}
