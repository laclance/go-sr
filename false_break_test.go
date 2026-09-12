package sr

import (
	"testing"
	"time"
)

func TestSRZone_FalseBreaksDoNotReanchorWindowDuringSustainedBreakout(t *testing.T) {
	cases := []struct {
		name    string
		zone    Level
		outside float64
		reentry float64
	}{
		{
			name:    "resistance",
			zone:    Level{Top: 101.0, Bottom: 99.0, IsHigh: true},
			outside: 102.0,
			reentry: 100.0,
		},
		{
			name:    "support",
			zone:    Level{Top: 101.0, Bottom: 99.0, IsHigh: false},
			outside: 98.0,
			reentry: 100.0,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			candles := makeFlatCandles(16, 100.0, time.Date(2024, 3, 4, 0, 0, 0, 0, time.UTC))
			for i := 7; i <= 12; i++ {
				candles[i].Close = tc.outside
			}
			candles[13].Close = tc.reentry

			if got := countFalseBreaks(tc.zone, candles, 5); got != 0 {
				t.Fatalf("expected sustained breakout to remain a genuine break, got %d false breaks", got)
			}
		})
	}
}
