package sr

import (
	"testing"
	"time"
)

func TestBuildZones_FinalGeometryContainsAllMemberPivots(t *testing.T) {
	candles := makeFlatCandles(40, 100, time.Date(2024, 4, 5, 0, 0, 0, 0, time.UTC))
	pivots := []srPivot{
		{Index: 10, ConfirmedAtIndex: 14, Time: candles[10].CloseTime, Price: 100.0, IsHigh: false, Timeframe: "5m", MergeWidth: 0.1, BounceATR: 0.5},
		{Index: 20, ConfirmedAtIndex: 24, Time: candles[20].CloseTime, Price: 101.9, IsHigh: false, Timeframe: "5m", MergeWidth: 2.0, BounceATR: 0.6},
	}

	zones := buildZones(pivots, candles, 30)
	if len(zones) != 1 {
		t.Fatalf("expected pivots to cluster into 1 zone, got %d", len(zones))
	}

	zone := zones[0]
	if zone.Strength != len(pivots) {
		t.Fatalf("zone strength: want %d, got %d", len(pivots), zone.Strength)
	}
	if len(zone.SourcePivotIndexes) != len(pivots) {
		t.Fatalf("source pivot indexes: want %d entries, got %d", len(pivots), len(zone.SourcePivotIndexes))
	}
	if len(zone.Pivots) != len(pivots) {
		t.Fatalf("pivot metadata: want %d entries, got %d", len(pivots), len(zone.Pivots))
	}

	for i, want := range pivots {
		if zone.SourcePivotIndexes[i] != want.Index {
			t.Errorf("source pivot index %d: want %d, got %d", i, want.Index, zone.SourcePivotIndexes[i])
		}
		got := zone.Pivots[i]
		if got.Index != want.Index || got.Price != want.Price {
			t.Errorf("pivot metadata %d: want index=%d price=%v, got index=%d price=%v", i, want.Index, want.Price, got.Index, got.Price)
		}
	}

	for _, p := range zone.Pivots {
		if p.Price < zone.Bottom || p.Price > zone.Top {
			t.Errorf("pivot price %v is outside zone [%v, %v]", p.Price, zone.Bottom, zone.Top)
		}
	}
}
