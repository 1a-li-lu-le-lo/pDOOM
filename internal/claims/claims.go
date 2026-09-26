// Copyright NU Cybernetics. p(DOOM) — research prototype.

// Package claims is the rule-based candidate claim extraction skeleton of the
// ingestion pipeline. A sentence becomes a candidate claim when it contains a
// quantity (a percentage, or a number with a recognised unit) AND a hedging or
// probability verb. Candidates carry status "candidate" and
// model_use_status "excluded": nothing here can feed the model without a human
// review that promotes it into a new snapshot.
//
// LLM classification is out of scope and explicitly off (ClassifierMode).
package claims

import (
	"regexp"
	"strconv"
	"strings"
	"unicode"

	"github.com/1a-li-lu-le-lo/pdoom/internal/schema"
)

// ClassifierMode documents the only extraction mode this package implements.
const ClassifierMode = "rule_based"

// LLMClassificationEnabled is false by design: retrieved content is untrusted
// data and is never sent to a language model by the ingestion pipeline.
const LLMClassificationEnabled = false

// Limits keep extraction bounded on hostile input.
const (
	MaxCandidatesPerDocument = 50
	MaxSentenceChars         = 600
	MaxSentences             = 2000
)

// Candidate is a candidate claim extracted from one sentence.
type Candidate struct {
	Text               string                `json:"text"`
	QuantitativeValue  float64               `json:"quantitative_value"`
	Unit               string                `json:"unit"`
	HedgeTerm          string                `json:"hedge_term"`
	DirectQuotePointer string                `json:"direct_quote_pointer"`
	Status             schema.ClaimStatus    `json:"status"`
	ModelUseStatus     schema.ModelUseStatus `json:"model_use_status"`
	Method             string                `json:"method"`
}

// hedgeTerms are the hedging / probability words (lowercase, matched on word
// boundaries; plural and tense variants listed explicitly).
var hedgeTerms = []string{
	"estimate", "estimates", "estimated", "estimating",
	"predict", "predicts", "predicted", "prediction", "predictions",
	"forecast", "forecasts", "forecasted",
	"expect", "expects", "expected",
	"projected", "projection", "projections",
	"suggest", "suggests", "suggested",
	"indicate", "indicates", "indicated",
	"likely", "unlikely", "probability", "probabilities", "chance", "odds",
	"could", "may", "might", "would",
	"believe", "believes", "believed",
	"reported", "reportedly",
	"find", "finds", "found",
	"measured", "measures",
	"shows", "showed",
	"assume", "assumes", "assumed",
	"roughly", "approximately",
}

// units maps a lowercase, singularised unit token to its canonical unit label.
var units = map[string]string{
	"%": "%", "percent": "%", "pct": "%",
	"pp": "percentage_points", "percentage point": "percentage_points",
	"x": "multiplier", "×": "multiplier", "fold": "multiplier",
	"year": "years", "yr": "years", "month": "months", "week": "weeks", "day": "days",
	"hour": "hours", "hr": "hours", "minute": "minutes", "min": "minutes", "second": "seconds", "sec": "seconds",
	"token": "tokens", "parameter": "parameters", "param": "parameters",
	"flop": "FLOP", "flops": "FLOP", "petaflop": "petaFLOP", "exaflop": "exaFLOP",
	"gpu": "GPUs", "chip": "chips", "accelerator": "accelerators",
	"gw": "GW", "gigawatt": "GW", "mw": "MW", "megawatt": "MW", "kwh": "kWh", "mwh": "MWh", "gwh": "GWh", "twh": "TWh",
	"usd": "USD", "dollar": "USD", "eur": "EUR", "euro": "EUR", "gbp": "GBP", "pound": "GBP",
	"people": "people", "person": "people", "respondent": "people", "participant": "people", "researcher": "people",
	"expert": "people", "forecaster": "people", "worker": "people", "user": "people",
	"paper": "papers", "incident": "incidents", "model": "models", "task": "tasks", "point": "points",
}

// multipliers scale a number and are folded into the unit label when no other
// unit follows ("10 billion" → 1e10, unit "count").
var multipliers = map[string]float64{
	"thousand": 1e3, "k": 1e3, "million": 1e6, "m": 1e6, "mn": 1e6,
	"billion": 1e9, "bn": 1e9, "b": 1e9, "trillion": 1e12, "tn": 1e12,
}

var currencies = map[string]string{"$": "USD", "us$": "USD", "€": "EUR", "£": "GBP"}

var reSentenceEnd = regexp.MustCompile(`([.!?]+)(\s+|$)`)

// SplitSentences splits text into sentences on ., ! and ? followed by
// whitespace. Whitespace is collapsed; very long sentences are cut at
// MaxSentenceChars; at most MaxSentences are returned.
func SplitSentences(text string) []string {
	clean := strings.Join(strings.Fields(text), " ")
	if clean == "" {
		return nil
	}
	var out []string
	start := 0
	for _, m := range reSentenceEnd.FindAllStringIndex(clean, -1) {
		s := strings.TrimSpace(clean[start:m[1]])
		if s != "" {
			out = append(out, truncate(s))
		}
		start = m[1]
		if len(out) >= MaxSentences {
			return out
		}
	}
	if rest := strings.TrimSpace(clean[start:]); rest != "" && len(out) < MaxSentences {
		out = append(out, truncate(rest))
	}
	return out
}

func truncate(s string) string {
	if len(s) <= MaxSentenceChars {
		return s
	}
	cut := s[:MaxSentenceChars]
	for len(cut) > 0 && !utf8Boundary(cut) {
		cut = cut[:len(cut)-1]
	}
	return cut
}

func utf8Boundary(s string) bool {
	return strings.ToValidUTF8(s, "") == s
}

// Quantity is a number with a unit found in a sentence.
type Quantity struct {
	Value float64
	Unit  string
}

// FindQuantity returns the first quantity of a sentence: a percentage
// ("45%", "45 percent") or a number followed by a recognised unit or preceded
// by a currency symbol. Bare numbers (years, counts without a unit) are not
// quantities. ok is false when none is found.
func FindQuantity(sentence string) (q Quantity, ok bool) {
	tokens := strings.Fields(sentence)
	for i, tok := range tokens {
		raw := strings.Trim(tok, "()[]{},;:\"'“”‘’")
		if raw == "" {
			continue
		}
		cur := ""
		lower := strings.ToLower(raw)
		for sym, code := range currencies {
			if strings.HasPrefix(lower, sym) {
				cur = code
				raw = raw[len(sym):]
				break
			}
		}
		if raw == "" || !startsWithDigit(raw) {
			continue
		}
		num, suffix := splitNumber(raw)
		v, err := strconv.ParseFloat(strings.ReplaceAll(num, ",", ""), 64)
		if err != nil || v < 0 {
			continue
		}
		// Suffix glued to the number: 45%, 10x, 3.5k.
		suffix = strings.TrimRight(suffix, ".!?")
		if suffix != "" {
			if u, ok := unitFor(suffix); ok {
				return Quantity{Value: v, Unit: u}, true
			}
			if mul, ok := multipliers[strings.ToLower(suffix)]; ok {
				v *= mul
				if cur != "" {
					return Quantity{Value: v, Unit: cur}, true
				}
				if i+1 < len(tokens) {
					if u, ok := unitFor(cleanWord(tokens[i+1])); ok {
						return Quantity{Value: v, Unit: u}, true
					}
				}
				return Quantity{Value: v, Unit: "count"}, true
			}
			continue
		}
		next := ""
		if i+1 < len(tokens) {
			next = cleanWord(tokens[i+1])
		}
		if mul, ok := multipliers[strings.ToLower(next)]; ok {
			v *= mul
			if cur != "" {
				return Quantity{Value: v, Unit: cur}, true
			}
			if i+2 < len(tokens) {
				if u, ok := unitFor(cleanWord(tokens[i+2])); ok {
					return Quantity{Value: v, Unit: u}, true
				}
			}
			return Quantity{Value: v, Unit: "count"}, true
		}
		if cur != "" {
			return Quantity{Value: v, Unit: cur}, true
		}
		if next != "" {
			if u, ok := unitFor(next); ok {
				return Quantity{Value: v, Unit: u}, true
			}
			// two-word units: "percentage points"
			if i+2 < len(tokens) {
				two := strings.ToLower(next + " " + cleanWord(tokens[i+2]))
				if u, ok := units[singular(two)]; ok {
					return Quantity{Value: v, Unit: u}, true
				}
			}
		}
	}
	return Quantity{}, false
}

func startsWithDigit(s string) bool {
	r := []rune(s)
	return len(r) > 0 && unicode.IsDigit(r[0])
}

// splitNumber separates the leading numeric part (digits, commas, one decimal
// point) from whatever is glued to it.
func splitNumber(s string) (num, suffix string) {
	i := 0
	seenDot := false
	for i < len(s) {
		c := s[i]
		switch {
		case c >= '0' && c <= '9':
		case c == ',' && i+1 < len(s) && s[i+1] >= '0' && s[i+1] <= '9':
		case c == '.' && !seenDot && i+1 < len(s) && s[i+1] >= '0' && s[i+1] <= '9':
			seenDot = true
		default:
			return s[:i], s[i:]
		}
		i++
	}
	return s, ""
}

func cleanWord(tok string) string {
	return strings.Trim(tok, "()[]{},;:.!?\"'“”‘’")
}

func singular(w string) string {
	w = strings.ToLower(w)
	if strings.HasSuffix(w, "ies") && len(w) > 3 {
		return w[:len(w)-3] + "y"
	}
	if strings.HasSuffix(w, "s") && !strings.HasSuffix(w, "ss") && len(w) > 1 {
		return w[:len(w)-1]
	}
	return w
}

func unitFor(tok string) (string, bool) {
	if tok == "" {
		return "", false
	}
	if u, ok := units[strings.ToLower(tok)]; ok {
		return u, true
	}
	if u, ok := units[singular(tok)]; ok {
		return u, true
	}
	return "", false
}

// FindHedge returns the first hedging / probability term of a sentence, or "".
// A capitalised "May" is treated as the month, not the modal verb.
func FindHedge(sentence string) string {
	for _, tok := range strings.Fields(sentence) {
		orig := strings.Trim(tok, "()[]{},;:.!?\"'“”‘’")
		if orig == "May" {
			continue
		}
		w := strings.ToLower(orig)
		for _, h := range hedgeTerms {
			if w == h {
				return h
			}
		}
	}
	return ""
}

// Extract returns the candidate claims of a text. Sentence numbers are
// 1-based positions in SplitSentences(text) and are recorded as
// direct_quote_pointer "sentence N" so a reviewer can locate the quote.
func Extract(text string) []Candidate {
	var out []Candidate
	for i, s := range SplitSentences(text) {
		q, ok := FindQuantity(s)
		if !ok {
			continue
		}
		h := FindHedge(s)
		if h == "" {
			continue
		}
		out = append(out, Candidate{
			Text:               s,
			QuantitativeValue:  q.Value,
			Unit:               q.Unit,
			HedgeTerm:          h,
			DirectQuotePointer: "sentence " + strconv.Itoa(i+1),
			Status:             schema.ClaimStatus("candidate"),
			ModelUseStatus:     schema.ModelUseExcluded,
			Method:             ClassifierMode,
		})
		if len(out) >= MaxCandidatesPerDocument {
			break
		}
	}
	return out
}
