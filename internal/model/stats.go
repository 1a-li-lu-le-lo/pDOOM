// Copyright NU Cybernetics. p(DOOM) — research prototype.

package model

import (
	"math"
	"sort"

	"github.com/1a-li-lu-le-lo/pdoom/internal/schema"
)

// Quantile returns the nearest-rank p-quantile of xs (0 < p ≤ 1). xs is sorted
// in place. An empty slice yields 0.
func Quantile(xs []float64, p float64) float64 {
	if len(xs) == 0 {
		return 0
	}
	sort.Float64s(xs)
	k := int(math.Ceil(p*float64(len(xs)))) - 1
	if k < 0 {
		k = 0
	}
	if k >= len(xs) {
		k = len(xs) - 1
	}
	return xs[k]
}

// Median is the conventional median (mean of the two middle values for even n).
func Median(xs []float64) float64 {
	if len(xs) == 0 {
		return 0
	}
	s := append([]float64(nil), xs...)
	sort.Float64s(s)
	n := len(s)
	if n%2 == 1 {
		return s[n/2]
	}
	return (s[n/2-1] + s[n/2]) / 2
}

// Mean is the arithmetic mean (0 for an empty slice).
func Mean(xs []float64) float64 {
	if len(xs) == 0 {
		return 0
	}
	sum := 0.0
	for _, x := range xs {
		sum += x
	}
	return sum / float64(len(xs))
}

// TrimmedMean drops fraction f from each end (by count, rounded down) and
// averages the rest. When fewer than five values are available it falls back
// to the median, as documented in docs/method/aggregation.md.
func TrimmedMean(xs []float64, f float64) float64 {
	if len(xs) < 5 {
		return Median(xs)
	}
	s := append([]float64(nil), xs...)
	sort.Float64s(s)
	k := int(math.Floor(f * float64(len(s))))
	return Mean(s[k : len(s)-k])
}

// WeightedMean returns Σ w x / Σ w (0 when the weights sum to 0).
func WeightedMean(xs, ws []float64) float64 {
	num, den := 0.0, 0.0
	for i := range xs {
		num += ws[i] * xs[i]
		den += ws[i]
	}
	if den == 0 {
		return 0
	}
	return num / den
}

// IQRLogOdds is the interquartile range of the values on the log-odds scale; it
// drives the disagreement label.
func IQRLogOdds(xs []float64) float64 {
	if len(xs) < 2 {
		return 0
	}
	lo := make([]float64, len(xs))
	for i, x := range xs {
		lo[i] = Logit(x)
	}
	return Quantile(append([]float64(nil), lo...), 0.75) - Quantile(append([]float64(nil), lo...), 0.25)
}

// Summarize computes the five quantiles and the mean of a sample.
func Summarize(samples []float64) schema.OutcomeEstimate {
	s := append([]float64(nil), samples...)
	sort.Float64s(s)
	q := func(p float64) float64 { return Quantile(s, p) }
	return schema.OutcomeEstimate{P05: q(0.05), P25: q(0.25), P50: q(0.5), P75: q(0.75), P95: q(0.95), Mean: Mean(s)}
}

// DisagreementLabel maps an IQR on the log-odds scale to a label
// (docs/method/uncertainty.md): <0.5 low, <1 moderate, <2 high, else extreme.
func DisagreementLabel(iqr float64) schema.UncertaintyLabel {
	switch {
	case iqr < 0.5:
		return schema.UncertaintyLow
	case iqr < 1.0:
		return schema.UncertaintyModerate
	case iqr < 2.0:
		return schema.UncertaintyHigh
	}
	return schema.UncertaintyExtreme
}

// UncertaintyLabelFromScore maps the 0–100 uncertainty score to a label:
// <25 low, <50 moderate, <75 high, else extreme.
func UncertaintyLabelFromScore(score float64) schema.UncertaintyLabel {
	switch {
	case score < 25:
		return schema.UncertaintyLow
	case score < 50:
		return schema.UncertaintyModerate
	case score < 75:
		return schema.UncertaintyHigh
	}
	return schema.UncertaintyExtreme
}

func round6(x float64) float64 { return math.Round(x*1e6) / 1e6 }
func round4(x float64) float64 { return math.Round(x*1e4) / 1e4 }
func clamp01(x float64) float64 {
	if x < 0 {
		return 0
	}
	if x > 1 {
		return 1
	}
	return x
}
