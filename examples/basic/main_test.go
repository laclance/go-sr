package main

import (
	"testing"

	sr "github.com/laclance/go-sr"
)

func TestMain(t *testing.T) {
	main()
}

func TestMustComputePanicsOnInvalidMode(t *testing.T) {
	t.Helper()

	defer func() {
		if recover() == nil {
			t.Fatal("expected panic for invalid mode")
		}
	}()

	mustCompute(sr.Options{
		Timeframe: "5m",
		Lookback:  120,
		Mode:      sr.Mode("invalid"),
	})
}
