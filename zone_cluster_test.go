package sr

import (
	"math"
	"reflect"
	"testing"
)

func TestClusterPriceSortedPivotsMatchesReference(t *testing.T) {
	tests := []struct {
		name   string
		pivots []srPivot
	}{
		{name: "empty"},
		{name: "one large cluster", pivots: testPriceSortedPivotsOneCluster(64)},
		{name: "distinct clusters", pivots: testPriceSortedPivotsDistinct(64)},
		{
			name: "mixed cluster widths",
			pivots: []srPivot{
				{Index: 0, Price: 100.00, MergeWidth: 0.30},
				{Index: 1, Price: 100.10, MergeWidth: 0.05},
				{Index: 2, Price: 100.20, MergeWidth: 0.50},
				{Index: 3, Price: 100.55, MergeWidth: 0.05},
				{Index: 4, Price: 101.50, MergeWidth: 0.20},
				{Index: 5, Price: 101.60, MergeWidth: 0.40},
				{Index: 6, Price: 103.00, MergeWidth: 0.10},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			want := referenceClusterPriceSortedPivots(tt.pivots)
			got := clusterPriceSortedPivots(tt.pivots)
			if !reflect.DeepEqual(got, want) {
				t.Fatalf("cluster mismatch\n got: %#v\nwant: %#v", got, want)
			}
		})
	}
}

func TestClusterPriceSortedPivotsPreservesPriceOrder(t *testing.T) {
	pivots := testPriceSortedPivotsOneCluster(64)
	clusters := clusterPriceSortedPivots(pivots)

	flattened := make([]srPivot, 0, len(pivots))
	for _, cluster := range clusters {
		for i := 1; i < len(cluster); i++ {
			if cluster[i].Price < cluster[i-1].Price {
				t.Fatalf("cluster is not price-sorted at %d: %v < %v", i, cluster[i].Price, cluster[i-1].Price)
			}
		}
		flattened = append(flattened, cluster...)
	}
	if !reflect.DeepEqual(flattened, pivots) {
		t.Fatalf("clustering reordered pivots")
	}
}

func referenceClusterPriceSortedPivots(sorted []srPivot) [][]srPivot {
	if len(sorted) == 0 {
		return nil
	}

	clusters := [][]srPivot{{sorted[0]}}
	for _, p := range sorted[1:] {
		last := &clusters[len(clusters)-1]
		clusterMedianPrice := medianPivotPrice(*last)
		clusterMedianWidth := medianPivotWidth(*last)
		threshold := math.Max(clusterMedianWidth, p.MergeWidth)
		if math.Abs(p.Price-clusterMedianPrice) <= threshold {
			*last = append(*last, p)
			continue
		}
		clusters = append(clusters, []srPivot{p})
	}
	return clusters
}

func testPriceSortedPivotsOneCluster(n int) []srPivot {
	pivots := make([]srPivot, n)
	for i := range pivots {
		pivots[i] = srPivot{
			Index:      i,
			Price:      100 + float64(i)*0.00001,
			MergeWidth: 1 + float64(i%17)*0.01,
		}
	}
	return pivots
}

func testPriceSortedPivotsDistinct(n int) []srPivot {
	pivots := make([]srPivot, n)
	for i := range pivots {
		pivots[i] = srPivot{
			Index:      i,
			Price:      float64(i) * 10,
			MergeWidth: 0.1,
		}
	}
	return pivots
}
