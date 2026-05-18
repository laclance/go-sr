package sr

import "math"

const (
	rsiPeriod    = 14
	avgVolPeriod = 20
)

func computeATR(candles []Candle, period int) float64 {
	if len(candles) < period+1 {
		return 0
	}
	start := len(candles) - period
	sum := 0.0
	for i := start; i < len(candles); i++ {
		high := candles[i].High
		low := candles[i].Low
		prevClose := candles[i-1].Close
		tr := high - low
		if v := math.Abs(high - prevClose); v > tr {
			tr = v
		}
		if v := math.Abs(low - prevClose); v > tr {
			tr = v
		}
		sum += tr
	}
	return sum / float64(period)
}

func computeAvgVolume(candles []Candle, period, n int) float64 {
	start := n - period - 1
	if start < 0 {
		start = 0
	}
	end := n - 1
	if end <= start {
		return 0
	}
	sum := 0.0
	count := 0
	for i := start; i < end; i++ {
		sum += candles[i].Volume
		count++
	}
	return sum / float64(count)
}
