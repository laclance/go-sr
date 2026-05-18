package sr

import (
	"encoding/csv"
	"math"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"testing"
	"time"
)

func loadBTCCSVFixture(t *testing.T) []Candle {
	t.Helper()

	_, filename, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("runtime.Caller failed")
	}
	path := filepath.Join(filepath.Dir(filename), "testdata", "btc_5m.csv")

	f, err := os.Open(path)
	if err != nil {
		t.Fatalf("open btc csv: %v", err)
	}
	defer f.Close()

	rows, err := csv.NewReader(f).ReadAll()
	if err != nil {
		t.Fatalf("read btc csv: %v", err)
	}

	candles := make([]Candle, 0, len(rows))
	for _, row := range rows {
		if len(row) < 6 {
			continue
		}
		openTimeMs, err := strconv.ParseInt(row[0], 10, 64)
		if err != nil {
			continue
		}
		open, _ := strconv.ParseFloat(row[1], 64)
		high, _ := strconv.ParseFloat(row[2], 64)
		low, _ := strconv.ParseFloat(row[3], 64)
		closePrice, _ := strconv.ParseFloat(row[4], 64)
		volume, _ := strconv.ParseFloat(row[5], 64)
		openTime := time.UnixMilli(openTimeMs).UTC()
		candles = append(candles, Candle{
			OpenTime:  openTime,
			CloseTime: openTime.Add(5 * time.Minute),
			Open:      open,
			High:      high,
			Low:       low,
			Close:     closePrice,
			Volume:    volume,
		})
	}
	return candles
}

type btcSnapshot struct {
	name      string
	timeframe string
	mode      Mode
	aggCount  int
	firstOpen string
	lastClose string
	levels    int
	raw       int
	nearSup   bool
	nearRes   bool
	sup       float64
	res       float64
	supStr    int
	resStr    int
	supScore  float64
	resScore  float64
}

func TestBTCCSVRegressionSnapshots(t *testing.T) {
	candles := loadBTCCSVFixture(t)
	prefix := candles[:900]

	snapshots := []btcSnapshot{
		func() btcSnapshot {
			levels := Compute(prefix, Options{Timeframe: "5m", Lookback: 50, Mode: ModeZones})
			return btcSnapshot{
				name:      "5m_zone",
				timeframe: "5m",
				mode:      ModeZones,
				aggCount:  len(prefix),
				firstOpen: prefix[0].OpenTime.UTC().Format(time.RFC3339),
				lastClose: prefix[len(prefix)-1].CloseTime.UTC().Format(time.RFC3339),
				levels:    len(levels.Levels),
				raw:       len(levels.RawZones),
				nearSup:   levels.NearSupport,
				nearRes:   levels.NearResistance,
				sup:       levels.NearestSupport,
				res:       levels.NearestResistance,
				supStr:    levels.NearestSupportStrength,
				resStr:    levels.NearestResistanceStrength,
				supScore:  levels.NearestSupportScore,
				resScore:  levels.NearestResistanceScore,
			}
		}(),
		func() btcSnapshot {
			agg := AggregateCandlesToTimeframe(prefix, "5m", "15m")
			levels := Compute(agg, Options{Timeframe: "15m", Lookback: 50, Mode: ModeZones})
			return btcSnapshot{
				name:      "15m_zone",
				timeframe: "15m",
				mode:      ModeZones,
				aggCount:  len(agg),
				firstOpen: agg[0].OpenTime.UTC().Format(time.RFC3339),
				lastClose: agg[len(agg)-1].CloseTime.UTC().Format(time.RFC3339),
				levels:    len(levels.Levels),
				raw:       len(levels.RawZones),
				nearSup:   levels.NearSupport,
				nearRes:   levels.NearResistance,
				sup:       levels.NearestSupport,
				res:       levels.NearestResistance,
				supStr:    levels.NearestSupportStrength,
				resStr:    levels.NearestResistanceStrength,
				supScore:  levels.NearestSupportScore,
				resScore:  levels.NearestResistanceScore,
			}
		}(),
		func() btcSnapshot {
			agg := AggregateCandlesToTimeframe(prefix, "5m", "1h")
			levels := Compute(agg, Options{Timeframe: "1h", Lookback: 50, Mode: ModeZones})
			return btcSnapshot{
				name:      "1h_zone",
				timeframe: "1h",
				mode:      ModeZones,
				aggCount:  len(agg),
				firstOpen: agg[0].OpenTime.UTC().Format(time.RFC3339),
				lastClose: agg[len(agg)-1].CloseTime.UTC().Format(time.RFC3339),
				levels:    len(levels.Levels),
				raw:       len(levels.RawZones),
				nearSup:   levels.NearSupport,
				nearRes:   levels.NearResistance,
				sup:       levels.NearestSupport,
				res:       levels.NearestResistance,
				supStr:    levels.NearestSupportStrength,
				resStr:    levels.NearestResistanceStrength,
				supScore:  levels.NearestSupportScore,
				resScore:  levels.NearestResistanceScore,
			}
		}(),
		func() btcSnapshot {
			levels := Compute(prefix, Options{Timeframe: "5m", Lookback: 50, Mode: ModeLegacy, Tolerance: 0.002})
			return btcSnapshot{
				name:      "5m_legacy",
				timeframe: "5m",
				mode:      ModeLegacy,
				aggCount:  len(prefix),
				firstOpen: prefix[0].OpenTime.UTC().Format(time.RFC3339),
				lastClose: prefix[len(prefix)-1].CloseTime.UTC().Format(time.RFC3339),
				levels:    len(levels.Levels),
				raw:       len(levels.RawZones),
				nearSup:   levels.NearSupport,
				nearRes:   levels.NearResistance,
				sup:       levels.NearestSupport,
				res:       levels.NearestResistance,
				supStr:    levels.NearestSupportStrength,
				resStr:    levels.NearestResistanceStrength,
				supScore:  levels.NearestSupportScore,
				resScore:  levels.NearestResistanceScore,
			}
		}(),
	}

	want := map[string]btcSnapshot{
		"5m_zone": {
			name:      "5m_zone",
			timeframe: "5m",
			mode:      ModeZones,
			aggCount:  900,
			firstOpen: "2026-01-01T00:00:00Z",
			lastClose: "2026-01-04T03:00:00Z",
			levels:    2,
			raw:       6,
			nearSup:   false,
			nearRes:   false,
			sup:       91032.2,
			res:       91279.495,
			supStr:    2,
			resStr:    2,
			supScore:  8.744184809243702,
			resScore:  9.191089038509844,
		},
		"15m_zone": {
			name:      "15m_zone",
			timeframe: "15m",
			mode:      ModeZones,
			aggCount:  300,
			firstOpen: "2026-01-01T00:00:00Z",
			lastClose: "2026-01-04T03:00:00Z",
			levels:    1,
			raw:       7,
			nearSup:   false,
			nearRes:   false,
			sup:       90208.775,
			res:       0,
			supStr:    2,
			resStr:    0,
			supScore:  7.328657452288571,
			resScore:  0,
		},
		"1h_zone": {
			name:      "1h_zone",
			timeframe: "1h",
			mode:      ModeZones,
			aggCount:  75,
			firstOpen: "2026-01-01T00:00:00Z",
			lastClose: "2026-01-04T03:00:00Z",
			levels:    0,
			raw:       6,
			nearSup:   false,
			nearRes:   false,
			sup:       0,
			res:       0,
			supStr:    0,
			resStr:    0,
			supScore:  0,
			resScore:  0,
		},
		"5m_legacy": {
			name:      "5m_legacy",
			timeframe: "5m",
			mode:      ModeLegacy,
			aggCount:  900,
			firstOpen: "2026-01-01T00:00:00Z",
			lastClose: "2026-01-04T03:00:00Z",
			levels:    4,
			raw:       4,
			nearSup:   true,
			nearRes:   true,
			sup:       91032.2,
			res:       91292,
			supStr:    2,
			resStr:    1,
			supScore:  0,
			resScore:  0,
		},
	}

	for _, snapshot := range snapshots {
		expected, ok := want[snapshot.name]
		if !ok {
			t.Fatalf("missing expected snapshot for %s", snapshot.name)
		}
		assertBTCSnapshot(t, snapshot, expected)
	}
}

func assertBTCSnapshot(t *testing.T, got, want btcSnapshot) {
	t.Helper()

	if got.timeframe != want.timeframe {
		t.Fatalf("%s timeframe: got %q want %q", got.name, got.timeframe, want.timeframe)
	}
	if got.mode != want.mode {
		t.Fatalf("%s mode: got %q want %q", got.name, got.mode, want.mode)
	}
	if got.aggCount != want.aggCount {
		t.Fatalf("%s aggCount: got %d want %d", got.name, got.aggCount, want.aggCount)
	}
	if got.firstOpen != want.firstOpen || got.lastClose != want.lastClose {
		t.Fatalf("%s boundaries: got %s -> %s want %s -> %s", got.name, got.firstOpen, got.lastClose, want.firstOpen, want.lastClose)
	}
	if got.levels != want.levels || got.raw != want.raw {
		t.Fatalf("%s counts: got levels=%d raw=%d want levels=%d raw=%d", got.name, got.levels, got.raw, want.levels, want.raw)
	}
	if got.nearSup != want.nearSup || got.nearRes != want.nearRes {
		t.Fatalf("%s proximity: got support=%t resistance=%t want support=%t resistance=%t", got.name, got.nearSup, got.nearRes, want.nearSup, want.nearRes)
	}
	if !almostEqual(got.sup, want.sup) || !almostEqual(got.res, want.res) {
		t.Fatalf("%s nearest levels: got support=%.12f resistance=%.12f want support=%.12f resistance=%.12f", got.name, got.sup, got.res, want.sup, want.res)
	}
	if got.supStr != want.supStr || got.resStr != want.resStr {
		t.Fatalf("%s strengths: got support=%d resistance=%d want support=%d resistance=%d", got.name, got.supStr, got.resStr, want.supStr, want.resStr)
	}
	if !almostEqual(got.supScore, want.supScore) || !almostEqual(got.resScore, want.resScore) {
		t.Fatalf("%s scores: got support=%.12f resistance=%.12f want support=%.12f resistance=%.12f", got.name, got.supScore, got.resScore, want.supScore, want.resScore)
	}
}

func almostEqual(a, b float64) bool {
	return math.Abs(a-b) <= 1e-6
}
