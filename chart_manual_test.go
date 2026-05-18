package sr

import (
	"fmt"
	"html"
	"math"
	"os"
	"strconv"
	"strings"
	"testing"
)

type manualSRChartOptions struct {
	Source      string
	Timeframe   string
	Mode        Mode
	Lookback    int
	Limit       int
	Window      int
	MinStrength int
}

func TestGenerateManualSRChart(t *testing.T) {
	outPath := os.Getenv("GO_SR_CHART")
	if outPath == "" {
		t.Skip("set GO_SR_CHART=/tmp/go-sr-chart.html to generate a manual SR chart")
	}

	opts := manualSRChartOptions{
		Source:      "testdata/btc_5m.csv",
		Timeframe:   manualChartEnvString("GO_SR_CHART_TIMEFRAME", "5m"),
		Mode:        Mode(manualChartEnvString("GO_SR_CHART_MODE", string(ModeZones))),
		Lookback:    manualChartEnvInt(t, "GO_SR_CHART_LOOKBACK", 50),
		Limit:       manualChartEnvInt(t, "GO_SR_CHART_LIMIT", 900),
		Window:      manualChartEnvInt(t, "GO_SR_CHART_WINDOW", 240),
		MinStrength: manualChartEnvInt(t, "GO_SR_CHART_MIN_STRENGTH", 2),
	}

	mode, err := normalizeMode(opts.Mode)
	if err != nil {
		t.Fatalf("invalid GO_SR_CHART_MODE: %v", err)
	}
	opts.Mode = mode

	candles := loadBTCCSVFixture(t)
	if opts.Limit > 0 && len(candles) > opts.Limit {
		candles = candles[len(candles)-opts.Limit:]
	}
	computeCandles := candles
	if opts.Timeframe != "5m" {
		computeCandles = AggregateCandlesToTimeframe(candles, "5m", opts.Timeframe)
		if len(computeCandles) == 0 {
			t.Fatalf("no candles after aggregating 5m to %s", opts.Timeframe)
		}
	}

	levels, err := Compute(computeCandles, Options{
		Timeframe:   opts.Timeframe,
		Lookback:    opts.Lookback,
		Mode:        opts.Mode,
		Tolerance:   0.002,
		MinStrength: opts.MinStrength,
	})
	if err != nil {
		t.Fatalf("compute SR levels: %v", err)
	}

	displayCandles := computeCandles
	if opts.Window > 0 && len(displayCandles) > opts.Window {
		displayCandles = displayCandles[len(displayCandles)-opts.Window:]
	}

	if err := os.WriteFile(outPath, []byte(renderManualSRChart(displayCandles, levels, opts)), 0o644); err != nil {
		t.Fatalf("write chart: %v", err)
	}
	t.Logf("wrote %s", outPath)
}

func manualChartEnvString(key, fallback string) string {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return fallback
	}
	return value
}

func manualChartEnvInt(t *testing.T, key string, fallback int) int {
	t.Helper()
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return fallback
	}
	parsed, err := strconv.Atoi(value)
	if err != nil {
		t.Fatalf("invalid %s=%q: %v", key, value, err)
	}
	return parsed
}

func renderManualSRChart(candles []Candle, levels Levels, opts manualSRChartOptions) string {
	const (
		width      = 1280.0
		height     = 760.0
		plotLeft   = 78.0
		plotTop    = 42.0
		plotRight  = 1060.0
		plotBottom = 675.0
	)

	minPrice, maxPrice := manualChartPriceRange(candles, levels)
	priceSpan := maxPrice - minPrice
	if priceSpan <= 0 {
		priceSpan = math.Max(1, maxPrice*0.01)
		maxPrice += priceSpan / 2
		minPrice -= priceSpan / 2
	}
	padding := priceSpan * 0.06
	if padding <= 0 {
		padding = 1
	}
	minPrice -= padding
	maxPrice += padding

	yForPrice := func(price float64) float64 {
		return plotTop + ((maxPrice-price)/(maxPrice-minPrice))*(plotBottom-plotTop)
	}

	var b strings.Builder
	fmt.Fprintf(&b, "<!doctype html>\n<html><head><meta charset=\"utf-8\">\n")
	fmt.Fprintf(&b, "<title>%s %s SR chart</title>\n", html.EscapeString(opts.Timeframe), html.EscapeString(string(opts.Mode)))
	fmt.Fprintf(&b, "<style>%s</style>\n</head><body>\n", manualChartCSS())
	fmt.Fprintf(&b, "<main><h1>%s %s SR chart</h1>\n", html.EscapeString(opts.Timeframe), html.EscapeString(string(opts.Mode)))
	fmt.Fprintf(&b, "<p class=\"meta\">source=%s candles=%d displayed=%d lookback=%d qualified=%d raw=%d nearest_support=%.2f nearest_resistance=%.2f</p>\n",
		html.EscapeString(opts.Source), opts.Limit, len(candles), opts.Lookback, len(levels.Levels), len(levels.RawZones), levels.NearestSupport, levels.NearestResistance)
	fmt.Fprintf(&b, "<svg viewBox=\"0 0 %.0f %.0f\" role=\"img\" aria-label=\"Support resistance chart\">\n", width, height)
	fmt.Fprintf(&b, "<rect x=\"0\" y=\"0\" width=\"%.0f\" height=\"%.0f\" class=\"canvas\"/>\n", width, height)
	fmt.Fprintf(&b, "<rect x=\"%.0f\" y=\"%.0f\" width=\"%.0f\" height=\"%.0f\" class=\"plot\"/>\n", plotLeft, plotTop, plotRight-plotLeft, plotBottom-plotTop)

	for i := 0; i <= 6; i++ {
		price := minPrice + (maxPrice-minPrice)*float64(i)/6
		y := yForPrice(price)
		fmt.Fprintf(&b, "<line x1=\"%.1f\" y1=\"%.1f\" x2=\"%.1f\" y2=\"%.1f\" class=\"grid\"/>\n", plotLeft, y, plotRight, y)
		fmt.Fprintf(&b, "<text x=\"%.1f\" y=\"%.1f\" class=\"axis\">%.2f</text>\n", plotLeft-10, y+4, price)
	}

	for _, level := range levels.RawZones {
		manualChartZoneSVG(&b, level, yForPrice, plotLeft, plotRight, false)
	}
	for _, level := range levels.Levels {
		manualChartZoneSVG(&b, level, yForPrice, plotLeft, plotRight, true)
	}

	step := (plotRight - plotLeft) / math.Max(1, float64(len(candles)))
	bodyWidth := math.Max(2, step*0.58)
	for i, candle := range candles {
		x := plotLeft + float64(i)*step + step/2
		manualChartCandleSVG(&b, candle, x, bodyWidth, yForPrice)
	}

	if len(candles) > 0 {
		last := candles[len(candles)-1]
		y := yForPrice(last.Close)
		fmt.Fprintf(&b, "<line x1=\"%.1f\" y1=\"%.1f\" x2=\"%.1f\" y2=\"%.1f\" class=\"last-close\"/>\n", plotLeft, y, plotRight, y)
		fmt.Fprintf(&b, "<text x=\"%.1f\" y=\"%.1f\" class=\"last-label\">last %.2f</text>\n", plotRight+8, y+4, last.Close)
	}

	for _, level := range levels.Levels {
		y := yForPrice(level.Price)
		side := "S"
		if level.IsHigh {
			side = "R"
		}
		fmt.Fprintf(&b, "<text x=\"%.1f\" y=\"%.1f\" class=\"level-label %s\">%s %.2f str=%d score=%.2f</text>\n",
			plotRight+8, y+4, manualChartSideClass(level), side, level.Price, level.Strength, level.Score)
	}

	manualChartNearestSVG(&b, levels.NearestSupport, "nearest support", yForPrice, plotLeft, plotRight, false)
	manualChartNearestSVG(&b, levels.NearestResistance, "nearest resistance", yForPrice, plotLeft, plotRight, true)

	fmt.Fprintf(&b, "</svg></main></body></html>\n")
	return b.String()
}

func manualChartPriceRange(candles []Candle, levels Levels) (float64, float64) {
	minPrice := math.Inf(1)
	maxPrice := math.Inf(-1)
	for _, candle := range candles {
		minPrice = math.Min(minPrice, candle.Low)
		maxPrice = math.Max(maxPrice, candle.High)
	}
	for _, level := range append(append([]Level(nil), levels.RawZones...), levels.Levels...) {
		minPrice = math.Min(minPrice, math.Min(level.Bottom, level.Price))
		maxPrice = math.Max(maxPrice, math.Max(level.Top, level.Price))
	}
	if math.IsInf(minPrice, 1) || math.IsInf(maxPrice, -1) {
		return 0, 1
	}
	return minPrice, maxPrice
}

func manualChartZoneSVG(b *strings.Builder, level Level, yForPrice func(float64) float64, left, right float64, qualified bool) {
	topY := yForPrice(level.Top)
	bottomY := yForPrice(level.Bottom)
	if bottomY < topY {
		topY, bottomY = bottomY, topY
	}
	height := math.Max(1, bottomY-topY)
	class := "raw-zone"
	if qualified {
		class = "qualified-zone"
	}
	fmt.Fprintf(b, "<rect x=\"%.1f\" y=\"%.1f\" width=\"%.1f\" height=\"%.1f\" class=\"%s %s\"/>\n",
		left, topY, right-left, height, class, manualChartSideClass(level))
	fmt.Fprintf(b, "<line x1=\"%.1f\" y1=\"%.1f\" x2=\"%.1f\" y2=\"%.1f\" class=\"zone-center %s\"/>\n",
		left, yForPrice(level.Price), right, yForPrice(level.Price), manualChartSideClass(level))
}

func manualChartCandleSVG(b *strings.Builder, candle Candle, x, bodyWidth float64, yForPrice func(float64) float64) {
	class := "candle up"
	if candle.Close < candle.Open {
		class = "candle down"
	} else if candle.Close == candle.Open {
		class = "candle flat"
	}

	highY := yForPrice(candle.High)
	lowY := yForPrice(candle.Low)
	openY := yForPrice(candle.Open)
	closeY := yForPrice(candle.Close)
	bodyY := math.Min(openY, closeY)
	bodyHeight := math.Max(1, math.Abs(closeY-openY))

	fmt.Fprintf(b, "<g class=\"%s\"><title>%s O %.2f H %.2f L %.2f C %.2f</title>",
		class, html.EscapeString(candle.OpenTime.UTC().Format("2006-01-02 15:04")), candle.Open, candle.High, candle.Low, candle.Close)
	fmt.Fprintf(b, "<line x1=\"%.1f\" y1=\"%.1f\" x2=\"%.1f\" y2=\"%.1f\" class=\"wick\"/>", x, highY, x, lowY)
	fmt.Fprintf(b, "<rect x=\"%.1f\" y=\"%.1f\" width=\"%.1f\" height=\"%.1f\" class=\"body\"/></g>\n", x-bodyWidth/2, bodyY, bodyWidth, bodyHeight)
}

func manualChartNearestSVG(b *strings.Builder, price float64, label string, yForPrice func(float64) float64, left, right float64, high bool) {
	if price == 0 {
		return
	}
	y := yForPrice(price)
	class := "nearest support"
	if high {
		class = "nearest resistance"
	}
	fmt.Fprintf(b, "<line x1=\"%.1f\" y1=\"%.1f\" x2=\"%.1f\" y2=\"%.1f\" class=\"%s\"/>\n", left, y, right, y, class)
	fmt.Fprintf(b, "<text x=\"%.1f\" y=\"%.1f\" class=\"nearest-label\">%s %.2f</text>\n", left+8, y-6, html.EscapeString(label), price)
}

func manualChartSideClass(level Level) string {
	if level.IsHigh {
		return "resistance"
	}
	return "support"
}

func manualChartCSS() string {
	return `
body { margin: 0; background: #f7f4ee; color: #1c2430; font-family: ui-sans-serif, system-ui, -apple-system, BlinkMacSystemFont, "Segoe UI", sans-serif; }
main { max-width: 1280px; margin: 0 auto; padding: 20px 24px 28px; }
h1 { margin: 0 0 6px; font-size: 22px; font-weight: 700; }
.meta { margin: 0 0 14px; color: #4a5565; font-size: 13px; }
svg { width: 100%; height: auto; border: 1px solid #d8d2c6; background: #fffdf8; }
.canvas { fill: #fffdf8; }
.plot { fill: #ffffff; stroke: #cfd6df; stroke-width: 1; }
.grid { stroke: #e7e2d8; stroke-width: 1; }
.axis { text-anchor: end; font-size: 11px; fill: #697386; }
.candle .wick { stroke-width: 1.2; }
.candle .body { stroke-width: 1; }
.candle.up .wick, .candle.up .body { stroke: #0f8f5f; fill: #15a873; }
.candle.down .wick, .candle.down .body { stroke: #b8363f; fill: #d94a54; }
.candle.flat .wick, .candle.flat .body { stroke: #697386; fill: #8792a2; }
.raw-zone { opacity: 0.16; }
.qualified-zone { opacity: 0.30; }
.raw-zone.support, .qualified-zone.support { fill: #20a464; }
.raw-zone.resistance, .qualified-zone.resistance { fill: #dc3946; }
.zone-center { stroke-width: 1; stroke-dasharray: 5 5; opacity: 0.55; }
.zone-center.support { stroke: #107a4b; }
.zone-center.resistance { stroke: #b22531; }
.last-close { stroke: #1d4ed8; stroke-width: 1.4; stroke-dasharray: 4 4; }
.last-label, .level-label, .nearest-label { font-size: 11px; fill: #263241; }
.level-label.support { fill: #0d7648; }
.level-label.resistance { fill: #aa2330; }
.nearest { stroke-width: 2; stroke-dasharray: 2 3; }
.nearest.support { stroke: #0d7648; }
.nearest.resistance { stroke: #aa2330; }
`
}
