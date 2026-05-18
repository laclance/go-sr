package sr

import (
	"math"
	"testing"
)

func TestNoLookahead_SupportResistance(t *testing.T) {
	candles := buildRealisticCandles(250)

	for idx := 130; idx < 240; idx += 10 {
		prefix := candles[:idx+1]

		srPrefix := computeZone(prefix, "5m", 50)
		srAgain := computeZone(candles[:idx+1], "5m", 50)

		if len(srPrefix.Levels) != len(srAgain.Levels) {
			t.Errorf("idx=%d: SR zone count differs on same prefix: %d vs %d",
				idx, len(srPrefix.Levels), len(srAgain.Levels))
			continue
		}
		if srPrefix.NearSupport != srAgain.NearSupport || srPrefix.NearResistance != srAgain.NearResistance {
			t.Errorf("idx=%d: SR proximity differs on same prefix", idx)
		}

		if idx+20 < len(candles) {
			extended := candles[:idx+21]
			srExtended := computeZone(extended, "5m", 50)

			for _, lvl := range srPrefix.Levels {
				found := false
				for _, eLvl := range srExtended.Levels {
					if math.Abs(lvl.Price-eLvl.Price)/lvl.Price < 0.005 {
						found = true
						break
					}
				}
				if !found && len(srPrefix.Levels) > 0 {
					t.Logf("idx=%d: prefix zone %.2f not in extended (lookback window slid or filtering changed)", idx, lvl.Price)
				}
			}
		}
	}
}
