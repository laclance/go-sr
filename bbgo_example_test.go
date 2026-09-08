package sr

import (
	"os"
	"strings"
	"testing"
)

func TestBBGOExampleUsesWarmupReadiness(t *testing.T) {
	const (
		lookback         = 120
		maxClosedCandles = 240
	)

	warmup := WarmupCandles(lookback, ModeZones)
	if warmup <= lookback {
		t.Fatalf("zone warmup: got %d, want greater than lookback %d", warmup, lookback)
	}
	if maxClosedCandles <= warmup {
		t.Fatalf("BBGO history cap: got %d, want greater than warmup %d", maxClosedCandles, warmup)
	}

	// adapter.go is deliberately build-excluded so the root module does not
	// depend on BBGO. Check the copyable source directly instead.
	source, err := os.ReadFile("examples/bbgo/adapter.go")
	if err != nil {
		t.Fatalf("read BBGO adapter: %v", err)
	}
	text := string(source)

	for _, want := range []string{
		"warmup := sr.WarmupCandles(lookback, sr.ModeZones)",
		"if len(s.closedCandles) < warmup {",
		"Lookback:    lookback,",
	} {
		if !strings.Contains(text, want) {
			t.Fatalf("BBGO adapter missing %q", want)
		}
	}
	if strings.Contains(text, "if len(s.closedCandles) < lookback {") {
		t.Fatal("BBGO adapter still gates readiness on raw lookback")
	}
}
