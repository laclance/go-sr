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
	if err != nil {
		return 0
	}
	if normalized == ModeZones {
		return pivotWindow
	}
	return legacyPivotWindow
}

// WarmupCandles returns the minimum closed-candle history needed before a
// bounded support/resistance calculation can be considered fully warmed up.
// It returns 0 when lookback <= 0 or mode is invalid because no finite warmup
// size can be provided.
func WarmupCandles(lookback int, mode Mode) int {
	window := pivotWindowForMode(mode)
	if window == 0 || lookback <= 0 {
		return 0
	}
	return lookback + 2*window + 10
}

// RequiredKlineLimit returns the raw kline fetch size needed to build an S/R
// bundle for targetInterval from a baseInterval stream. The returned size
// includes alignment slack for UTC target buckets and one extra live candle
// because exchange REST responses usually include the currently forming bar.
// It returns 0 when no finite fetch size can be provided, including for a
// non-positive lookback or invalid mode/interval combination.
func RequiredKlineLimit(baseInterval, targetInterval string, lookback int, mode Mode) int {
	baseDur := intervalDuration(baseInterval)
	targetDur := intervalDuration(targetInterval)
	if baseDur <= 0 || targetDur <= 0 || targetDur <= baseDur || targetDur%baseDur != 0 {
		return 0
	}

	warmup := WarmupCandles(lookback, mode)
	if warmup == 0 {
		return 0
	}

	ratio := int(targetDur / baseDur)
	closedRequired := warmup*ratio + (ratio - 1)
	return closedRequired + 1
}

// AggregateCandlesToTimeframe rolls a closed-candle slice into a higher
// timeframe using fixed-duration UTC buckets anchored at 1970-01-01T00:00:00Z.
// Any leading or trailing partial bucket is dropped.
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
		bucketStart := aggregateBucketStart(candle.OpenTime, toDur)
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

// aggregateBucketStart floors an open time to the fixed UTC bucket grid whose
// origin is the Unix epoch.
func aggregateBucketStart(openTime time.Time, bucketDur time.Duration) time.Time {
	bucketSeconds := int64(bucketDur / time.Second)
	unixSeconds := openTime.UTC().Unix()
	remainder := unixSeconds % bucketSeconds
	if remainder < 0 {
		remainder += bucketSeconds
	}
	return time.Unix(unixSeconds-remainder, 0).UTC()
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
