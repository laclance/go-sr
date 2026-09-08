package sr

import (
	"math"
	"sort"
)

// clusterPriceSortedPivots clusters pivots in ascending price order. buildZones
// establishes that ordering before calling this helper, so each cluster remains
// price-sorted as members are appended.
func clusterPriceSortedPivots(sorted []srPivot) [][]srPivot {
	if len(sorted) == 0 {
		return nil
	}

	clusters := [][]srPivot{{sorted[0]}}
	sortedWidths := make([]float64, 1, len(sorted))
	sortedWidths[0] = sorted[0].MergeWidth

	for _, p := range sorted[1:] {
		last := &clusters[len(clusters)-1]
		clusterMedianPrice := medianPriceSortedPivots(*last)
		clusterMedianWidth := medianSortedFloat64(sortedWidths)
		threshold := math.Max(clusterMedianWidth, p.MergeWidth)
		if math.Abs(p.Price-clusterMedianPrice) <= threshold {
			*last = append(*last, p)
			sortedWidths = insertSortedFloat64(sortedWidths, p.MergeWidth)
			continue
		}

		clusters = append(clusters, []srPivot{p})
		sortedWidths = sortedWidths[:1]
		sortedWidths[0] = p.MergeWidth
	}

	return clusters
}

func medianPriceSortedPivots(pivots []srPivot) float64 {
	if len(pivots) == 0 {
		return 0
	}
	mid := len(pivots) / 2
	if len(pivots)%2 == 1 {
		return pivots[mid].Price
	}
	return (pivots[mid-1].Price + pivots[mid].Price) / 2
}

func medianSortedFloat64(values []float64) float64 {
	if len(values) == 0 {
		return 0
	}
	mid := len(values) / 2
	if len(values)%2 == 1 {
		return values[mid]
	}
	return (values[mid-1] + values[mid]) / 2
}

func insertSortedFloat64(values []float64, value float64) []float64 {
	index := sort.SearchFloat64s(values, value)
	values = append(values, 0)
	copy(values[index+1:], values[index:])
	values[index] = value
	return values
}
