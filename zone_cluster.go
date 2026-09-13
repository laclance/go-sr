package sr

import (
	"math"
	"sort"
)

// clusterPriceSortedPivots clusters pivots in ascending price order. buildZones
// establishes that ordering before calling this helper, so each returned cluster
// remains price-sorted.
func clusterPriceSortedPivots(sorted []srPivot) [][]srPivot {
	if len(sorted) == 0 {
		return nil
	}

	// Keep cluster storage independent from the caller while allocating pivot
	// storage once. Full-slice expressions keep appends to one cluster from
	// overwriting the next cluster's backing storage.
	members := append([]srPivot(nil), sorted...)
	var clusters [][]srPivot
	clusterStart := 0

	sortedWidths := make([]float64, 1, len(members))
	sortedWidths[0] = members[0].MergeWidth

	for i := 1; i < len(members); i++ {
		p := members[i]
		current := members[clusterStart:i:i]
		clusterMedianPrice := medianPriceSortedPivots(current)
		clusterMedianWidth := medianSortedFloat64(sortedWidths)
		threshold := math.Max(clusterMedianWidth, p.MergeWidth)
		if math.Abs(p.Price-clusterMedianPrice) <= threshold {
			sortedWidths = insertSortedFloat64(sortedWidths, p.MergeWidth)
			continue
		}

		clusters = append(clusters, current)
		clusterStart = i
		sortedWidths = sortedWidths[:1]
		sortedWidths[0] = p.MergeWidth
	}

	end := len(members)
	clusters = append(clusters, members[clusterStart:end:end])
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
