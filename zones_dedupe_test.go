package sr

import "testing"

func TestDedupeZonesBySide_RechecksReplacementAgainstEarlierZones(t *testing.T) {
	zones := []Level{
		{
			Price:              100,
			Bottom:             98,
			Top:                102,
			Strength:           2,
			Score:              10,
			IsHigh:             false,
			SourcePivotIndexes: []int{1, 2},
		},
		{
			Price:              103.25,
			Bottom:             102.5,
			Top:                104,
			Strength:           2,
			Score:              1,
			IsHigh:             false,
			SourcePivotIndexes: []int{3, 4},
		},
		{
			Price:              103.375,
			Bottom:             101.75,
			Top:                105,
			Strength:           2,
			Score:              20,
			IsHigh:             false,
			SourcePivotIndexes: []int{5, 6},
		},
	}

	got := dedupeZonesBySide(zones)
	if len(got) != 1 {
		t.Fatalf("expected replacement chain to leave one zone, got %d: %+v", len(got), got)
	}
	if got[0].Price != 103.375 || got[0].Score != 20 {
		t.Fatalf("unexpected dedupe winner: %+v", got[0])
	}

	for i := 0; i < len(got); i++ {
		for j := i + 1; j < len(got); j++ {
			if zonesOverlap(got[i], got[j]) {
				t.Fatalf("dedupe retained overlapping zones: %+v and %+v", got[i], got[j])
			}
		}
	}
}
