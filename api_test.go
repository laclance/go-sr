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
