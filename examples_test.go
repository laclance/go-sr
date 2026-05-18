package sr

import (
	"fmt"
	"time"
)

func ExampleCompute_zone() {
	levels, err := Compute(buildSRCategoryCandles(), Options{
		Timeframe: "5m",
		Lookback:  120,
		Mode:      ModeZones,
	})
	if err != nil {
		fmt.Println(err)
		return
	}

	fmt.Println(levels.Timeframe, len(levels.Levels) >= 2, levels.NearestSupport > 0, levels.NearestResistance > 0)
	// Output: 5m true true true
}

func ExampleCompute_legacy() {
	levels, err := Compute(buildSRCategoryCandles(), Options{
		Timeframe: "5m",
		Lookback:  120,
		Mode:      ModeLegacy,
		Tolerance: 0.002,
	})
	if err != nil {
		fmt.Println(err)
		return
	}

	fmt.Println(levels.Timeframe, len(levels.Levels) >= 2, levels.NearestSupport > 0, levels.NearestResistance > 0)
	// Output: 5m true true true
}

func ExampleAggregateCandlesToTimeframe() {
	start := time.Date(2024, 4, 1, 0, 0, 0, 0, time.UTC)
	candles := make5mCandles(start, []struct {
		open   float64
		high   float64
		low    float64
		close  float64
		volume float64
	}{
		{100, 101, 99, 100.5, 10},
		{100.5, 102, 100, 101.5, 20},
		{101.5, 103, 101, 102.5, 30},
	})

	agg := AggregateCandlesToTimeframe(candles, "5m", "15m")
	fmt.Println(len(agg), agg[0].Open, agg[0].Close, agg[0].Volume)
	// Output: 1 100 102.5 60
}
