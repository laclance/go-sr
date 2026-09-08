package strategy

import "github.com/c9s/bbgo/pkg/bbgo"

// bbgoStrategyCompatibility supplies the ID method normally provided by the
// consuming BBGO strategy while promoting the real adapter's Run method.
type bbgoStrategyCompatibility struct {
	*Strategy
}

func (*bbgoStrategyCompatibility) ID() string { return "go-sr-example" }

var (
	_ bbgo.ExchangeSessionSubscriber = (*Strategy)(nil)
	_ bbgo.SingleExchangeStrategy    = (*bbgoStrategyCompatibility)(nil)
)
