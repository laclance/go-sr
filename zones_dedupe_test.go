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

func TestFilterZones_PreservesNestedSliceIndependenceFromRawZones(t *testing.T) {
	raw := []Level{
		{
			Price:              100,
			Bottom:             99,
			Top:                101,
			Strength:           2,
			Score:              10,
			IsHigh:             false,
			SourcePivotIndexes: []int{1, 2},
			Pivots:             []PivotInfo{{Index: 1}, {Index: 2}},
		},
		{
			Price:              110,
			Bottom:             109,
			Top:                111,
			Strength:           2,
			Score:              9,
			IsHigh:             false,
			SourcePivotIndexes: []int{3, 4},
			Pivots:             []PivotInfo{{Index: 3}, {Index: 4}},
		},
	}

	filtered := filterZones(raw, 2)
	if len(filtered) != 2 {
		t.Fatalf("expected two filtered zones, got %d", len(filtered))
	}

	filtered[0].SourcePivotIndexes[0] = 999
	filtered[0].Pivots[0].Index = 999

	if raw[0].SourcePivotIndexes[0] != 1 {
		t.Fatalf("filtered SourcePivotIndexes aliases raw zone: %+v", raw[0].SourcePivotIndexes)
	}
	if raw[0].Pivots[0].Index != 1 {
		t.Fatalf("filtered Pivots aliases raw zone: %+v", raw[0].Pivots)
	}
	if filtered[1].SourcePivotIndexes[0] != 3 || filtered[1].Pivots[0].Index != 3 {
		t.Fatalf("filtered zones alias each other: %+v", filtered)
	}

	raw[1].SourcePivotIndexes[0] = 777
	raw[1].Pivots[0].Index = 777
	if filtered[1].SourcePivotIndexes[0] != 3 || filtered[1].Pivots[0].Index != 3 {
		t.Fatalf("raw zone mutation leaked into filtered result: %+v", filtered[1])
	}
}
