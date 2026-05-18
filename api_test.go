package sr

import (
	"reflect"
	"testing"
	"time"
)

func TestEmptyLevels_ReturnsTimeframeOnly(t *testing.T) {
	got := EmptyLevels("15m")
	if got.Timeframe != "15m" {
		t.Fatalf("expected timeframe label to be preserved, got %q", got.Timeframe)
	}
	if len(got.Levels) != 0 || len(got.RawZones) != 0 {
		t.Fatalf("expected empty levels/raw zones, got %+v", got)
	}
	if got.NearSupport || got.NearResistance {
		t.Fatalf("expected zero-value proximity flags, got %+v", got)
	}
}

func TestCompute_ZoneShortInputReturnsEmptyLevels(t *testing.T) {
	candles := makeFlatCandles(8, 100.0, time.Date(2024, 4, 10, 0, 0, 0, 0, time.UTC))
	got := computeZone(candles, "5m", 50)
	if !reflect.DeepEqual(got, EmptyLevels("5m")) {
		t.Fatalf("expected zone mode short input to return empty bundle, got %+v", got)
	}
}

func TestCompute_ZoneLookbackZeroAtPivotBoundary(t *testing.T) {
	closeTime := time.Date(2024, 4, 10, 0, 0, 0, 0, time.UTC)
	short, err := Compute(makeFlatCandles(pivotWindow*2, 100.0, closeTime), Options{
		Timeframe: "5m",
		Lookback:  0,
		Mode:      ModeZones,
	})
	if err != nil {
		t.Fatalf("unexpected Compute error: %v", err)
	}
	if !reflect.DeepEqual(short, EmptyLevels("5m")) {
		t.Fatalf("expected below pivot boundary to return empty bundle, got %+v", short)
	}

	atBoundary := makeFlatCandles(pivotWindow*2+1, 100.0, closeTime)
	atBoundary[pivotWindow].High = 120
	atBoundary[pivotWindow].Low = 80
	got, err := Compute(atBoundary, Options{
		Timeframe:   "5m",
		Lookback:    0,
		Mode:        ModeZones,
		MinStrength: 1,
	})
	if err != nil {
		t.Fatalf("unexpected Compute error: %v", err)
	}
	if len(got.RawZones) != 2 {
		t.Fatalf("expected boundary input to scan the interior pivot, got %d raw zones: %+v", len(got.RawZones), got.RawZones)
	}
	if findZoneBySource(got.RawZones, []int{pivotWindow}, true) == nil {
		t.Fatalf("expected high pivot at boundary index %d, got %+v", pivotWindow, got.RawZones)
	}
	if findZoneBySource(got.RawZones, []int{pivotWindow}, false) == nil {
		t.Fatalf("expected low pivot at boundary index %d, got %+v", pivotWindow, got.RawZones)
	}
}

func TestCompute_UnknownModeReturnsError(t *testing.T) {
	got, err := Compute(makeFlatCandles(12, 100, time.Date(2024, 4, 20, 0, 0, 0, 0, time.UTC)), Options{
		Timeframe: "5m",
		Mode:      Mode("zones"),
	})
	if err == nil {
		t.Fatal("expected error for unknown mode")
	}
	if err.Error() != `sr: unknown Mode "zones"` {
		t.Fatalf("unexpected error: %v", err)
	}
	if !reflect.DeepEqual(got, EmptyLevels("5m")) {
		t.Fatalf("expected empty bundle on unknown mode, got %+v", got)
	}
}

func TestCompute_LegacyShortInputReturnsEmptyLevels(t *testing.T) {
	candles := makeFlatCandles(10, 100.0, time.Date(2024, 4, 11, 0, 0, 0, 0, time.UTC))
	got := computeLegacy(candles, "5m", 50, 0.002)
	if !reflect.DeepEqual(got, EmptyLevels("5m")) {
		t.Fatalf("expected legacy mode short input to return empty bundle, got %+v", got)
	}
}

func TestCompute_LegacyDefaultToleranceMatchesExplicitFallback(t *testing.T) {
	candles := buildSRCategoryCandles()
	implicit := computeLegacy(candles, "5m", 120, 0)
	explicit := computeLegacy(candles, "5m", 120, 0.002)
	if !reflect.DeepEqual(implicit, explicit) {
		t.Fatalf("expected implicit and explicit default legacy tolerance to match")
	}
}

func TestComputeSRLegacy_InternalFallbackAndWindowEdges(t *testing.T) {
	candles := buildSRCategoryCandles()
	implicit := computeSRLegacy(candles, "5m", 120, 0)
	explicit := computeSRLegacy(candles, "5m", 120, 0.002)
	if !reflect.DeepEqual(implicit, explicit) {
		t.Fatalf("expected internal legacy tolerance fallback to match explicit tolerance")
	}

	fullWindow := computeSRLegacy(makeFlatCandles(20, 100, time.Date(2024, 4, 18, 0, 0, 0, 0, time.UTC)), "5m", 100, 0.002)
	if fullWindow.Timeframe != "5m" {
		t.Fatalf("expected timeframe to survive full-window legacy calculation, got %+v", fullWindow)
	}

	empty := computeSRLegacy(makeFlatCandles(11, 100, time.Date(2024, 4, 19, 0, 0, 0, 0, time.UTC)), "5m", 1, 0.002)
	if !reflect.DeepEqual(empty, EmptyLevels("5m")) {
		t.Fatalf("expected empty levels when legacy lookback leaves no scan range, got %+v", empty)
	}
}
