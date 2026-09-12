package main

import (
	"fmt"
	"math"
	"time"

	sr "github.com/laclance/go-sr"
)

func main() {
	levels := mustCompute(sr.Options{
		Timeframe:   "5m",
		Lookback:    120,
		Mode:        sr.ModeZones,
		MinStrength: 2,
	})

	fmt.Printf("levels=%d support=%.2f resistance=%.2f\n",
		len(levels.Levels),
		levels.NearestSupport,
		levels.NearestResistance,
	)
}

func mustCompute(opts sr.Options) sr.Levels {
	levels, err := sr.Compute(demoCandles(), opts)
	if err != nil {
		panic(err)
	}
	return levels
}

// demoCandles returns deterministic closed 5m candles with repeated swing lows
// near 49,000 and swing highs near 51,000. Replace this function with candles
// from your own exchange, broker, backtest fixture, or market-data pipeline.
func demoCandles() []sr.Candle {
	candles := make([]sr.Candle, 160)
	t := time.Date(2024, 2, 1, 0, 0, 0, 0, time.UTC)

	for i := range candles {
		base := 50000.0 + math.Sin(float64(i)/10.0)*60.0
		open := base - 20
		closePrice := base + 20
		if i%2 == 1 {
			open, closePrice = closePrice, open
		}

		high := math.Max(open, closePrice) + 70
		low := math.Min(open, closePrice) - 70

		switch i {
		case 20, 52, 84, 116:
			open = 50010
			closePrice = 50070
			high = 50130
			low = 49000
		case 36, 68, 100, 132:
			open = 49990
			closePrice = 49940
			high = 51000
			low = 49870
		case 159:
			open = 49990
			closePrice = 50010
			high = 50080
			low = 49920
		}

		candles[i] = sr.Candle{
			OpenTime:  t,
			CloseTime: t.Add(5 * time.Minute),
			Open:      open,
			High:      high,
			Low:       low,
			Close:     closePrice,
			Volume:    1000 + float64(i%5)*25,
		}
		t = t.Add(5 * time.Minute)
	}

	return candles
}
