package sr

import "time"

// srPivot holds per-swing data captured at pivot confirmation time. Keep its
// fields identical to PivotInfo; pivotInfo relies on direct conversion.
type srPivot struct {
	Index             int
	ConfirmedAtIndex  int
	Time              time.Time
	Price             float64
	IsHigh            bool
	Timeframe         string
	ATRSnapshot       float64
	AvgVolumeSnapshot float64
	Volume            float64
	VolumeRatio       float64
	MergeWidth        float64
	BounceATR         float64
}

func findPivotHighs(candles []Candle, timeframe string, lookback int) []srPivot {
	return findPivots(candles, timeframe, lookback, true)
}

func findPivotLows(candles []Candle, timeframe string, lookback int) []srPivot {
	return findPivots(candles, timeframe, lookback, false)
}

func findPivots(candles []Candle, timeframe string, lookback int, isHigh bool) []srPivot {
	window := newSRLookbackWindow(len(candles), lookback)
	if window.PivotEnd <= window.PivotStart {
		return nil
	}

	var pivots []srPivot
	for i := window.PivotStart; i < window.PivotEnd; i++ {
		if !isConfirmedPivot(candles, i, isHigh) {
			continue
		}

		confirmedAtIndex := i + pivotWindow
		snapshotCandles := candles[:confirmedAtIndex+1]
		atrSnapshot := computeATR(snapshotCandles, rsiPeriod)
		avgVolumeSnapshot := computeAvgVolume(snapshotCandles, avgVolPeriod, len(snapshotCandles))
		price := candles[i].Low
		if isHigh {
			price = candles[i].High
		}

		volumeRatio := 0.0
		if avgVolumeSnapshot > 0 {
			volumeRatio = candles[i].Volume / avgVolumeSnapshot
		}

		pivot := srPivot{
			Index:             i,
			ConfirmedAtIndex:  confirmedAtIndex,
			Time:              candles[i].CloseTime,
			Price:             price,
			IsHigh:            isHigh,
			Timeframe:         timeframe,
			ATRSnapshot:       atrSnapshot,
			AvgVolumeSnapshot: avgVolumeSnapshot,
			Volume:            candles[i].Volume,
			VolumeRatio:       volumeRatio,
		}
		pivot.MergeWidth = pivotMergeWidth(pivot)
		pivot.BounceATR = pivotBounceATR(candles, pivot)

		pivots = append(pivots, pivot)
	}
	return pivots
}

func isConfirmedPivot(candles []Candle, idx int, isHigh bool) bool {
	for j := idx - pivotWindow; j <= idx+pivotWindow; j++ {
		if j == idx {
			continue
		}
		if isHigh {
			if candles[j].High >= candles[idx].High {
				return false
			}
			continue
		}
		if candles[j].Low <= candles[idx].Low {
			return false
		}
	}
	return true
}

func pivotMergeWidth(p srPivot) float64 {
	width := p.ATRSnapshot * 0.25
	priceFloor := p.Price * 0.001
	if width < priceFloor {
		width = priceFloor
	}
	return width
}

func pivotBounceATR(candles []Candle, p srPivot) float64 {
	if p.ATRSnapshot <= 0 {
		return 0
	}

	best := 0.0
	end := p.Index + pivotWindow
	if end >= len(candles) {
		end = len(candles) - 1
	}
	for j := p.Index + 1; j <= end; j++ {
		if p.IsHigh {
			if move := p.Price - candles[j].Low; move > best {
				best = move
			}
			continue
		}
		if move := candles[j].High - p.Price; move > best {
			best = move
		}
	}

	return best / p.ATRSnapshot
}

func pivotInfo(p srPivot) PivotInfo {
	return PivotInfo(p)
}
