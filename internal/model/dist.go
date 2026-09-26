// Copyright NU Cybernetics. p(DOOM) — research prototype.

package model

import (
	"math"

	"github.com/1a-li-lu-le-lo/pdoom/internal/schema"
)

// z95 is the standard-normal quantile for 0.95, used to fit a logit-normal from
// a 5th/95th percentile pair.
const z95 = 1.6448536269514722

// Logit maps a probability to the log-odds scale; inputs are clamped away from 0 and 1.
func Logit(p float64) float64 {
	const eps = 1e-9
	if p < eps {
		p = eps
	}
	if p > 1-eps {
		p = 1 - eps
	}
	return math.Log(p / (1 - p))
}

// Sigmoid maps log-odds back to a probability.
func Sigmoid(x float64) float64 { return 1 / (1 + math.Exp(-x)) }

// LogitNormal is a distribution on (0,1) whose logit is normal.
type LogitNormal struct {
	Mu    float64
	Sigma float64
}

// LogitNormalFromTri fits a logit-normal to a {p05, p50, p95} specification:
// mu = logit(p50), sigma = (logit(p95) − logit(p05)) / (2 · 1.6448536269514722).
// The median is reproduced exactly; the 5th and 95th percentiles are matched
// on average when the specification is symmetric on the logit scale.
func LogitNormalFromTri(t schema.TriQuantile) LogitNormal {
	return LogitNormal{Mu: Logit(t.P50), Sigma: (Logit(t.P95) - Logit(t.P05)) / (2 * z95)}
}

// At evaluates the distribution at a standard-normal deviate.
func (d LogitNormal) At(z float64) float64 { return Sigmoid(d.Mu + d.Sigma*z) }

// Shifted returns the distribution with its logit-scale location moved by delta.
func (d LogitNormal) Shifted(delta float64) LogitNormal {
	return LogitNormal{Mu: d.Mu + delta, Sigma: d.Sigma}
}

// Tri returns the {p05,p50,p95} implied by the distribution.
func (d LogitNormal) Tri() schema.TriQuantile {
	return schema.TriQuantile{P05: d.At(-z95), P50: d.At(0), P95: d.At(z95)}
}
