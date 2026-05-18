package sr

import (
	"reflect"
	"testing"
	"time"
)

func TestCandleGeometryMethods(t *testing.T) {
	bull := Candle{Open: 100, High: 112, Low: 96, Close: 108}
	if bull.Body() != 8 {
		t.Fatalf("bull body: got %.2f want 8", bull.Body())
	}
	if bull.Range() != 16 {
		t.Fatalf("bull range: got %.2f want 16", bull.Range())
	}
	if !bull.IsBullish() || bull.IsBearish() {
		t.Fatalf("bullish helpers inconsistent: %+v", bull)
	}
	if bull.UpperWick() != 4 {
		t.Fatalf("bull upper wick: got %.2f want 4", bull.UpperWick())
	}
	if bull.LowerWick() != 4 {
		t.Fatalf("bull lower wick: got %.2f want 4", bull.LowerWick())
	}
	if bull.MidPoint() != 104 {
		t.Fatalf("bull midpoint: got %.2f want 104", bull.MidPoint())
	}

	bear := Candle{Open: 108, High: 112, Low: 96, Close: 100}
	if bear.Body() != 8 {
		t.Fatalf("bear body: got %.2f want 8", bear.Body())
	}
	if bear.IsBullish() || !bear.IsBearish() {
		t.Fatalf("bearish helpers inconsistent: %+v", bear)
	}
	if bear.UpperWick() != 4 {
		t.Fatalf("bear upper wick: got %.2f want 4", bear.UpperWick())
	}
	if bear.LowerWick() != 4 {
		t.Fatalf("bear lower wick: got %.2f want 4", bear.LowerWick())
	}
}

func TestNewSRLookbackWindow_ClampsBounds(t *testing.T) {
	full := newSRLookbackWindow(10, -1)
	if full.Start != 0 || full.PivotStart != pivotWindow || full.PivotEnd != 6 || full.ScanLen != 10 {
		t.Fatalf("unexpected full lookback window: %+v", full)
	}

	short := newSRLookbackWindow(5, 1)
	if short.Start != 4 || short.PivotStart != 4 || short.PivotEnd != 4 || short.ScanLen != 1 {
		t.Fatalf("unexpected short lookback window: %+v", short)
	}
}

func TestComputeAvgVolume_WindowSelection(t *testing.T) {
	candles := []Candle{
		{Volume: 10},
		{Volume: 20},
		{Volume: 30},
		{Volume: 40},
	}

	if got := computeAvgVolume(candles, 2, len(candles)); got != 25 {
		t.Fatalf("avg volume over trailing history: got %.2f want 25", got)
	}
	if got := computeAvgVolume(candles[:1], 20, 1); got != 0 {
		t.Fatalf("expected zero avg volume for insufficient history, got %.2f", got)
	}
}

func TestCompareIntSlices(t *testing.T) {
	cases := []struct {
		name string
		a    []int
		b    []int
		want int
	}{
		{name: "equal", a: []int{1, 2}, b: []int{1, 2}, want: 0},
		{name: "left smaller", a: []int{1, 2}, b: []int{1, 3}, want: -1},
		{name: "left larger", a: []int{2}, b: []int{1, 9}, want: 1},
		{name: "shorter prefix", a: []int{1}, b: []int{1, 2}, want: -1},
		{name: "longer prefix", a: []int{1, 2, 3}, b: []int{1, 2}, want: 1},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := compareIntSlices(tc.a, tc.b); got != tc.want {
				t.Fatalf("compareIntSlices(%v, %v): got %d want %d", tc.a, tc.b, got, tc.want)
			}
		})
	}
}

func TestPreferZoneForDedup_TieBreaks(t *testing.T) {
	base := Level{
		Price:              100,
		Top:                101,
		Bottom:             99,
		Strength:           2,
		Score:              3,
		IsHigh:             false,
		LastTouchIndex:     10,
		SourcePivotIndexes: []int{2, 4},
	}

	if !preferZoneForDedup(Level{Score: 4}, base) {
		t.Fatal("higher score should win dedup tie-break")
	}
	if !preferZoneForDedup(Level{Score: 3, Strength: 3}, base) {
		t.Fatal("higher strength should win when score is tied")
	}
	if !preferZoneForDedup(Level{Score: 3, Strength: 2, LastTouchIndex: 11}, base) {
		t.Fatal("more recent touch should win when score and strength are tied")
	}
	if !preferZoneForDedup(Level{Score: 3, Strength: 2, LastTouchIndex: 10, Price: 99}, base) {
		t.Fatal("lower price should win when prior fields are tied")
	}
	if !preferZoneForDedup(Level{Score: 3, Strength: 2, LastTouchIndex: 10, Price: 100, SourcePivotIndexes: []int{1, 9}}, base) {
		t.Fatal("lexicographically smaller pivot index slice should win final tie-break")
	}
}

func TestSortSRLevels_TieBreaks(t *testing.T) {
	levels := []Level{
		{Price: 100, IsHigh: true, LastTouchIndex: 10, Strength: 2, Timeframe: "1h", SourcePivotIndexes: []int{2}},
		{Price: 100, IsHigh: false, LastTouchIndex: 10, Strength: 2, Timeframe: "1h", SourcePivotIndexes: []int{2}},
		{Price: 100, IsHigh: false, LastTouchIndex: 8, Strength: 2, Timeframe: "1h", SourcePivotIndexes: []int{2}},
		{Price: 100, IsHigh: false, LastTouchIndex: 8, Strength: 1, Timeframe: "15m", SourcePivotIndexes: []int{3}},
		{Price: 100, IsHigh: false, LastTouchIndex: 8, Strength: 1, Timeframe: "15m", SourcePivotIndexes: []int{1}},
	}

	sortSRLevels(levels)

	want := []Level{
		{Price: 100, IsHigh: false, LastTouchIndex: 8, Strength: 1, Timeframe: "15m", SourcePivotIndexes: []int{1}},
		{Price: 100, IsHigh: false, LastTouchIndex: 8, Strength: 1, Timeframe: "15m", SourcePivotIndexes: []int{3}},
		{Price: 100, IsHigh: false, LastTouchIndex: 8, Strength: 2, Timeframe: "1h", SourcePivotIndexes: []int{2}},
		{Price: 100, IsHigh: false, LastTouchIndex: 10, Strength: 2, Timeframe: "1h", SourcePivotIndexes: []int{2}},
		{Price: 100, IsHigh: true, LastTouchIndex: 10, Strength: 2, Timeframe: "1h", SourcePivotIndexes: []int{2}},
	}

	if !reflect.DeepEqual(levels, want) {
		t.Fatalf("unexpected sort order:\n got: %+v\nwant: %+v", levels, want)
	}
}

func TestDedupeZonesBySide_PrefersBestOverlappingZone(t *testing.T) {
	zones := []Level{
		{
			Price:              100,
			Top:                101,
			Bottom:             99,
			Strength:           2,
			Score:              3,
			IsHigh:             false,
			LastTouchIndex:     10,
			SourcePivotIndexes: []int{1, 3},
			Pivots:             []PivotInfo{{Index: 1}, {Index: 3}},
		},
		{
			Price:              100.2,
			Top:                101.2,
			Bottom:             99.2,
			Strength:           3,
			Score:              5,
			IsHigh:             false,
			LastTouchIndex:     12,
			SourcePivotIndexes: []int{2, 4},
			Pivots:             []PivotInfo{{Index: 2}, {Index: 4}},
		},
	}

	got := dedupeZonesBySide(zones)
	if len(got) != 1 {
		t.Fatalf("expected one deduped zone, got %d", len(got))
	}
	if got[0].Score != 5 || got[0].Strength != 3 || got[0].LastTouchIndex != 12 {
		t.Fatalf("unexpected deduped winner: %+v", got[0])
	}
	if !reflect.DeepEqual(got[0].SourcePivotIndexes, []int{2, 4}) {
		t.Fatalf("expected clone to preserve source pivots, got %v", got[0].SourcePivotIndexes)
	}
}

func TestCloneSRLevel_DeepCopiesSlices(t *testing.T) {
	level := Level{
		Price:              100,
		SourcePivotIndexes: []int{1, 2},
		Pivots:             []PivotInfo{{Index: 1, Time: time.Unix(0, 0)}, {Index: 2, Time: time.Unix(1, 0)}},
	}

	cloned := cloneSRLevel(level)
	cloned.SourcePivotIndexes[0] = 99
	cloned.Pivots[0].Index = 88

	if level.SourcePivotIndexes[0] != 1 {
		t.Fatalf("source pivot indexes were not deep-copied: %+v", level.SourcePivotIndexes)
	}
	if level.Pivots[0].Index != 1 {
		t.Fatalf("pivot diagnostics were not deep-copied: %+v", level.Pivots)
	}
}
