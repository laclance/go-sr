package sr

import "testing"

var (
	benchmarkZoneClusters [][]srPivot
	benchmarkLegacyLevels []Level
)

func BenchmarkZoneClustering(b *testing.B) {
	cases := []struct {
		name   string
		pivots []srPivot
	}{
		{name: "one_large_cluster/bounded_120", pivots: testPriceSortedPivotsOneCluster(120)},
		{name: "one_large_cluster/all_history_2000", pivots: testPriceSortedPivotsOneCluster(2000)},
		{name: "distinct_clusters/bounded_120", pivots: testPriceSortedPivotsDistinct(120)},
		{name: "distinct_clusters/all_history_2000", pivots: testPriceSortedPivotsDistinct(2000)},
	}

	for _, tc := range cases {
		b.Run(tc.name, func(b *testing.B) {
			b.ReportAllocs()
			for i := 0; i < b.N; i++ {
				benchmarkZoneClusters = clusterPriceSortedPivots(tc.pivots)
			}
		})
	}
}

func BenchmarkLegacyClustering(b *testing.B) {
	cases := []struct {
		name string
		size int
	}{
		{name: "many_clusters/bounded_120", size: 120},
		{name: "many_clusters/all_history_5000", size: 5000},
	}

	for _, tc := range cases {
		prices := benchmarkDistinctLegacyPrices(tc.size)
		b.Run(tc.name, func(b *testing.B) {
			b.ReportAllocs()
			for i := 0; i < b.N; i++ {
				benchmarkLegacyLevels = clusterLevels(prices, true, 0.01, "1m")
			}
		})
	}
}

func benchmarkDistinctLegacyPrices(n int) []float64 {
	prices := make([]float64, n)
	for i := range prices {
		prices[i] = float64(i)
	}
	return prices
}
