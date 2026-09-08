package sr

import "testing"

func TestZoneProximity_ZeroPriceLevelUsesStrengthForPresence(t *testing.T) {
	zones := []Level{{
		Price:    0,
		Top:      1,
		Bottom:   -1,
		Strength: 2,
		Score:    5,
		IsHigh:   false,
	}}

	nearSup, _, nearestSup, _, supDist, _, supStr, _, supScore, _ := detectZoneProximity(zones, 0)
	if !nearSup {
		t.Fatal("expected zero-price support to be near")
	}
	if nearestSup != 0 || supDist != 0 {
		t.Fatalf("expected zero-price support at zero distance, got price=%v distance=%v", nearestSup, supDist)
	}
	if supStr != 2 || supScore != 5 {
		t.Fatalf("expected positive metadata to indicate a found zero-price support, got strength=%d score=%v", supStr, supScore)
	}
}

func TestLegacyProximity_ZeroPriceLevelUsesStrengthForPresence(t *testing.T) {
	levels := []Level{{Price: 0, Strength: 3}}

	nearSup, _, nearestSup, _, supDist, _, supStr, _ := legacyProximity(levels, 0, 0)
	if !nearSup {
		t.Fatal("expected zero-price support to be near")
	}
	if nearestSup != 0 || supDist != 0 {
		t.Fatalf("expected zero-price support at zero distance, got price=%v distance=%v", nearestSup, supDist)
	}
	if supStr != 3 {
		t.Fatalf("expected positive strength to indicate a found zero-price support, got %d", supStr)
	}
}
