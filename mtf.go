package sr

import (
	"fmt"
	"strconv"
	"time"
)

func normalizeMode(mode Mode) (Mode, error) {
	switch mode {
	case "", ModeLegacy:
		return ModeLegacy, nil
	case ModeZones:
		return ModeZones, nil
	default:
		return "", fmt.Errorf("sr: unknown Mode %q", mode)
	}
}

func pivotWindowForMode(mode Mode) int {
	normalized, err := normalizeMode(mode)
	if err == nil && normalized == ModeZones {
		return pivotWindow
	}
	return legacyPivotWindow
}

// WarmupCandles returns the minimum closed-candle history needed before a
// support/resistance calculation can be considered fully warmed up.
func WarmupCandles(lookback int, mode Mode) int {
	if lookback < 0 {
		lookback = 0
	}
	return lookback + 2*pivotWindowForMode(mode) + 10
}

// RequiredKlineLimit returns the raw kline fetch size needed to build an S/R
// bundle for targetInterval from a baseInterval stream. The returned size
// includes one extra live candle because exchange REST responses usually
// include the currently forming bar.
func RequiredKlineLimit(baseInterval, targetInterval string, lookback int, mode Mode) int {
	baseDur := intervalDuration(baseInterval)
	targetDur := intervalDuration(targetInterval)
	if baseDur <= 0 || targetDur <= 0 || targetDur <= baseDur || targetDur%baseDur != 0 {
		return 0
	}

	closedRequired := WarmupCandles(lookback, mode) * int(targetDur/baseDur)
	return closedRequired + 1
}

// AggregateCandlesToTimeframe rolls a closed-candle slice into a higher
// timeframe using UTC-aligned buckets. Any leading or trailing partial bucket
// is dropped.
func AggregateCandlesToTimeframe(candles []Candle, fromInterval, toInterval string) []Candle {
	fromDur := intervalDuration(fromInterval)
	toDur := intervalDuration(toInterval)
	if len(candles) == 0 || fromDur <= 0 || toDur <= 0 || toDur <= fromDur || toDur%fromDur != 0 {
		return nil
	}

	bucketSize := int(toDur / fromDur)
	type bucket struct {
		start   time.Time
		end     time.Time
		candles []Candle
		seen    map[time.Time]struct{}
	}

	var (
		out     []Candle
		current *bucket
	)

	flush := func() {
		if current == nil || len(current.candles) != bucketSize {
			return
		}
		first := current.candles[0]
		last := current.candles[len(current.candles)-1]
		for i, candle := range current.candles {
			expectedOpen := current.start.Add(time.Duration(i) * fromDur)
			if !candle.OpenTime.UTC().Equal(expectedOpen) {
				return
			}
		}

		agg := Candle{
			OpenTime:  current.start,
			CloseTime: current.end,
			Open:      first.Open,
			High:      first.High,
			Low:       first.Low,
			Close:     last.Close,
			Volume:    0,
		}
		for _, candle := range current.candles {
			if candle.High > agg.High {
				agg.High = candle.High
			}
			if candle.Low < agg.Low {
				agg.Low = candle.Low
			}
			agg.Volume += candle.Volume
		}
		out = append(out, agg)
	}

	for _, candle := range candles {
		bucketStart := candle.OpenTime.UTC().Truncate(toDur)
		bucketEnd := bucketStart.Add(toDur)
		if current == nil || !current.start.Equal(bucketStart) {
			flush()
			current = &bucket{
				start: bucketStart,
				end:   bucketEnd,
				seen:  make(map[time.Time]struct{}, bucketSize),
			}
		}
		openTime := candle.OpenTime.UTC()
		if _, ok := current.seen[openTime]; ok {
			continue
		}
		current.seen[openTime] = struct{}{}
		current.candles = append(current.candles, candle)
	}
	flush()

	return out
}

func intervalDuration(interval string) time.Duration {
	if len(interval) < 2 {
		return 0
	}

	unit := interval[len(interval)-1]
	n, err := strconv.Atoi(interval[:len(interval)-1])
	if err != nil || n <= 0 {
		return 0
	}

	switch unit {
	case 'm':
		return time.Duration(n) * time.Minute
	case 'h':
		return time.Duration(n) * time.Hour
	case 'd':
		return time.Duration(n) * 24 * time.Hour
	default:
		return 0
	}
}
