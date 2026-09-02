//go:build ignore

// Copy this adapter into a BBGO strategy package and remove the build constraint.
// It is excluded here so github.com/laclance/go-sr does not depend on BBGO.
package strategy

import (
	"context"
	"log"

	"github.com/c9s/bbgo/pkg/bbgo"
	"github.com/c9s/bbgo/pkg/types"
	sr "github.com/laclance/go-sr"
)

const (
	timeframe        = "5m"
	lookback         = 120
	maxClosedCandles = 240
)

type Strategy struct {
	Symbol string `json:"symbol"`

	closedCandles []sr.Candle
}

func (s *Strategy) Subscribe(session *bbgo.ExchangeSession) {
	session.Subscribe(types.KLineChannel, s.Symbol, types.SubscribeOptions{
		Interval: types.Interval(timeframe),
	})
}

func (s *Strategy) Run(
	_ context.Context,
	_ bbgo.OrderExecutor,
	session *bbgo.ExchangeSession,
) error {
	// Fix the capacity up front so this slice cannot grow beyond the bound below.
	s.closedCandles = make([]sr.Candle, 0, maxClosedCandles)

	session.MarketDataStream.OnKLineClosed(func(kline types.KLine) {
		if kline.Symbol != s.Symbol || kline.Interval != types.Interval(timeframe) {
			return
		}
		// OnKLineClosed is the important boundary. This guard is defense in depth:
		// never pass a still-open kline to go-sr.
		if !kline.Closed {
			return
		}

		s.closedCandles = appendClosedCandle(
			s.closedCandles,
			candleFromBBGO(kline),
		)
		if len(s.closedCandles) < lookback {
			return
		}

		levels, err := sr.Compute(s.closedCandles, sr.Options{
			Timeframe:   "5m",
			Lookback:    120,
			Mode:        sr.ModeZones,
			MinStrength: 2,
		})
		if err != nil {
			log.Printf("go-sr compute failed: %v", err)
			return
		}

		log.Printf(
			"go-sr support=%.8f resistance=%.8f near_support=%t near_resistance=%t",
			levels.NearestSupport,
			levels.NearestResistance,
			levels.NearSupport,
			levels.NearResistance,
		)
	})

	return nil
}

func candleFromBBGO(kline types.KLine) sr.Candle {
	return sr.Candle{
		OpenTime:  kline.StartTime.Time(),
		CloseTime: kline.EndTime.Time(),
		Open:      kline.Open.Float64(),
		High:      kline.High.Float64(),
		Low:       kline.Low.Float64(),
		Close:     kline.Close.Float64(),
		Volume:    kline.Volume.Float64(),
	}
}

func appendClosedCandle(candles []sr.Candle, candle sr.Candle) []sr.Candle {
	if len(candles) < maxClosedCandles {
		return append(candles, candle)
	}

	copy(candles, candles[1:])
	candles[len(candles)-1] = candle
	return candles
}
