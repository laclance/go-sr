package sr

import (
	"math"
	"time"
)

// Candle represents a single OHLCV closed candlestick.
type Candle struct {
	OpenTime  time.Time
	CloseTime time.Time
	Open      float64
	High      float64
	Low       float64
	Close     float64
	Volume    float64
}

func (c Candle) Body() float64 {
	return math.Abs(c.Close - c.Open)
}

func (c Candle) Range() float64 {
	return c.High - c.Low
}

func (c Candle) IsBullish() bool {
	return c.Close > c.Open
}

func (c Candle) IsBearish() bool {
	return c.Close < c.Open
}

func (c Candle) UpperWick() float64 {
	top := c.Close
	if c.Open > c.Close {
		top = c.Open
	}
	return c.High - top
}

func (c Candle) LowerWick() float64 {
	bottom := c.Open
	if c.Close < c.Open {
		bottom = c.Close
	}
	return bottom - c.Low
}

func (c Candle) MidPoint() float64 {
	return (c.High + c.Low) / 2
}

// PivotInfo captures the local snapshot data used to build a support or
// resistance zone.
type PivotInfo struct {
	Index             int       `json:"index"`
	ConfirmedAtIndex  int       `json:"confirmed_at_index"`
	Time              time.Time `json:"time"`
	Price             float64   `json:"price"`
	IsHigh            bool      `json:"is_high"`
	Timeframe         string    `json:"timeframe"`
	ATRSnapshot       float64   `json:"atr_snapshot"`
	AvgVolumeSnapshot float64   `json:"avg_volume_snapshot"`
	Volume            float64   `json:"volume"`
	VolumeRatio       float64   `json:"volume_ratio"`
	MergeWidth        float64   `json:"merge_width"`
	BounceATR         float64   `json:"bounce_atr"`
}

// Level represents a support or resistance zone built from clustered swing pivots.
type Level struct {
	Price              float64
	Top                float64
	Bottom             float64
	Strength           int
	Score              float64
	IsHigh             bool
	Timeframe          string
	LastTouchIndex     int
	SourcePivotIndexes []int
	Pivots             []PivotInfo
}

// Levels holds computed support/resistance data for the current candle series.
type Levels struct {
	Timeframe string

	Levels                    []Level
	RawZones                  []Level
	NearSupport               bool
	NearResistance            bool
	NearestSupport            float64
	NearestResistance         float64
	NearestSupportDistance    float64
	NearestResistanceDistance float64
	NearestSupportStrength    int
	NearestResistanceStrength int
	NearestSupportScore       float64
	NearestResistanceScore    float64
}

// Mode selects the support/resistance algorithm.
type Mode string

const (
	ModeLegacy Mode = "legacy"
	ModeZones  Mode = "zone"
)

// Options configures a support/resistance computation.
type Options struct {
	Timeframe string
	Lookback  int
	Mode      Mode
	Tolerance float64
	// MinStrength filters zone-mode raw zones by Strength.
	// 0 or negative -> use the default of 2 (back-compat).
	// 1+           -> require Strength >= MinStrength.
	// No effect on Mode=Legacy.
	MinStrength int
}
