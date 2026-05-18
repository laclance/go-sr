package sr

import (
	"math"
	"sort"
)

func buildZones(pivots []srPivot, candles []Candle, lookback int) []Level {
	if len(pivots) == 0 {
		return nil
	}

	window := newSRLookbackWindow(len(candles), lookback)
	sorted := append([]srPivot(nil), pivots...)
	sort.Slice(sorted, func(i, j int) bool {
		if sorted[i].Price != sorted[j].Price {
			return sorted[i].Price < sorted[j].Price
		}
		return sorted[i].Index < sorted[j].Index
	})

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

	zones := make([]Level, 0, len(clusters))
	for _, cluster := range clusters {
		zones = append(zones, buildZone(cluster, candles, window.ScanLen))
	}
	sortSRLevels(zones)
	return zones
}

func buildZone(cluster []srPivot, candles []Candle, scanLen int) Level {
	members := append([]srPivot(nil), cluster...)
	sort.Slice(members, func(i, j int) bool {
		if members[i].Index != members[j].Index {
			return members[i].Index < members[j].Index
		}
		if members[i].Price != members[j].Price {
			return members[i].Price < members[j].Price
		}
		return members[i].ConfirmedAtIndex < members[j].ConfirmedAtIndex
	})

	center := medianPivotPrice(members)
	width := medianPivotWidth(members)
	lastTouchIndex := members[len(members)-1].Index

	sourcePivotIndexes := make([]int, 0, len(members))
	pivotInfos := make([]PivotInfo, 0, len(members))
	for _, p := range members {
		sourcePivotIndexes = append(sourcePivotIndexes, p.Index)
		pivotInfos = append(pivotInfos, pivotInfo(p))
	}

	zone := Level{
		Price:              center,
		Top:                center + width/2,
		Bottom:             center - width/2,
		Strength:           len(members),
		IsHigh:             members[0].IsHigh,
		Timeframe:          members[0].Timeframe,
		LastTouchIndex:     lastTouchIndex,
		SourcePivotIndexes: sourcePivotIndexes,
		Pivots:             pivotInfos,
	}
	zone.Score = scoreZone(zone, members, candles, scanLen)
	return zone
}

// filterZones drops raw zones below the strength/score thresholds and dedups
// by overlap. minStrength <= 0 falls back to the historical default of 2 for
// back-compat with callers that don't set Options.MinStrength.
func filterZones(zones []Level, minStrength int) []Level {
	threshold := minStrength
	if threshold <= 0 {
		threshold = 2
	}
	var strong []Level
	for _, z := range zones {
		if z.Strength >= threshold && z.Score > 0 {
			strong = append(strong, cloneSRLevel(z))
		}
	}

	if len(strong) <= 1 {
		sortSRLevels(strong)
		return strong
	}

	var supports []Level
	var resistances []Level
	for _, z := range strong {
		if z.IsHigh {
			resistances = append(resistances, z)
			continue
		}
		supports = append(supports, z)
	}

	supports = dedupeZonesBySide(supports)
	resistances = dedupeZonesBySide(resistances)

	filtered := append(supports, resistances...)
	sortSRLevels(filtered)
	return filtered
}

func dedupeZonesBySide(zones []Level) []Level {
	if len(zones) <= 1 {
		sortSRLevels(zones)
		return zones
	}

	sortSRLevels(zones)

	deduped := []Level{cloneSRLevel(zones[0])}
	for _, z := range zones[1:] {
		last := &deduped[len(deduped)-1]
		if zonesOverlap(*last, z) {
			if preferZoneForDedup(z, *last) {
				*last = cloneSRLevel(z)
			}
			continue
		}
		deduped = append(deduped, cloneSRLevel(z))
	}
	return deduped
}

func zonesOverlap(a, b Level) bool {
	return a.IsHigh == b.IsHigh && b.Bottom <= a.Top
}

func preferZoneForDedup(candidate, current Level) bool {
	if candidate.Score != current.Score {
		return candidate.Score > current.Score
	}
	if candidate.Strength != current.Strength {
		return candidate.Strength > current.Strength
	}
	if candidate.LastTouchIndex != current.LastTouchIndex {
		return candidate.LastTouchIndex > current.LastTouchIndex
	}
	if candidate.Price != current.Price {
		return candidate.Price < current.Price
	}
	return compareIntSlices(candidate.SourcePivotIndexes, current.SourcePivotIndexes) < 0
}

func medianPivotPrice(pivots []srPivot) float64 {
	values := make([]float64, 0, len(pivots))
	for _, p := range pivots {
		values = append(values, p.Price)
	}
	return medianFloat64(values)
}

func medianPivotWidth(pivots []srPivot) float64 {
	values := make([]float64, 0, len(pivots))
	for _, p := range pivots {
		values = append(values, p.MergeWidth)
	}
	return medianFloat64(values)
}

func medianPivotBounceATR(pivots []srPivot) float64 {
	values := make([]float64, 0, len(pivots))
	for _, p := range pivots {
		values = append(values, p.BounceATR)
	}
	return medianFloat64(values)
}

func medianFloat64(values []float64) float64 {
	if len(values) == 0 {
		return 0
	}
	sorted := append([]float64(nil), values...)
	sort.Float64s(sorted)
	mid := len(sorted) / 2
	if len(sorted)%2 == 1 {
		return sorted[mid]
	}
	return (sorted[mid-1] + sorted[mid]) / 2
}

func sortSRLevels(levels []Level) {
	sort.Slice(levels, func(i, j int) bool {
		if levels[i].Price != levels[j].Price {
			return levels[i].Price < levels[j].Price
		}
		if levels[i].IsHigh != levels[j].IsHigh {
			return !levels[i].IsHigh && levels[j].IsHigh
		}
		if levels[i].LastTouchIndex != levels[j].LastTouchIndex {
			return levels[i].LastTouchIndex < levels[j].LastTouchIndex
		}
		if levels[i].Strength != levels[j].Strength {
			return levels[i].Strength < levels[j].Strength
		}
		if levels[i].Timeframe != levels[j].Timeframe {
			return levels[i].Timeframe < levels[j].Timeframe
		}
		return compareIntSlices(levels[i].SourcePivotIndexes, levels[j].SourcePivotIndexes) < 0
	})
}

func compareIntSlices(a, b []int) int {
	n := len(a)
	if len(b) < n {
		n = len(b)
	}
	for i := 0; i < n; i++ {
		if a[i] < b[i] {
			return -1
		}
		if a[i] > b[i] {
			return 1
		}
	}
	switch {
	case len(a) < len(b):
		return -1
	case len(a) > len(b):
		return 1
	default:
		return 0
	}
}

func cloneSRLevels(levels []Level) []Level {
	if len(levels) == 0 {
		return nil
	}
	cloned := make([]Level, 0, len(levels))
	for _, level := range levels {
		cloned = append(cloned, cloneSRLevel(level))
	}
	return cloned
}

func appendSRLevels(a, b []Level) []Level {
	levels := make([]Level, 0, len(a)+len(b))
	levels = append(levels, a...)
	levels = append(levels, b...)
	return levels
}

func cloneSRLevel(level Level) Level {
	cloned := level
	cloned.SourcePivotIndexes = append([]int(nil), level.SourcePivotIndexes...)
	cloned.Pivots = append([]PivotInfo(nil), level.Pivots...)
	return cloned
}
