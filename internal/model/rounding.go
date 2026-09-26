// Copyright NU Cybernetics. p(DOOM) — research prototype.

package model

import (
	"fmt"
	"math"

	"github.com/1a-li-lu-le-lo/pdoom/internal/schema"
)

// RoundingStep returns the display rounding step in percentage points for an
// uncertainty label (build-spec §0.8): extreme 5, high 2, moderate 1, low 1.
func RoundingStep(rules schema.RoundingRules, label schema.UncertaintyLabel) int {
	switch label {
	case schema.UncertaintyExtreme:
		return nz(rules.Extreme, 5)
	case schema.UncertaintyHigh:
		return nz(rules.High, 2)
	case schema.UncertaintyModerate:
		return nz(rules.Moderate, 1)
	}
	return nz(rules.Low, 1)
}

func nz(v, d int) int {
	if v <= 0 {
		return d
	}
	return v
}

// RoundingRuleName names the rule used, e.g. "nearest_5".
func RoundingRuleName(step int) string { return fmt.Sprintf("nearest_%d", step) }

// RoundForDisplay renders a probability as a rounded percentage string. It
// never shows decimals; values that round below one point show "<1%" and
// values that round above ninety-nine show ">99%".
func RoundForDisplay(p float64, step int) string {
	if step <= 0 {
		step = 1
	}
	pct := p * 100
	rounded := math.Round(pct/float64(step)) * float64(step)
	switch {
	case p > 0 && rounded < 1:
		return "<1%"
	case p < 1 && rounded > 99:
		return ">99%"
	case rounded <= 0:
		return "0%"
	case rounded >= 100:
		return "100%"
	}
	return fmt.Sprintf("%d%%", int(rounded))
}

// FormatInterval renders a low–high pair with an en dash, e.g. "3%–30%".
func FormatInterval(lo, hi float64, step int) string {
	return RoundForDisplay(lo, step) + "–" + RoundForDisplay(hi, step)
}
