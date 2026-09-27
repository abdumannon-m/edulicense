package app

import (
	"fmt"
	"html/template"
	"math"
	"strings"
)

// Server-rendered SVG charts for the finance dashboard. They scale with the
// container (viewBox + width 100%) and use the admin colour tokens.

type ChartSeries struct {
	Label  string
	Class  string // fin-chart__bar--income, --expense, --profit ...
	Values []float64
}

func niceMax(value float64) float64 {
	if value <= 0 {
		return 1
	}
	exp := math.Pow(10, math.Floor(math.Log10(value)))
	for _, step := range []float64{1, 2, 2.5, 5, 10} {
		if step*exp >= value {
			return step * exp
		}
	}
	return 10 * exp
}

func compactNumber(value float64) string {
	abs := math.Abs(value)
	sign := ""
	if value < 0 {
		sign = "−"
	}
	switch {
	case abs >= 1e9:
		return sign + trimZero(fmt.Sprintf("%.1f", abs/1e9)) + "B"
	case abs >= 1e6:
		return sign + trimZero(fmt.Sprintf("%.1f", abs/1e6)) + "M"
	case abs >= 1e3:
		return sign + trimZero(fmt.Sprintf("%.1f", abs/1e3)) + "k"
	}
	return sign + trimZero(fmt.Sprintf("%.0f", abs))
}

func trimZero(value string) string {
	return strings.TrimSuffix(value, ".0")
}

// BarChart draws grouped vertical bars. Negative values go below the axis.
func BarChart(labels []string, series []ChartSeries, prefix string) template.HTML {
	const width, height = 640.0, 240.0
	const left, right, top, bottom = 48.0, 8.0, 12.0, 28.0
	if len(labels) == 0 || len(series) == 0 {
		return ""
	}
	maxValue, minValue := 0.0, 0.0
	for _, s := range series {
		for _, v := range s.Values {
			maxValue = math.Max(maxValue, v)
			minValue = math.Min(minValue, v)
		}
	}
	upper := niceMax(maxValue)
	lower := 0.0
	if minValue < 0 {
		lower = -niceMax(-minValue)
	}
	plotH := height - top - bottom
	plotW := width - left - right
	y := func(v float64) float64 { return top + (upper-v)/(upper-lower)*plotH }
	var b strings.Builder
	fmt.Fprintf(&b, `<svg class="fin-chart" viewBox="0 0 %.0f %.0f" role="img" preserveAspectRatio="xMidYMid meet">`, width, height)
	for i := 0; i <= 4; i++ {
		v := lower + (upper-lower)*float64(i)/4
		fmt.Fprintf(&b, `<line class="fin-chart__grid" x1="%.1f" x2="%.1f" y1="%.1f" y2="%.1f"/>`, left, width-right, y(v), y(v))
		fmt.Fprintf(&b, `<text class="fin-chart__axis" x="%.1f" y="%.1f" text-anchor="end">%s%s</text>`, left-6, y(v)+4, prefix, compactNumber(v))
	}
	fmt.Fprintf(&b, `<line class="fin-chart__zero" x1="%.1f" x2="%.1f" y1="%.1f" y2="%.1f"/>`, left, width-right, y(0), y(0))
	group := plotW / float64(len(labels))
	barW := math.Min(28, group*0.7/float64(len(series)))
	for i, label := range labels {
		x0 := left + group*float64(i) + (group-barW*float64(len(series)))/2
		for j, s := range series {
			if i >= len(s.Values) {
				continue
			}
			v := s.Values[i]
			yTop, yBottom := y(math.Max(v, 0)), y(math.Min(v, 0))
			h := math.Max(yBottom-yTop, 0)
			if v != 0 && h < 1 {
				h = 1
			}
			fmt.Fprintf(&b, `<rect class="fin-chart__bar %s" x="%.1f" y="%.1f" width="%.1f" height="%.1f" rx="2"><title>%s %s: %s%s</title></rect>`,
				template.HTMLEscapeString(s.Class), x0+barW*float64(j), yTop, barW-2, h,
				template.HTMLEscapeString(s.Label), template.HTMLEscapeString(label), prefix, template.HTMLEscapeString(GroupDigits(v, 0)))
		}
		fmt.Fprintf(&b, `<text class="fin-chart__axis" x="%.1f" y="%.1f" text-anchor="middle">%s</text>`, left+group*(float64(i)+0.5), height-8, template.HTMLEscapeString(label))
	}
	b.WriteString(`</svg>`)
	return template.HTML(b.String())
}

// LineChart draws one line with dots; points after dashedFrom are dashed
// (used for forecast months).
func LineChart(labels []string, values []float64, dashedFrom int, prefix string, width float64) template.HTML {
	if width < 320 {
		width = 640
	}
	const height = 220.0
	const left, right, top, bottom = 48.0, 12.0, 14.0, 28.0
	if len(values) == 0 {
		return ""
	}
	maxValue, minValue := 0.0, 0.0
	for _, v := range values {
		maxValue = math.Max(maxValue, v)
		minValue = math.Min(minValue, v)
	}
	upper := niceMax(maxValue)
	lower := 0.0
	if minValue < 0 {
		lower = -niceMax(-minValue)
	}
	plotH := height - top - bottom
	plotW := width - left - right
	step := plotW
	if len(values) > 1 {
		step = plotW / float64(len(values)-1)
	}
	x := func(i int) float64 {
		if len(values) == 1 {
			return left + plotW/2
		}
		return left + step*float64(i)
	}
	y := func(v float64) float64 { return top + (upper-v)/(upper-lower)*plotH }
	var b strings.Builder
	fmt.Fprintf(&b, `<svg class="fin-chart" viewBox="0 0 %.0f %.0f" role="img" preserveAspectRatio="xMidYMid meet">`, width, height)
	for i := 0; i <= 4; i++ {
		v := lower + (upper-lower)*float64(i)/4
		fmt.Fprintf(&b, `<line class="fin-chart__grid" x1="%.1f" x2="%.1f" y1="%.1f" y2="%.1f"/>`, left, width-right, y(v), y(v))
		fmt.Fprintf(&b, `<text class="fin-chart__axis" x="%.1f" y="%.1f" text-anchor="end">%s%s</text>`, left-6, y(v)+4, prefix, compactNumber(v))
	}
	if lower < 0 {
		fmt.Fprintf(&b, `<line class="fin-chart__zero fin-chart__zero--danger" x1="%.1f" x2="%.1f" y1="%.1f" y2="%.1f"/>`, left, width-right, y(0), y(0))
	}
	var solid, dashed []string
	for i, v := range values {
		point := fmt.Sprintf("%.1f,%.1f", x(i), y(v))
		if dashedFrom < 0 || i <= dashedFrom {
			solid = append(solid, point)
		}
		if dashedFrom >= 0 && i >= dashedFrom {
			dashed = append(dashed, point)
		}
	}
	if len(solid) > 1 {
		fmt.Fprintf(&b, `<polyline class="fin-chart__line" points="%s"/>`, strings.Join(solid, " "))
	}
	if len(dashed) > 1 {
		fmt.Fprintf(&b, `<polyline class="fin-chart__line fin-chart__line--dashed" points="%s"/>`, strings.Join(dashed, " "))
	}
	for i, v := range values {
		class := "fin-chart__dot"
		if v < 0 {
			class += " fin-chart__dot--danger"
		}
		if dashedFrom >= 0 && i > dashedFrom {
			class += " fin-chart__dot--forecast"
		}
		label := ""
		if i < len(labels) {
			label = labels[i]
		}
		fmt.Fprintf(&b, `<circle class="%s" cx="%.1f" cy="%.1f" r="4"><title>%s: %s%s</title></circle>`, class, x(i), y(v), template.HTMLEscapeString(label), prefix, template.HTMLEscapeString(GroupDigits(v, 0)))
		fmt.Fprintf(&b, `<text class="fin-chart__axis" x="%.1f" y="%.1f" text-anchor="middle">%s</text>`, x(i), height-8, template.HTMLEscapeString(label))
	}
	b.WriteString(`</svg>`)
	return template.HTML(b.String())
}
