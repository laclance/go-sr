package sr

import (
	"reflect"
	"testing"
)

func TestCompute_LegacyNonPositiveLookbackUsesFullHistory(t *testing.T) {
	candles := buildSRCategoryCandles()
	full := computeLegacy(candles, "5m", len(candles), 0.002)
	if len(full.Levels) == 0 {
		t.Fatal("expected deterministic fixture to produce legacy levels")
	}

	for _, lookback := range []int{0, -1, -50} {
		got := computeLegacy(candles, "5m", lookback, 0.002)
		if !reflect.DeepEqual(got, full) {
			t.Fatalf("legacy lookback %d should match full-history computation\ngot:  %+v\nwant: %+v", lookback, got, full)
		}
	}
}

func TestCompute_LegacyPositiveLookbackRemainsWindowed(t *testing.T) {
	candles := buildSRCategoryCandles()
	const lookback = 80

	got := computeLegacy(candles, "5m", lookback, 0.002)
	if len(got.Levels) == 0 {
		t.Fatal("expected deterministic fixture to produce legacy levels")
	}

	// Preserve the legacy positive-lookback contract: scan the final lookback
	// pivot candidates while retaining the preceding pivot context.
	start := len(candles) - lookback - legacyPivotWindow
	window := candles[start:]
	want := computeLegacy(window, "5m", len(window), 0.002)
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("positive legacy lookback changed window semantics\ngot:  %+v\nwant: %+v", got, want)
	}
}

func TestCompute_ZoneNonPositiveLookbackUsesFullHistory(t *testing.T) {
	candles := buildSRCategoryCandles()
	full := computeZone(candles, "5m", len(candles))
	if len(full.Levels) == 0 {
		t.Fatal("expected deterministic fixture to produce zone levels")
	}

	for _, lookback := range []int{0, -1, -50} {
		got := computeZone(candles, "5m", lookback)
		if !reflect.DeepEqual(got, full) {
			t.Fatalf("zone lookback %d should match full-history computation\ngot:  %+v\nwant: %+v", lookback, got, full)
		}
	}
}

func TestModeHelpers_UnknownModeReturnsZero(t *testing.T) {
	invalid := Mode("invalid")
	if got := WarmupCandles(50, invalid); got != 0 {
		t.Fatalf("invalid-mode warmup: got %d want 0", got)
	}
	if got := RequiredKlineLimit("5m", "1h", 50, invalid); got != 0 {
		t.Fatalf("invalid-mode required limit: got %d want 0", got)
	}
}
