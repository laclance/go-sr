package sr

import "math"

const legacyPivotWindow = 5

func computeSRLegacy(candles []Candle, timeframe string, lookback int, tolerance float64) Levels {
	n := len(candles)
	if n < legacyPivotWindow*2+1 {
		return EmptyLevels(timeframe)
	}
	if tolerance <= 0 {
		tolerance = 0.002
	}

	start := n - lookback
	if start < legacyPivotWindow {
		start = legacyPivotWindow
	}
	end := n - legacyPivotWindow
	if end <= start {
		return EmptyLevels(timeframe)
	}

	lastClose := candles[n-1].Close
	tolAbs := lastClose * tolerance

	var highs, lows []float64
	for i := start; i < end; i++ {
		if isSwingHigh(candles, i, legacyPivotWindow) {
			highs = append(highs, candles[i].High)
		}
		if isSwingLow(candles, i, legacyPivotWindow) {
			lows = append(lows, candles[i].Low)
		}
	}

	levels := clusterLevels(highs, true, tolAbs, timeframe)
	levels = append(levels, clusterLevels(lows, false, tolAbs, timeframe)...)
	sortSRLevels(levels)

	rawZones := cloneSRLevels(levels)

	nearSup, nearRes, nearestSup, nearestRes, supDist, resDist, supStr, resStr := legacyProximity(levels, lastClose, tolAbs)

	return Levels{
		Timeframe:                 timeframe,
		Levels:                    levels,
		RawZones:                  rawZones,
		NearSupport:               nearSup,
		NearResistance:            nearRes,
		NearestSupport:            nearestSup,
		NearestResistance:         nearestRes,
		NearestSupportDistance:    supDist,
		NearestResistanceDistance: resDist,
		NearestSupportStrength:    supStr,
		NearestResistanceStrength: resStr,
		NearestSupportScore:       0,
		NearestResistanceScore:    0,
	}
}

func isSwingHigh(candles []Candle, i, k int) bool {
	n := len(candles)
	for j := i - k; j <= i+k; j++ {
		if j == i || j < 0 || j >= n {
			continue
		}
		if candles[j].High >= candles[i].High {
			return false
		}
	}
	return true
}

func isSwingLow(candles []Candle, i, k int) bool {
	n := len(candles)
	for j := i - k; j <= i+k; j++ {
		if j == i || j < 0 || j >= n {
			continue
		}
		if candles[j].Low <= candles[i].Low {
			return false
		}
	}
	return true
}

func clusterLevels(prices []float64, isHigh bool, tolAbs float64, timeframe string) []Level {
	type entry struct {
		sum   float64
		count int
	}
	var entries []entry

	for _, p := range prices {
		merged := false
		for i := range entries {
			center := entries[i].sum / float64(entries[i].count)
			if math.Abs(p-center) <= tolAbs {
				entries[i].sum += p
				entries[i].count++
				merged = true
				break
			}
		}
		if !merged {
			entries = append(entries, entry{sum: p, count: 1})
		}
	}

	levels := make([]Level, 0, len(entries))
	for _, e := range entries {
		center := e.sum / float64(e.count)
		levels = append(levels, Level{
			Price:     center,
			Top:       center + tolAbs,
			Bottom:    center - tolAbs,
			Strength:  e.count,
			Score:     0,
			IsHigh:    isHigh,
			Timeframe: timeframe,
		})
	}
	return levels
}

func legacyProximity(levels []Level, price, tolAbs float64) (
	nearSup, nearRes bool,
	nearestSup, nearestRes float64,
	supDist, resDist float64,
	supStr, resStr int,
) {
	nearestSupDist := math.MaxFloat64
	nearestResDist := math.MaxFloat64

	for _, lvl := range levels {
		if lvl.Price <= price {
			dist := price - lvl.Price
			if dist < nearestSupDist {
				nearestSupDist = dist
				nearestSup = lvl.Price
				supDist = dist
				supStr = lvl.Strength
				nearSup = dist <= tolAbs
			}
		} else {
			dist := lvl.Price - price
			if dist < nearestResDist {
				nearestResDist = dist
				nearestRes = lvl.Price
				resDist = dist
				resStr = lvl.Strength
				nearRes = dist <= tolAbs
			}
		}
	}
	return
}
