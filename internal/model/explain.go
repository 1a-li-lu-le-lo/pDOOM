// Copyright NU Cybernetics. p(DOOM) — research prototype.

package model

import (
	"fmt"

	"github.com/1a-li-lu-le-lo/pdoom/internal/schema"
	"github.com/1a-li-lu-le-lo/pdoom/internal/snapshot"
)

// Explain builds the "Why this number?" contributions for the signal-based
// indexes. Each row carries the signal's contribution in index points, the
// sources behind the observation, its confidence, and how much the index would
// move if the signal were removed. It also states plainly that causal
// attribution of probability deltas is not available in v0.
func Explain(s *snapshot.Snapshot, indexes []schema.IndexValue, asOf string) schema.DriversExplained {
	out := schema.DriversExplained{Kind: "drivers_explained", SchemaVersion: 1, AsOf: asOf, Items: []schema.DriverContribution{}}
	signals := s.SignalIndex()
	weightsFor := map[schema.IndexID]map[string]float64{
		schema.IndexCapabilityPressure:        s.ModelSpec.IndexWeights.CapabilityPressure,
		schema.IndexControlStrength:           s.ModelSpec.IndexWeights.ControlStrength,
		schema.IndexAgenticInfrastructureRisk: s.ModelSpec.IndexWeights.AgenticInfrastructureRisk,
	}
	for _, iv := range indexes {
		w, ok := weightsFor[iv.IndexID]
		if !ok || iv.Value == nil {
			continue
		}
		for _, c := range iv.Components {
			o, found := observationFor(s, c.SignalID)
			if !found {
				continue
			}
			without := signalIndex(s, iv.IndexID, w, asOf, c.SignalID)
			sens := 0.0
			if without.Value != nil {
				sens = round4(*iv.Value - *without.Value)
			}
			direction := "raises_pressure"
			if iv.IndexID == schema.IndexControlStrength {
				direction = "strengthens_control"
			}
			family := ""
			if ref, ok := signals[c.SignalID]; ok {
				family = ref.Family
			}
			out.Items = append(out.Items, schema.DriverContribution{
				IndexID: iv.IndexID, Driver: c.SignalID, DriverFamily: family, Direction: direction, Magnitude: c.Contribution,
				SourceIDs: o.SourceIDs, Confidence: o.Confidence, ModelRole: fmt.Sprintf("weighted component (weight %.2f, tier %d)", c.Weight, c.Tier),
				LastUpdated: o.AsOf, Sensitivity: sens, Counterevidence: o.Counterevidence,
			})
		}
	}
	out.Notes = []string{
		"Contributions are shares of a 0–100 index, not probabilities.",
		"Sensitivity is the change in the index when the signal is removed and the remaining weights are renormalized.",
		"Causal attribution of probability changes to individual drivers is not available in this model version; the research-mode estimate depends on judgment-based parameters documented in the model specification.",
	}
	return out
}
