package sr

import "testing"

func TestPivotMergeWidth_NegativePriceUsesMagnitudeFloor(t *testing.T) {
	got := pivotMergeWidth(srPivot{Price: -100, ATRSnapshot: 0})
	if got != 0.1 {
		t.Fatalf("expected magnitude-based price floor 0.1, got %v", got)
	}
}

func TestDetectZoneProximity_NegativePriceZeroWidthUsesMagnitudeFallback(t *testing.T) {
	zones := []Level{{Price: -100.1, Top: -100.1, Bottom: -100.1, Strength: 1, IsHigh: false}}

	nearSup, nearRes, nearestSup, _, supDist, _, _, _, _, _ := detectZoneProximity(zones, -100)
	if !nearSup || nearRes {
		t.Fatalf("expected only nearby support for negative zero-width zone, got support=%t resistance=%t", nearSup, nearRes)
	}
	if nearestSup != -100.1 || supDist != 0.1 {
		t.Fatalf("unexpected nearest support result: price=%v distance=%v", nearestSup, supDist)
	}
}
