package sr

func scoreZone(zone Level, members []srPivot, candles []Candle, scanLen int) float64 {
	if scanLen < 1 {
		scanLen = 1
	}

	touchScore := float64(zone.Strength) * 1.5
	bounceScore := medianPivotBounceATR(members) * 2.0

	lastIndexInPrefix := len(candles) - 1
	ageBars := lastIndexInPrefix - zone.LastTouchIndex
	freshnessScore := 2.0 * (1 - float64(ageBars)/float64(scanLen))
	if freshnessScore < 0 {
		freshnessScore = 0
	}

	zoneEstablishedAt := members[0].ConfirmedAtIndex
	for _, p := range members[1:] {
		if p.ConfirmedAtIndex > zoneEstablishedAt {
			zoneEstablishedAt = p.ConfirmedAtIndex
		}
	}

	falseBreakPenalty := 1.5 * float64(countFalseBreaks(zone, candles, zoneEstablishedAt))
	return touchScore + bounceScore + freshnessScore - falseBreakPenalty
}

// countFalseBreaks scans only the bars available in the current prefix and only
// after the zone's current membership could have existed.
func countFalseBreaks(zone Level, candles []Candle, zoneEstablishedAt int) int {
	start := zoneEstablishedAt + 1
	if start < 0 {
		start = 0
	}
	if start >= len(candles) {
		return 0
	}

	count := 0
	for i := start; i < len(candles); i++ {
		if zone.IsHigh {
			if candles[i].Close <= zone.Top {
				continue
			}
			if reentry := findZoneReentry(candles, i, zone.Top, true); reentry >= 0 {
				count++
				i = reentry
			}
			continue
		}

		if candles[i].Close >= zone.Bottom {
			continue
		}
		if reentry := findZoneReentry(candles, i, zone.Bottom, false); reentry >= 0 {
			count++
			i = reentry
		}
	}
	return count
}

func findZoneReentry(candles []Candle, breakoutIndex int, boundary float64, isHigh bool) int {
	end := breakoutIndex + pivotWindow
	if end >= len(candles) {
		end = len(candles) - 1
	}
	for j := breakoutIndex + 1; j <= end; j++ {
		if isHigh {
			if candles[j].Close <= boundary {
				return j
			}
			continue
		}
		if candles[j].Close >= boundary {
			return j
		}
	}
	return -1
}
