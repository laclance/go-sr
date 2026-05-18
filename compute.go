package sr

import "math"

const (
	pivotWindow = 4
)

// Compute detects support/resistance levels for the given candle prefix.
func Compute(candles []Candle, opts Options) Levels {
	switch normalizeMode(opts.Mode) {
	case ModeZones:
		return computeZones(candles, opts.Timeframe, opts.Lookback, opts.MinStrength)
	default:
		tolerance := opts.Tolerance
		if tolerance <= 0 {
			tolerance = 0.002
		}
		return computeSRLegacy(candles, opts.Timeframe, opts.Lookback, tolerance)
	}
}

// EmptyLevels returns the zero-value bundle for a timeframe label.
func EmptyLevels(timeframe string) Levels {
	return Levels{Timeframe: timeframe}
}

func computeZones(candles []Candle, timeframe string, lookback int, minStrength int) Levels {
	n := len(candles)
	if n < pivotWindow*2+1 {
		return EmptyLevels(timeframe)
	}

	highPivots := findPivotHighs(candles, timeframe, lookback)
	lowPivots := findPivotLows(candles, timeframe, lookback)

	rawHighZones := buildZones(highPivots, candles, lookback)
	rawLowZones := buildZones(lowPivots, candles, lookback)

	rawAll := append(cloneSRLevels(rawLowZones), cloneSRLevels(rawHighZones)...)
	sortSRLevels(rawAll)

	highZones := filterZones(rawHighZones, minStrength)
	lowZones := filterZones(rawLowZones, minStrength)
	all := append(cloneSRLevels(lowZones), cloneSRLevels(highZones)...)
	sortSRLevels(all)

	nearSup, nearRes, nearestSup, nearestRes, supDist, resDist, supStr, resStr, supScore, resScore :=
		detectZoneProximity(all, candles[n-1].Close)

	return Levels{
		Timeframe:                 timeframe,
		Levels:                    all,
		RawZones:                  rawAll,
		NearSupport:               nearSup,
		NearResistance:            nearRes,
		NearestSupport:            nearestSup,
		NearestResistance:         nearestRes,
		NearestSupportDistance:    supDist,
		NearestResistanceDistance: resDist,
		NearestSupportStrength:    supStr,
		NearestResistanceStrength: resStr,
		NearestSupportScore:       supScore,
		NearestResistanceScore:    resScore,
	}
}

type srLookbackWindow struct {
	Start      int
	PivotStart int
	PivotEnd   int
	ScanLen    int
}

func newSRLookbackWindow(n, lookback int) srLookbackWindow {
	if lookback <= 0 || lookback > n {
		lookback = n
	}
	start := n - lookback
	if start < 0 {
		start = 0
	}
	pivotStart := start
	if pivotStart < pivotWindow {
		pivotStart = pivotWindow
	}
	pivotEnd := n - pivotWindow
	if pivotEnd < pivotStart {
		pivotEnd = pivotStart
	}
	scanLen := n - start
	if scanLen < 1 {
		scanLen = 1
	}
	return srLookbackWindow{
		Start:      start,
		PivotStart: pivotStart,
		PivotEnd:   pivotEnd,
		ScanLen:    scanLen,
	}
}

// detectZoneProximity determines if the current price is near any qualified
// support/resistance zone. A zone counts as near when the close is within 2x
// its half-width.
func detectZoneProximity(zones []Level, close float64) (
	nearSup, nearRes bool,
	nearestSup, nearestRes float64,
	supDistance, resDistance float64,
	supStrength, resStrength int,
	supScore, resScore float64,
) {
	nearestSupDist := math.MaxFloat64
	nearestResDist := math.MaxFloat64

	type best struct {
		zone Level
		dist float64
	}
	var bestSupp, bestResi *best

	for _, z := range zones {
		if z.Price <= close {
			dist := close - z.Price
			if dist < nearestSupDist {
				nearestSupDist = dist
				bestSupp = &best{zone: z, dist: dist}
			}
			continue
		}

		dist := z.Price - close
		if dist < nearestResDist {
			nearestResDist = dist
			bestResi = &best{zone: z, dist: dist}
		}
	}

	if bestSupp != nil {
		z := bestSupp.zone
		nearestSup = z.Price
		supDistance = bestSupp.dist
		supStrength = z.Strength
		supScore = z.Score
		zoneRadius := (z.Top - z.Bottom) / 2
		if zoneRadius <= 0 {
			zoneRadius = close * 0.001
		}
		if bestSupp.dist <= zoneRadius*2 {
			nearSup = true
		}
	}

	if bestResi != nil {
		z := bestResi.zone
		nearestRes = z.Price
		resDistance = bestResi.dist
		resStrength = z.Strength
		resScore = z.Score
		zoneRadius := (z.Top - z.Bottom) / 2
		if zoneRadius <= 0 {
			zoneRadius = close * 0.001
		}
		if bestResi.dist <= zoneRadius*2 {
			nearRes = true
		}
	}

	return
}
