package sr

import (
	"fmt"
	"math"
)

func isFiniteFloat(value float64) bool {
	return !math.IsNaN(value) && !math.IsInf(value, 0)
}

func candleValuesFinite(candle Candle) bool {
	return isFiniteFloat(candle.Open) &&
		isFiniteFloat(candle.High) &&
		isFiniteFloat(candle.Low) &&
		isFiniteFloat(candle.Close) &&
		isFiniteFloat(candle.Volume)
}

func validateFiniteLevels(levels Levels) error {
	if !isFiniteFloat(levels.NearestSupport) ||
		!isFiniteFloat(levels.NearestResistance) ||
		!isFiniteFloat(levels.NearestSupportDistance) ||
		!isFiniteFloat(levels.NearestResistanceDistance) ||
		!isFiniteFloat(levels.NearestSupportScore) ||
		!isFiniteFloat(levels.NearestResistanceScore) {
		return fmt.Errorf("sr: computed non-finite proximity result")
	}

	for _, level := range levels.Levels {
		if !levelValuesFinite(level) {
			return fmt.Errorf("sr: computed non-finite level")
		}
	}
	for _, level := range levels.RawZones {
		if !levelValuesFinite(level) {
			return fmt.Errorf("sr: computed non-finite raw zone")
		}
	}
	return nil
}

func levelValuesFinite(level Level) bool {
	if !isFiniteFloat(level.Price) ||
		!isFiniteFloat(level.Top) ||
		!isFiniteFloat(level.Bottom) ||
		!isFiniteFloat(level.Score) {
		return false
	}

	for _, pivot := range level.Pivots {
		if !isFiniteFloat(pivot.Price) ||
			!isFiniteFloat(pivot.ATRSnapshot) ||
			!isFiniteFloat(pivot.AvgVolumeSnapshot) ||
			!isFiniteFloat(pivot.Volume) ||
			!isFiniteFloat(pivot.VolumeRatio) ||
			!isFiniteFloat(pivot.MergeWidth) ||
			!isFiniteFloat(pivot.BounceATR) {
			return false
		}
	}
	return true
}
