package sr

import (
	"math"
	"reflect"
	"testing"
	"time"
)

type fuzzIntervalPair struct {
	from, to       string
	fromDur, toDur time.Duration
	bucketSize     int
}

var fuzzIntervalPairs = []fuzzIntervalPair{
	{"1m", "5m", time.Minute, 5 * time.Minute, 5},
	{"1m", "15m", time.Minute, 15 * time.Minute, 15},
	{"5m", "15m", 5 * time.Minute, 15 * time.Minute, 3},
	{"5m", "1h", 5 * time.Minute, time.Hour, 12},
	{"15m", "1h", 15 * time.Minute, time.Hour, 4},
	{"1h", "4h", time.Hour, 4 * time.Hour, 4},
}

func FuzzAggregateCandlesToTimeframe(f *testing.F) {
	for _, seed := range [][5]int{
		{0, 0, 0, 0, 0}, {1, 0, 1, 1, 0}, {4, 0, 0, 2, 0},
		{5, 0, 0, 3, 0}, {6, 0, 0, 4, 0}, {15, 1, 0, 5, 0},
		{7, 2, 1, 6, 0}, {24, 3, 0, 7, 0}, {9, 4, 2, 8, 2},
		{12, 5, 1, 9, 3},
	} {
		f.Add(seed[0], seed[1], seed[2], seed[3], seed[4])
	}

	f.Fuzz(func(t *testing.T, countSeed, pairSeed, alignmentSeed, shapeSeed, gapSeed int) {
		pair := fuzzIntervalPairs[fuzzBoundedInt(pairSeed, len(fuzzIntervalPairs))]
		candles := fuzzAggregateCandles(fuzzBoundedInt(countSeed, 257), pair, alignmentSeed, shapeSeed, gapSeed)
		before := append(candles[:0:0], candles...)
		got := AggregateCandlesToTimeframe(candles, pair.from, pair.to)

		if !reflect.DeepEqual(got, AggregateCandlesToTimeframe(candles, pair.from, pair.to)) {
			t.Fatal("aggregation is not deterministic")
		}
		if !reflect.DeepEqual(candles, before) {
			t.Fatal("aggregation mutated input")
		}
		if len(got) > len(candles) {
			t.Fatalf("aggregation expanded input: %d -> %d", len(candles), len(got))
		}

		expected := fuzzCompleteBuckets(candles, pair)
		if len(got) != len(expected) {
			t.Fatalf("complete bucket count mismatch for %s -> %s: got=%d want=%d", pair.from, pair.to, len(got), len(expected))
		}
		for i, candle := range got {
			if i > 0 && !got[i-1].OpenTime.Before(candle.OpenTime) {
				t.Fatalf("output is not strictly chronological")
			}
			if !fuzzCandleFinite(candle) || candle.High < candle.Low || candle.High < candle.Open || candle.High < candle.Close || candle.Low > candle.Open || candle.Low > candle.Close || candle.Volume < 0 {
				t.Fatalf("invalid aggregated candle: %+v", candle)
			}
			if !candle.OpenTime.Equal(fuzzBucketStart(candle.OpenTime, pair.toDur)) || !candle.CloseTime.Equal(candle.OpenTime.Add(pair.toDur)) {
				t.Fatalf("misaligned aggregated candle: %+v", candle)
			}
			assertFuzzAggregateBucket(t, candle, expected[i])
		}
	})
}

func FuzzComputeInvariants(f *testing.F) {
	for _, seed := range []struct {
		count, pattern, mode, lookback int
		phase                          int64
	}{
		{0, 0, 0, 0, 0}, {1, 1, 1, 2, 1}, {8, 2, 2, 4, 2},
		{9, 3, 2, 5, 3}, {10, 4, 1, 6, 5}, {11, 5, 1, 7, 8},
		{54, 6, 1, 9, 13}, {55, 7, 1, 10, 21}, {56, 8, 1, 11, 34},
		{65, 9, 2, 9, 55}, {66, 10, 2, 10, 89}, {67, 3, 0, 11, 144},
		{120, 4, 2, 0, 177}, {120, 8, 1, 1, 199}, {160, 4, 2, 8, 233},
	} {
		f.Add(seed.count, seed.pattern, seed.phase, seed.mode, seed.lookback)
	}

	f.Fuzz(func(t *testing.T, countSeed, patternSeed int, phaseSeed int64, modeSeed, lookbackSeed int) {
		count := fuzzBoundedInt(countSeed, 321)
		pattern := fuzzBoundedInt(patternSeed, 11)
		mode := []Mode{"", ModeLegacy, ModeZones}[fuzzBoundedInt(modeSeed, 3)]
		candles := fuzzComputePatternCandles(count, pattern, phaseSeed)
		before := append(candles[:0:0], candles...)
		opts := Options{
			Timeframe:   "5m",
			Lookback:    fuzzComputeLookback(lookbackSeed, count),
			Mode:        mode,
			Tolerance:   0.002,
			MinStrength: fuzzBoundedInt(fuzzBoundedInt(patternSeed, 4)+fuzzBoundedInt(lookbackSeed, 4), 4),
		}

		levels, err := Compute(candles, opts)
		if err != nil {
			t.Fatalf("unexpected Compute error: %v", err)
		}
		repeated, err := Compute(candles, opts)
		if err != nil || !reflect.DeepEqual(levels, repeated) {
			t.Fatalf("Compute is not deterministic: err=%v", err)
		}
		if !reflect.DeepEqual(candles, before) {
			t.Fatal("Compute mutated input")
		}

		normalizedMode := mode
		if mode == "" {
			normalizedMode = ModeLegacy
			legacyOpts := opts
			legacyOpts.Mode = ModeLegacy
			legacy, legacyErr := Compute(candles, legacyOpts)
			if legacyErr != nil || !reflect.DeepEqual(levels, legacy) {
				t.Fatalf("zero-value mode differs from explicit legacy: err=%v", legacyErr)
			}
		}
		assertFuzzLevels(t, levels, candles, normalizedMode)
	})
}

func TestCompute_FuzzRegression_NearestResistanceNotBelowClose(t *testing.T) {
	candles := fuzzComputeCandles(184, 19)
	levels, err := Compute(candles, Options{Timeframe: "5m", Lookback: 50, Mode: ModeZones, Tolerance: 0.002})
	if err != nil {
		t.Fatalf("unexpected Compute error: %v", err)
	}
	closePrice := candles[len(candles)-1].Close
	if levels.NearestSupport != 0 && levels.NearestSupport > closePrice {
		t.Fatalf("nearest support above price: support=%.2f close=%.2f", levels.NearestSupport, closePrice)
	}
	if levels.NearestResistance != 0 && levels.NearestResistance < closePrice {
		t.Fatalf("nearest resistance below price: resistance=%.2f close=%.2f", levels.NearestResistance, closePrice)
	}
}

type fuzzExpectedBucket struct {
	start   time.Time
	sources []Candle
}

func fuzzAggregateCandles(count int, pair fuzzIntervalPair, alignmentSeed, shapeSeed, gapSeed int) []Candle {
	base := time.Date(2024, 4, 1, 0, 0, 0, 0, time.UTC)
	start := base.Add(time.Duration(fuzzBoundedInt(alignmentSeed, pair.bucketSize)) * pair.fromDur)
	candles := make([]Candle, 0, count)
	slot := 0
	for i := 0; i < count; i++ {
		slot += fuzzGapBefore(i, pair.bucketSize, gapSeed)
		openTime := start.Add(time.Duration(slot) * pair.fromDur)
		level := 20 + float64(fuzzBoundedInt(shapeSeed, 2000))/10
		center := level + float64(((i+fuzzBoundedInt(shapeSeed, 17))%13)-6)*0.25 + float64(i%7)*0.1
		body := 0.01 + float64(fuzzBoundedInt(shapeSeed+i*3, 40))/20
		wick := float64(fuzzBoundedInt(shapeSeed+i*5, 30)) / 20
		open, closePrice := center-body/2, center+body/2
		if (i+fuzzBoundedInt(shapeSeed, 2))%2 == 1 {
			open, closePrice = closePrice, open
		}
		candles = append(candles, Candle{
			OpenTime: openTime, CloseTime: openTime.Add(pair.fromDur), Open: open,
			High: math.Max(open, closePrice) + wick, Low: math.Min(open, closePrice) - wick,
			Close: closePrice, Volume: float64(fuzzBoundedInt(shapeSeed+i*7, 500)) / 10,
		})
		slot++
	}
	return candles
}

func fuzzGapBefore(index, bucketSize, gapSeed int) int {
	if index == 0 {
		return 0
	}
	switch fuzzBoundedInt(gapSeed, 4) {
	case 1:
		if index == 1 {
			return 1
		}
	case 2:
		if index == bucketSize-1 {
			return 1
		}
	case 3:
		if index%(bucketSize+1) == 0 {
			return 2
		}
	}
	return 0
}

func fuzzCompleteBuckets(candles []Candle, pair fuzzIntervalPair) []fuzzExpectedBucket {
	var out []fuzzExpectedBucket
	var current fuzzExpectedBucket
	flush := func() {
		if len(current.sources) != pair.bucketSize {
			return
		}
		for i, candle := range current.sources {
			if !candle.OpenTime.Equal(current.start.Add(time.Duration(i) * pair.fromDur)) {
				return
			}
		}
		out = append(out, current)
	}
	for _, candle := range candles {
		start := fuzzBucketStart(candle.OpenTime, pair.toDur)
		if len(current.sources) == 0 {
			current.start = start
		}
		if !current.start.Equal(start) {
			flush()
			current = fuzzExpectedBucket{start: start}
		}
		current.sources = append(current.sources, candle)
	}
	flush()
	return out
}

func assertFuzzAggregateBucket(t *testing.T, got Candle, expected fuzzExpectedBucket) {
	t.Helper()
	wantOpen, wantHigh, wantLow := expected.sources[0].Open, expected.sources[0].High, expected.sources[0].Low
	wantClose := expected.sources[len(expected.sources)-1].Close
	var wantVolume float64
	for _, candle := range expected.sources {
		wantHigh = math.Max(wantHigh, candle.High)
		wantLow = math.Min(wantLow, candle.Low)
		wantVolume += candle.Volume
	}
	if !got.OpenTime.Equal(expected.start) || got.Open != wantOpen || got.High != wantHigh || got.Low != wantLow || got.Close != wantClose || !fuzzFloatEqual(got.Volume, wantVolume) {
		t.Fatalf("aggregate mismatch: got=%+v", got)
	}
}

func fuzzBucketStart(openTime time.Time, bucketDur time.Duration) time.Time {
	seconds := int64(bucketDur / time.Second)
	unix := openTime.UTC().Unix()
	return time.Unix(unix-(unix%seconds), 0).UTC()
}

func fuzzComputeLookback(seed, count int) int {
	values := []int{-1, 0, 1, 2, 4, 8, max(1, count/2), max(1, count-1), count, count + 1, 50, 100}
	return values[fuzzBoundedInt(seed, len(values))]
}

func fuzzComputePatternCandles(count, pattern int, phaseSeed int64) []Candle {
	candles := make([]Candle, count)
	start := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	phase := int(phaseSeed % 31)
	if phase < 0 {
		phase = -phase
	}
	for i := range candles {
		var open, high, low, closePrice float64
		switch pattern {
		case 0:
			open, high, low, closePrice = 100, 100, 100, 100
		case 1:
			center := 10 + float64(i)*0.25
			open, closePrice, high, low = center-0.1, center+0.1, center+0.15, center-0.15
		case 2:
			center := 200 - float64(i)*0.25
			open, closePrice, high, low = center+0.1, center-0.1, center+0.15, center-0.15
		case 3:
			center := 100 + float64((i+phase)%5)
			open, closePrice = center-0.75, center+0.75
			if i%2 == 1 {
				open, closePrice = closePrice, open
			}
			high, low = math.Max(open, closePrice)+0.25, math.Min(open, closePrice)-0.25
		case 4:
			center := 100 + math.Sin(float64(i+phase)/4)*2
			open, closePrice = center+math.Sin(float64(i+phase)/7)*0.2, center+math.Cos(float64(i+phase)/9)*0.2
			high, low = math.Max(open, closePrice)+0.1, math.Min(open, closePrice)-0.1
		case 5:
			center := 100 + float64((i+phase)/7%5)*8
			if (i+phase)%14 >= 7 {
				center -= 16
			}
			open, closePrice, high, low = center-1, center+1, center+2, center-2
		case 6:
			center := 100 + float64((i+phase)%3)*0.25
			open, closePrice, high, low = center-0.1, center+0.1, 101, 99
		case 7:
			center := 1 + float64((i+phase)%5)*1e-9
			open, closePrice, high, low = center-1e-10, center+1e-10, center+2e-10, center-2e-10
		case 8:
			center := 50 + float64((i+phase)%4)
			open, closePrice, high, low = center-0.25, center+0.25, center+0.5, center-0.5
			if i%3 == 0 {
				open, closePrice = center, center
			}
		case 9:
			center := 0.001 + float64((i+phase)%7)*0.00001
			open, closePrice, high, low = center-0.000002, center+0.000002, center+0.000004, center-0.000004
		case 10:
			center := -100 + float64((i+phase)%9)*0.25
			open, closePrice = center-0.1, center+0.1
			if i%2 == 1 {
				open, closePrice = closePrice, open
			}
			high, low = math.Max(open, closePrice)+0.05, math.Min(open, closePrice)-0.05
		}
		openTime := start.Add(time.Duration(i) * 5 * time.Minute)
		candles[i] = Candle{OpenTime: openTime, CloseTime: openTime.Add(5 * time.Minute), Open: open, High: high, Low: low, Close: closePrice, Volume: float64((i+phase)%11) * 10}
	}
	return candles
}

func assertFuzzLevels(t *testing.T, levels Levels, candles []Candle, mode Mode) {
	t.Helper()
	if levels.Timeframe != "5m" || len(levels.RawZones) < len(levels.Levels) || !fuzzLevelsFinite(levels) {
		t.Fatalf("invalid result bundle: %+v", levels)
	}
	for _, group := range []struct {
		name   string
		levels []Level
	}{{"levels", levels.Levels}, {"raw zones", levels.RawZones}} {
		for i, level := range group.levels {
			if i > 0 && group.levels[i-1].Price > level.Price {
				t.Fatalf("%s not sorted", group.name)
			}
			if level.Top < level.Bottom || level.Price < level.Bottom || level.Price > level.Top || level.Strength <= 0 {
				t.Fatalf("invalid %s geometry: %+v", group.name, level)
			}
			if mode == ModeZones {
				if level.LastTouchIndex < 0 || level.LastTouchIndex >= len(candles) || level.Strength != len(level.SourcePivotIndexes) || level.Strength != len(level.Pivots) {
					t.Fatalf("invalid %s metadata: %+v", group.name, level)
				}
				for _, index := range level.SourcePivotIndexes {
					if index < 0 || index >= len(candles) {
						t.Fatalf("source pivot index out of bounds: %d", index)
					}
				}
				for _, pivot := range level.Pivots {
					if pivot.Index < 0 || pivot.Index >= len(candles) || pivot.ConfirmedAtIndex < pivot.Index || pivot.ConfirmedAtIndex >= len(candles) {
						t.Fatalf("invalid pivot metadata: %+v", pivot)
					}
				}
			}
		}
	}
	if len(candles) == 0 {
		return
	}
	closePrice := candles[len(candles)-1].Close
	if levels.NearestSupport != 0 && levels.NearestSupport > closePrice {
		t.Fatalf("nearest support above close")
	}
	if levels.NearestResistance != 0 && levels.NearestResistance < closePrice {
		t.Fatalf("nearest resistance below close")
	}
	if levels.NearestSupportDistance < 0 || levels.NearestResistanceDistance < 0 || levels.NearestSupportStrength < 0 || levels.NearestResistanceStrength < 0 {
		t.Fatalf("negative nearest metadata")
	}
}

func fuzzLevelsFinite(levels Levels) bool {
	for _, v := range []float64{levels.NearestSupport, levels.NearestResistance, levels.NearestSupportDistance, levels.NearestResistanceDistance, levels.NearestSupportScore, levels.NearestResistanceScore} {
		if !fuzzFloatFinite(v) {
			return false
		}
	}
	for _, group := range [][]Level{levels.Levels, levels.RawZones} {
		for _, level := range group {
			if !fuzzFloatFinite(level.Price) || !fuzzFloatFinite(level.Top) || !fuzzFloatFinite(level.Bottom) || !fuzzFloatFinite(level.Score) {
				return false
			}
			for _, pivot := range level.Pivots {
				for _, v := range []float64{pivot.Price, pivot.ATRSnapshot, pivot.AvgVolumeSnapshot, pivot.Volume, pivot.VolumeRatio, pivot.MergeWidth, pivot.BounceATR} {
					if !fuzzFloatFinite(v) {
						return false
					}
				}
			}
		}
	}
	return true
}

func fuzzCandleFinite(c Candle) bool {
	return fuzzFloatFinite(c.Open) && fuzzFloatFinite(c.High) && fuzzFloatFinite(c.Low) && fuzzFloatFinite(c.Close) && fuzzFloatFinite(c.Volume)
}
func fuzzFloatFinite(v float64) bool { return !math.IsNaN(v) && !math.IsInf(v, 0) }
func fuzzFloatEqual(a, b float64) bool {
	return math.Abs(a-b) <= 1e-12*math.Max(1, math.Max(math.Abs(a), math.Abs(b)))
}
func fuzzBoundedInt(value, limit int) int {
	if limit <= 0 {
		return 0
	}
	value %= limit
	if value < 0 {
		value += limit
	}
	return value
}

func fuzzComputeCandles(count int, phaseSeed int64) []Candle {
	candles := make([]Candle, count)
	start := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	price := 50000.0
	for i := range candles {
		phase := float64((int64(i)+phaseSeed)%31) / 5
		price += math.Sin(phase)*15 + float64((i%5)-2)*6
		if price < 100 {
			price = 100
		}
		open, closePrice := price-10, price+10
		if i%2 == 1 {
			open, closePrice = closePrice, open
		}
		openTime := start.Add(time.Duration(i) * 5 * time.Minute)
		candles[i] = Candle{OpenTime: openTime, CloseTime: openTime.Add(5 * time.Minute), Open: open, High: math.Max(open, closePrice) + 5, Low: math.Min(open, closePrice) - 5, Close: closePrice, Volume: 100 + float64(i%9)}
	}
	return candles
}
