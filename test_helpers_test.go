package sr

import (
	"math"
	"reflect"
	"testing"
	"time"
)

func computeZone(candles []Candle, timeframe string, lookback int) Levels {
	return mustCompute(candles, Options{
		Timeframe: timeframe,
		Lookback:  lookback,
		Mode:      ModeZones,
	})
}

func computeLegacy(candles []Candle, timeframe string, lookback int, tolerance float64) Levels {
	return mustCompute(candles, Options{
		Timeframe: timeframe,
		Lookback:  lookback,
		Mode:      ModeLegacy,
		Tolerance: tolerance,
	})
}

func mustCompute(candles []Candle, opts Options) Levels {
	levels, err := Compute(candles, opts)
	if err != nil {
		panic(err)
	}
	return levels
}

func make5mCandles(start time.Time, values []struct {
	open   float64
	high   float64
	low    float64
	close  float64
	volume float64
}) []Candle {
	candles := make([]Candle, 0, len(values))
	for i, v := range values {
		openTime := start.Add(time.Duration(i) * 5 * time.Minute).UTC()
		candles = append(candles, Candle{
			OpenTime:  openTime,
			CloseTime: openTime.Add(5 * time.Minute),
			Open:      v.open,
			High:      v.high,
			Low:       v.low,
			Close:     v.close,
			Volume:    v.volume,
		})
	}
	return candles
}

// buildSRCandles produces a series with repeated swing highs and lows so that
// both legacy and zone SR algorithms have something to detect.
func buildSRCandles(n int) []Candle {
	candles := make([]Candle, n)
	t0 := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	var price float64
	for i := 0; i < n; i++ {
		phase := float64(i) / 15.0
		price = 50000.0 + math.Sin(phase)*500.0 + float64(i)*0.5
		body := 80.0 + float64(i%5)*10
		wick := 30.0
		open := price - body/2
		close_ := price + body/2
		candles[i] = Candle{
			OpenTime:  t0,
			CloseTime: t0.Add(5 * time.Minute),
			Open:      open,
			High:      math.Max(open, close_) + wick,
			Low:       math.Min(open, close_) - wick,
			Close:     close_,
			Volume:    1000 + float64(i%20)*50,
		}
		t0 = t0.Add(5 * time.Minute)
	}
	return candles
}

// buildSRCategoryCandles produces repeated, clearly separated swing lows near
// 49000 and swing highs near 51000 so category/type assertions can be strict.
func buildSRCategoryCandles() []Candle {
	candles := make([]Candle, 160)
	t0 := time.Date(2024, 2, 1, 0, 0, 0, 0, time.UTC)

	for i := range candles {
		base := 50000.0 + math.Sin(float64(i)/10.0)*60.0
		open := base - 20
		close_ := base + 20
		if i%2 == 1 {
			open, close_ = close_, open
		}
		high := math.Max(open, close_) + 70
		low := math.Min(open, close_) - 70

		switch i {
		case 20, 52, 84, 116:
			open = 50010
			close_ = 50070
			high = 50130
			low = 49000
		case 36, 68, 100, 132:
			open = 49990
			close_ = 49940
			high = 51000
			low = 49870
		case 159:
			open = 49990
			close_ = 50010
			high = 50080
			low = 49920
		}

		candles[i] = Candle{
			OpenTime:  t0,
			CloseTime: t0.Add(5 * time.Minute),
			Open:      open,
			High:      high,
			Low:       low,
			Close:     close_,
			Volume:    1000 + float64(i%5)*25,
		}
		t0 = t0.Add(5 * time.Minute)
	}

	return candles
}

// buildRealisticCandles creates a series with enough variation to trigger SR detection.
func buildRealisticCandles(n int) []Candle {
	candles := make([]Candle, n)
	t0 := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	price := 50000.0
	for i := 0; i < n; i++ {
		phase := i % 60
		var delta float64
		switch {
		case phase < 20:
			delta = 30 + float64(i%7)*5
		case phase < 30:
			delta = -10 + float64(i%5)*2
		case phase < 50:
			delta = -25 - float64(i%6)*4
		default:
			delta = float64(i%9)*3 - 10
		}
		price += delta
		if price < 30000 {
			price = 30000
		}

		body := 50 + float64(i%10)*10
		wick := 20 + float64(i%7)*5

		var open, close_ float64
		if i%3 != 0 {
			open = price - body/2
			close_ = price + body/2
		} else {
			open = price + body/2
			close_ = price - body/2
		}

		candles[i] = Candle{
			OpenTime:  t0,
			CloseTime: t0.Add(5 * time.Minute),
			Open:      open,
			High:      math.Max(open, close_) + wick,
			Low:       math.Min(open, close_) - wick,
			Close:     close_,
			Volume:    1000 + float64(i%20)*100,
		}
		t0 = t0.Add(5 * time.Minute)
	}
	return candles
}

// makeFlatCandles builds n identical flat candles ending at closeTime.
func makeFlatCandles(n int, basePrice float64, closeTime time.Time) []Candle {
	candles := make([]Candle, n)
	for i := range candles {
		ct := closeTime.Add(-time.Duration(n-1-i) * time.Minute)
		candles[i] = Candle{
			OpenTime:  ct.Add(-time.Minute),
			CloseTime: ct,
			Open:      basePrice,
			High:      basePrice * 1.002,
			Low:       basePrice * 0.998,
			Close:     basePrice * 1.001,
			Volume:    100,
		}
	}
	return candles
}

func countLevelsNear(levels []Level, price, tol float64, isHigh bool) int {
	count := 0
	for _, lvl := range levels {
		if lvl.IsHigh == isHigh && math.Abs(lvl.Price-price) <= tol {
			count++
		}
	}
	return count
}

func findLevelNear(levels []Level, price, tol float64, isHigh bool) *Level {
	for i := range levels {
		if levels[i].IsHigh == isHigh && math.Abs(levels[i].Price-price) <= tol {
			return &levels[i]
		}
	}
	return nil
}

func findLevelByPrice(levels []Level, price float64) *Level {
	for i := range levels {
		if levels[i].Price == price {
			return &levels[i]
		}
	}
	return nil
}

func findPivotByIndex(pivots []srPivot, index int) *srPivot {
	for i := range pivots {
		if pivots[i].Index == index {
			return &pivots[i]
		}
	}
	return nil
}

func findZoneBySource(levels []Level, sourcePivotIndexes []int, isHigh bool) *Level {
	for i := range levels {
		if levels[i].IsHigh == isHigh && reflect.DeepEqual(levels[i].SourcePivotIndexes, sourcePivotIndexes) {
			return &levels[i]
		}
	}
	return nil
}

func assertZoneOrdering(t *testing.T, levels []Level) {
	t.Helper()
	for i := 1; i < len(levels); i++ {
		prev := levels[i-1]
		curr := levels[i]
		if prev.Price > curr.Price {
			t.Fatalf("zones out of price order: %.2f before %.2f", prev.Price, curr.Price)
		}
		if prev.Price == curr.Price && prev.IsHigh && !curr.IsHigh {
			t.Fatalf("support/resistance tie order violated at price %.2f", curr.Price)
		}
	}
	for _, level := range levels {
		for i := 1; i < len(level.SourcePivotIndexes); i++ {
			if level.SourcePivotIndexes[i-1] > level.SourcePivotIndexes[i] {
				t.Fatalf("source pivot indexes out of order in zone %.2f: %v", level.Price, level.SourcePivotIndexes)
			}
		}
		for i := 1; i < len(level.Pivots); i++ {
			if level.Pivots[i-1].Index > level.Pivots[i].Index {
				t.Fatalf("pivot diagnostics out of order in zone %.2f: %+v", level.Price, level.Pivots)
			}
		}
	}
}
