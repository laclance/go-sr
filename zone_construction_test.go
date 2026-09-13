package sr

import (
	"reflect"
	"testing"
)

func TestClusterPriceSortedPivotsOwnsStorage(t *testing.T) {
	pivots := []srPivot{
		{Index: 1, Price: 100, MergeWidth: 0.1},
		{Index: 2, Price: 200, MergeWidth: 0.1},
	}
	clusters := clusterPriceSortedPivots(pivots)
	if len(clusters) != 2 {
		t.Fatalf("expected two clusters, got %d", len(clusters))
	}

	clusters[0][0].Price = 101
	if pivots[0].Price != 100 {
		t.Fatalf("cluster mutation leaked into input: got input price %v", pivots[0].Price)
	}

	second := clusters[1][0]
	clusters[0] = append(clusters[0], srPivot{Index: 3, Price: 150, MergeWidth: 0.1})
	if clusters[1][0] != second {
		t.Fatalf("append to first cluster modified second cluster: got %+v want %+v", clusters[1][0], second)
	}
}

func TestBuildZonesUsesPriceMedianBeforeIndexOrdering(t *testing.T) {
	candles := make([]Candle, 30)
	for i := range candles {
		candles[i] = Candle{Open: 100, High: 101, Low: 99, Close: 100, Volume: 1000}
	}
	pivots := []srPivot{
		{Index: 15, ConfirmedAtIndex: 19, Price: 99, IsHigh: false, Timeframe: "5m", MergeWidth: 5},
		{Index: 5, ConfirmedAtIndex: 9, Price: 101, IsHigh: false, Timeframe: "5m", MergeWidth: 5},
		{Index: 10, ConfirmedAtIndex: 14, Price: 100, IsHigh: false, Timeframe: "5m", MergeWidth: 5},
	}

	zones := buildZones(pivots, candles, 0)
	if len(zones) != 1 {
		t.Fatalf("expected one zone, got %d", len(zones))
	}
	if zones[0].Price != 100 {
		t.Fatalf("zone center: got %v want 100", zones[0].Price)
	}

	wantIndexes := []int{5, 10, 15}
	if !reflect.DeepEqual(zones[0].SourcePivotIndexes, wantIndexes) {
		t.Fatalf("source pivot order: got %v want %v", zones[0].SourcePivotIndexes, wantIndexes)
	}
	for i, want := range wantIndexes {
		if zones[0].Pivots[i].Index != want {
			t.Fatalf("pivot metadata index %d: got %d want %d", i, zones[0].Pivots[i].Index, want)
		}
	}
}

func TestBuildZoneDoesNotMutateInput(t *testing.T) {
	candles := make([]Candle, 20)
	for i := range candles {
		candles[i] = Candle{Open: 100, High: 101, Low: 99, Close: 100, Volume: 1000}
	}
	cluster := []srPivot{
		{Index: 8, ConfirmedAtIndex: 12, Price: 101, IsHigh: false, Timeframe: "5m", MergeWidth: 2},
		{Index: 4, ConfirmedAtIndex: 8, Price: 99, IsHigh: false, Timeframe: "5m", MergeWidth: 2},
		{Index: 6, ConfirmedAtIndex: 10, Price: 100, IsHigh: false, Timeframe: "5m", MergeWidth: 2},
	}
	before := append([]srPivot(nil), cluster...)

	_ = buildZone(cluster, candles, len(candles))

	if !reflect.DeepEqual(cluster, before) {
		t.Fatalf("buildZone mutated input:\n got: %#v\nwant: %#v", cluster, before)
	}
}
