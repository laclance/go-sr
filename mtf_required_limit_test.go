package sr

import (
	"fmt"
	"testing"
	"time"
)

func TestRequiredKlineLimit_PreservesWarmupAfterUTCAlignment(t *testing.T) {
	cases := []struct {
		base   string
		target string
	}{
		{base: "5m", target: "15m"},
		{base: "5m", target: "1h"},
		{base: "1h", target: "1d"},
	}
	modes := []Mode{ModeLegacy, ModeZones}
	const lookback = 2
	alignedStart := time.Date(2024, 4, 1, 0, 0, 0, 0, time.UTC)

	for _, tc := range cases {
		for _, mode := range modes {
			t.Run(fmt.Sprintf("%s_to_%s_%s", tc.base, tc.target, mode), func(t *testing.T) {
				baseDur := intervalDuration(tc.base)
				targetDur := intervalDuration(tc.target)
				ratio := int(targetDur / baseDur)
				limit := RequiredKlineLimit(tc.base, tc.target, lookback, mode)
				warmup := WarmupCandles(lookback, mode)

				for offset := 0; offset < ratio; offset++ {
					t.Run(fmt.Sprintf("offset_%d", offset), func(t *testing.T) {
						start := alignedStart.Add(time.Duration(offset) * baseDur)
						candles := make([]Candle, limit)
						for i := range candles {
							openTime := start.Add(time.Duration(i) * baseDur)
							price := 100 + float64(i)
							candles[i] = Candle{
								OpenTime:  openTime,
								CloseTime: openTime.Add(baseDur),
								Open:      price,
								High:      price + 1,
								Low:       price - 1,
								Close:     price + 0.5,
								Volume:    1,
							}
						}

						closed := candles[:len(candles)-1]
						aggregated := AggregateCandlesToTimeframe(closed, tc.base, tc.target)
						if len(aggregated) < warmup {
							t.Fatalf("got %d complete %s candles after excluding live candle, want at least %d", len(aggregated), tc.target, warmup)
						}
					})
				}
			})
		}
	}
}
